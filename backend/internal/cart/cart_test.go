package cart_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/auth"
	"github.com/nullopt-t/prim-backend/internal/cart"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/catalog/product/variant"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/object"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type dummyProductService struct {
	products map[uuid.UUID]*model.Product
}

func (d *dummyProductService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	if p, ok := d.products[id]; ok {
		return p, nil
	}
	return &model.Product{
		ID:     id,
		Title:  "Test Headphones",
		Status: model.PublicationStatusPublished,
	}, nil
}

// GetBundleComposition reports every dummy product as an ordinary listing: the
// cart only treats a product as a kit when a composition exists.
func (d *dummyProductService) GetBundleComposition(_ context.Context, _ uuid.UUID) (*model.BundleComposition, error) {
	return nil, database.ErrNotFound
}

// dummySellerReader keeps all cart fixtures platform-owned by reporting an
// active seller for any id (the dummy products carry no seller).
type dummySellerReader struct{}

func (dummySellerReader) GetByID(ctx context.Context, id uuid.UUID) (*model.Seller, error) {
	return &model.Seller{ID: id, Status: model.SellerStatusActive}, nil
}

type CartHTTPTestSuite struct {
	suite.Suite
	pgContainer      *postgres.PostgresContainer
	db               *database.DB
	txRunner         database.Runner
	cartRepo         *cart.CartRepository
	inventoryRepo    *inventory.InventoryRepository
	inventoryService *inventory.InventoryService
	variantRepo      *variant.VariantRepository
	variantService   *variant.VariantService
	productService   *dummyProductService
	cartService      *cart.CartService
	sessionService   *auth.SessionService
	handler          *cart.CartHandler
	router           *gin.Engine
	categoryID       uuid.UUID
	productID        uuid.UUID
	redisContainer   *redisContainer.RedisContainer
	redisClient      *redis.Client
}

func (s *CartHTTPTestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_cart_http_test"),
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
		DBName:     "prim_cart_http_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	schemaBytes, err := os.ReadFile("../../migrations/000001_init.up.sql")
	require.NoError(s.T(), err)
	_, err = db.ExecContext(ctx, string(schemaBytes))
	require.NoError(s.T(), err)

	logger := log.NewConsoleLogger()
	s.txRunner = database.NewTxRunner(s.db, logger)

	s.inventoryRepo = inventory.NewRepository()
	s.inventoryService = inventory.NewService(logger, s.txRunner, s.inventoryRepo, storage.NoopProvider{}, nil)

	s.variantRepo = variant.NewRepository()
	s.variantService = variant.NewService(logger, s.txRunner, object.NewRepository(), storage.NoopProvider{}, s.variantRepo, s.inventoryRepo, s.inventoryService)

	s.categoryID = uuid.New()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO product_categories (id, name)
		VALUES ($1, $2)
	`, s.categoryID, "Audio Electronics")
	require.NoError(s.T(), err)

	s.productID = uuid.New()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type, variation_axes)
		VALUES ($1, $2, $3, $4, 'published', 'variable', ARRAY['color']::text[])
	`, s.productID, s.categoryID, "anc-headphones-"+uuid.NewString(), "ANC Headphones")
	require.NoError(s.T(), err)

	s.productService = &dummyProductService{
		products: map[uuid.UUID]*model.Product{
			s.productID: {
				ID:     s.productID,
				Title:  "ANC Headphones",
				Status: model.PublicationStatusPublished,
			},
		},
	}

	s.cartRepo = cart.NewRepository()
	s.cartService = cart.NewService(s.txRunner, s.cartRepo, s.variantRepo, s.productService, dummySellerReader{}, s.inventoryRepo, object.NewRepository(), storage.NoopProvider{})
	s.handler = cart.NewHandler(s.cartService)

	// Start Redis Container
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

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.ErrorHandler(logger))

	s.sessionService = auth.NewSessionService(s.redisClient, logger, time.Hour)
	cartRouter := cart.NewRouter(s.handler, s.sessionService)
	v1 := s.router.Group("/api/v1")
	cartRouter.MapRoutes(v1)
}

