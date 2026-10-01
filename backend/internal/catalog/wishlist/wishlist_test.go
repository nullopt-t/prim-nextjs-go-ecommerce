package wishlist_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/wishlist"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type mockProductService struct {
	products map[uuid.UUID]*model.Product
}

func (m *mockProductService) GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return nil, nil
}

type mockObjectService struct{}

func (m *mockObjectService) GetObjectURL(ctx context.Context, bucket, key string) string {
	return fmt.Sprintf("http://localhost:9000/%s/%s", bucket, key)
}

type WishlistHTTPTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *database.DB
	txRunner    database.Runner
	repo        *wishlist.WishlistRepository
	service     *wishlist.WishlistService
	handler     *wishlist.WishlistHandler
	router      *gin.Engine
	jwtManager  *jwt.JWTManager
	secrets     config.Secrets

	categoryID uuid.UUID
	product1ID uuid.UUID
	product2ID uuid.UUID
	product3ID uuid.UUID
	prodMock   *mockProductService
}

func (s *WishlistHTTPTestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_wishlist_http_test"),
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
		DBName:     "prim_wishlist_http_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	schemaBytes, err := os.ReadFile("../../../migrations/000001_init_sechema.up.sql")
	require.NoError(s.T(), err)
	_, err = db.Exec(ctx, string(schemaBytes))
	require.NoError(s.T(), err)

	s.txRunner = database.NewTxRunner(s.db)
	s.repo = wishlist.NewWishlistRepository()

	s.prodMock = &mockProductService{
		products: make(map[uuid.UUID]*model.Product),
	}

	logger := log.NewConsoleLogger()
	s.service = wishlist.NewService(s.txRunner, s.repo, s.prodMock, &mockObjectService{}, logger)
	s.handler = wishlist.NewHandler(s.service)

	s.secrets = config.Secrets{
		JwtAccessTokenSecretKey:  "wishlist_test_access_secret_key_32b",
		JwtRefreshTokenSecretKey: "wishlist_test_refresh_secret_key_32",
	}
	s.jwtManager = jwt.NewJWTManager(s.secrets)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.ErrorHandler(logger))

	wishlistRouter := wishlist.NewRouter(s.handler, s.secrets)
	apiGroup := s.router.Group("/api/v1")
	wishlistRouter.MapRoutes(apiGroup)

	// Seed catalog data
	s.categoryID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_categories (id, name)
		VALUES ($1, $2)
	`, s.categoryID, "Electronics")
	require.NoError(s.T(), err)

	brandID := uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_brands (id, name)
		VALUES ($1, $2)
	`, brandID, "Apple")
	require.NoError(s.T(), err)

	// Product 1
	s.product1ID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO products (id, brand_id, category_id, slug, title, description, status, product_type)
		VALUES ($1, $2, $3, $4, $5, $6, 'published', 'simple')
	`, s.product1ID, brandID, s.categoryID, "macbook-pro-16", "MacBook Pro 16", "M3 Max Powerhouse")
	require.NoError(s.T(), err)

	v1ID := uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_variants (id, product_id, sku, title, price, crossed_out_price, currency, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, 'USD', true)
	`, v1ID, s.product1ID, "SKU-MBP-16", "Space Black", int64(249900), int64(299900))
	require.NoError(s.T(), err)

	// Add inventory ledger for product 1
	_, err = s.db.Exec(ctx, `
		INSERT INTO inventory_ledgers (id, variant_id, quantity, reason)
		VALUES ($1, $2, 10, 'restock')
	`, uuid.New(), v1ID)
	require.NoError(s.T(), err)

	s.prodMock.products[s.product1ID] = &model.Product{
		ID:     s.product1ID,
		Status: model.PublicationStatusPublished,
		Title:  "MacBook Pro 16",
	}

	// Product 2
	s.product2ID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO products (id, brand_id, category_id, slug, title, description, status, product_type)
		VALUES ($1, $2, $3, $4, $5, $6, 'published', 'simple')
	`, s.product2ID, brandID, s.categoryID, "iphone-16-pro", "iPhone 16 Pro", "Titanium Design")
	require.NoError(s.T(), err)

	v2ID := uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_variants (id, product_id, sku, title, price, currency, is_default)
		VALUES ($1, $2, $3, $4, $5, 'USD', true)
	`, v2ID, s.product2ID, "SKU-IPH-16", "Desert Titanium", int64(99900))
	require.NoError(s.T(), err)

	s.prodMock.products[s.product2ID] = &model.Product{
		ID:     s.product2ID,
		Status: model.PublicationStatusPublished,
		Title:  "iPhone 16 Pro",
	}

	// Product 3 (Draft status)
	s.product3ID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type)
		VALUES ($1, $2, $3, $4, 'draft', 'simple')
	`, s.product3ID, s.categoryID, "unreleased-device", "Unreleased Device")
	require.NoError(s.T(), err)

	s.prodMock.products[s.product3ID] = &model.Product{
		ID:     s.product3ID,
		Status: model.PublicationStatusDraft,
		Title:  "Unreleased Device",
	}
}

func (s *WishlistHTTPTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *WishlistHTTPTestSuite) createTestUser() (uuid.UUID, string) {
	userID := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO users (id, email, role, status)
		VALUES ($1, $2, 'customer', 'active')
	`, userID, fmt.Sprintf("user-%s@example.com", uuid.NewString()[:8]))
	s.Require().NoError(err)

	roleVal := "customer"
	token, err := s.jwtManager.GenerateAccessToken(&jwt.UserClaims{
		UserID:   userID,
		UserRole: &roleVal,
	})
	s.Require().NoError(err)
	return userID, token
}

