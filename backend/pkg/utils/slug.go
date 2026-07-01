package utils

import (
	"regexp"
	"strings"
	"unicode"
)

var dashRegex = regexp.MustCompile(`-+`)

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	var b strings.Builder
	b.Grow(len(s))

	lastWasDash := false

	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastWasDash = false

		case unicode.IsSpace(r) || r == '-' || r == '_':
			if !lastWasDash {
				b.WriteByte('-')
				lastWasDash = true
			}
		}
	}

	slug := dashRegex.ReplaceAllString(b.String(), "-")
	return strings.Trim(slug, "-")
}