func (s *CartHTTPTestSuite) TearDownSuite() {
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

func (s *CartHTTPTestSuite) createGuestSession() string {
	session, err := s.sessionService.CreateSession(
		context.Background(),
		uuid.Nil,
		"",
		model.SessionTypeGuest,
		"cart-test",
		"127.0.0.1",
	)
	s.Require().NoError(err)
	return session.ID
}

func (s *CartHTTPTestSuite) createVariant(sku string, price int64, stock int) uuid.UUID {
	priceVal := price
	skuVal := sku
	v, err := s.variantService.CreateVariant(context.Background(), &variant.CreateVariantInput{
		ProductID:  s.productID,
		SKU:        &skuVal,
		Title:      "Black / Matte",
		Price:      &priceVal,
		Attributes: map[string]any{"color": sku},
	})
	s.Require().NoError(err)

	if stock > 0 {
		_, err := s.inventoryService.AdjustStock(context.Background(), inventory.AdjustStockInput{
			VariantID: v.ID,
			Quantity:  stock,
			Reason:    "restock",
		})
		s.Require().NoError(err)
	}

	return v.ID
}

// Strict Envelope Structure for Success Responses
type StrictDataEnvelope[T any] struct {
	Data T `json:"data"`
}

// Strict Envelope Structure for Error Responses
type StrictErrorEnvelope struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []api.FieldError `json:"details,omitempty"`
}

func (s *CartHTTPTestSuite) TestHTTP_GetCart_InitialEmptySchema() {
	sessionID := s.createGuestSession()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// Status Code Assertion
	s.Require().Equal(http.StatusOK, w.Code)
	s.Equal("application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// Strict JSON Structure Assertion
	var resp StrictDataEnvelope[cart.CartResponse]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err, "Response body must strictly conform to DataResponse{data=CartResponse}")

	// Assert Top-Level CartResponse Fields
	_, parseErr := uuid.Parse(resp.Data.ID)
	s.Require().NoError(parseErr, "Cart ID must be a valid UUID string")
	s.Equal("EGP", resp.Data.Currency)
	s.Equal(0, resp.Data.ItemCount)
	s.NotZero(resp.Data.UpdatedAt)
	s.NotNil(resp.Data.Items)
	s.Empty(resp.Data.Items)

	// Assert CartSummaryResponse Fields
	s.Equal(int64(0), resp.Data.Summary.Subtotal)
	s.Equal(int64(0), resp.Data.Summary.Discount)
	s.Equal(int64(0), resp.Data.Summary.Shipping)
	s.Equal(int64(0), resp.Data.Summary.Tax)
	s.Equal(int64(0), resp.Data.Summary.Total)
}

