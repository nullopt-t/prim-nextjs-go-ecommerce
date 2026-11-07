package history_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/history"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type fakeSessionStore struct {
	userID uuid.UUID
}

func (f fakeSessionStore) GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error) {
	if sessionID != "history-test-session" {
		return nil, fmt.Errorf("unknown session")
	}
	return &model.UserSession{
		ID:       sessionID,
		UserID:   f.userID,
		UserRole: model.CustomerRole,
		Type:     model.SessionTypeAuthenticated,
	}, nil
}

type allowAllLimiter struct{}

func (allowAllLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	return true, nil
}

type denyAllLimiter struct{}

func (denyAllLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	return false, nil
}

type HistoryTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *database.DB
	runner      database.Runner
	service     *history.Service
	router      *gin.Engine

	userID     uuid.UUID
	categoryID uuid.UUID
	product1ID uuid.UUID
	product2ID uuid.UUID
	product3ID uuid.UUID
}

func (s *HistoryTestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_history_test"),
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
		DBName:     "prim_history_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	for _, migration := range []string{
		"000001_init_sechema.up.sql",
		"000003_seller_domain.up.sql",
		"000005_user_product_views.up.sql",
	} {
		raw, err := os.ReadFile("../../migrations/" + migration)
		require.NoError(s.T(), err)
		_, err = db.ExecContext(ctx, string(raw))
		require.NoError(s.T(), err)
	}

	s.runner = database.NewTxRunner(db)
	repo := history.NewRepository()
	s.service = history.NewService(s.runner, repo, storage.NoopProvider{})

	s.userID = uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, full_name, role)
		VALUES ($1, $2, $3, 'customer')
	`, s.userID, "history@example.com", "History Customer")
	require.NoError(s.T(), err)

	s.categoryID = uuid.New()
	_, err = db.ExecContext(ctx, `INSERT INTO product_categories (id, name) VALUES ($1, $2)`, s.categoryID, "History")
	require.NoError(s.T(), err)

	s.product1ID = s.createProduct("history-product-1", "History Product 1", model.PublicationStatusPublished, nil, nil)
	s.product2ID = s.createProduct("history-product-2", "History Product 2", model.PublicationStatusPublished, nil, nil)
	s.product3ID = s.createProduct("history-product-3", "History Product 3", model.PublicationStatusPublished, nil, nil)

	variantID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO product_variants (id, sku, product_id, title, price, currency)
		VALUES ($1, $2, $3, $4, 249900, 'EGP')
	`, variantID, "HISTORY-SKU-1", s.product1ID, "Default")
	require.NoError(s.T(), err)
	_, err = db.ExecContext(ctx, `UPDATE products SET default_variant_id = $1 WHERE id = $2`, variantID, s.product1ID)
	require.NoError(s.T(), err)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.ErrorHandler(log.NewNoopLogger()))

	handler := history.NewHandler(s.service)
	router := history.NewRouter(
		handler,
		fakeSessionStore{userID: s.userID},
		allowAllLimiter{},
		config.RateLimitConfig{
			HistoryWriteRequests: 120,
			HistoryWriteWindow:   time.Minute,
		},
	)
	router.MapRoutes(s.router.Group("/api/v1"))
}

func (s *HistoryTestSuite) SetupTest() {
	_, err := s.db.ExecContext(context.Background(), `TRUNCATE user_product_views`)
	require.NoError(s.T(), err)
}

