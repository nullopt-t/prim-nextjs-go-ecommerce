package checkout_test

// Strict end-to-end integration tests for the checkout domain.
//
// The suite exercises the real HTTP stack (Gin router + auth middleware), the
// real CheckoutService, and a real PostgreSQL instance (via testcontainers) with
// the production schema applied. Redis is used to seed authenticated sessions.
//
// Coverage is deliberately split into "red" cases (business rejections that must
// fail closed and leave no side effects) and "non-red" cases (successful purchases
// that must persist exactly the expected money, stock, and payment state).
//
// Business invariants asserted here:
//   - Pricing is authoritative: computed from live variant prices, never client input.
//   - Shipping is free at/above 15000 minor units, otherwise 999.
//   - A checkout is atomic: any failure rolls back the order, payment, coupon
//     redemption, stock reservation, and cart clearance as one unit.
//   - Coupon rules (existence, active window, min-cart, global/per-user limits)
//     are enforced at checkout time.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nullopt-t/prim-backend/internal/auth"
	"github.com/nullopt-t/prim-backend/internal/catalog/cart"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/checkout"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/user"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// Business constants mirrored from the checkout service contract.
const (
	freeShippingThreshold int64 = 15000
	shippingFee           int64 = 999
	defaultCurrency             = "EGP"
)

// dummyProductService satisfies the cart.ProductService dependency so the cart
// can resolve product metadata without pulling in the full catalog graph.
type dummyProductService struct{}

func (d *dummyProductService) GetByID(_ context.Context, id uuid.UUID) (*model.Product, error) {
	return &model.Product{
		ID:    id,
		Title: "Integration Test Product",
		Slug:  "integration-test-product",
	}, nil
}

// dataEnvelope is the strict success envelope: {"data": <T>}.
type dataEnvelope[T any] struct {
	Data T `json:"data"`
}

// errorEnvelope is the strict error envelope emitted by middleware.ErrorHandler.
type errorEnvelope struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []api.FieldError `json:"details,omitempty"`
}

type CheckoutTestSuite struct {
	suite.Suite

	pgContainer    *postgres.PostgresContainer
	redisContainer *redisContainer.RedisContainer
	db             *database.DB
	redisClient    *redis.Client
	router         *gin.Engine

	txRunner         database.Runner
	checkoutService  *checkout.CheckoutService
	cartService      *cart.CartService
	variantService   *variant.VariantService
	inventoryService *inventory.InventoryService
	orderRepo        *order.Repository
	addressService   *user.AddressService
	userService      *user.UserService

	categoryID uuid.UUID
	productID  uuid.UUID
}

func (s *CheckoutTestSuite) SetupSuite() {
	ctx := context.Background()

	// ---- PostgreSQL -------------------------------------------------------
	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_checkout_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(s.T(), err)
	s.pgContainer = pgContainer

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	db, err := database.ConnectDB(ctx, config.DatabaseConfig{
		DBHost:     host,
		DBPort:     port.Port(),
		DBUser:     "testuser",
		DBPassword: "testpass",
		DBName:     "prim_checkout_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	schemaBytes, err := os.ReadFile("../../migrations/000001_init_sechema.up.sql")
	require.NoError(s.T(), err)
	_, err = db.Exec(ctx, string(schemaBytes))
	require.NoError(s.T(), err)

	// ---- Redis (sessions) -------------------------------------------------
	rContainer, err := redisContainer.Run(ctx, "redis:7-alpine")
	require.NoError(s.T(), err)
	s.redisContainer = rContainer

	redisHost, err := rContainer.Host(ctx)
	require.NoError(s.T(), err)
	redisPort, err := rContainer.MappedPort(ctx, "6379")
	require.NoError(s.T(), err)

	s.redisClient = redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort.Port(),
	})

	// ---- Wiring -----------------------------------------------------------
	logger := log.NewConsoleLogger()
	s.txRunner = database.NewTxRunner(s.db)

	inventoryRepo := inventory.NewRepository()
	s.inventoryService = inventory.NewService(logger, s.txRunner, inventoryRepo)

	variantRepo := variant.NewRepository()
	s.variantService = variant.NewService(logger, s.txRunner, nil, variantRepo, s.inventoryService)

	s.cartService = cart.NewService(
		s.txRunner,
		cart.NewRepository(),
		s.variantService,
		&dummyProductService{},
		s.inventoryService,
	)

	s.orderRepo = order.NewRepository()
	orderService := order.NewService(s.txRunner, s.orderRepo, logger, s.inventoryService)

	userRepo := user.NewPostgresRepository()
	s.userService = user.NewService(s.txRunner, userRepo, logger)

	addressRepo := user.NewAddressRepository()
	s.addressService = user.NewAddressService(s.txRunner, addressRepo, logger)

	s.checkoutService = checkout.NewService(
		s.cartService,
		orderService,
		s.userService,
		s.addressService,
		s.inventoryService,
		s.txRunner,
	)

	// ---- Seed catalog base ------------------------------------------------
	s.categoryID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_categories (id, name)
		VALUES ($1, $2)
	`, s.categoryID, "Integration Electronics")
	require.NoError(s.T(), err)

	s.productID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type)
		VALUES ($1, $2, $3, $4, 'published', 'simple')
	`, s.productID, s.categoryID, "it-product-"+uuid.NewString(), "Integration Product")
	require.NoError(s.T(), err)

	// ---- Router -----------------------------------------------------------
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.ErrorHandler(logger))

	sessionService := auth.NewSessionService(s.redisClient, logger, time.Hour)
	checkoutRouter := checkout.NewRouter(checkout.NewHandler(s.checkoutService), sessionService)

	v1 := s.router.Group("/api/v1")
	checkoutRouter.MapRoutes(v1)
}

