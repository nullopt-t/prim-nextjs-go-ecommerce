package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// NoCache sets Cache-Control: no-store, private on every response.
// Apply to all authenticated routes to prevent proxies and browsers from
// caching user-specific or sensitive data (RFC 9205 §4.9).
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, private")
		c.Next()
	}
}

// PublicCache sets Cache-Control: public, max-age=<seconds> and Vary: Accept-Encoding.
// Apply to fully public, read-only endpoints where caching is safe (RFC 9205 §4.9).
func PublicCache(maxAgeSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAgeSeconds))
		c.Header("Vary", "Accept-Encoding")
		c.Next()
	}
}
