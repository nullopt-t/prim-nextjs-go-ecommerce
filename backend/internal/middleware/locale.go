package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyLocale = "locale"
	LocaleAR         = "ar"
	LocaleEN         = "en"
	DefaultLocale    = LocaleEN
)

// Locale returns a Gin middleware that extracts the client's language preference
// from the Accept-Language or X-App-Locale headers and normalizes it to "ar" or "en".
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = c.GetHeader("X-App-Locale")
		}

		resolvedLocale := DefaultLocale
		normalized := strings.ToLower(strings.TrimSpace(lang))
		if strings.HasPrefix(normalized, "ar") {
			resolvedLocale = LocaleAR
		} else if strings.HasPrefix(normalized, "en") {
			resolvedLocale = LocaleEN
		}

		c.Set(ContextKeyLocale, resolvedLocale)
		if c.Request != nil {
			ctx := context.WithValue(c.Request.Context(), contextKey(ContextKeyLocale), resolvedLocale)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

type contextKey string

// GetLocale extracts the normalized locale from gin.Context, falling back to DefaultLocale.
func GetLocale(c *gin.Context) string {
	if val, exists := c.Get(ContextKeyLocale); exists {
		if loc, ok := val.(string); ok && loc != "" {
			return loc
		}
	}
	return DefaultLocale
}

// GetLocaleFromContext extracts the locale from standard context.Context.
func GetLocaleFromContext(ctx context.Context) string {
	if val := ctx.Value(contextKey(ContextKeyLocale)); val != nil {
		if loc, ok := val.(string); ok && loc != "" {
			return loc
		}
	}
	return DefaultLocale
}
