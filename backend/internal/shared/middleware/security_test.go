package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// TestSecurityHeadersExcept verifies that the API responses carry the strict
// lockdown headers while the Swagger documentation paths receive a relaxed CSP
// so the interactive docs can load their own assets.
func TestSecurityHeadersExcept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.SecurityHeadersExcept())
	router.GET("/api/v1/products", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/api/v1/docs/public/*any", func(c *gin.Context) { c.Status(http.StatusOK) })

	t.Run("api responses are locked down", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
		router.ServeHTTP(w, req)

		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
		}
		if got := w.Header().Get("Content-Security-Policy"); got != "default-src 'none'; frame-ancestors 'none'" {
			t.Errorf("CSP = %q, want the strict API policy", got)
		}
		if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
			t.Errorf("X-Frame-Options = %q, want DENY", got)
		}
	})

	t.Run("swagger docs get a self-only CSP", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/docs/public/index.html", nil)
		router.ServeHTTP(w, req)

		if got := w.Header().Get("Content-Security-Policy"); got != "default-src 'self'; frame-ancestors 'none'" {
			t.Errorf("CSP = %q, want the relaxed docs policy", got)
		}
		// Hardening unrelated to script loading still applies.
		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
		}
	})
}
