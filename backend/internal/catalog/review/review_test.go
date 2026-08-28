package review_test

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
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/review"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ReviewHTTPTestSuite struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	db          *database.DB
	txRunner    database.Runner
	reviewRepo  *review.ReviewRepository
	service     *review.ReviewService
	handler     *review.ReviewHandler
	router      *gin.Engine
	jwtManager  *jwt.JWTManager
	secrets     config.Secrets

	categoryID uuid.UUID
	productID  uuid.UUID
	variantID  uuid.UUID
}

func (s *ReviewHTTPTestSuite) SetupSuite() {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("prim_review_http_test"),
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
		DBName:     "prim_review_http_test",
	})
	require.NoError(s.T(), err)
	s.db = db

	schemaBytes, err := os.ReadFile("../../../migrations/000001_init_sechema.up.sql")
	require.NoError(s.T(), err)
	_, err = db.Exec(ctx, string(schemaBytes))
	require.NoError(s.T(), err)

	s.txRunner = database.NewTxRunner(s.db)
	s.reviewRepo = review.NewReviewRepository()
	s.service = review.NewService(s.txRunner, s.reviewRepo)
	s.handler = review.NewHandler(s.service)

	s.secrets = config.Secrets{
		JwtAccessTokenSecretKey:  "review_test_access_secret_key_32_bytes",
		JwtRefreshTokenSecretKey: "review_test_refresh_secret_key_32_bytes",
	}
	s.jwtManager = jwt.NewJWTManager(s.secrets)

	logger := log.NewConsoleLogger()
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.ErrorHandler(logger))

	reviewRouter := review.NewRouter(s.handler, s.secrets)
	v1 := s.router.Group("/api/v1")
	reviewRouter.MapRoutes(v1)

	// Seed product hierarchy
	s.categoryID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_categories (id, public_id, name)
		VALUES ($1, $2, $3)
	`, s.categoryID, uuid.New(), "Review Category")
	require.NoError(s.T(), err)

	s.productID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO products (id, category_id, slug, title, status, product_type)
		VALUES ($1, $2, $3, $4, 'published', 'simple')
	`, s.productID, s.categoryID, "anc-headphones-pro", "ANC Headphones Pro")
	require.NoError(s.T(), err)

	s.variantID = uuid.New()
	_, err = s.db.Exec(ctx, `
		INSERT INTO product_variants (id, product_id, sku, title, price, is_default)
		VALUES ($1, $2, $3, $4, 29999, true)
	`, s.variantID, s.productID, "SKU-ANC-PRO-BLK", "Black Edition")
	require.NoError(s.T(), err)
}

