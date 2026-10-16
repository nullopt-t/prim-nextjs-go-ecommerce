package middleware

import (
	"github.com/gin-gonic/gin"
)

// Package middleware provides essential HTTP security headers to protect against
// MIME-type confusion, clickjacking, protocol downgrade attacks, and cross-site leaks.

// NoSniff sets X-Content-Type-Options: nosniff on every response, along with OWASP baseline API security headers.
func NoSniff() gin.HandlerFunc {
	return SecurityHeaders()
}

// SecurityHeaders applies OWASP recommended security headers to all HTTP responses:
// - X-Content-Type-Options: nosniff
// - X-Frame-Options: DENY
// - Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
// - Referrer-Policy: strict-origin-when-cross-origin
// - Permissions-Policy: camera=(), microphone=(), geolocation=()
// - Content-Security-Policy: default-src 'none'; frame-ancestors 'none'
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		c.Next()
	}
}