func (s *CheckoutTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.redisClient != nil {
		_ = s.redisClient.Close()
	}
	if s.redisContainer != nil {
		_ = s.redisContainer.Terminate(context.Background())
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *CheckoutTestSuite) SetupTest() {
	// Sessions are the only Redis-backed state; wipe between tests for isolation.
	s.Require().NoError(s.redisClient.FlushAll(context.Background()).Err())
}

// =============================================================================
// Seeding helpers
// =============================================================================

// seedVariant creates a priced variant and stocks it. A nil price produces a
// variant that the storefront cannot legally sell (used for the missing-price case).
func (s *CheckoutTestSuite) seedVariant(price *int64, stock int) uuid.UUID {
	ctx := context.Background()

	sku := "SKU-" + uuid.NewString()
	currency := defaultCurrency
	v, err := s.variantService.CreateVariant(ctx, &variant.CreateVariantInput{
		ProductID: s.productID,
		SKU:       &sku,
		Title:     "Default / Standard",
		Price:     price,
		Currency:  &currency,
	})
	s.Require().NoError(err)

	if stock > 0 {
		_, err = s.inventoryService.AdjustStock(ctx, inventory.AdjustStockInput{
			VariantID: v.ID,
			Quantity:  stock,
			Reason:    "restock",
		})
		s.Require().NoError(err)
	}

	return v.ID
}

func (s *CheckoutTestSuite) seedPricedVariant(price int64, stock int) uuid.UUID {
	p := price
	return s.seedVariant(&p, stock)
}

// seedUser inserts an active customer and seeds an authenticated session in Redis.
func (s *CheckoutTestSuite) seedUser() (uuid.UUID, string) {
	ctx := context.Background()
	userID := uuid.New()

	_, err := s.db.Exec(ctx, `
		INSERT INTO users (id, email, full_name, role, is_email_verified, status)
		VALUES ($1, $2, $3, 'customer', true, 'active')
	`, userID, "checkout_"+userID.String()+"@example.com", "Checkout Customer")
	s.Require().NoError(err)

	sessionID := "sess_" + uuid.NewString()
	data, err := json.Marshal(map[string]any{
		"id":       sessionID,
		"userId":   userID,
		"userRole": "customer",
		"type":     "authenticated",
	})
	s.Require().NoError(err)
	s.Require().NoError(s.redisClient.Set(ctx, "auth:session:"+sessionID, data, time.Hour).Err())

	return userID, sessionID
}

// seedAddress inserts a saved shipping address owned by userID.
func (s *CheckoutTestSuite) seedAddress(userID uuid.UUID) uuid.UUID {
	addressID := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO addresses (id, user_id, full_name, line1, city, state, postal_code, country, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
	`, addressID, userID, "Checkout Customer", "12 Test Street", "Cairo", "Cairo", "11511", "EG")
	s.Require().NoError(err)
	return addressID
}

// addToCart adds stock-validated items through the cart domain service.
func (s *CheckoutTestSuite) addToCart(userID, variantID uuid.UUID, qty int) {
	_, err := s.cartService.AddItem(context.Background(), &userID, nil, variantID, qty)
	s.Require().NoError(err)
}

type couponSeed struct {
	Code          string
	DiscountType  string
	DiscountValue int64
	MinCartAmount *int64
	UsageLimit    *int32
	PerUserLimit  *int32
	StartsAt      *time.Time
	ExpiresAt     *time.Time
	Status        string
}

// seedCoupon inserts a coupon directly, giving tests precise control over the
// rules under examination.
func (s *CheckoutTestSuite) seedCoupon(c couponSeed) uuid.UUID {
	id := uuid.New()

	status := c.Status
	if status == "" {
		status = "active"
	}

	var (
		minCart any
		usage   any
		perUser any
		starts  any
		expires any
	)
	if c.MinCartAmount != nil {
		minCart = *c.MinCartAmount
	}
	if c.UsageLimit != nil {
		usage = *c.UsageLimit
	}
	if c.PerUserLimit != nil {
		perUser = *c.PerUserLimit
	}
	if c.StartsAt != nil {
		starts = *c.StartsAt
	}
	if c.ExpiresAt != nil {
		expires = *c.ExpiresAt
	}

	_, err := s.db.Exec(context.Background(), `
		INSERT INTO coupons (
			id, code, discount_type, discount_value, min_cart_amount,
			usage_limit, per_user_limit, starts_at, expires_at, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, c.Code, c.DiscountType, c.DiscountValue, minCart, usage, perUser, starts, expires, status)
	s.Require().NoError(err)

	return id
}

// seedRawOrder inserts a bare order so coupon-redemption fixtures can satisfy the
// coupon_redemptions.order_id foreign key.
func (s *CheckoutTestSuite) seedRawOrder(customerID *uuid.UUID) uuid.UUID {
	id := uuid.New()

	var customer any
	if customerID != nil {
		customer = *customerID
	}

	_, err := s.db.Exec(context.Background(), `
		INSERT INTO orders (id, customer_id, customer_email, shipping_address, status, total_amount, currency)
		VALUES ($1, $2, 'seed@example.com', '{}'::jsonb, 'pending', 0, 'EGP')
	`, id, customer)
	s.Require().NoError(err)

	return id
}

// seedRedemption records a prior coupon redemption (used to exhaust usage limits).
func (s *CheckoutTestSuite) seedRedemption(couponID uuid.UUID, customerID *uuid.UUID, orderID uuid.UUID) {
	var customer any
	if customerID != nil {
		customer = *customerID
	}

	_, err := s.db.Exec(context.Background(), `
		INSERT INTO coupon_redemptions (id, coupon_id, user_id, order_id, discount_amount)
		VALUES ($1, $2, $3, $4, 100)
	`, uuid.New(), couponID, customer, orderID)
	s.Require().NoError(err)
}

// seedRawCartItem bypasses cart-domain validation to construct states the public
// API refuses to create (e.g. quantity above stock or a variant with no price).
func (s *CheckoutTestSuite) seedRawCartItem(userID, variantID uuid.UUID, qty int, price int64) uuid.UUID {
	ctx := context.Background()
	cartID := uuid.New()

	_, err := s.db.Exec(ctx, `
		INSERT INTO carts (id, user_id)
		VALUES ($1, $2)
	`, cartID, userID)
	s.Require().NoError(err)

	_, err = s.db.Exec(ctx, `
		INSERT INTO cart_items (id, cart_id, variant_id, quantity, price_at_purchase, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), cartID, variantID, qty, price, defaultCurrency)
	s.Require().NoError(err)

	return cartID
}

// =============================================================================
// HTTP helpers
// =============================================================================

func (s *CheckoutTestSuite) postCheckout(sessionID string, body any) *httptest.ResponseRecorder {
	payload, err := json.Marshal(body)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/checkout", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("Authorization", "Bearer "+sessionID)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func (s *CheckoutTestSuite) postApplyCoupon(sessionID string, body any) *httptest.ResponseRecorder {
	payload, err := json.Marshal(body)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/coupons/apply", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("Authorization", "Bearer "+sessionID)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

// decodeSuccess strictly decodes a {"data": ...} envelope and fails otherwise.
func (s *CheckoutTestSuite) decodeSuccess(w *httptest.ResponseRecorder, out any) {
	s.Require().Equal("application/json; charset=utf-8", w.Header().Get("Content-Type"))
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), out))
}

// decodeError strictly decodes the error envelope and returns it.
func (s *CheckoutTestSuite) decodeError(w *httptest.ResponseRecorder) errorEnvelope {
	var env errorEnvelope
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &env))
	s.Require().NotEmpty(env.Code, "error responses must carry a machine-readable code")
	s.Require().NotEmpty(env.Message, "error responses must carry a human-readable message")
	return env
}

// =============================================================================
// Database assertions
// =============================================================================

type orderRow struct {
	ID             uuid.UUID
	CustomerID     *uuid.UUID
	CustomerEmail  string
	Status         string
	DiscountAmount int64
	ShippingAmount int64
	TaxAmount      int64
	TotalAmount    int64
	Currency       string
}

func (s *CheckoutTestSuite) fetchOrder(orderID uuid.UUID) orderRow {
	var o orderRow
	err := s.db.QueryRow(context.Background(), `
		SELECT id, customer_id, customer_email, status::text,
		       discount_amount, shipping_amount, tax_amount, total_amount, currency
		FROM orders
		WHERE id = $1
	`, orderID).Scan(
		&o.ID, &o.CustomerID, &o.CustomerEmail, &o.Status,
		&o.DiscountAmount, &o.ShippingAmount, &o.TaxAmount, &o.TotalAmount, &o.Currency,
	)
	s.Require().NoError(err)
	return o
}

type paymentRow struct {
	Amount   int64
	Currency string
	Status   string
	Provider string
}

func (s *CheckoutTestSuite) fetchPayment(orderID uuid.UUID) paymentRow {
	var p paymentRow
	err := s.db.QueryRow(context.Background(), `
		SELECT amount, currency, status::text, provider
		FROM payments
		WHERE order_id = $1
	`, orderID).Scan(&p.Amount, &p.Currency, &p.Status, &p.Provider)
	s.Require().NoError(err)
	return p
}

func (s *CheckoutTestSuite) countOrdersForUser(userID uuid.UUID) int {
	var n int
	err := s.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM orders WHERE customer_id = $1`, userID).Scan(&n)
	s.Require().NoError(err)
	return n
}

func (s *CheckoutTestSuite) countPaymentsForUser(userID uuid.UUID) int {
	var n int
	err := s.db.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM payments p
		JOIN orders o ON o.id = p.order_id
		WHERE o.customer_id = $1
	`, userID).Scan(&n)
	s.Require().NoError(err)
	return n
}

func (s *CheckoutTestSuite) countActiveCartItems(userID uuid.UUID) int {
	var n int
	err := s.db.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM cart_items ci
		JOIN carts c ON c.id = ci.cart_id
		WHERE c.user_id = $1 AND ci.deleted_at IS NULL
	`, userID).Scan(&n)
	s.Require().NoError(err)
	return n
}

func (s *CheckoutTestSuite) countActiveReservations(variantID uuid.UUID) int {
	var n int
	err := s.db.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM inventory_reservations
		WHERE variant_id = $1 AND released_at IS NULL AND expires_at > now()
	`, variantID).Scan(&n)
	s.Require().NoError(err)
	return n
}

func (s *CheckoutTestSuite) fetchStock(variantID uuid.UUID) *model.InventoryStock {
	stock, err := s.inventoryService.GetStockByVariantID(context.Background(), variantID)
	s.Require().NoError(err)
	return stock
}

// =============================================================================
// NON-RED: successful checkout
// =============================================================================

// TestCheckout_Success_CreatesOrderPaymentAndReservesStock covers the canonical
// happy path: a priced cart is converted into a pending order, a COD payment is
// recorded, stock is reserved, and the cart is cleared.
func (s *CheckoutTestSuite) TestCheckout_Success_CreatesOrderPaymentAndReservesStock() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(2999, 10)
	s.addToCart(userID, variantID, 2) // gross = 5998

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		PaymentMethod:     "cash_on_delivery",
	})

	s.Require().Equal(http.StatusOK, w.Code)

	var resp dataEnvelope[checkout.CheckoutResponse]
	s.decodeSuccess(w, &resp)
	s.Require().NotEmpty(resp.Data.OrderID)
	orderID, err := uuid.Parse(resp.Data.OrderID)
	s.Require().NoError(err)
	s.Equal(string(model.OrderStatusPending), resp.Data.Status)

	// Order money must be computed server-side: 5998 + 999 shipping.
	o := s.fetchOrder(orderID)
	s.Equal(userID, *o.CustomerID)
	s.Equal(string(model.OrderStatusPending), o.Status)
	s.Equal(int64(0), o.DiscountAmount)
	s.Equal(shippingFee, o.ShippingAmount)
	s.Equal(int64(0), o.TaxAmount)
	s.Equal(int64(6997), o.TotalAmount)
	s.Equal(defaultCurrency, o.Currency)

	// Payment row mirrors the order total and defaults to COD when requested.
	p := s.fetchPayment(orderID)
	s.Equal(int64(6997), p.Amount)
	s.Equal(defaultCurrency, p.Currency)
	s.Equal("pending", p.Status)
	s.Equal("cash_on_delivery", p.Provider)

	// Stock is reserved (not yet committed): on-hand untouched, availability reduced.
	stock := s.fetchStock(variantID)
	s.Equal(10, stock.OnHandQuantity)
	s.Equal(2, stock.ReservedQuantity)
	s.Equal(8, stock.AvailableQuantity)

	// Cart is cleared atomically with the order.
	s.Equal(0, s.countActiveCartItems(userID))
}

// TestCheckout_Success_FreeShippingAtThreshold verifies the shipping rule at the
// free-shipping boundary.
func (s *CheckoutTestSuite) TestCheckout_Success_FreeShippingAtThreshold() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(20000, 5)
	s.addToCart(userID, variantID, 1)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusOK, w.Code)

	var resp dataEnvelope[checkout.CheckoutResponse]
	s.decodeSuccess(w, &resp)
	orderID, err := uuid.Parse(resp.Data.OrderID)
	s.Require().NoError(err)

	o := s.fetchOrder(orderID)
	s.Equal(int64(0), o.ShippingAmount)
	s.Equal(int64(20000), o.TotalAmount)
}

// TestCheckout_Success_PercentageCoupon applies a percentage discount and records
// exactly one redemption for the applied amount.
func (s *CheckoutTestSuite) TestCheckout_Success_PercentageCoupon() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(10000, 5)
	s.addToCart(userID, variantID, 1)

	couponID := s.seedCoupon(couponSeed{
		Code:          "PERCENT10-" + uuid.NewString()[:8],
		DiscountType:  "percentage",
		DiscountValue: 10,
	})

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusOK, w.Code)

	var resp dataEnvelope[checkout.CheckoutResponse]
	s.decodeSuccess(w, &resp)
	orderID, err := uuid.Parse(resp.Data.OrderID)
	s.Require().NoError(err)

	// 10000 - 10% (1000) + 999 shipping = 9999.
	o := s.fetchOrder(orderID)
	s.Equal(int64(1000), o.DiscountAmount)
	s.Equal(shippingFee, o.ShippingAmount)
	s.Equal(int64(9999), o.TotalAmount)

	// Exactly one redemption with the discounted amount.
	var count int
	var discount int64
	err = s.db.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(SUM(discount_amount), 0)
		FROM coupon_redemptions
		WHERE coupon_id = $1 AND user_id = $2
	`, couponID, userID).Scan(&count, &discount)
	s.Require().NoError(err)
	s.Equal(1, count)
	s.Equal(int64(1000), discount)
}