func (s *CartHTTPTestSuite) TestHTTP_AddItem_SuccessAndStrictItemSchema() {
	sessionID := s.createGuestSession()
	variantID := s.createVariant("SKU-MATTE-BLK", 2999, 10)

	bodyBytes, _ := json.Marshal(cart.AddItemRequest{
		VariantID: variantID,
		Quantity:  2,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)

	var resp StrictDataEnvelope[cart.CartResponse]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal(2, resp.Data.ItemCount)
	s.Require().Len(resp.Data.Items, 1)

	// Strict CartItemResponse Validation
	item := resp.Data.Items[0]
	itemID, itemIDErr := strconv.ParseInt(item.ID, 10, 64)
	s.Require().NoError(itemIDErr, "CartItem ID must be a numeric database identity")
	s.Greater(itemID, int64(0))
	s.Equal(s.productID.String(), item.ProductID)
	s.Equal(variantID.String(), item.VariantID)
	s.Equal("ANC Headphones - Black / Matte", item.Title)
	s.Equal(2, item.Quantity)
	s.Equal(int64(2999), item.UnitPrice)
	s.Equal(int64(5998), item.Subtotal)
	s.True(item.InStock, "Item must be in stock since available inventory is 10 >= 2")

	// Strict CartSummaryResponse Validation
	s.Equal(int64(5998), resp.Data.Summary.Subtotal)
	s.Equal(int64(999), resp.Data.Summary.Shipping)
	s.Equal(int64(6997), resp.Data.Summary.Total)
}

func (s *CartHTTPTestSuite) TestHTTP_AddItem_InvalidPayloads() {
	sessionID := s.createGuestSession()

	// Case 1: Missing variantId
	badBody1 := []byte(`{"quantity": 2}`)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(badBody1))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusBadRequest, w1.Code)

	var errResp1 StrictErrorEnvelope
	s.Require().NoError(json.Unmarshal(w1.Body.Bytes(), &errResp1))
	s.NotEmpty(errResp1.Code)
	s.NotEmpty(errResp1.Message)

	// Case 2: Quantity <= 0
	badBody2 := []byte(`{"variantId": "00000000-0000-0000-0000-000000000001", "quantity": 0}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(badBody2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusBadRequest, w2.Code)

	// Case 3: Non-existent variant UUID
	badBody3, _ := json.Marshal(cart.AddItemRequest{
		VariantID: uuid.New(),
		Quantity:  1,
	})
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(badBody3))
	req3.Header.Set("Content-Type", "application/json")
	req3.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w3 := httptest.NewRecorder()
	s.router.ServeHTTP(w3, req3)
	s.Require().Equal(http.StatusNotFound, w3.Code)

	var errResp3 StrictErrorEnvelope
	s.Require().NoError(json.Unmarshal(w3.Body.Bytes(), &errResp3))
	s.Equal("VARIANT_NOT_FOUND", errResp3.Code)
}

func (s *CartHTTPTestSuite) TestHTTP_UpdateItemQuantity_SuccessAndInsufficientStockValidation() {
	sessionID := s.createGuestSession()
	variantID := s.createVariant("SKU-SILVER", 1500, 5)

	// 1. Add 2 items (valid: 2 <= 5)
	addBody, _ := json.Marshal(cart.AddItemRequest{
		VariantID: variantID,
		Quantity:  2,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var addResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &addResp))
	itemID := addResp.Data.Items[0].ID

	// 2. Reject adding more than available stock (exceeds remaining stock)
	exceedAddBody, _ := json.Marshal(cart.AddItemRequest{
		VariantID: variantID,
		Quantity:  10,
	})
	reqExceed := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(exceedAddBody))
	reqExceed.Header.Set("Content-Type", "application/json")
	reqExceed.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wExceed := httptest.NewRecorder()
	s.router.ServeHTTP(wExceed, reqExceed)
	s.Require().Equal(http.StatusBadRequest, wExceed.Code)

	var exceedErrResp StrictErrorEnvelope
	s.Require().NoError(json.Unmarshal(wExceed.Body.Bytes(), &exceedErrResp))
	s.Equal("INSUFFICIENT_INVENTORY", exceedErrResp.Code)

	// 3. Reject update when quantity > available stock (8 > 5)
	updateExceedBody, _ := json.Marshal(cart.UpdateQuantityRequest{
		Quantity: 8,
	})
	reqUpdateExceed := httptest.NewRequest(http.MethodPatch, "/api/v1/cart/items/"+itemID, bytes.NewReader(updateExceedBody))
	reqUpdateExceed.Header.Set("Content-Type", "application/json")
	reqUpdateExceed.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wUpdateExceed := httptest.NewRecorder()
	s.router.ServeHTTP(wUpdateExceed, reqUpdateExceed)

	s.Require().Equal(http.StatusBadRequest, wUpdateExceed.Code)
	var updateErrResp StrictErrorEnvelope
	s.Require().NoError(json.Unmarshal(wUpdateExceed.Body.Bytes(), &updateErrResp))
	s.Equal("INSUFFICIENT_INVENTORY", updateErrResp.Code)

	// 4. Successful update within stock limit (update to 4 <= 5)
	validUpdateBody, _ := json.Marshal(cart.UpdateQuantityRequest{
		Quantity: 4,
	})
	reqValidUpdate := httptest.NewRequest(http.MethodPatch, "/api/v1/cart/items/"+itemID, bytes.NewReader(validUpdateBody))
	reqValidUpdate.Header.Set("Content-Type", "application/json")
	reqValidUpdate.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wValidUpdate := httptest.NewRecorder()
	s.router.ServeHTTP(wValidUpdate, reqValidUpdate)

	s.Require().Equal(http.StatusOK, wValidUpdate.Code)

	var updateResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wValidUpdate.Body.Bytes(), &updateResp))

	s.Equal(4, updateResp.Data.ItemCount)
	s.Require().Len(updateResp.Data.Items, 1)
	s.Equal(4, updateResp.Data.Items[0].Quantity)
	s.Equal(int64(1500), updateResp.Data.Items[0].UnitPrice)
	s.Equal(int64(6000), updateResp.Data.Items[0].Subtotal)
	s.True(updateResp.Data.Items[0].InStock)
}

func (s *CartHTTPTestSuite) TestHTTP_RemoveItemAndClearCart() {
	sessionID := s.createGuestSession()
	v1 := s.createVariant("SKU-1", 1000, 10)
	v2 := s.createVariant("SKU-2", 2000, 10)

	// Add item 1
	b1, _ := json.Marshal(cart.AddItemRequest{VariantID: v1, Quantity: 1})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(b1))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusOK, w1.Code)

	var resp1 StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(w1.Body.Bytes(), &resp1))
	item1ID := resp1.Data.Items[0].ID

	// Add item 2
	b2, _ := json.Marshal(cart.AddItemRequest{VariantID: v2, Quantity: 2})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusOK, w2.Code)

	// 1. DELETE /api/v1/cart/items/:id (Remove item 1)
	reqRemove := httptest.NewRequest(http.MethodDelete, "/api/v1/cart/items/"+item1ID, nil)
	reqRemove.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wRemove := httptest.NewRecorder()
	s.router.ServeHTTP(wRemove, reqRemove)

	s.Require().Equal(http.StatusOK, wRemove.Code)

	var removeResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wRemove.Body.Bytes(), &removeResp))
	s.Require().Len(removeResp.Data.Items, 1)
	s.Equal(v2.String(), removeResp.Data.Items[0].VariantID)
	s.Equal(int64(4000), removeResp.Data.Summary.Subtotal)
	s.Equal(int64(999), removeResp.Data.Summary.Shipping)
	s.Equal(int64(4999), removeResp.Data.Summary.Total)

	// 2. DELETE /api/v1/cart (Clear remaining items)
	reqClear := httptest.NewRequest(http.MethodDelete, "/api/v1/cart", nil)
	reqClear.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wClear := httptest.NewRecorder()
	s.router.ServeHTTP(wClear, reqClear)

	s.Require().Equal(http.StatusNoContent, wClear.Code)
	s.Empty(wClear.Body.Bytes(), "204 No Content response must have an empty body")

	// 3. Verify cart is empty via GET
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	reqGet.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wGet := httptest.NewRecorder()
	s.router.ServeHTTP(wGet, reqGet)

	s.Require().Equal(http.StatusOK, wGet.Code)
	var finalResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wGet.Body.Bytes(), &finalResp))
	s.Empty(finalResp.Data.Items)
	s.Equal(0, finalResp.Data.ItemCount)
	s.Equal(int64(0), finalResp.Data.Summary.Total)
}

func (s *CartHTTPTestSuite) createAuthenticatedUser() (uuid.UUID, string) {
	userID := uuid.New()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO users (id, email, full_name, role, is_email_verified, status)
		VALUES ($1, $2, $3, 'customer', true, 'active')
	`, userID, "user_"+userID.String()+"@example.com", "Test User")
	s.Require().NoError(err)

	session, err := s.sessionService.CreateSession(
		context.Background(),
		userID,
		"customer",
		model.SessionTypeAuthenticated,
		"cart-test",
		"127.0.0.1",
	)
	s.Require().NoError(err)

	return userID, session.ID
}

