package e2e

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
)

// =============================================================================
// Tier 4: Real-World Scenarios
// =============================================================================

// TestTier4_RealWorld_CompleteCashOnDeliveryCheckout simulates a real storefront
// customer purchasing items via Cash On Delivery.
func TestTier4_RealWorld_CompleteCashOnDeliveryCheckout(t *testing.T) {
	runner := NewMockRunner(1)
	cartRepo := NewMockCartRepository()
	invRepo := NewMockInventoryRepository()
	userRepo := NewMockUserRepository()
	addrRepo := NewMockAddressRepository()
	orderRepo := NewMockOrderRepository()

	userID := uuid.New()
	cartID := uuid.New()
	addrID := uuid.New()
	var1 := uuid.New()
	var2 := uuid.New()
	orderID := uuid.New()

	// Seed User & Address
	name := "Alice Shopper"
	userRepo.AddUser(&model.User{
		ID:         userID,
		Phone: "alice@example.com",
		FullName:   name,
		Status:     model.StatusActive,
		Role:       model.CustomerRole,
	})

	postal := "11511"
	addrRepo.AddAddress(&model.UserAddress{
		ID:         addrID,
		UserID:     userID,
		Line1:      "22 Nile Corniche",
		City:       "Cairo",
		Country:    "EG",
		PostalCode: &postal,
	})

	// Seed Inventory
	invRepo.SetStock(var1, 10, 0)
	invRepo.SetStock(var2, 5, 0)

	// Seed Cart
	cartRepo.SetCart(&model.Cart{
		ID:     cartID,
		UserID: &userID,
		Items: []model.CartItem{
			{ID: 1, CartID: cartID, VariantID: var1, Quantity: 1, PriceAtPurchase: 5000},
			{ID: 2, CartID: cartID, VariantID: var2, Quantity: 2, PriceAtPurchase: 3000},
		},
	})

	// Execute COD Checkout in single transaction (Rule 2)
	err := runner.WithTx(context.Background(), func(ctx context.Context) error {
		// 1. Validate Customer
		user, err := userRepo.GetByID(ctx, userID)
		if err != nil || user.Status != model.StatusActive {
			return errors.New("invalid customer")
		}

		// 2. Validate Address
		if _, err := addrRepo.GetByIDAndUser(ctx, addrID, userID); err != nil {
			return errors.New("invalid address")
		}

		// 3. Lock & Deduct Inventory (COD commits sale immediately)
		if err := invRepo.CreateLedger(ctx, &model.InventoryLedger{
			VariantID: var1,
			Quantity:  -1,
			Reason:    model.InventoryReasonSale,
		}); err != nil {
			return err
		}
		if err := invRepo.CreateLedger(ctx, &model.InventoryLedger{
			VariantID: var2,
			Quantity:  -2,
			Reason:    model.InventoryReasonSale,
		}); err != nil {
			return err
		}

		// 4. Create Order (Gross = 5000 + 6000 = 11000; Shipping = 999; COD Fee = 500; Total = 12499)
		gross := int64(11000)
		shipping := int64(999)
		codFee := int64(500)
		total := gross + shipping + codFee

		if err := orderRepo.CreateOrder(ctx, &model.Order{
			ID:             orderID,
			CustomerID:     &userID,
			ShippingAmount: shipping,
			CODFee:         codFee,
			TotalAmount:    total,
			PaymentMethod:  "cash_on_delivery",
			Status:         model.OrderStatusPending,
		}); err != nil {
			return err
		}

		// 5. Clear Cart
		return cartRepo.ClearCart(ctx, cartID)
	})

	require.NoError(t, err)

	// Post-condition assertions
	order, err := orderRepo.GetOrderByID(context.Background(), orderID)
	require.NoError(t, err)
	assert.Equal(t, int64(12499), order.TotalAmount)
	assert.Equal(t, "cash_on_delivery", order.PaymentMethod)

	c, _ := cartRepo.GetCartByUserID(context.Background(), userID)
	assert.Empty(t, c.Items, "Cart must be emptied")

	ledgers := invRepo.Ledgers()
	require.Len(t, ledgers, 2)
	assert.Equal(t, -1, ledgers[0].Quantity)
	assert.Equal(t, -2, ledgers[1].Quantity)

	// COD orders leave zero active reservations
	unreleased, _ := invRepo.GetUnreleasedReservationsForOrder(context.Background(), orderID)
	assert.Empty(t, unreleased)
	assert.False(t, runner.ConnectionHopDetected())
}