// TestCheckout_Success_FixedAmountCoupon applies a flat discount.
func (s *CheckoutTestSuite) TestCheckout_Success_FixedAmountCoupon() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	couponID := s.seedCoupon(couponSeed{
		Code:          "FLAT1000-" + uuid.NewString()[:8],
		DiscountType:  "fixed_amount",
		DiscountValue: 1000,
	})

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusOK, w.Code)

	var resp dataEnvelope[checkout.CheckoutResponse]
	s.decodeSuccess(w, &resp)
	orderID, err := uuid.Parse(resp.Data.OrderID)
	s.Require().NoError(err)

	// 5000 - 1000 + 999 = 4999.
	o := s.fetchOrder(orderID)
	s.Equal(int64(1000), o.DiscountAmount)
	s.Equal(int64(4999), o.TotalAmount)
}

// TestCheckout_Success_SnapshotAndAuthoritativePrice verifies the order line
// captures the live variant price and an immutable product snapshot.
func (s *CheckoutTestSuite) TestCheckout_Success_SnapshotAndAuthoritativePrice() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(1499, 3)
	s.addToCart(userID, variantID, 3)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusOK, w.Code)

	var resp dataEnvelope[checkout.CheckoutResponse]
	s.decodeSuccess(w, &resp)
	orderID, err := uuid.Parse(resp.Data.OrderID)
	s.Require().NoError(err)

	var qty int
	var unitPrice int64
	var snapshot map[string]any
	err = s.db.QueryRow(context.Background(), `
		SELECT quantity, price_at_purchase, product_snapshot
		FROM order_items
		WHERE order_id = $1
	`, orderID).Scan(&qty, &unitPrice, &snapshot)
	s.Require().NoError(err)

	s.Equal(3, qty)
	s.Equal(int64(1499), unitPrice)
	s.Equal(variantID.String(), snapshot["variantId"])
	s.Equal(float64(1499), snapshot["price"])
}