func (s *WishlistHTTPTestSuite) TestAddToWishlist_Success() {
	_, token := s.createTestUser()

	body, _ := json.Marshal(map[string]any{
		"productId": s.product1ID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusCreated, w.Code)

	var resp struct {
		Data wishlist.WishlistItemResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	s.Equal(s.product1ID.String(), resp.Data.ProductID)
	s.Equal("MacBook Pro 16", resp.Data.Product.Title)
	s.Equal("macbook-pro-16", resp.Data.Product.Slug)
	s.Equal("Apple", *resp.Data.Product.BrandName)
	s.Equal("Electronics", *resp.Data.Product.CategoryName)
	s.True(resp.Data.Product.InStock)
	s.Require().NotNil(resp.Data.Product.Price)
	s.Equal("$2499.00", *resp.Data.Product.Price)
	s.Require().NotNil(resp.Data.Product.OriginalPrice)
	s.Equal("$2999.00", *resp.Data.Product.OriginalPrice)
}

func (s *WishlistHTTPTestSuite) TestAddToWishlist_DuplicateConflict() {
	_, token := s.createTestUser()

	body, _ := json.Marshal(map[string]any{
		"productId": s.product1ID.String(),
	})

	// First addition
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusCreated, w1.Code)

	// Second addition
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusConflict, w2.Code)
}

func (s *WishlistHTTPTestSuite) TestAddToWishlist_ProductNotFound() {
	_, token := s.createTestUser()

	body, _ := json.Marshal(map[string]any{
		"productId": uuid.NewString(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusNotFound, w.Code)
}

func (s *WishlistHTTPTestSuite) TestAddToWishlist_UnpublishedProduct() {
	_, token := s.createTestUser()

	body, _ := json.Marshal(map[string]any{
		"productId": s.product3ID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusBadRequest, w.Code)
}

func (s *WishlistHTTPTestSuite) TestAddToWishlist_Unauthorized() {
	body, _ := json.Marshal(map[string]any{
		"productId": s.product1ID.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)
	s.Require().Equal(http.StatusUnauthorized, w.Code)
}

func (s *WishlistHTTPTestSuite) TestGetMyWishlist_And_Count() {
	_, token := s.createTestUser()

	// Add product 1
	body1, _ := json.Marshal(map[string]any{"productId": s.product1ID.String()})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body1))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusCreated, w1.Code)

	// Add product 2
	body2, _ := json.Marshal(map[string]any{"productId": s.product2ID.String()})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body2))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusCreated, w2.Code)

	// Test GET /wishlist/count
	countReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist/count", nil)
	countReq.Header.Set("Authorization", "Bearer "+token)
	countW := httptest.NewRecorder()
	s.router.ServeHTTP(countW, countReq)
	s.Require().Equal(http.StatusOK, countW.Code)

	var countResp struct {
		Data wishlist.WishlistCountResponse `json:"data"`
	}
	err := json.Unmarshal(countW.Body.Bytes(), &countResp)
	s.Require().NoError(err)
	s.Equal(2, countResp.Data.Count)

	// Test GET /wishlist
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist?page=1&pageSize=10", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listW := httptest.NewRecorder()
	s.router.ServeHTTP(listW, listReq)
	s.Require().Equal(http.StatusOK, listW.Code)

	var listResp struct {
		Data []wishlist.WishlistItemResponse `json:"data"`
		Meta pagination.Page                 `json:"meta"`
	}
	err = json.Unmarshal(listW.Body.Bytes(), &listResp)
	s.Require().NoError(err)
	s.Equal(2, len(listResp.Data))
	s.Equal(2, listResp.Meta.TotalItems)
}

