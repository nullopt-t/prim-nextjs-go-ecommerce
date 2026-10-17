package setting

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api"
)

// MaintenanceMiddleware rejects storefront traffic with HTTP 503 while
// maintenance mode is enabled. Administrative, authentication, portal, settings,
// and health endpoints stay reachable so operators can still sign in and turn
// maintenance mode back off.
func MaintenanceMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isMaintenanceExempt(c.Request.URL.Path) {
			c.Next()
			return
		}

		cfg, err := svc.GetConfig(c.Request.Context())
		if err == nil && cfg.MaintenanceMode {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, api.ErrorResponse{
				Code:    "MAINTENANCE_MODE",
				Message: cfg.MaintenanceMessage,
			})
			return
		}

		c.Next()
	}
}

// isMaintenanceExempt reports whether a request path must remain available while
// the store is in maintenance mode.
func isMaintenanceExempt(path string) bool {
	switch {
	case strings.HasPrefix(path, "/api/v1/admin"),
		strings.HasPrefix(path, "/api/v1/auth"),
		strings.HasPrefix(path, "/api/v1/settings"),
		strings.HasPrefix(path, "/api/v1/portal"),
		strings.HasPrefix(path, "/health"):
		return true
	default:
		return false
	}
}