// TestTier4_RealWorld_CompleteOnlineCardPaymentCheckout simulates a checkout where
// stock is temporarily held, then committed when payment webhook arrives.
func TestTier4_RealWorld_CompleteOnlineCardPaymentCheckout(t *testing.T) {
	runner := NewMockRunner(1)
	orderRepo := NewMockOrderRepository()
	invRepo := NewMockInventoryRepository()

	varID := uuid.New()
	orderID := uuid.New()
	invRepo.SetStock(varID, 10, 0)

	// Stage 1: Checkout places order and stock hold
	err := runner.WithTx(context.Background(), func(ctx context.Context) error {
		if err := orderRepo.CreateOrder(ctx, &model.Order{
			ID:            orderID,
			TotalAmount:   7500,
			PaymentMethod: "credit_card",
			Status:        model.OrderStatusPending,
		}); err != nil {
			return err
		}

		return invRepo.CreateReservation(ctx, &model.InventoryReservation{
			ID:        uuid.New(),
			VariantID: varID,
			OrderID:   &orderID,
			Quantity:  1,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		})
	})
	require.NoError(t, err)

	unreleased, _ := invRepo.GetUnreleasedReservationsForOrder(context.Background(), orderID)
	require.Len(t, unreleased, 1)

	// Stage 2: Payment Webhook arrives (payment captured) -> commit hold to sale
	err = runner.WithTx(context.Background(), func(ctx context.Context) error {
		if err := orderRepo.UpdateOrderStatus(ctx, orderID, model.OrderStatusProcessing); err != nil {
			return err
		}
		if _, err := invRepo.ReleaseReservationsForOrder(ctx, orderID); err != nil {
			return err
		}
		return invRepo.CreateLedger(ctx, &model.InventoryLedger{
			VariantID: varID,
			Quantity:  -1,
			Reason:    model.InventoryReasonSale,
		})
	})
	require.NoError(t, err)

	order, _ := orderRepo.GetOrderByID(context.Background(), orderID)
	assert.Equal(t, model.OrderStatusProcessing, order.Status)

	unreleasedAfterCapture, _ := invRepo.GetUnreleasedReservationsForOrder(context.Background(), orderID)
	assert.Empty(t, unreleasedAfterCapture)

	ledgers := invRepo.Ledgers()
	require.Len(t, ledgers, 1)
	assert.Equal(t, -1, ledgers[0].Quantity)
	assert.Equal(t, model.InventoryReasonSale, ledgers[0].Reason)
}

// TestTier4_RealWorld_OrderCancellationAndInventoryRestockFlow tests customer order
// cancellation and automatic restocking via consumer inventory repository.
func TestTier4_RealWorld_OrderCancellationAndInventoryRestockFlow(t *testing.T) {
	runner := NewMockRunner(1)
	orderRepo := NewMockOrderRepository()
	invRepo := NewMockInventoryRepository()
	auditRepo := NewMockAuditRepository()

	orderID := uuid.New()
	varID := uuid.New()
	customerID := uuid.New()

	require.NoError(t, orderRepo.CreateOrder(context.Background(), &model.Order{
		ID:          orderID,
		CustomerID:  &customerID,
		TotalAmount: 9000,
		Status:      model.OrderStatusProcessing,
	}))

	// Cancel Order & Restock inside single transaction
	err := runner.WithTx(context.Background(), func(ctx context.Context) error {
		// Verify current order state
		o, err := orderRepo.GetOrderByID(ctx, orderID)
		if err != nil {
			return err
		}
		if o.Status != model.OrderStatusProcessing && o.Status != model.OrderStatusPending {
			return errors.New("cannot cancel non-cancellable order")
		}

		// Transition status
		if err := orderRepo.UpdateOrderStatus(ctx, orderID, model.OrderStatusCancelled); err != nil {
			return err
		}

		// Restock units
		if err := invRepo.CreateLedger(ctx, &model.InventoryLedger{
			VariantID: varID,
			Quantity:  2,
			Reason:    model.InventoryReasonReturn,
		}); err != nil {
			return err
		}

		// Record audit
		return auditRepo.Insert(ctx, &model.AuditLog{
			ID:         1,
			ActorID:    &customerID,
			Action:     "order.cancel",
			EntityType: "order",
		})
	})

	require.NoError(t, err)

	order, _ := orderRepo.GetOrderByID(context.Background(), orderID)
	assert.Equal(t, model.OrderStatusCancelled, order.Status)

	ledgers := invRepo.Ledgers()
	require.Len(t, ledgers, 1)
	assert.Equal(t, model.InventoryReasonReturn, ledgers[0].Reason)
	assert.Equal(t, 2, ledgers[0].Quantity)

	assert.Len(t, auditRepo.Entries(), 1)
	assert.Equal(t, "order.cancel", auditRepo.Entries()[0].Action)
}