func (s *WishlistHTTPTestSuite) TestCheckInWishlist() {
	_, token := s.createTestUser()

	// Initially false
	checkReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist/check/"+s.product1ID.String(), nil)
	checkReq.Header.Set("Authorization", "Bearer "+token)
	checkW := httptest.NewRecorder()
	s.router.ServeHTTP(checkW, checkReq)
	s.Require().Equal(http.StatusOK, checkW.Code)

	var checkResp struct {
		Data wishlist.WishlistCheckResponse `json:"data"`
	}
	err := json.Unmarshal(checkW.Body.Bytes(), &checkResp)
	s.Require().NoError(err)
	s.False(checkResp.Data.InWishlist)
	s.Nil(checkResp.Data.WishlistItemID)

	// Add product 1
	body, _ := json.Marshal(map[string]any{"productId": s.product1ID.String()})
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	addReq.Header.Set("Authorization", "Bearer "+token)
	addReq.Header.Set("Content-Type", "application/json")
	addW := httptest.NewRecorder()
	s.router.ServeHTTP(addW, addReq)
	s.Require().Equal(http.StatusCreated, addW.Code)

	// Now check again -> true
	checkReq2 := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist/check/"+s.product1ID.String(), nil)
	checkReq2.Header.Set("Authorization", "Bearer "+token)
	checkW2 := httptest.NewRecorder()
	s.router.ServeHTTP(checkW2, checkReq2)
	s.Require().Equal(http.StatusOK, checkW2.Code)

	err = json.Unmarshal(checkW2.Body.Bytes(), &checkResp)
	s.Require().NoError(err)
	s.True(checkResp.Data.InWishlist)
	s.NotNil(checkResp.Data.WishlistItemID)
}

func (s *WishlistHTTPTestSuite) TestRemoveItem_ByID() {
	_, token := s.createTestUser()

	// Add product 1
	body, _ := json.Marshal(map[string]any{"productId": s.product1ID.String()})
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	addReq.Header.Set("Authorization", "Bearer "+token)
	addReq.Header.Set("Content-Type", "application/json")
	addW := httptest.NewRecorder()
	s.router.ServeHTTP(addW, addReq)
	s.Require().Equal(http.StatusCreated, addW.Code)

	var addResp struct {
		Data wishlist.WishlistItemResponse `json:"data"`
	}
	_ = json.Unmarshal(addW.Body.Bytes(), &addResp)

	// Delete by item ID
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist/items/"+addResp.Data.ID, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delW := httptest.NewRecorder()
	s.router.ServeHTTP(delW, delReq)
	s.Require().Equal(http.StatusOK, delW.Code)

	// Check it's gone
	checkReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist/check/"+s.product1ID.String(), nil)
	checkReq.Header.Set("Authorization", "Bearer "+token)
	checkW := httptest.NewRecorder()
	s.router.ServeHTTP(checkW, checkReq)
	s.Require().Equal(http.StatusOK, checkW.Code)

	var checkResp struct {
		Data wishlist.WishlistCheckResponse `json:"data"`
	}
	_ = json.Unmarshal(checkW.Body.Bytes(), &checkResp)
	s.False(checkResp.Data.InWishlist)
}