func (s *ReviewHTTPTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func (s *ReviewHTTPTestSuite) createTestUser(role string) (uuid.UUID, string) {
	userID := uuid.New()
	_, err := s.db.Exec(context.Background(), `
		INSERT INTO users (id, email, role, status)
		VALUES ($1, $2, $3, 'active')
	`, userID, fmt.Sprintf("user-%s@example.com", uuid.NewString()[:8]), role)
	s.Require().NoError(err)

	roleVal := role
	token, err := s.jwtManager.GenerateAccessToken(&jwt.UserClaims{
		UserID:   userID,
		UserRole: &roleVal,
	})
	s.Require().NoError(err)
	return userID, token
}

func (s *ReviewHTTPTestSuite) createDeliveredOrder(userID uuid.UUID) uuid.UUID {
	ctx := context.Background()
	orderID := uuid.New()
	orderItemID := uuid.New()

	addrJSON := `{"street": "123 Main St", "city": "Cairo", "state": "Cairo", "country": "EG"}`

	_, err := s.db.Exec(ctx, `
		INSERT INTO orders (id, customer_id, customer_email, shipping_address, billing_address, status, total_amount, currency)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, 'delivered', 29999, 'USD')
	`, orderID, userID, "customer@example.com", addrJSON, addrJSON)
	s.Require().NoError(err)

	snapshotJSON := `{"title": "ANC Headphones Pro", "price": 29999}`

	_, err = s.db.Exec(ctx, `
		INSERT INTO order_items (id, order_id, variant_id, quantity, price_at_purchase, product_snapshot)
		VALUES ($1, $2, $3, 1, 29999, $4::jsonb)
	`, orderItemID, orderID, s.variantID, snapshotJSON)
	s.Require().NoError(err)

	return orderItemID
}

type StrictDataEnvelope[T any] struct {
	Data T `json:"data"`
}

type StrictPaginatedEnvelope[T any] struct {
	Data []T `json:"data"`
	Meta any `json:"meta"`
}

type StrictErrorEnvelope struct {
	Code    string           `json:"code"`
	Message string           `json:"message"`
	Details []api.FieldError `json:"details,omitempty"`
}

func (s *ReviewHTTPTestSuite) TestHTTP_01_CreateReview_SuccessAndValidation() {
	userID, token := s.createTestUser("customer")
	orderItemID := s.createDeliveredOrder(userID)

	title := "Outstanding noise cancellation"
	body := "Exceeded my expectations on airplane flights."

	// 1. Success Create Review (Rating = 5)
	reqBody, _ := json.Marshal(review.CreateReviewRequest{
		ProductID:   s.productID.String(),
		OrderItemID: orderItemID.String(),
		Rating:      5,
		Title:       &title,
		Body:        &body,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Require().Equal(http.StatusCreated, w.Code)
	s.Equal("application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var resp StrictDataEnvelope[review.ReviewResponse]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	s.Require().NoError(err)

	// Strict ReviewResponse assertions
	_, parseErr := uuid.Parse(resp.Data.ID)
	s.Require().NoError(parseErr, "Review ID must be valid UUID")
	s.Equal(s.productID.String(), resp.Data.ProductID)
	s.Equal(userID.String(), resp.Data.UserID)
	s.Equal(orderItemID.String(), resp.Data.OrderItemID)
	s.Equal(int16(5), resp.Data.Rating)
	s.Equal(&title, resp.Data.Title)
	s.Equal(&body, resp.Data.Body)
	s.Equal("pending", resp.Data.Status, "Newly created review must start in pending moderation")
	s.NotEmpty(resp.Data.CreatedAt)
	s.NotEmpty(resp.Data.UpdatedAt)

	// 2. Reject Duplicate Review for same Order Item
	wDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(reqBody))
	reqDup.Header.Set("Content-Type", "application/json")
	reqDup.Header.Set("Authorization", "Bearer "+token)
	s.router.ServeHTTP(wDup, reqDup)

	s.Require().Equal(http.StatusConflict, wDup.Code)
	var dupErr StrictErrorEnvelope
	s.Require().NoError(json.Unmarshal(wDup.Body.Bytes(), &dupErr))
	s.Equal("REVIEW_ALREADY_EXISTS", dupErr.Code)

	// 3. Reject Invalid Rating (Rating = 6 > 5)
	badRatingBody, _ := json.Marshal(review.CreateReviewRequest{
		ProductID:   s.productID.String(),
		OrderItemID: uuid.NewString(),
		Rating:      6,
	})
	wRating := httptest.NewRecorder()
	reqRating := httptest.NewRequest(http.MethodPost, "/api/v1/reviews", bytes.NewReader(badRatingBody))
	reqRating.Header.Set("Content-Type", "application/json")
	reqRating.Header.Set("Authorization", "Bearer "+token)
	s.router.ServeHTTP(wRating, reqRating)

	s.Require().Equal(http.StatusBadRequest, wRating.Code)
}

func (s *ReviewHTTPTestSuite) TestHTTP_02_CustomerGetAndUpdateMyReview() {
	userID, token := s.createTestUser("customer")
	orderItemID := s.createDeliveredOrder(userID)

	title := "Initial review"
	body := "Good headphones."
	createdReview, err := s.service.CreateReview(context.Background(), review.CreateReviewInput{
		ProductID:   s.productID,
		UserID:      userID,
		OrderItemID: orderItemID,
		Rating:      4,
		Title:       &title,
		Body:        &body,
	})
	s.Require().NoError(err)

	// 1. GET /api/v1/reviews/me
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+token)
	wMe := httptest.NewRecorder()
	s.router.ServeHTTP(wMe, reqMe)

	s.Require().Equal(http.StatusOK, wMe.Code)
	var meResp StrictPaginatedEnvelope[review.ReviewResponse]
	s.Require().NoError(json.Unmarshal(wMe.Body.Bytes(), &meResp))
	s.Require().NotEmpty(meResp.Data)
	s.Equal(createdReview.ID.String(), meResp.Data[0].ID)

	// 2. GET /api/v1/reviews/:id
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/"+createdReview.ID.String(), nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	wGet := httptest.NewRecorder()
	s.router.ServeHTTP(wGet, reqGet)

	s.Require().Equal(http.StatusOK, wGet.Code)
	var getResp StrictDataEnvelope[review.ReviewResponse]
	s.Require().NoError(json.Unmarshal(wGet.Body.Bytes(), &getResp))
	s.Equal(int16(4), getResp.Data.Rating)

	// 3. PATCH /api/v1/reviews/:id (Update Rating to 5)
	newRating := int16(5)
	newTitle := "Updated: Best headphones ever"
	updateBody, _ := json.Marshal(review.UpdateReviewRequest{
		Rating: &newRating,
		Title:  &newTitle,
	})
	reqPatch := httptest.NewRequest(http.MethodPatch, "/api/v1/reviews/"+createdReview.ID.String(), bytes.NewReader(updateBody))
	reqPatch.Header.Set("Content-Type", "application/json")
	reqPatch.Header.Set("Authorization", "Bearer "+token)
	wPatch := httptest.NewRecorder()
	s.router.ServeHTTP(wPatch, reqPatch)

	s.Require().Equal(http.StatusOK, wPatch.Code)
	var patchResp StrictDataEnvelope[review.ReviewResponse]
	s.Require().NoError(json.Unmarshal(wPatch.Body.Bytes(), &patchResp))
	s.Equal(int16(5), patchResp.Data.Rating)
	s.Equal(&newTitle, patchResp.Data.Title)

	// 4. DELETE /api/v1/reviews/:id
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/v1/reviews/"+createdReview.ID.String(), nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	wDel := httptest.NewRecorder()
	s.router.ServeHTTP(wDel, reqDel)

	s.Require().Equal(http.StatusOK, wDel.Code)

	// Verify it's gone
	reqGetDeleted := httptest.NewRequest(http.MethodGet, "/api/v1/reviews/"+createdReview.ID.String(), nil)
	reqGetDeleted.Header.Set("Authorization", "Bearer "+token)
	wGetDeleted := httptest.NewRecorder()
	s.router.ServeHTTP(wGetDeleted, reqGetDeleted)
	s.Require().Equal(http.StatusNotFound, wGetDeleted.Code)
}

func (s *ReviewHTTPTestSuite) TestHTTP_03_AdminModerationLifecycle() {
	_, adminToken := s.createTestUser("admin")
	customerID, _ := s.createTestUser("customer")
	orderItemID := s.createDeliveredOrder(customerID)

	title := "Review pending moderation"
	rv, err := s.service.CreateReview(context.Background(), review.CreateReviewInput{
		ProductID:   s.productID,
		UserID:      customerID,
		OrderItemID: orderItemID,
		Rating:      5,
		Title:       &title,
	})
	s.Require().NoError(err)

	// 1. GET /api/v1/admin/reviews?status=pending
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reviews?status=pending", nil)
	reqList.Header.Set("Authorization", "Bearer "+adminToken)
	wList := httptest.NewRecorder()
	s.router.ServeHTTP(wList, reqList)

	s.Require().Equal(http.StatusOK, wList.Code)
	var listResp StrictPaginatedEnvelope[review.ReviewResponse]
	s.Require().NoError(json.Unmarshal(wList.Body.Bytes(), &listResp))
	s.NotEmpty(listResp.Data)

	// 2. PATCH /api/v1/admin/reviews/:id/status -> approve
	statusBody, _ := json.Marshal(review.UpdateReviewStatusRequest{
		Status: "approved",
	})
	reqStatus := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/reviews/"+rv.ID.String()+"/status", bytes.NewReader(statusBody))
	reqStatus.Header.Set("Content-Type", "application/json")
	reqStatus.Header.Set("Authorization", "Bearer "+adminToken)
	wStatus := httptest.NewRecorder()
	s.router.ServeHTTP(wStatus, reqStatus)

	s.Require().Equal(http.StatusOK, wStatus.Code)
	var msgResp api.MessageResponse
	s.Require().NoError(json.Unmarshal(wStatus.Body.Bytes(), &msgResp))
	s.Equal("status updated successfully", msgResp.Message)

	// Verify status updated via GET /admin/reviews/:id
	reqAdminGet := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reviews/"+rv.ID.String(), nil)
	reqAdminGet.Header.Set("Authorization", "Bearer "+adminToken)
	wAdminGet := httptest.NewRecorder()
	s.router.ServeHTTP(wAdminGet, reqAdminGet)
	s.Require().Equal(http.StatusOK, wAdminGet.Code)

	var adminGetResp StrictDataEnvelope[review.ReviewResponse]
	s.Require().NoError(json.Unmarshal(wAdminGet.Body.Bytes(), &adminGetResp))
	s.Equal("approved", adminGetResp.Data.Status)

	// 3. Admin Delete
	reqAdminDel := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/reviews/"+rv.ID.String(), nil)
	reqAdminDel.Header.Set("Authorization", "Bearer "+adminToken)
	wAdminDel := httptest.NewRecorder()
	s.router.ServeHTTP(wAdminDel, reqAdminDel)
	s.Require().Equal(http.StatusOK, wAdminDel.Code)
}

func (s *ReviewHTTPTestSuite) TestHTTP_04_RatingSummaryComputation() {
	// Create multiple approved reviews and test summary calculation
	for i := 1; i <= 5; i++ {
		custID, _ := s.createTestUser("customer")
		orderItemID := s.createDeliveredOrder(custID)
		rv, err := s.service.CreateReview(context.Background(), review.CreateReviewInput{
			ProductID:   s.productID,
			UserID:      custID,
			OrderItemID: orderItemID,
			Rating:      int16(i),
		})
		s.Require().NoError(err)

		// Approve review
		err = s.service.UpdateReviewStatus(context.Background(), rv.ID, model.ReviewStatusApproved)
		s.Require().NoError(err)
	}

	summary, err := s.service.GetRatingSummary(context.Background(), s.productID)
	s.Require().NoError(err)
	s.Equal(float64(3.0), summary.AverageRating)
	s.Equal(5, summary.ReviewCount)
	s.Equal(1, summary.Distribution[1])
	s.Equal(1, summary.Distribution[2])
	s.Equal(1, summary.Distribution[3])
	s.Equal(1, summary.Distribution[4])
	s.Equal(1, summary.Distribution[5])
}

func TestReviewHTTPTestSuite(t *testing.T) {
	suite.Run(t, new(ReviewHTTPTestSuite))
}
