package e2e

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// =============================================================================
// Tier 3: Cross-Feature Combinations & Multi-Domain Interactions
// =============================================================================

// TestTier3_Product_Variant_Storage_MediaResolution exercises the catalog read path:
// Product + Variant + Inventory Stock + Storage URL resolution composed on 1 connection.
func TestTier3_Product_Variant_Storage_MediaResolution(t *testing.T) {
	runner := NewMockRunner(1)
	prodRepo := NewMockProductRepository()
	varRepo := NewMockVariantRepository()
	invRepo := NewMockInventoryRepository()
	storageProvider := NewMockStorageProvider("https://cdn.prim.com")

	prodID := uuid.New()
	varID := uuid.New()
	sku := "SKU-SNEAKER-42"
	price := int64(12999)

	prodRepo.AddProduct(&model.Product{
		ID:          prodID,
		Title:       "Performance Sneaker",
		Status:      model.PublicationStatusPublished,
		ProductType: model.ProductTypeVariable,
	})

	varRepo.AddVariant(&model.ProductVariant{
		ID:        varID,
		ProductID: prodID,
		SKU:       sku,
		Title:     "Size 42 / Obsidian",
		Price:     &price,
	})

	invRepo.SetStock(varID, 50, 5) // 45 available

	type ProductDTO struct {
		ID        uuid.UUID
		Title     string
		ImageURL  string
		VariantID uuid.UUID
		SKU       string
		Price     int64
		Available int
	}

	var result *ProductDTO

	// Single connection (WithExecutor) reads across Product, Variant, Inventory, Storage
	err := runner.WithExecutor(context.Background(), func(exec database.QueryExecutor) error {
		p, err := prodRepo.GetByID(context.Background(), exec, prodID)
		if err != nil {
			return CanonicalHandleError(err, "product")
		}

		v, err := varRepo.GetByID(context.Background(), exec, varID)
		if err != nil {
			return CanonicalHandleError(err, "variant")
		}

		stock, err := invRepo.GetStock(context.Background(), exec, varID)
		if err != nil {
			return CanonicalHandleError(err, "inventory")
		}

		// Direct Storage Resolution (Rule 3)
		imageURL := storageProvider.GetPublicURL("products", "sneaker_thumb.webp")

		result = &ProductDTO{
			ID:        p.ID,
			Title:     p.Title,
			ImageURL:  imageURL,
			VariantID: v.ID,
			SKU:       v.SKU,
			Price:     *v.Price,
			Available: stock.AvailableQuantity,
		}
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Performance Sneaker", result.Title)
	assert.Equal(t, "https://cdn.prim.com/products/sneaker_thumb.webp", result.ImageURL)
	assert.Equal(t, 45, result.Available)
	assert.Equal(t, int64(12999), result.Price)

	withDB, withTx, _, _ := runner.Stats()
	assert.Equal(t, 1, withDB)
	assert.Equal(t, 0, withTx)
	assert.False(t, runner.ConnectionHopDetected())
}

// TestTier3_Order_Inventory_StateTransition_CommitAndRestock exercises order lifecycle
// interacting with consumer inventory repository (stock hold, commit, and compensating restock).
func TestTier3_Order_Inventory_StateTransition_CommitAndRestock(t *testing.T) {
	runner := NewMockRunner(1)
	orderRepo := NewMockOrderRepository()
	invRepo := NewMockInventoryRepository()

	orderID := uuid.New()
	varID := uuid.New()
	invRepo.SetStock(varID, 20, 0)

	// Step 1: Create Order & Hold Reservation on Tx
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		if err := orderRepo.CreateOrder(ctx, tx, &model.Order{
			ID:          orderID,
			TotalAmount: 5000,
			Status:      model.OrderStatusPending,
		}); err != nil {
			return err
		}

		return invRepo.CreateReservation(ctx, tx, &model.InventoryReservation{
			ID:        uuid.New(),
			VariantID: varID,
			OrderID:   &orderID,
			Quantity:  2,
			ExpiresAt: time.Now().Add(time.Hour),
		})
	})
	require.NoError(t, err)

	unreleased, _ := invRepo.GetUnreleasedReservationsForOrder(context.Background(), nil, orderID)
	require.Len(t, unreleased, 1)

	// Step 2: Payment Confirmed -> Commit reservation to finalized sale
	err = runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		// Update status to processing
		if err := orderRepo.UpdateOrderStatus(ctx, tx, orderID, model.OrderStatusProcessing); err != nil {
			return err
		}
		// Release reservation hold and record sale ledger
		if _, err := invRepo.ReleaseReservationsForOrder(ctx, tx, orderID); err != nil {
			return err
		}
		return invRepo.CreateLedger(ctx, tx, &model.InventoryLedger{
			VariantID: varID,
			Quantity:  -2,
			Reason:    model.InventoryReasonSale,
		})
	})
	require.NoError(t, err)

	unreleasedAfterCommit, _ := invRepo.GetUnreleasedReservationsForOrder(context.Background(), nil, orderID)
	assert.Empty(t, unreleasedAfterCommit, "Reservation holds must be cleared upon sale commitment")

	// Step 3: Order Cancelled -> Compensating return ledger restores stock
	err = runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		if err := orderRepo.UpdateOrderStatus(ctx, tx, orderID, model.OrderStatusCancelled); err != nil {
			return err
		}
		// Write compensating return ledger
		return invRepo.CreateLedger(ctx, tx, &model.InventoryLedger{
			VariantID: varID,
			Quantity:  2,
			Reason:    model.InventoryReasonReturn,
		})
	})
	require.NoError(t, err)

	ledgers := invRepo.Ledgers()
	require.Len(t, ledgers, 2)
	assert.Equal(t, model.InventoryReasonSale, ledgers[0].Reason)
	assert.Equal(t, -2, ledgers[0].Quantity)
	assert.Equal(t, model.InventoryReasonReturn, ledgers[1].Reason)
	assert.Equal(t, 2, ledgers[1].Quantity)
	assert.False(t, runner.ConnectionHopDetected())
}

