package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestLocaleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		acceptLanguage string
		xAppLocale     string
		expectedLocale string
	}{
		{
			name:           "Default to English when no header provided",
			acceptLanguage: "",
			xAppLocale:     "",
			expectedLocale: "en",
		},
		{
			name:           "Arabic via Accept-Language ar",
			acceptLanguage: "ar",
			expectedLocale: "ar",
		},
		{
			name:           "Arabic via Accept-Language with country code ar-EG",
			acceptLanguage: "ar-EG,ar;q=0.9,en;q=0.8",
			expectedLocale: "ar",
		},
		{
			name:           "English via Accept-Language en-US",
			acceptLanguage: "en-US,en;q=0.9",
			expectedLocale: "en",
		},
		{
			name:           "Fallback to X-App-Locale header if Accept-Language is empty",
			acceptLanguage: "",
			xAppLocale:     "ar",
			expectedLocale: "ar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(middleware.Locale())

			var resolvedLocale string
			r.GET("/test", func(c *gin.Context) {
				resolvedLocale = middleware.GetLocale(c)
				c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tc.acceptLanguage)
			}
			if tc.xAppLocale != "" {
				req.Header.Set("X-App-Locale", tc.xAppLocale)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tc.expectedLocale, resolvedLocale)
		})
	}
}