// TestTier4_RealWorld_MediaUploadWithCompensatingCleanupOnDBFailure tests Rule 3:
// Upload before commit, compensating delete on DB rollback to avoid orphaned blobs.
func TestTier4_RealWorld_MediaUploadWithCompensatingCleanupOnDBFailure(t *testing.T) {
	runner := NewMockRunner(1)
	storageProvider := NewMockStorageProvider("https://cdn.prim.com")
	objRepo := NewMockObjectRepository()

	bucket := "product-images"
	key := "catalog/hoodie_front.png"
	fileData := bytes.NewReader([]byte("fake-png-binary-content"))

	// Step 1: External storage upload (outside DB transaction)
	err := storageProvider.Upload(context.Background(), bucket, key, fileData, int64(fileData.Len()), "image/png")
	require.NoError(t, err)
	assert.True(t, storageProvider.HasObject(bucket, key), "Blob must exist after upload")

	// Step 2: Attempt DB transaction that fails (e.g. unique constraint conflict)
	dbErr := runner.WithTx(context.Background(), func(ctx context.Context) error {
		_ = objRepo.Create(ctx, &model.Object{
			ID:     uuid.New(),
			Bucket: bucket,
			Key:    key,
		})
		return errors.New("simulated DB constraint failure")
	})
	require.Error(t, dbErr)

	// Step 3: Compensation - cleanup orphaned blob on DB failure (Rule 3)
	cleanupErr := storageProvider.Delete(context.Background(), bucket, key)
	require.NoError(t, cleanupErr)

	assert.False(t, storageProvider.HasObject(bucket, key), "Compensating delete must remove orphaned blob from storage")
	uploads, deletes, _ := storageProvider.Stats()
	assert.Equal(t, 1, uploads)
	assert.Equal(t, 1, deletes)
}

// TestTier4_RealWorld_MediaUploadAndURLResolutionSuccess verifies the success path
// for media uploads and in-memory URL generation without DB queries.
func TestTier4_RealWorld_MediaUploadAndURLResolutionSuccess(t *testing.T) {
	runner := NewMockRunner(1)
	storageProvider := NewMockStorageProvider("https://cdn.prim.com")
	objRepo := NewMockObjectRepository()

	bucket := "product-images"
	key := "catalog/banner_hero.webp"
	fileData := bytes.NewReader([]byte("hero-banner-image-bytes"))
	objID := uuid.New()

	// 1. Upload
	err := storageProvider.Upload(context.Background(), bucket, key, fileData, int64(fileData.Len()), "image/webp")
	require.NoError(t, err)

	// 2. Persist in DB
	err = runner.WithTx(context.Background(), func(ctx context.Context) error {
		return objRepo.Create(ctx, &model.Object{
			ID:          objID,
			Bucket:      bucket,
			Key:         key,
			ContentType: "image/webp",
			Status:      model.ObjectStatusUploaded,
		})
	})
	require.NoError(t, err)

	// 3. Resolve Public URL (Rule 3: in-memory, no DB, no network)
	publicURL := storageProvider.GetPublicURL(bucket, key)
	assert.Equal(t, "https://cdn.prim.com/product-images/catalog/banner_hero.webp", publicURL)
	assert.True(t, storageProvider.HasObject(bucket, key))
}

// TestTier4_RealWorld_AdminPrivilegedActionAuditTrail tests that privileged admin
// settings updates atomically record audit accountability trails.
func TestTier4_RealWorld_AdminPrivilegedActionAuditTrail(t *testing.T) {
	runner := NewMockRunner(1)
	auditRepo := NewMockAuditRepository()

	superadminID := uuid.New()
	settingKey := "store.maintenance_mode"
	settingValue := "true"

	err := runner.WithTx(context.Background(), func(ctx context.Context) error {
		// Perform mutation and audit log insertion together
		return auditRepo.Insert(ctx, &model.AuditLog{
			ID:         101,
			ActorID:    &superadminID,
			ActorRole:  string(model.SuperRole),
			Action:     "setting.maintenance_mode.enable",
			EntityType: "setting",
			EntityID:   settingKey,
			After:      []byte(settingValue),
		})
	})

	require.NoError(t, err)
	entries := auditRepo.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "setting.maintenance_mode.enable", entries[0].Action)
	assert.Equal(t, string(model.SuperRole), entries[0].ActorRole)
	assert.Equal(t, superadminID, *entries[0].ActorID)
	assert.Equal(t, "true", string(entries[0].After))
}