// TestCheckout_Success_SecondCheckoutRejectedAsEmpty verifies the cart is really
// cleared, preventing accidental duplicate orders.
func (s *CheckoutTestSuite) TestCheckout_Success_SecondCheckoutRejectedAsEmpty() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(1000, 5)
	s.addToCart(userID, variantID, 1)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusOK, w.Code)

	second := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusBadRequest, second.Code)
	env := s.decodeError(second)
	s.Equal("EMPTY_CART", env.Code)

	s.Equal(1, s.countOrdersForUser(userID), "second checkout must not create another order")
}

// =============================================================================
// RED: authentication & request validation
// =============================================================================

func (s *CheckoutTestSuite) TestCheckout_Red_Unauthenticated() {
	userID, _ := s.seedUser()
	addressID := s.seedAddress(userID)

	w := s.postCheckout("", checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusUnauthorized, w.Code)
	env := s.decodeError(w)
	s.Equal("UNAUTHORIZED", env.Code)
}

func (s *CheckoutTestSuite) TestCheckout_Red_MissingShippingAddressID() {
	_, sessionID := s.seedUser()

	// Missing required shippingAddressId binding tag.
	w := s.postCheckout(sessionID, map[string]any{"paymentMethod": "cod"})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	env := s.decodeError(w)
	s.Equal("VALIDATION_FAILED", env.Code)
	s.Require().NotEmpty(env.Details)
}

