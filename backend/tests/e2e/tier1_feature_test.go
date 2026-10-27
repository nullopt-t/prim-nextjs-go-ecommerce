package e2e

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// =============================================================================
// Tier 1: Feature Group A (F1 & F12) - Dead Twin Elimination & Direct Repo on Tx
// =============================================================================

func TestTier1_F1_F12_DirectObjectRepositoryOnTx(t *testing.T) {
	runner := NewMockRunner(1)
	repo := NewMockObjectRepository()

	objID := uuid.New()
	targetObj := &model.Object{
		ID:          objID,
		Bucket:      "storefront",
		Key:         "catalog/item1.png",
		FileSize:    2048,
		ContentType: "image/png",
		Status:      model.ObjectStatusUploaded,
	}

	// Directly call repository on tx - no ObjectService.CreateObjectWithTx needed
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		return repo.Create(ctx, tx, targetObj)
	})
	require.NoError(t, err)

	saved, err := repo.GetByID(context.Background(), nil, objID)
	require.NoError(t, err)
	assert.Equal(t, "storefront", saved.Bucket)
	assert.Equal(t, "catalog/item1.png", saved.Key)
	_, _, commits, _ := runner.Stats()
	assert.Equal(t, 1, commits, "Transaction must commit cleanly")
}

func TestTier1_F1_F12_DirectVariantRepositoryOnTx(t *testing.T) {
	runner := NewMockRunner(1)
	repo := NewMockVariantRepository()

	varID := uuid.New()
	sku := "SKU-TEST-100"
	repo.AddVariant(&model.ProductVariant{
		ID:        varID,
		ProductID: uuid.New(),
		SKU:       sku,
		Title:     "Blue / Large",
	})

	// Directly call repository on tx - no VariantService.GetVariantByIDWithTx needed
	var retrieved *model.ProductVariant
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		var err error
		retrieved, err = repo.GetByID(ctx, tx, varID)
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "SKU-TEST-100", retrieved.SKU)
}

func TestTier1_F1_F12_DirectProductRepositoryOnTx(t *testing.T) {
	runner := NewMockRunner(1)
	repo := NewMockProductRepository()

	prodID := uuid.New()
	repo.AddProduct(&model.Product{
		ID:          prodID,
		Title:       "Decoupled Hoodie",
		Slug:        "decoupled-hoodie",
		Status:      model.PublicationStatusDraft,
		ProductType: model.ProductTypeSimple,
	})

	// Direct call on tx - no ProductService.GetByIDWithTx needed
	var retrieved *model.Product
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		var err error
		retrieved, err = repo.GetByID(ctx, tx, prodID)
		return err
	})
	require.NoError(t, err)
	assert.Equal(t, "Decoupled Hoodie", retrieved.Title)
}

func TestTier1_F1_F12_DirectUserRepositoryOnTx(t *testing.T) {
	runner := NewMockRunner(1)
	userRepo := NewMockUserRepository()

	userID := uuid.New()
	name := "Decoupled Tester"
	userRepo.AddUser(&model.User{
		ID:         userID,
		Identifier: "test@example.com",
		FullName:   &name,
		Role:       model.CustomerRole,
		Status:     model.StatusActive,
	})

	// Direct call on tx - no UserService.GetUserByIDWithTx needed
	var found *model.User
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		var err error
		found, err = userRepo.GetByID(ctx, tx, userID)
		return err
	})
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", found.Identifier)
}

func TestTier1_F1_F12_DirectAddressRepositoryOnTx(t *testing.T) {
	runner := NewMockRunner(1)
	addrRepo := NewMockAddressRepository()

	userID := uuid.New()
	addrID := uuid.New()
	postal := "12345"
	addrRepo.AddAddress(&model.UserAddress{
		ID:         addrID,
		UserID:     userID,
		FullName:   "Decoupled Buyer",
		PostalCode: &postal,
	})

	// Direct call on tx - no AddressService.GetAddressWithTx needed
	var addr *model.UserAddress
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		var err error
		addr, err = addrRepo.GetByIDAndUser(ctx, tx, addrID, userID)
		return err
	})
	require.NoError(t, err)
	assert.Equal(t, "Decoupled Buyer", addr.FullName)
}

