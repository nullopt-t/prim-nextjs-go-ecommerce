package middleware

import (
	"strings"

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
	return SecurityHeadersExcept()
}

// isDocsPath matches request paths that must NOT receive the strict API
// Content-Security-Policy. The Swagger UI serves HTML that loads its own JS and
// CSS assets, so `default-src 'none'` would render the docs unusable. Only the
// CSP is relaxed for these paths; the remaining hardening headers still apply.
func isDocsPath(path string) bool {
	return strings.HasPrefix(path, "/api/v1/docs")
}

// SecurityHeadersExcept applies the OWASP API header set to every response
// except the Swagger documentation paths, which need a permissive CSP to load
// their UI assets. This keeps the docs browsable without weakening the API.
func SecurityHeadersExcept() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// Swagger UI is an HTML app; a `default-src 'none'` CSP would block its
		// bundled scripts and styles. Serve it with a self-only policy instead.
		if isDocsPath(c.Request.URL.Path) {
			c.Header("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		} else {
			c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		}

		c.Next()
	}
}