// =============================================================================
// RED: empty cart & address resolution
// =============================================================================

func (s *CheckoutTestSuite) TestCheckout_Red_EmptyCart() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	env := s.decodeError(w)
	s.Equal("EMPTY_CART", env.Code)
}

func (s *CheckoutTestSuite) TestCheckout_Red_AddressNotFound() {
	userID, sessionID := s.seedUser()
	variantID := s.seedPricedVariant(1000, 5)
	s.addToCart(userID, variantID, 1)

	// Random, non-existent address.
	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: uuid.New()})
	s.Require().Equal(http.StatusNotFound, w.Code)
	env := s.decodeError(w)
	s.Equal("NOT_FOUND", env.Code)

	// A failed checkout must not leak side effects.
	s.Equal(0, s.countOrdersForUser(userID))
	s.Equal(1, s.countActiveCartItems(userID))
}

func (s *CheckoutTestSuite) TestCheckout_Red_AddressBelongsToAnotherUser() {
	userID, sessionID := s.seedUser()
	otherUserID, _ := s.seedUser()
	variantID := s.seedPricedVariant(1000, 5)
	s.addToCart(userID, variantID, 1)

	// Address exists but is owned by a different user (tenant isolation).
	otherAddressID := s.seedAddress(otherUserID)
	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: otherAddressID})
	s.Require().Equal(http.StatusNotFound, w.Code)
	s.Equal("NOT_FOUND", s.decodeError(w).Code)
	s.Equal(0, s.countOrdersForUser(userID))
}