// =============================================================================
// Tier 1: Feature Group B (F2) - Product & Variant Decoupling
// =============================================================================

func TestTier1_F2_PublishProduct_RequiresAtLeastOneVariantViaRepo(t *testing.T) {
	runner := NewMockRunner(1)
	prodRepo := NewMockProductRepository()
	variantRepo := NewMockVariantRepository()

	prodID := uuid.New()
	prodRepo.AddProduct(&model.Product{
		ID:     prodID,
		Title:  "Zero-Variant Jacket",
		Status: model.PublicationStatusDraft,
	})

	// Simulate PublishProduct: checks variantRepo.CountByProductID (no VariantService injection)
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		count, err := variantRepo.CountByProductID(ctx, tx, prodID)
		if err != nil {
			return err
		}
		if count == 0 {
			return apierr.BadRequestError("Cannot publish a product with no variants")
		}
		return prodRepo.UpdateStatus(ctx, tx, prodID, model.PublicationStatusPublished)
	})

	require.Error(t, err)
	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
}

func TestTier1_F2_PublishProduct_TransitionsToPublishedWhenVariantsExist(t *testing.T) {
	runner := NewMockRunner(1)
	prodRepo := NewMockProductRepository()
	variantRepo := NewMockVariantRepository()

	prodID := uuid.New()
	prodRepo.AddProduct(&model.Product{
		ID:     prodID,
		Title:  "Stocked Jacket",
		Status: model.PublicationStatusDraft,
	})
	sku := "SKU-JACKET-1"
	variantRepo.AddVariant(&model.ProductVariant{
		ID:        uuid.New(),
		ProductID: prodID,
		SKU:       sku,
	})

	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		count, err := variantRepo.CountByProductID(ctx, tx, prodID)
		if err != nil {
			return err
		}
		if count == 0 {
			return apierr.BadRequestError("Cannot publish a product with no variants")
		}
		return prodRepo.UpdateStatus(ctx, tx, prodID, model.PublicationStatusPublished)
	})

	require.NoError(t, err)
	p, err := prodRepo.GetByID(context.Background(), nil, prodID)
	require.NoError(t, err)
	assert.Equal(t, model.PublicationStatusPublished, p.Status)
}

func TestTier1_F2_Product_ResolvesStockViaInventoryRepo(t *testing.T) {
	runner := NewMockRunner(1)
	inventoryRepo := NewMockInventoryRepository()

	var1 := uuid.New()
	var2 := uuid.New()
	inventoryRepo.SetStock(var1, 20, 5)
	inventoryRepo.SetStock(var2, 10, 0)

	// Single WithExecutor reading stock for multiple variants via consumer repository
	var stockMap map[uuid.UUID]*model.InventoryStock
	err := runner.WithExecutor(context.Background(), func(exec database.QueryExecutor) error {
		var err error
		stockMap, err = inventoryRepo.GetStockForVariants(context.Background(), exec, []uuid.UUID{var1, var2})
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, 15, stockMap[var1].AvailableQuantity)
	assert.Equal(t, 10, stockMap[var2].AvailableQuantity)
	withDB, withTx, _, _ := runner.Stats()
	assert.Equal(t, 1, withDB)
	assert.Equal(t, 0, withTx)
}

func TestTier1_F2_Product_DirectStorageURLResolution(t *testing.T) {
	storageProvider := NewMockStorageProvider("https://assets.prim.com")
	bucket := "catalog"
	key := "products/shoes.webp"

	// Rule 3: Asset URL generated directly in-memory via storageProvider.GetPublicURL
	url := storageProvider.GetPublicURL(bucket, key)
	assert.Equal(t, "https://assets.prim.com/catalog/products/shoes.webp", url)
	_, _, lookups := storageProvider.Stats()
	assert.Equal(t, 1, lookups)
}

func TestTier1_F2_Product_InventoryStockLockOnSingleTx(t *testing.T) {
	runner := NewMockRunner(1)
	inventoryRepo := NewMockInventoryRepository()
	varID := uuid.New()

	// Locking row and updating stock strictly within a single WithTx
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		if err := inventoryRepo.LockVariantForUpdate(ctx, tx, varID); err != nil {
			return err
		}
		return inventoryRepo.CreateLedger(ctx, tx, &model.InventoryLedger{
			VariantID: varID,
			Quantity:  -2,
			Reason:    model.InventoryReasonSale,
		})
	})

	require.NoError(t, err)
	assert.Len(t, inventoryRepo.Ledgers(), 1)
	assert.Equal(t, -2, inventoryRepo.Ledgers()[0].Quantity)
}

