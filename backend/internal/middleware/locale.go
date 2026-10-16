package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

// Package middleware provides HTTP request localization.
// It detects the preferred locale (Arabic or English) from incoming headers,
// stores it in the Gin context and standard request context, and provides helper
// functions to retrieve the resolved locale across services and handlers.

const (
	// ContextKeyLocale is the context key string used in gin.Context.
	ContextKeyLocale = "locale"
	// LocaleAR represents the Arabic locale identifier.
	LocaleAR = "ar"
	// LocaleEN represents the English locale identifier.
	LocaleEN = "en"
	// DefaultLocale defines the fallback locale when no supported preference is provided.
	DefaultLocale = LocaleEN
)

// Locale returns a Gin middleware that extracts the client's language preference
// prioritizing X-App-Locale over Accept-Language, normalizes it to "ar" or "en",
// sets Vary and Content-Language response headers, and attaches the locale to the request context.
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		resolvedLocale := DefaultLocale

		// 1. Explicit application-level preference selected by user takes highest precedence
		if appLocale := strings.ToLower(strings.TrimSpace(c.GetHeader("X-App-Locale"))); appLocale != "" {
			switch {
			case strings.HasPrefix(appLocale, LocaleAR):
				resolvedLocale = LocaleAR
			case strings.HasPrefix(appLocale, LocaleEN):
				resolvedLocale = LocaleEN
			}
		} else if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
			// 2. Parse RFC 9110 comma-separated language tags and match supported locales
			resolvedLocale = parseAcceptLanguage(acceptLang)
		}

		// Store resolved locale in Gin context for handler access
		c.Set(ContextKeyLocale, resolvedLocale)

		// Set caching and language response headers
		c.Header("Vary", "Accept-Language, X-App-Locale")
		c.Header("Content-Language", resolvedLocale)

		// Propagate resolved locale to standard context.Context for database, service, and template calls
		if c.Request != nil {
			ctx := context.WithValue(c.Request.Context(), contextKey(ContextKeyLocale), resolvedLocale)
			c.Request = c.Request.WithContext(ctx)
		}

		// Continue handler chain
		c.Next()
	}
}

// parseAcceptLanguage parses standard Accept-Language headers (e.g. "fr-FR, ar;q=0.9, en;q=0.8")
// and returns the best supported locale ("ar" or "en"), falling back to DefaultLocale ("en").
func parseAcceptLanguage(header string) string {
	parts := strings.Split(header, ",")
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if idx := strings.Index(tag, ";"); idx != -1 {
			tag = strings.TrimSpace(tag[:idx])
		}
		tag = strings.ToLower(tag)
		if strings.HasPrefix(tag, LocaleAR) {
			return LocaleAR
		}
		if strings.HasPrefix(tag, LocaleEN) {
			return LocaleEN
		}
	}
	return DefaultLocale
}

// contextKey defines an unexported private type for context keys to prevent collisions.
type contextKey string

// GetLocale extracts the normalized locale from gin.Context, falling back to DefaultLocale ("en")
// if the key is missing or not a valid string.
func GetLocale(c *gin.Context) string {
	if val, exists := c.Get(ContextKeyLocale); exists {
		if loc, ok := val.(string); ok && loc != "" {
			return loc
		}
	}
	return DefaultLocale
}

// GetLocaleFromContext extracts the normalized locale from standard context.Context.
// This is used by background jobs, repository layers, and email renderers where gin.Context is unavailable.
func GetLocaleFromContext(ctx context.Context) string {
	if ctx == nil {
		return DefaultLocale
	}
	if val := ctx.Value(contextKey(ContextKeyLocale)); val != nil {
		if loc, ok := val.(string); ok && loc != "" {
			return loc
		}
	}
	return DefaultLocale
}