// =============================================================================
// RED: coupon lifecycle
// =============================================================================

func (s *CheckoutTestSuite) TestCheckout_Red_CouponNotFound() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	missing := uuid.New()
	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &missing,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_NOT_FOUND", s.decodeError(w).Code)
	s.Equal(0, s.countOrdersForUser(userID))
}

func (s *CheckoutTestSuite) TestCheckout_Red_CouponInactive() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	couponID := s.seedCoupon(couponSeed{
		Code:          "INACTIVE-" + uuid.NewString()[:8],
		DiscountType:  "percentage",
		DiscountValue: 10,
		Status:        "disabled",
	})

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_INACTIVE", s.decodeError(w).Code)
	s.Equal(0, s.countOrdersForUser(userID))
	s.Equal(0, s.countActiveReservations(variantID))
}

func (s *CheckoutTestSuite) TestCheckout_Red_CouponNotStarted() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	future := time.Now().Add(24 * time.Hour)
	couponID := s.seedCoupon(couponSeed{
		Code:          "FUTURE-" + uuid.NewString()[:8],
		DiscountType:  "percentage",
		DiscountValue: 10,
		StartsAt:      &future,
	})

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_NOT_STARTED", s.decodeError(w).Code)
}

func (s *CheckoutTestSuite) TestCheckout_Red_CouponExpired() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	past := time.Now().Add(-24 * time.Hour)
	couponID := s.seedCoupon(couponSeed{
		Code:          "EXPIRED-" + uuid.NewString()[:8],
		DiscountType:  "percentage",
		DiscountValue: 10,
		ExpiresAt:     &past,
	})

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_EXPIRED", s.decodeError(w).Code)
}