func (s *HistoryTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *HistoryTestSuite) createProduct(
	slug string,
	title string,
	status model.PublicationStatus,
	sellerID *uuid.UUID,
	thumbnailID *uuid.UUID,
) uuid.UUID {
	s.T().Helper()

	id := uuid.New()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO products (
			id, category_id, slug, title, status, product_type, seller_id, thumbnail_object_id
		)
		VALUES ($1, $2, $3, $4, $5, 'simple', $6, $7)
	`, id, s.categoryID, slug, title, status, sellerID, thumbnailID)
	require.NoError(s.T(), err)
	return id
}

func (s *HistoryTestSuite) request(method, path string, body any, authenticated bool) *httptest.ResponseRecorder {
	s.T().Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	if authenticated {
		req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "history-test-session"})
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	return rec
}

func (s *HistoryTestSuite) list(limit int) []history.ViewedProductItem {
	s.T().Helper()

	rec := s.request(http.MethodGet, fmt.Sprintf("/api/v1/user/history/products?limit=%d", limit), nil, true)
	require.Equal(s.T(), http.StatusOK, rec.Code, rec.Body.String())

	var payload struct {
		Data []history.ViewedProductItem `json:"data"`
	}
	require.NoError(s.T(), json.Unmarshal(rec.Body.Bytes(), &payload))
	return payload.Data
}

func (s *HistoryTestSuite) TestRequiresAuthentication() {
	rec := s.request(http.MethodGet, "/api/v1/user/history/products", nil, false)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *HistoryTestSuite) TestWritesAreRateLimited() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler(log.NewNoopLogger()))
	history.NewRouter(
		history.NewHandler(s.service),
		fakeSessionStore{userID: s.userID},
		denyAllLimiter{},
		config.RateLimitConfig{
			HistoryWriteRequests: 1,
			HistoryWriteWindow:   time.Minute,
		},
	).MapRoutes(router.Group("/api/v1"))

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/user/history/products/"+s.product1ID.String(),
		nil,
	)
	req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "history-test-session"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	s.Equal(http.StatusTooManyRequests, rec.Code)
}

func (s *HistoryTestSuite) TestRecordIsIdempotentAndListsNewestFirst() {
	s.Require().Equal(http.StatusNoContent, s.request(
		http.MethodPost, "/api/v1/user/history/products/"+s.product1ID.String(), nil, true,
	).Code)
	time.Sleep(2 * time.Millisecond)
	s.Require().Equal(http.StatusNoContent, s.request(
		http.MethodPost, "/api/v1/user/history/products/"+s.product2ID.String(), nil, true,
	).Code)
	time.Sleep(2 * time.Millisecond)
	s.Require().Equal(http.StatusNoContent, s.request(
		http.MethodPost, "/api/v1/user/history/products/"+s.product1ID.String(), nil, true,
	).Code)

	items := s.list(10)
	s.Require().Len(items, 2)
	s.Equal(s.product1ID, items[0].ProductID)
	s.Equal(s.product2ID, items[1].ProductID)
	s.Equal("History Product 1", items[0].Title)
	s.Equal("history-product-1", items[0].Slug)
	s.Equal(int64(249900), items[0].BasePrice)
	s.Equal("EGP", items[0].Currency)

	var count int
	s.Require().NoError(s.db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM user_product_views WHERE user_id = $1`, s.userID,
	).Scan(&count))
	s.Equal(2, count)
}

func (s *HistoryTestSuite) TestListHidesProductsThatAreNoLongerVisible() {
	archived := s.createProduct("history-archived", "Archived", model.PublicationStatusArchived, nil, nil)
	draft := s.createProduct("history-draft", "Draft", model.PublicationStatusDraft, nil, nil)

	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_product_views (user_id, product_id, viewed_at)
		VALUES ($1, $2, now()), ($1, $3, now()), ($1, $4, now())
	`, s.userID, archived, draft, s.product1ID)
	s.Require().NoError(err)

	items := s.list(10)
	s.Require().Len(items, 1)
	s.Equal(s.product1ID, items[0].ProductID)
}

func (s *HistoryTestSuite) TestListHidesInactiveSellerProducts() {
	ctx := context.Background()
	sellerUserID := uuid.New()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (id, email, full_name, role)
		VALUES ($1, $2, $3, 'customer')
	`, sellerUserID, "inactive-seller@example.com", "Inactive Seller")
	s.Require().NoError(err)

	sellerID := uuid.New()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO sellers (id, user_id, store_name, slug, status)
		VALUES ($1, $2, $3, $4, 'suspended')
	`, sellerID, sellerUserID, "Inactive Store", "inactive-store")
	s.Require().NoError(err)

	productID := s.createProduct("history-seller-product", "Seller Product", model.PublicationStatusPublished, &sellerID, nil)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO user_product_views (user_id, product_id, viewed_at)
		VALUES ($1, $2, now())
	`, s.userID, productID)
	s.Require().NoError(err)

	items := s.list(10)
	s.Empty(items)
}

func (s *HistoryTestSuite) TestSyncMergesGuestSlugsAndIds() {
	s.createProduct("history-sync-archived", "Archived", model.PublicationStatusArchived, nil, nil)

	rec := s.request(http.MethodPost, "/api/v1/user/history/products/sync", map[string]any{
		"slugs": []string{"history-product-3", "history-product-1"},
	}, true)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	var syncPayload struct {
		Data history.SyncHistoryResponse `json:"data"`
	}
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &syncPayload))
	s.Equal(2, syncPayload.Data.Synced)

	items := s.list(10)
	s.Require().Len(items, 2)
	s.Equal(s.product3ID, items[0].ProductID)
	s.Equal(s.product1ID, items[1].ProductID)

	rec = s.request(http.MethodPost, "/api/v1/user/history/products/sync", map[string]any{
		"productIds": []string{s.product2ID.String()},
	}, true)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	items = s.list(10)
	s.Require().Len(items, 3)
	s.Equal(s.product2ID, items[0].ProductID)

	rec = s.request(http.MethodPost, "/api/v1/user/history/products/sync", map[string]any{
		"slugs": []string{"missing-product", "history-sync-archived"},
	}, true)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &syncPayload))
	s.Equal(0, syncPayload.Data.Synced)
}