// =============================================================================
// Tier 1: Feature Group C (F3 & F4) - Supporting Domains Error Mapping & Threading
// =============================================================================

func TestTier1_F3_Object_HandleError_NotFound(t *testing.T) {
	err := CanonicalHandleError(pgx.ErrNoRows, "Media")
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.Status)
	assert.Contains(t, apiErr.Error(), "Media not found")
}

func TestTier1_F3_Object_HandleError_Conflict(t *testing.T) {
	err := CanonicalHandleError(database.ErrConflict, "Object")
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusConflict, apiErr.Status)
}

func TestTier1_F3_Setting_HandleError_DatabaseFailure(t *testing.T) {
	driverErr := errors.New("pq: connection terminated unexpectedly")
	err := CanonicalHandleError(driverErr, "load settings")
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusInternalServerError, apiErr.Status)
	assert.False(t, strings.Contains(apiErr.Error(), "pq:"), "Database driver internals must not be leaked")
}

func TestTier1_F4_Object_CreateObject_SingleWithTx(t *testing.T) {
	runner := NewMockRunner(1)
	repo := NewMockObjectRepository()

	// Rule 2 & Feature F4: Mutating operations use WithTx
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		return repo.Create(ctx, tx, &model.Object{
			ID:     uuid.New(),
			Bucket: "media",
			Key:    "banner.jpg",
			Status: model.ObjectStatusUploaded,
		})
	})

	require.NoError(t, err)
	withDB, withTx, commits, _ := runner.Stats()
	assert.Equal(t, 0, withDB)
	assert.Equal(t, 1, withTx)
	assert.Equal(t, 1, commits)
	assert.False(t, runner.ConnectionHopDetected())
}

func TestTier1_F4_Audit_Record_SingleConnection(t *testing.T) {
	runner := NewMockRunner(1)
	repo := NewMockAuditRepository()

	actorID := uuid.New()
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		return repo.Insert(ctx, tx, &model.AuditLog{
			ID:         1,
			ActorID:    &actorID,
			Action:     "UPDATE_SETTING",
			EntityType: "settings",
		})
	})

	require.NoError(t, err)
	assert.Len(t, repo.Entries(), 1)
	assert.Equal(t, "UPDATE_SETTING", repo.Entries()[0].Action)
	assert.False(t, runner.ConnectionHopDetected())
}

// =============================================================================
// Tier 1: Feature Group D (F5, F6 & F7) - Order Domain Decoupling & Error Handling
// =============================================================================

func TestTier1_F5_Order_Cancellation_ReleasesHoldViaConsumerInventoryRepo(t *testing.T) {
	runner := NewMockRunner(1)
	inventoryRepo := NewMockInventoryRepository()
	orderID := uuid.New()
	resID := uuid.New()

	require.NoError(t, inventoryRepo.CreateReservation(context.Background(), nil, &model.InventoryReservation{
		ID:        resID,
		OrderID:   &orderID,
		Quantity:  3,
		ExpiresAt: time.Now().Add(time.Hour),
	}))

	// Order cancellation calls consumer inventoryRepo.ReleaseReservationsForOrder on held qe
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		released, err := inventoryRepo.ReleaseReservationsForOrder(ctx, tx, orderID)
		if err != nil {
			return CanonicalHandleError(err, "cancel order")
		}
		assert.Equal(t, int64(1), released)
		return nil
	})

	require.NoError(t, err)
	unreleased, _ := inventoryRepo.GetUnreleasedReservationsForOrder(context.Background(), nil, orderID)
	assert.Empty(t, unreleased, "All reservations must be released")
}