func (s *CartHTTPTestSuite) TestHTTP_GuestCart_CookieGenerationAndPersistence() {
	// A read-only request without a session must not mint a session or a cart.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusOK, w.Code)
	var resp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	s.Empty(resp.Data.Items)
	s.Empty(w.Result().Cookies(), "GET without a session must not create one")

	// A mutating request without a session mints a guest session and sets the
	// canonical "ps" cookie.
	variantID := s.createVariant("SKU-COOKIE-GUEST", 1200, 10)
	addBody, _ := json.Marshal(cart.AddItemRequest{
		VariantID: variantID,
		Quantity:  3,
	})
	reqAdd := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(addBody))
	reqAdd.Header.Set("Content-Type", "application/json")
	wAdd := httptest.NewRecorder()
	s.router.ServeHTTP(wAdd, reqAdd)

	s.Require().Equal(http.StatusOK, wAdd.Code)
	var addResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wAdd.Body.Bytes(), &addResp))
	s.Equal(3, addResp.Data.ItemCount)
	s.NotEmpty(addResp.Data.ID, "A cart ID should be created once items are added")

	var sessionCookie *http.Cookie
	for _, cookie := range wAdd.Result().Cookies() {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	s.Require().NotNil(sessionCookie, "ps cookie must be set when a guest session is created")
	s.NotEmpty(sessionCookie.Value)
	s.True(sessionCookie.HttpOnly)

	// Verify GET with the guest cookie returns the same cart and items.
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	reqGet.AddCookie(sessionCookie)
	wGet := httptest.NewRecorder()
	s.router.ServeHTTP(wGet, reqGet)

	s.Require().Equal(http.StatusOK, wGet.Code)
	var getResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wGet.Body.Bytes(), &getResp))
	s.Equal(3, getResp.Data.ItemCount)
	s.Require().Len(getResp.Data.Items, 1)
	s.Equal(variantID.String(), getResp.Data.Items[0].VariantID)
}