func (s *WishlistHTTPTestSuite) TestRemoveItem_ByProductID() {
	_, token := s.createTestUser()

	// Add product 2
	body, _ := json.Marshal(map[string]any{"productId": s.product2ID.String()})
	addReq := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body))
	addReq.Header.Set("Authorization", "Bearer "+token)
	addReq.Header.Set("Content-Type", "application/json")
	addW := httptest.NewRecorder()
	s.router.ServeHTTP(addW, addReq)
	s.Require().Equal(http.StatusCreated, addW.Code)

	// Delete by product ID
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist/products/"+s.product2ID.String(), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delW := httptest.NewRecorder()
	s.router.ServeHTTP(delW, delReq)
	s.Require().Equal(http.StatusOK, delW.Code)

	// Deleting again -> 404
	delReq2 := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist/products/"+s.product2ID.String(), nil)
	delReq2.Header.Set("Authorization", "Bearer "+token)
	delW2 := httptest.NewRecorder()
	s.router.ServeHTTP(delW2, delReq2)
	s.Require().Equal(http.StatusNotFound, delW2.Code)
}

func (s *WishlistHTTPTestSuite) TestClearWishlist() {
	_, token := s.createTestUser()

	// Add product 1 and product 2
	body1, _ := json.Marshal(map[string]any{"productId": s.product1ID.String()})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body1))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusCreated, w1.Code)

	body2, _ := json.Marshal(map[string]any{"productId": s.product2ID.String()})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body2))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusCreated, w2.Code)

	// Clear wishlist
	clearReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist", nil)
	clearReq.Header.Set("Authorization", "Bearer "+token)
	clearW := httptest.NewRecorder()
	s.router.ServeHTTP(clearW, clearReq)
	s.Require().Equal(http.StatusOK, clearW.Code)

	var msg api.MessageResponse
	err := json.Unmarshal(clearW.Body.Bytes(), &msg)
	s.Require().NoError(err)
	s.Equal("Wishlist cleared successfully", msg.Message)

	// Verify count is 0
	countReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist/count", nil)
	countReq.Header.Set("Authorization", "Bearer "+token)
	countW := httptest.NewRecorder()
	s.router.ServeHTTP(countW, countReq)
	s.Require().Equal(http.StatusOK, countW.Code)

	var countResp struct {
		Data wishlist.WishlistCountResponse `json:"data"`
	}
	_ = json.Unmarshal(countW.Body.Bytes(), &countResp)
	s.Equal(0, countResp.Data.Count)
}

func (s *WishlistHTTPTestSuite) TestClearWishlist_WithQueryProductId() {
	_, token := s.createTestUser()

	// Add product 1 and product 2
	body1, _ := json.Marshal(map[string]any{"productId": s.product1ID.String()})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body1))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)
	s.Require().Equal(http.StatusCreated, w1.Code)

	body2, _ := json.Marshal(map[string]any{"productId": s.product2ID.String()})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(body2))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)
	s.Require().Equal(http.StatusCreated, w2.Code)

	// Delete product 1 via DELETE /wishlist?productId=...
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist?productId="+s.product1ID.String(), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delW := httptest.NewRecorder()
	s.router.ServeHTTP(delW, delReq)
	s.Require().Equal(http.StatusOK, delW.Code)

	// Count should now be 1
	countReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist/count", nil)
	countReq.Header.Set("Authorization", "Bearer "+token)
	countW := httptest.NewRecorder()
	s.router.ServeHTTP(countW, countReq)
	s.Require().Equal(http.StatusOK, countW.Code)

	var countResp struct {
		Data wishlist.WishlistCountResponse `json:"data"`
	}
	_ = json.Unmarshal(countW.Body.Bytes(), &countResp)
	s.Equal(1, countResp.Data.Count)
}

func TestWishlistHTTPTestSuite(t *testing.T) {
	suite.Run(t, new(WishlistHTTPTestSuite))
}