func TestTier1_F5_Order_Cancellation_RestocksViaConsumerLedger(t *testing.T) {
	runner := NewMockRunner(1)
	inventoryRepo := NewMockInventoryRepository()
	varID := uuid.New()

	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		return inventoryRepo.CreateLedger(ctx, tx, &model.InventoryLedger{
			VariantID: varID,
			Quantity:  2,
			Reason:    model.InventoryReasonReturn,
		})
	})

	require.NoError(t, err)
	ledgers := inventoryRepo.Ledgers()
	require.Len(t, ledgers, 1)
	assert.Equal(t, model.InventoryReasonReturn, ledgers[0].Reason)
	assert.Equal(t, 2, ledgers[0].Quantity)
}

func TestTier1_F6_Order_CreateOrderWithTx_TakesExplicitTx(t *testing.T) {
	tx, fakePgx := NewTestTx()
	orderRepo := NewMockOrderRepository()
	orderID := uuid.New()

	// Simulate OrderService.CreateOrderWithTx taking explicit *database.Tx
	createOrderWithTx := func(ctx context.Context, callerTx *database.Tx, in *model.Order) error {
		if callerTx == nil {
			return errors.New("nil transaction provided")
		}
		return orderRepo.CreateOrder(ctx, callerTx, in)
	}

	err := createOrderWithTx(context.Background(), tx, &model.Order{
		ID:          orderID,
		TotalAmount: 4999,
		Status:      model.OrderStatusPending,
	})

	require.NoError(t, err)
	require.False(t, fakePgx.committed, "Caller transaction must remain uncommitted until caller decides")
	saved, err := orderRepo.GetOrderByID(context.Background(), nil, orderID)
	require.NoError(t, err)
	assert.Equal(t, int64(4999), saved.TotalAmount)
}

func TestTier1_F7_Order_HandleError_TrackingNotFound(t *testing.T) {
	err := CanonicalHandleError(database.ErrNotFound, "update order tracking")
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.Status)
}

func TestTier1_F7_Order_HandleError_ForeignKeyViolation(t *testing.T) {
	err := CanonicalHandleError(database.ErrForeignKeyViolation, "create order line item")
	require.Error(t, err)

	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.Contains(t, apiErr.Error(), "Referenced entity does not exist")
}

// =============================================================================
// Tier 1: Feature Group E (F8, F9, F10 & F11) - Checkout Domain Decoupling
// =============================================================================

func TestTier1_F8_Checkout_ValidatesUserViaConsumerRepo(t *testing.T) {
	runner := NewMockRunner(1)
	userRepo := NewMockUserRepository()

	userID := uuid.New()
	userRepo.AddUser(&model.User{
		ID:         userID,
		Identifier: "customer@example.com",
		Status:     model.StatusActive,
		Role:       model.CustomerRole,
	})

	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		u, err := userRepo.GetByID(ctx, tx, userID)
		if err != nil {
			return CanonicalHandleError(err, "customer")
		}
		if u.Status != model.StatusActive {
			return apierr.ForbiddenError("Account is suspended")
		}
		return nil
	})

	require.NoError(t, err)
}

func TestTier1_F8_Checkout_ValidatesAddressOwnershipViaConsumerRepo(t *testing.T) {
	runner := NewMockRunner(1)
	addrRepo := NewMockAddressRepository()

	ownerID := uuid.New()
	imposterID := uuid.New()
	addrID := uuid.New()
	addrRepo.AddAddress(&model.UserAddress{
		ID:     addrID,
		UserID: ownerID,
	})

	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		_, err := addrRepo.GetByIDAndUser(ctx, tx, addrID, imposterID)
		if err != nil {
			return CanonicalHandleError(err, "shipping address")
		}
		return nil
	})

	require.Error(t, err)
	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusNotFound, apiErr.Status)
}