func (s *CartHTTPTestSuite) TestHTTP_GuestCart_MultipleGuestsAreIsolated() {
	session1 := s.createGuestSession()
	session2 := s.createGuestSession()

	v1 := s.createVariant("SKU-ISO-1", 1000, 10)
	v2 := s.createVariant("SKU-ISO-2", 2000, 10)

	// Guest 1 adds v1
	b1, _ := json.Marshal(cart.AddItemRequest{VariantID: v1, Quantity: 1})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(b1))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: session1})
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusOK, w1.Code)

	// Guest 2 adds v2
	b2, _ := json.Marshal(cart.AddItemRequest{VariantID: v2, Quantity: 2})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: session2})
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusOK, w2.Code)

	// Fetch guest 1 cart
	reqGet1 := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	reqGet1.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: session1})
	wGet1 := httptest.NewRecorder()
	s.router.ServeHTTP(wGet1, reqGet1)
	var resp1 StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wGet1.Body.Bytes(), &resp1))
	s.Equal(1, resp1.Data.ItemCount)
	s.Require().Len(resp1.Data.Items, 1)
	s.Equal(v1.String(), resp1.Data.Items[0].VariantID)

	// Fetch guest 2 cart
	reqGet2 := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	reqGet2.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: session2})
	wGet2 := httptest.NewRecorder()
	s.router.ServeHTTP(wGet2, reqGet2)
	var resp2 StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wGet2.Body.Bytes(), &resp2))
	s.Equal(2, resp2.Data.ItemCount)
	s.Require().Len(resp2.Data.Items, 1)
	s.Equal(v2.String(), resp2.Data.Items[0].VariantID)

	s.NotEqual(resp1.Data.ID, resp2.Data.ID, "Different guest sessions must have distinct carts")
}

