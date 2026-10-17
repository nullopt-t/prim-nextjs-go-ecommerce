package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
)

type mockSessionStore struct {
	getSessionByIDFunc func(ctx context.Context, sessionID string) (*model.UserSession, error)
}

func (m *mockSessionStore) GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error) {
	return m.getSessionByIDFunc(ctx, sessionID)
}

func TestAuthenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := log.NewConsoleLogger()

	setupRouter := func(store middleware.SessionStore) *gin.Engine {
		r := gin.New()
		r.Use(gin.Recovery())
		r.Use(middleware.ErrorHandler(logger))
		r.Use(middleware.Authenticate(store))
		return r
	}

	t.Run("Missing cookie returns 401", func(t *testing.T) {
		store := &mockSessionStore{}
		r := setupRouter(store)
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Provider returns error returns 401", func(t *testing.T) {
		store := &mockSessionStore{
			getSessionByIDFunc: func(ctx context.Context, sessionID string) (*model.UserSession, error) {
				return nil, errors.New("invalid session")
			},
		}
		r := setupRouter(store)
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: "invalid-token"})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid session sets context and proceeds", func(t *testing.T) {
		expectedUserID := uuid.New()
		expectedSessionID := "sess-12345"
		store := &mockSessionStore{
			getSessionByIDFunc: func(ctx context.Context, sessionID string) (*model.UserSession, error) {
				return &model.UserSession{
					ID:       expectedSessionID,
					UserID:   expectedUserID,
					UserRole: "admin",
				}, nil
			},
		}
		r := setupRouter(store)
		var capturedUserID any
		var capturedSessionID any
		var capturedUserRole any
		r.GET("/test", func(c *gin.Context) {
			capturedUserID, _ = c.Get("userID")
			capturedSessionID, _ = c.Get("sessionID")
			capturedUserRole, _ = c.Get("userRole")
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: expectedSessionID})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedUserID, capturedUserID)
		assert.Equal(t, expectedSessionID, capturedSessionID)
		assert.Equal(t, "admin", capturedUserRole)
	})

	t.Run("OptionalAuthenticate allows unauthenticated request without error", func(t *testing.T) {
		store := &mockSessionStore{}
		r := gin.New()
		r.Use(middleware.OptionalAuthenticate(store))
		var capturedUserID any
		r.GET("/test-optional", func(c *gin.Context) {
			capturedUserID, _ = c.Get("userID")
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test-optional", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Nil(t, capturedUserID)
	})

	t.Run("OptionalAuthenticate sets user context when valid cookie is present", func(t *testing.T) {
		expectedUserID := uuid.New()
		expectedSessionID := "sess-opt-123"
		store := &mockSessionStore{
			getSessionByIDFunc: func(ctx context.Context, sessionID string) (*model.UserSession, error) {
				return &model.UserSession{
					ID:       expectedSessionID,
					UserID:   expectedUserID,
					UserRole: "customer",
				}, nil
			},
		}
		r := gin.New()
		r.Use(middleware.OptionalAuthenticate(store))
		var capturedUserID any
		r.GET("/test-optional", func(c *gin.Context) {
			capturedUserID, _ = c.Get("userID")
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test-optional", nil)
		req.AddCookie(&http.Cookie{Name: middleware.SessionCookieName, Value: expectedSessionID})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, expectedUserID, capturedUserID)
	})
}