func TestTier1_F8_Checkout_ReservesStockViaConsumerRepo(t *testing.T) {
	runner := NewMockRunner(1)
	inventoryRepo := NewMockInventoryRepository()

	varID := uuid.New()
	orderID := uuid.New()
	inventoryRepo.SetStock(varID, 10, 2) // available = 8

	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		stock, err := inventoryRepo.GetStock(ctx, tx, varID)
		if err != nil {
			return CanonicalHandleError(err, "stock")
		}
		if stock.AvailableQuantity < 5 {
			return apierr.BadRequestError("insufficient stock").WithCode("INSUFFICIENT_STOCK")
		}
		return inventoryRepo.CreateReservation(ctx, tx, &model.InventoryReservation{
			ID:        uuid.New(),
			VariantID: varID,
			OrderID:   &orderID,
			Quantity:  5,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		})
	})

	require.NoError(t, err)
	unreleased, _ := inventoryRepo.GetUnreleasedReservationsForOrder(context.Background(), nil, orderID)
	require.Len(t, unreleased, 1)
	assert.Equal(t, 5, unreleased[0].Quantity)
}

func TestTier1_F9_Checkout_DirectStorageResolutionInLineItems(t *testing.T) {
	storageProvider := NewMockStorageProvider("https://cdn.prim.com")
	bucket := "store-assets"
	key := "products/shoes_thumb.webp"

	// Resolves URL directly in-memory without cross-domain DB or service calls
	publicURL := storageProvider.GetPublicURL(bucket, key)
	assert.Equal(t, "https://cdn.prim.com/store-assets/products/shoes_thumb.webp", publicURL)
}

func TestTier1_F10_Checkout_ValidateCoupon_SingleWithExecutor(t *testing.T) {
	runner := NewMockRunner(1)
	couponRepo := NewMockCouponRepository()

	couponID := uuid.New()
	couponRepo.AddCoupon(&model.Coupon{
		ID:            couponID,
		Code:          "SAVE10",
		DiscountType:  model.CouponDiscountTypePercentage,
		DiscountValue: 10,
		Status:        model.CouponStatusActive,
	})

	// Rule 2 & Feature F10: ValidateCoupon executes within a single WithExecutor
	var discount int64
	err := runner.WithExecutor(context.Background(), func(exec database.QueryExecutor) error {
		c, err := couponRepo.GetByCode(context.Background(), exec, "SAVE10")
		if err != nil {
			return CanonicalHandleError(err, "coupon")
		}
		discount = 1000 * c.DiscountValue / 100
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, int64(100), discount)
	withDB, withTx, _, _ := runner.Stats()
	assert.Equal(t, 1, withDB)
	assert.Equal(t, 0, withTx)
	assert.False(t, runner.ConnectionHopDetected())
}

func TestTier1_F10_Checkout_SingleWithTxBundling(t *testing.T) {
	runner := NewMockRunner(1)
	cartRepo := NewMockCartRepository()
	orderRepo := NewMockOrderRepository()

	cartID := uuid.New()
	userID := uuid.New()
	orderID := uuid.New()
	cartRepo.SetCart(&model.Cart{
		ID:     cartID,
		UserID: &userID,
		Items:  []model.CartItem{{ID: 1, Quantity: 2}},
	})

	// Feature F10: Entire checkout executes within 1 WithTx without opening second connection
	err := runner.WithTx(context.Background(), func(ctx context.Context, tx database.QueryExecutor) error {
		// 1. Create order
		if err := orderRepo.CreateOrder(ctx, tx, &model.Order{ID: orderID, TotalAmount: 2000}); err != nil {
			return err
		}
		// 2. Clear cart
		return cartRepo.ClearCart(ctx, tx, cartID)
	})

	require.NoError(t, err)
	withDB, withTx, commits, _ := runner.Stats()
	assert.Equal(t, 0, withDB)
	assert.Equal(t, 1, withTx)
	assert.Equal(t, 1, commits)
	assert.False(t, runner.ConnectionHopDetected())
}

func TestTier1_F11_Checkout_HandleError_EmptyCart(t *testing.T) {
	cartRepo := NewMockCartRepository()
	userID := uuid.New()

	checkoutFn := func(ctx context.Context, qe database.QueryExecutor) error {
		c, err := cartRepo.GetCartByUserID(ctx, qe, userID)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				return apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
			}
			return CanonicalHandleError(err, "cart")
		}
		if len(c.Items) == 0 {
			return apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
		}
		return nil
	}

	err := checkoutFn(context.Background(), nil)
	require.Error(t, err)
	var apiErr *apierr.APIError
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, "EMPTY_CART", apiErr.Code)
}