func (s *CartHTTPTestSuite) TestHTTP_GuestCart_MergeIntoUserCart() {
	guestSessionID := s.createGuestSession()
	userID, token := s.createAuthenticatedUser()

	v1 := s.createVariant("SKU-MERGE-1", 1000, 20)
	v2 := s.createVariant("SKU-MERGE-2", 2000, 20)
	v3 := s.createVariant("SKU-MERGE-3", 3000, 20)

	// 1. Guest adds v1 (qty 2) and v2 (qty 1)
	b1, _ := json.Marshal(cart.AddItemRequest{VariantID: v1, Quantity: 2})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(b1))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: guestSessionID})
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusOK, w1.Code)

	b2, _ := json.Marshal(cart.AddItemRequest{VariantID: v2, Quantity: 1})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(b2))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: guestSessionID})
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusOK, w2.Code)

	// 2. User has an existing cart with v2 (qty 3) and v3 (qty 1)
	bUser1, _ := json.Marshal(cart.AddItemRequest{VariantID: v2, Quantity: 3})
	reqU1 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(bUser1))
	reqU1.Header.Set("Content-Type", "application/json")
	reqU1.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: token})
	wU1 := httptest.NewRecorder()
	s.router.ServeHTTP(wU1, reqU1)
	s.Require().Equal(http.StatusOK, wU1.Code)

	bUser2, _ := json.Marshal(cart.AddItemRequest{VariantID: v3, Quantity: 1})
	reqU2 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(bUser2))
	reqU2.Header.Set("Content-Type", "application/json")
	reqU2.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: token})
	wU2 := httptest.NewRecorder()
	s.router.ServeHTTP(wU2, reqU2)
	s.Require().Equal(http.StatusOK, wU2.Code)

	// 3. Merge guest cart into user cart via CartService
	err := s.cartService.MergeGuestCart(context.Background(), guestSessionID, userID)
	s.Require().NoError(err)

	// 4. Verify merged user cart contents:
	// - v1: qty 2 (from guest)
	// - v2: qty 1 (guest) + 3 (user) = 4
	// - v3: qty 1 (user)
	// Total items = 7 units, 3 distinct lines
	reqUserCart := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	reqUserCart.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: token})
	wUserCart := httptest.NewRecorder()
	s.router.ServeHTTP(wUserCart, reqUserCart)

	s.Require().Equal(http.StatusOK, wUserCart.Code)
	var userCartResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wUserCart.Body.Bytes(), &userCartResp))

	s.Equal(7, userCartResp.Data.ItemCount)
	s.Require().Len(userCartResp.Data.Items, 3)

	itemMap := make(map[string]cart.CartItemResponse)
	for _, it := range userCartResp.Data.Items {
		itemMap[it.VariantID] = it
	}

	s.Equal(2, itemMap[v1.String()].Quantity)
	s.Equal(int64(2000), itemMap[v1.String()].Subtotal)

	s.Equal(4, itemMap[v2.String()].Quantity)
	s.Equal(int64(8000), itemMap[v2.String()].Subtotal)

	s.Equal(1, itemMap[v3.String()].Quantity)
	s.Equal(int64(3000), itemMap[v3.String()].Subtotal)

	// Expected total = 2000 + 8000 + 3000 = 13000, plus 999 shipping (< 15000 threshold)
	s.Equal(int64(13000), userCartResp.Data.Summary.Subtotal)
	s.Equal(int64(999), userCartResp.Data.Summary.Shipping)
	s.Equal(int64(13999), userCartResp.Data.Summary.Total)

	// 5. Verify the guest cart record was deleted after merge
	reqGuestAfter := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	reqGuestAfter.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: guestSessionID})
	wGuestAfter := httptest.NewRecorder()
	s.router.ServeHTTP(wGuestAfter, reqGuestAfter)

	s.Require().Equal(http.StatusOK, wGuestAfter.Code)
	var guestAfterResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wGuestAfter.Body.Bytes(), &guestAfterResp))
	s.Empty(guestAfterResp.Data.Items)
	s.Equal(0, guestAfterResp.Data.ItemCount)
}

func (s *CartHTTPTestSuite) TestHTTP_GuestCart_MergeEmptyGuestCartDoesNotError() {
	guestSessionID := s.createGuestSession()
	userID, _ := s.createAuthenticatedUser()

	// Calling merge on a non-existent / empty guest cart session
	err := s.cartService.MergeGuestCart(context.Background(), guestSessionID, userID)
	s.Require().NoError(err)
}

