package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Package middleware provides HTTP caching control handlers.
// Proper cache headers ensure that client browsers, intermediate CDNs,
// and reverse proxies only cache responses when safe, preventing information leakage.

// NoCache sets Cache-Control: no-store, private on every response.
// Apply to all authenticated routes to prevent proxies and browsers from
// caching user-specific or sensitive data (RFC 9205 §4.9).
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Instruct all caching mechanisms (CDNs, shared proxies, and browser caches)
		// not to store any part of the request or response.
		c.Header("Cache-Control", "no-store, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// PublicCache sets Cache-Control: public, max-age=<seconds> and Vary headers.
// Apply to fully public, read-only endpoints where caching is safe (RFC 9205 §4.9).
// maxAgeSeconds specifies the maximum time in seconds that a resource is considered fresh.
func PublicCache(maxAgeSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxAgeSeconds <= 0 {
			c.Header("Cache-Control", "no-cache, must-revalidate")
		} else {
			c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAgeSeconds))
		}

		c.Header("Vary", "Accept-Encoding, Accept-Language, Origin")
		c.Next()
	}
}