func (s *HistoryTestSuite) TestSyncRejectsInvalidAndOversizedPayloads() {
	rec := s.request(http.MethodPost, "/api/v1/user/history/products/sync", map[string]any{
		"productIds": []string{"not-a-uuid"},
	}, true)
	s.Equal(http.StatusBadRequest, rec.Code)

	slugs := make([]string, history.MaxSyncItems+1)
	for i := range slugs {
		slugs[i] = fmt.Sprintf("history-extra-%d", i)
	}
	rec = s.request(http.MethodPost, "/api/v1/user/history/products/sync", map[string]any{
		"slugs": slugs,
	}, true)
	s.Equal(http.StatusBadRequest, rec.Code)

	rec = s.request(http.MethodPost, "/api/v1/user/history/products/sync", map[string]any{}, true)
	s.Equal(http.StatusBadRequest, rec.Code)
}

func (s *HistoryTestSuite) TestCapacityKeepsNewestFiftyViews() {
	ctx := context.Background()
	slugs := make([]string, 55)
	for i := range slugs {
		slugs[i] = fmt.Sprintf("history-cap-%02d", i)
		s.createProduct(slugs[i], fmt.Sprintf("Cap %d", i), model.PublicationStatusPublished, nil, nil)
	}

	synced, err := s.service.SyncGuestHistory(ctx, s.userID, nil, slugs)
	s.Require().NoError(err)
	s.Equal(55, synced)

	var count int
	s.Require().NoError(s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_product_views WHERE user_id = $1`, s.userID,
	).Scan(&count))
	s.Equal(history.MaxItems, count)

	items := s.list(history.MaxItems)
	s.Require().Len(items, history.MaxItems)
	s.Equal("history-cap-00", items[0].Slug)
	s.Equal("history-cap-49", items[len(items)-1].Slug)
}

func (s *HistoryTestSuite) TestDeleteAndClear() {
	for _, productID := range []uuid.UUID{s.product1ID, s.product2ID} {
		rec := s.request(http.MethodPost, "/api/v1/user/history/products/"+productID.String(), nil, true)
		s.Require().Equal(http.StatusNoContent, rec.Code)
	}

	rec := s.request(http.MethodDelete, "/api/v1/user/history/products/"+s.product1ID.String(), nil, true)
	s.Require().Equal(http.StatusNoContent, rec.Code)
	items := s.list(10)
	s.Require().Len(items, 1)
	s.Equal(s.product2ID, items[0].ProductID)

	rec = s.request(http.MethodDelete, "/api/v1/user/history/products/"+s.product1ID.String(), nil, true)
	s.Require().Equal(http.StatusNotFound, rec.Code)

	rec = s.request(http.MethodDelete, "/api/v1/user/history/products", nil, true)
	s.Require().Equal(http.StatusNoContent, rec.Code)
	s.Empty(s.list(10))
}

func (s *HistoryTestSuite) TestCascadeDeletesCleanViews() {
	ctx := context.Background()
	productID := s.createProduct("history-cascade-product", "Cascade", model.PublicationStatusPublished, nil, nil)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_product_views (user_id, product_id, viewed_at)
		VALUES ($1, $2, now())
	`, s.userID, productID)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, productID)
	s.Require().NoError(err)

	var count int
	s.Require().NoError(s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_product_views WHERE product_id = $1`, productID,
	).Scan(&count))
	s.Equal(0, count)

	cascadeUserID := uuid.New()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO users (id, email, full_name, role)
		VALUES ($1, $2, $3, 'customer')
	`, cascadeUserID, "cascade-user@example.com", "Cascade User")
	s.Require().NoError(err)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO user_product_views (user_id, product_id, viewed_at)
		VALUES ($1, $2, now())
	`, cascadeUserID, s.product1ID)
	s.Require().NoError(err)

	_, err = s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, cascadeUserID)
	s.Require().NoError(err)
	s.Require().NoError(s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_product_views WHERE user_id = $1`, cascadeUserID,
	).Scan(&count))
	s.Equal(0, count)
}

func (s *HistoryTestSuite) TestThumbnailIsResolvedFromStorage() {
	ctx := context.Background()
	objectID := uuid.New()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO storage_objects (id, bucket, object_key, content_type, file_size, status)
		VALUES ($1, 'catalog', 'products/history-cover.webp', 'image/webp', 1024, 'uploaded')
	`, objectID)
	s.Require().NoError(err)

	productID := s.createProduct("history-thumbnail-product", "Thumbnail", model.PublicationStatusPublished, nil, &objectID)
	rec := s.request(http.MethodPost, "/api/v1/user/history/products/"+productID.String(), nil, true)
	s.Require().Equal(http.StatusNoContent, rec.Code)

	items := s.list(10)
	s.Require().Len(items, 1)
	s.Require().NotNil(items[0].ThumbnailURL)
	s.Equal("http://localhost/catalog/products/history-cover.webp", *items[0].ThumbnailURL)
}

func TestHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(HistoryTestSuite))
}