// TestHTTP_UpdateItemQuantity_ZeroAndForeignItem pins that a quantity of zero
// (or negative) is rejected with INVALID_QUANTITY instead of deleting the line,
// and that an item from another cart is a 404, never a cross-cart mutation.
func (s *CartHTTPTestSuite) TestHTTP_UpdateItemQuantity_ZeroAndForeignItem() {
	sessionID := s.createGuestSession()
	variantID := s.createVariant("SKU-ZERO-QTY", 1000, 10)

	addBody, _ := json.Marshal(cart.AddItemRequest{VariantID: variantID, Quantity: 2})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(addBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusOK, w.Code)

	var addResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &addResp))
	itemID := addResp.Data.Items[0].ID

	for _, quantity := range []int{0, -1} {
		body, _ := json.Marshal(cart.UpdateQuantityRequest{Quantity: quantity})
		reqZero := httptest.NewRequest(http.MethodPatch, "/api/v1/cart/items/"+itemID, bytes.NewReader(body))
		reqZero.Header.Set("Content-Type", "application/json")
		reqZero.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
		wZero := httptest.NewRecorder()
		s.router.ServeHTTP(wZero, reqZero)

		s.Require().Equal(http.StatusBadRequest, wZero.Code)
		var errResp StrictErrorEnvelope
		s.Require().NoError(json.Unmarshal(wZero.Body.Bytes(), &errResp))
		s.Equal("INVALID_CART_QUANTITY", errResp.Code)
	}

	// The rejected updates must not have removed or changed the line.
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/cart", nil)
	reqGet.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wGet := httptest.NewRecorder()
	s.router.ServeHTTP(wGet, reqGet)
	var getResp StrictDataEnvelope[cart.CartResponse]
	s.Require().NoError(json.Unmarshal(wGet.Body.Bytes(), &getResp))
	s.Require().Len(getResp.Data.Items, 1)
	s.Equal(2, getResp.Data.Items[0].Quantity)

	// A second guest cannot mutate the first guest's line.
	otherSession := s.createGuestSession()
	body, _ := json.Marshal(cart.UpdateQuantityRequest{Quantity: 5})
	reqForeign := httptest.NewRequest(http.MethodPatch, "/api/v1/cart/items/"+itemID, bytes.NewReader(body))
	reqForeign.Header.Set("Content-Type", "application/json")
	reqForeign.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: otherSession})
	wForeign := httptest.NewRecorder()
	s.router.ServeHTTP(wForeign, reqForeign)

	s.Require().Equal(http.StatusNotFound, wForeign.Code)
	var foreignErr StrictErrorEnvelope
	s.Require().NoError(json.Unmarshal(wForeign.Body.Bytes(), &foreignErr))
	s.Equal("CART_ITEM_NOT_FOUND", foreignErr.Code)

	// And a missing item ID is also 404.
	bodyMissing, _ := json.Marshal(cart.UpdateQuantityRequest{Quantity: 1})
	reqMissing := httptest.NewRequest(http.MethodPatch, "/api/v1/cart/items/999999", bytes.NewReader(bodyMissing))
	reqMissing.Header.Set("Content-Type", "application/json")
	reqMissing.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: sessionID})
	wMissing := httptest.NewRecorder()
	s.router.ServeHTTP(wMissing, reqMissing)
	s.Require().Equal(http.StatusNotFound, wMissing.Code)
}

// TestMergeGuestCart_UserHasNoCart covers the branch that previously created an
// empty user cart, dropped every guest line, and left the guest cart alive.
func (s *CartHTTPTestSuite) TestMergeGuestCart_UserHasNoCart() {
	guestSessionID := s.createGuestSession()
	userID, _ := s.createAuthenticatedUser()
	variantID := s.createVariant("SKU-MERGE-NO-USER-CART", 1500, 10)

	guestCart, err := s.cartService.GetOrCreateCart(context.Background(), nil, &guestSessionID)
	s.Require().NoError(err)
	_, err = s.cartService.AddItem(context.Background(), guestCart.ID, variantID, 3)
	s.Require().NoError(err)

	s.Require().NoError(s.cartService.MergeGuestCart(context.Background(), guestSessionID, userID))

	userCart, err := s.cartService.GetCartByUserID(context.Background(), userID)
	s.Require().NoError(err)
	s.Require().NotEqual(guestCart.ID, userCart.ID)
	s.Require().Len(userCart.Items, 1)
	s.Equal(variantID, userCart.Items[0].VariantID)
	s.Equal(3, userCart.Items[0].Quantity)

	// The guest cart must be consumed, not left behind.
	_, err = s.cartService.GetCartBySessionID(context.Background(), guestSessionID)
	s.Require().Error(err)
}

// TestMergeGuestCart_RejectsCombinedQuantityAboveStock proves the merge does not
// create an over-stock cart: guest 2 + user 4 against stock 5 is rejected and
// both carts stay untouched.
func (s *CartHTTPTestSuite) TestMergeGuestCart_RejectsCombinedQuantityAboveStock() {
	guestSessionID := s.createGuestSession()
	userID, _ := s.createAuthenticatedUser()
	variantID := s.createVariant("SKU-MERGE-OVERSTOCK", 1000, 5)

	guestCart, err := s.cartService.GetOrCreateCart(context.Background(), nil, &guestSessionID)
	s.Require().NoError(err)
	_, err = s.cartService.AddItem(context.Background(), guestCart.ID, variantID, 2)
	s.Require().NoError(err)

	userCart, err := s.cartService.GetOrCreateCart(context.Background(), &userID, nil)
	s.Require().NoError(err)
	_, err = s.cartService.AddItem(context.Background(), userCart.ID, variantID, 4)
	s.Require().NoError(err)

	err = s.cartService.MergeGuestCart(context.Background(), guestSessionID, userID)
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusBadRequest, apiErr.Status)
	s.Equal("INSUFFICIENT_INVENTORY", apiErr.Code)

	reloadedUser, err := s.cartService.GetCartByID(context.Background(), userCart.ID)
	s.Require().NoError(err)
	s.Require().Len(reloadedUser.Items, 1)
	s.Equal(4, reloadedUser.Items[0].Quantity)

	reloadedGuest, err := s.cartService.GetCartBySessionID(context.Background(), guestSessionID)
	s.Require().NoError(err)
	s.Require().Len(reloadedGuest.Items, 1)
	s.Equal(2, reloadedGuest.Items[0].Quantity)
}