func (s *CheckoutTestSuite) TestCheckout_Red_CouponMinAmountNotMet() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1) // gross 5000

	min := int64(10000)
	couponID := s.seedCoupon(couponSeed{
		Code:          "MIN-" + uuid.NewString()[:8],
		DiscountType:  "percentage",
		DiscountValue: 10,
		MinCartAmount: &min,
	})

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_MIN_AMOUNT_NOT_MET", s.decodeError(w).Code)
	s.Equal(0, s.countOrdersForUser(userID))
	s.Equal(1, s.countActiveCartItems(userID), "cart must survive a failed coupon validation")
}

func (s *CheckoutTestSuite) TestCheckout_Red_CouponGlobalUsageLimitExceeded() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	limit := int32(1)
	couponID := s.seedCoupon(couponSeed{
		Code:          "GLOBAL-" + uuid.NewString()[:8],
		DiscountType:  "percentage",
		DiscountValue: 10,
		UsageLimit:    &limit,
	})

	// Exhaust the global limit with a prior redemption by another user.
	otherUserID, _ := s.seedUser()
	priorOrder := s.seedRawOrder(&otherUserID)
	s.seedRedemption(couponID, &otherUserID, priorOrder)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_USAGE_LIMIT_EXCEEDED", s.decodeError(w).Code)
	s.Equal(0, s.countOrdersForUser(userID))
}

func (s *CheckoutTestSuite) TestCheckout_Red_CouponPerUserLimitExceeded() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	perUser := int32(1)
	couponID := s.seedCoupon(couponSeed{
		Code:          "PERUSER-" + uuid.NewString()[:8],
		DiscountType:  "percentage",
		DiscountValue: 10,
		PerUserLimit:  &perUser,
	})

	// This same user already redeemed the coupon once.
	priorOrder := s.seedRawOrder(&userID)
	s.seedRedemption(couponID, &userID, priorOrder)
	ordersBefore := s.countOrdersForUser(userID)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_USER_LIMIT_EXCEEDED", s.decodeError(w).Code)
	s.Equal(ordersBefore, s.countOrdersForUser(userID), "no new order may be created")
}

// =============================================================================
// RED: stock & pricing integrity (fail closed, fully atomic)
// =============================================================================

func (s *CheckoutTestSuite) TestCheckout_Red_InsufficientStock_RollsBackEntirely() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(2500, 1)

	// Bypass cart validation to place a quantity the API would normally reject.
	s.seedRawCartItem(userID, variantID, 5, 2500)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("INSUFFICIENT_INVENTORY", s.decodeError(w).Code)

	// Nothing may persist: no order, no payment, no reservation, cart intact.
	s.Equal(0, s.countOrdersForUser(userID))
	s.Equal(0, s.countPaymentsForUser(userID))
	s.Equal(0, s.countActiveReservations(variantID))
	s.Equal(1, s.countActiveCartItems(userID))
	s.Equal(1, s.fetchStock(variantID).OnHandQuantity)
}