// TestTier3_Checkout_Cart_Coupon_Inventory_AtomicOrchestration exercises the full
// checkout orchestration bundling Cart, Coupon, Inventory, and Order in 1 transaction.
func TestTier3_Checkout_Cart_Coupon_Inventory_AtomicOrchestration(t *testing.T) {
	runner := NewMockRunner(1)
	cartRepo := NewMockCartRepository()
	couponRepo := NewMockCouponRepository()
	invRepo := NewMockInventoryRepository()
	varRepo := NewMockVariantRepository()
	orderRepo := NewMockOrderRepository()

	userID := uuid.New()
	cartID := uuid.New()
	varID := uuid.New()
	couponID := uuid.New()
	orderID := uuid.New()
	price := int64(10000)

	varRepo.AddVariant(&model.ProductVariant{
		ID:    varID,
		SKU:   "SKU-SHIRT-M",
		Price: &price,
	})

	invRepo.SetStock(varID, 15, 0) // 15 available

	minCart := int64(5000)
	couponRepo.AddCoupon(&model.Coupon{
		ID:            couponID,
		Code:          "SAVE20",
		DiscountType:  model.CouponDiscountTypePercentage,
		DiscountValue: 20,
		MinCartAmount: &minCart,
		Status:        model.CouponStatusActive,
	})

	cartRepo.SetCart(&model.Cart{
		ID:     cartID,
		UserID: &userID,
		Items: []model.CartItem{{
			ID:        1,
			CartID:    cartID,
			VariantID: varID,
			Quantity:  2, // Gross = 20000
		}},
	})

	// Execute complete checkout transaction within a single WithTx
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		// 1. Fetch Cart
		c, err := cartRepo.GetCartByUserID(ctx, tx, userID)
		if err != nil {
			return CanonicalHandleError(err, "cart")
		}
		if len(c.Items) == 0 {
			return apierr.BadRequestError("empty cart")
		}

		// 2. Lock variant & check stock
		if err := invRepo.LockVariantForUpdate(ctx, tx, varID); err != nil {
			return CanonicalHandleError(err, "variant lock")
		}
		stock, err := invRepo.GetStock(ctx, tx, varID)
		if err != nil {
			return CanonicalHandleError(err, "stock")
		}
		if stock.AvailableQuantity < 2 {
			return apierr.BadRequestError("out of stock")
		}

		// 3. Validate coupon
		coupon, err := couponRepo.GetByID(ctx, tx, couponID)
		if err != nil {
			return CanonicalHandleError(err, "coupon")
		}
		gross := int64(20000)
		if gross < *coupon.MinCartAmount {
			return apierr.BadRequestError("min cart not reached")
		}
		discount := gross * coupon.DiscountValue / 100 // 4000
		netTotal := gross - discount                   // 16000

		// 4. Place stock reservation
		if err := invRepo.CreateReservation(ctx, tx, &model.InventoryReservation{
			ID:        uuid.New(),
			VariantID: varID,
			OrderID:   &orderID,
			Quantity:  2,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}); err != nil {
			return CanonicalHandleError(err, "reservation")
		}

		// 5. Record coupon redemption
		if err := couponRepo.RecordRedemption(ctx, tx, &model.CouponRedemption{
			CouponID:       couponID,
			UserID:         userID,
			OrderID:        orderID,
			DiscountAmount: discount,
		}); err != nil {
			return CanonicalHandleError(err, "coupon redemption")
		}

		// 6. Create order
		if err := orderRepo.CreateOrder(ctx, tx, &model.Order{
			ID:             orderID,
			CustomerID:     &userID,
			CouponID:       &couponID,
			DiscountAmount: discount,
			TotalAmount:    netTotal,
			Status:         model.OrderStatusPending,
		}); err != nil {
			return CanonicalHandleError(err, "order")
		}

		// 7. Clear cart
		return cartRepo.ClearCart(ctx, tx, cartID)
	})

	require.NoError(t, err)

	// Invariants verification
	savedOrder, err := orderRepo.GetOrderByID(context.Background(), nil, orderID)
	require.NoError(t, err)
	assert.Equal(t, int64(16000), savedOrder.TotalAmount)
	assert.Equal(t, int64(4000), savedOrder.DiscountAmount)

	clearedCart, _ := cartRepo.GetCartByUserID(context.Background(), nil, userID)
	assert.Empty(t, clearedCart.Items, "Cart must be cleared atomically")

	redemptions, _ := couponRepo.GetRedemptionCount(context.Background(), nil, couponID)
	assert.Equal(t, 1, redemptions)

	assert.False(t, runner.ConnectionHopDetected())
	_, withTx, commits, _ := runner.Stats()
	assert.Equal(t, 1, withTx)
	assert.Equal(t, 1, commits)
}