// TestAddItem_ConcurrentSameVariantDoesNotOversell pins the cart-row lock: two
// simultaneous adds of 3 against stock 5 must result in one success and a final
// quantity of 3, never 6.
func (s *CartHTTPTestSuite) TestAddItem_ConcurrentSameVariantDoesNotOversell() {
	guestSessionID := s.createGuestSession()
	variantID := s.createVariant("SKU-CONCURRENT-ADD", 1000, 5)

	cart, err := s.cartService.GetOrCreateCart(context.Background(), nil, &guestSessionID)
	s.Require().NoError(err)

	start := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			_, addErr := s.cartService.AddItem(context.Background(), cart.ID, variantID, 3)
			results <- addErr
		}()
	}
	close(start)

	successes := 0
	insufficient := 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			successes++
			continue
		}
		var apiErr *apierr.APIError
		s.Require().ErrorAs(err, &apiErr)
		s.Equal("INSUFFICIENT_INVENTORY", apiErr.Code)
		insufficient++
	}
	// One add may also serialize after the other if it lands after the stock
	// change; what must never happen is both succeeding.
	s.GreaterOrEqual(successes, 1)
	s.LessOrEqual(successes, 1)
	s.Equal(1, insufficient)

	finalCart, err := s.cartService.GetCartByID(context.Background(), cart.ID)
	s.Require().NoError(err)
	s.Require().Len(finalCart.Items, 1)
	s.Equal(3, finalCart.Items[0].Quantity)
}

type suspendedSellerReader struct{}

func (suspendedSellerReader) GetByID(ctx context.Context, id uuid.UUID) (*model.Seller, error) {
	return &model.Seller{ID: id, Status: model.SellerStatusSuspended}, nil
}

// TestAddItem_RejectsUnavailableCatalog pins that unpublished products and
// products owned by a suspended seller cannot enter the cart.
func (s *CartHTTPTestSuite) TestAddItem_RejectsUnavailableCatalog() {
	guestSessionID := s.createGuestSession()
	variantID := s.createVariant("SKU-UNAVAILABLE", 1000, 10)

	cartSvc := cart.NewService(
		s.txRunner, s.cartRepo, s.variantRepo, s.productService, suspendedSellerReader{},
		s.inventoryRepo, object.NewRepository(), storage.NoopProvider{},
	)
	cartData, err := cartSvc.GetOrCreateCart(context.Background(), nil, &guestSessionID)
	s.Require().NoError(err)

	originalProduct := s.productService.products[s.productID]
	defer func() { s.productService.products[s.productID] = originalProduct }()

	sellerID := uuid.New()
	s.productService.products[s.productID] = &model.Product{
		ID: s.productID, Title: "Suspended Seller Product",
		Status: model.PublicationStatusPublished, SellerID: &sellerID,
	}

	_, err = cartSvc.AddItem(context.Background(), cartData.ID, variantID, 1)
	s.Require().Error(err)
	var apiErr *apierr.APIError
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusBadRequest, apiErr.Status)

	// Unpublished product, platform-owned.
	s.productService.products[s.productID] = &model.Product{
		ID: s.productID, Title: "Draft Product", Status: model.PublicationStatusDraft,
	}
	_, err = cartSvc.AddItem(context.Background(), cartData.ID, variantID, 1)
	s.Require().Error(err)
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(http.StatusBadRequest, apiErr.Status)
}

func TestCartHTTPTestSuite(t *testing.T) {
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker daemon unavailable; skipping PostgreSQL integration suite")
	}
	suite.Run(t, new(CartHTTPTestSuite))
}