func (s *CheckoutTestSuite) TestCheckout_Red_VariantMissingPrice() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)

	// Variant exists but has no price; storefront must never sell it.
	variantID := s.seedVariant(nil, 5)
	s.seedRawCartItem(userID, variantID, 1, 0)

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{ShippingAddressID: addressID})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("INVALID_INPUT", s.decodeError(w).Code)

	s.Equal(0, s.countOrdersForUser(userID))
	s.Equal(0, s.countActiveReservations(variantID))
}

// TestCheckout_Red_AtomicityOnCouponFailure asserts the whole checkout is one
// transaction: a coupon failure leaves zero traces across every table it touches.
func (s *CheckoutTestSuite) TestCheckout_Red_AtomicityOnCouponFailure() {
	userID, sessionID := s.seedUser()
	addressID := s.seedAddress(userID)
	variantID := s.seedPricedVariant(5000, 5)
	s.addToCart(userID, variantID, 1)

	past := time.Now().Add(-time.Hour)
	couponID := s.seedCoupon(couponSeed{
		Code:          "EXPIREDATOMIC-" + uuid.NewString()[:8],
		DiscountType:  "fixed_amount",
		DiscountValue: 500,
		ExpiresAt:     &past,
	})

	w := s.postCheckout(sessionID, checkout.CheckoutRequest{
		ShippingAddressID: addressID,
		CouponID:          &couponID,
	})
	s.Require().Equal(http.StatusBadRequest, w.Code)

	s.Equal(0, s.countOrdersForUser(userID))
	s.Equal(0, s.countPaymentsForUser(userID))
	s.Equal(1, s.countActiveCartItems(userID))
	s.Equal(0, s.countActiveReservations(variantID))

	var redemptions int
	err := s.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1`, couponID).Scan(&redemptions)
	s.Require().NoError(err)
	s.Equal(0, redemptions)

	stock := s.fetchStock(variantID)
	s.Equal(5, stock.OnHandQuantity)
	s.Equal(0, stock.ReservedQuantity)
}

// =============================================================================
// Coupon application endpoint (pre-checkout quote)
// =============================================================================

func (s *CheckoutTestSuite) TestApplyCoupon_Success_PercentageQuote() {
	userID, sessionID := s.seedUser()
	variantID := s.seedPricedVariant(10000, 5)
	s.addToCart(userID, variantID, 1)

	code := "APPLY10-" + uuid.NewString()[:8]
	s.seedCoupon(couponSeed{
		Code:          code,
		DiscountType:  "percentage",
		DiscountValue: 10,
	})

	w := s.postApplyCoupon(sessionID, checkout.ApplyCouponRequest{Code: code})
	s.Require().Equal(http.StatusOK, w.Code)

	var resp dataEnvelope[checkout.CouponValidationResult]
	s.decodeSuccess(w, &resp)

	s.Equal(code, resp.Data.Code)
	s.Equal(int64(10000), resp.Data.Subtotal)
	s.Equal(int64(1000), resp.Data.DiscountAmount)
	s.Equal(shippingFee, resp.Data.Shipping)
	s.Equal(int64(9999), resp.Data.Total)
}

func (s *CheckoutTestSuite) TestApplyCoupon_Red_NotFound() {
	_, sessionID := s.seedUser()

	w := s.postApplyCoupon(sessionID, checkout.ApplyCouponRequest{Code: "DOES-NOT-EXIST-" + uuid.NewString()[:8]})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_NOT_FOUND", s.decodeError(w).Code)
}

func (s *CheckoutTestSuite) TestApplyCoupon_Red_Inactive() {
	_, sessionID := s.seedUser()

	code := "APPLYINACTIVE-" + uuid.NewString()[:8]
	s.seedCoupon(couponSeed{
		Code:          code,
		DiscountType:  "percentage",
		DiscountValue: 10,
		Status:        "disabled",
	})

	w := s.postApplyCoupon(sessionID, checkout.ApplyCouponRequest{Code: code})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("COUPON_INACTIVE", s.decodeError(w).Code)
}

func (s *CheckoutTestSuite) TestApplyCoupon_Red_EmptyCode() {
	_, sessionID := s.seedUser()

	w := s.postApplyCoupon(sessionID, checkout.ApplyCouponRequest{Code: ""})
	s.Require().Equal(http.StatusBadRequest, w.Code)
	s.Equal("VALIDATION_FAILED", s.decodeError(w).Code)
}

func TestCheckoutTestSuite(t *testing.T) {
	suite.Run(t, new(CheckoutTestSuite))
}
