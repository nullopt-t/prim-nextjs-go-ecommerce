package middleware

import (
	"github.com/gin-gonic/gin"
)

// NoSniff sets X-Content-Type-Options: nosniff on every response.
// Apply to all routes to prevent proxies and browsers from
// MIME-sniffing attacks.
func NoSniff() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	}
}
