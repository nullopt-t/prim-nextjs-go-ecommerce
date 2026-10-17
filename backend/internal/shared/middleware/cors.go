package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// Package middleware provides Cross-Origin Resource Sharing (CORS) enforcement.
// CORS defines a mechanism for client web applications loaded in one domain to
// interact with resources in another domain securely.

// CORS returns a middleware that sets CORS headers based on an explicit allowlist.
// If the request Origin matches, it is echoed back and credentials are allowed.
// OPTIONS (preflight) requests are answered with 204 immediately, without auth.
func CORS(allowedOrigins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve the Origin request header sent by the browser
		origin := c.GetHeader("Origin")

		// Check if the origin is explicitly allowed (never allow "null" when credentials are enabled)
		isAllowed := origin != "" && origin != "null" && slices.Contains(allowedOrigins, origin)
		isWildcard := slices.Contains(allowedOrigins, "*")

		if isAllowed {
			// Echo origin back to satisfy browser CORS requirement for credentialed requests
			c.Header("Access-Control-Allow-Origin", origin)
			// Allow cookies, authorization headers, and client TLS certificates across origins
			c.Header("Access-Control-Allow-Credentials", "true")
			// Inform downstream caches that response varies depending on the incoming Origin
			c.Header("Vary", "Origin")
		} else if isWildcard && origin != "" && origin != "null" {
			// Wildcard origin allowed without credentials
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// Advertise permitted HTTP verbs for preflight negotiations
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Advertise permitted client request headers
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With, Accept-Language, X-App-Locale")

		// Cache preflight responses for 24 hours to reduce OPTIONS round trips
		c.Header("Access-Control-Max-Age", "86400")

		// Expose headers needed by frontend clients
		c.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Length, X-App-Locale")

		// Handle CORS preflight check immediately
		if c.Request.Method == http.MethodOptions {
			if !isAllowed && !isWildcard {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Proceed to subsequent middleware and handlers
		c.Next()
	}
}
