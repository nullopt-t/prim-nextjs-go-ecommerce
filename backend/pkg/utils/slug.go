package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Slugify converts an arbitrary string (such as a product title, category name, or brand)
// into a URL-friendly ASCII slug containing only lowercase alphanumeric runes and single hyphens.
// Common symbols (&, +, @, /) are converted to words or hyphens, accented Latin characters are
// decomposed to ASCII equivalents, consecutive hyphens are collapsed, and empty results
// fall back to a safe default ("n-a" or optional custom fallback).
//
// Parameters:
//   - s: The source string to slugify.
//   - fallback: Optional fallback string to return if s contains no valid alphanumeric runes.
//
// Returns:
//   - string: A sanitized slug without leading/trailing hyphens.
func Slugify(s string, fallback ...string) string {
	// Replace common symbols with word or separator equivalents before rune filtering
	s = strings.ReplaceAll(s, "&", " and ")
	s = strings.ReplaceAll(s, "+", " plus ")
	s = strings.ReplaceAll(s, "@", " at ")

	// Decompose accented characters into base character + combining mark (NFD)
	s = norm.NFD.String(s)

	var b strings.Builder
	b.Grow(len(s))

	lastWasDash := false

	for _, r := range s {
		switch {
		// Non-spacing combining marks (accents, tildes, umlauts) are discarded
		case unicode.Is(unicode.Mn, r):
			continue

		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			b.WriteRune(unicode.ToLower(r))
			lastWasDash = false

		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastWasDash = false

		case unicode.IsSpace(r) || r == '-' || r == '_' || r == '/' || r == '\\' || r == '|' || r == '.' || r == ':':
			if b.Len() > 0 && !lastWasDash {
				b.WriteByte('-')
				lastWasDash = true
			}
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		if len(fallback) > 0 && fallback[0] != "" {
			return fallback[0]
		}
		return "n-a"
	}

	return slug
}

