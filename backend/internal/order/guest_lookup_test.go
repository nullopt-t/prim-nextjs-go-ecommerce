package order

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestOrderHandler_GuestLookupValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ErrorHandler(log.NewNoopLogger()))
	handler := &OrderHandler{}
	router.GET("/orders/lookup", handler.GuestLookupOrder)

	// Missing parameters
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/orders/lookup", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Invalid email
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/lookup?email=not-an-email&orderNumber=10000000-0000-0000-0000-000000000001", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Invalid UUID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/orders/lookup?email=valid@example.com&orderNumber=not-a-uuid", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