// TestTier3_Checkout_User_Address_OwnershipEnforcement verifies checkout verifies
// user active status and address ownership on the caller's connection.
func TestTier3_Checkout_User_Address_OwnershipEnforcement(t *testing.T) {
	runner := NewMockRunner(1)
	userRepo := NewMockUserRepository()
	addrRepo := NewMockAddressRepository()

	validUser := uuid.New()
	stolenAddrID := uuid.New()
	legitUser := uuid.New()

	userRepo.AddUser(&model.User{
		ID:         validUser,
		Identifier: "buyer@example.com",
		Status:     model.StatusActive,
		Role:       model.CustomerRole,
	})

	// Address belongs to legitUser, not validUser
	addrRepo.AddAddress(&model.UserAddress{
		ID:       stolenAddrID,
		UserID:   legitUser,
		FullName: "Other Customer",
	})

	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		// Verify customer
		u, err := userRepo.GetByID(ctx, tx, validUser)
		if err != nil {
			return CanonicalHandleError(err, "customer")
		}
		if u.Status != model.StatusActive {
			return apierr.ForbiddenError("Inactive account")
		}

		// Verify address ownership
		_, err = addrRepo.GetByIDAndUser(ctx, tx, stolenAddrID, validUser)
		if err != nil {
			return CanonicalHandleError(err, "shipping address")
		}
		return nil
	})

	require.Error(t, err)
	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.Status)
	assert.Contains(t, apiErr.Error(), "shipping address not found")
}

// TestTier3_Settings_Audit_ConfigurationCacheInvalidation tests administrative updates
// that mutate settings in WithTx, write audit log, and invalidate cache.
func TestTier3_Settings_Audit_ConfigurationCacheInvalidation(t *testing.T) {
	runner := NewMockRunner(1)
	auditRepo := NewMockAuditRepository()

	cachedValue := int64(15000)
	actorID := uuid.New()

	updateSettingAndAudit := func(ctx context.Context, newThreshold int64) error {
		return runner.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
			// Update in memory or DB
			cachedValue = newThreshold

			// Record audit log entry
			return auditRepo.Insert(ctx, tx, &model.AuditLog{
				ID:         1,
				ActorID:    &actorID,
				Action:     "setting.free_shipping_threshold.update",
				EntityType: "setting",
			})
		})
	}

	err := updateSettingAndAudit(context.Background(), 20000)
	require.NoError(t, err)

	assert.Equal(t, int64(20000), cachedValue)
	assert.Len(t, auditRepo.Entries(), 1)
	assert.Equal(t, "setting.free_shipping_threshold.update", auditRepo.Entries()[0].Action)
	assert.False(t, runner.ConnectionHopDetected())
}
