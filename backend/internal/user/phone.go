package user

import (
	"regexp"
	"strings"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// egyptianMobilePattern matches Egyptian mobile numbers in national
// (01xxxxxxxxx) or international (+20 / 0020) form. Operator prefixes are
// 010, 011, 012 and 015. Spaces and dashes are stripped before matching.
var egyptianMobilePattern = regexp.MustCompile(`^(?:\+20|0020|0)1[0125]\d{8}$`)

// postalCodePattern matches Egyptian postal codes (five digits).
var postalCodePattern = regexp.MustCompile(`^\d{5}$`)

// toDigits strips the separators people type into phone inputs.
func toDigits(raw string) string {
	return strings.ReplaceAll(strings.ReplaceAll(raw, " ", ""), "-", "")
}

// normalizePhone validates an optional Egyptian phone number. It returns
// the stripped form to persist; an empty value is allowed (phone is
// optional) and returned as the empty string.
func normalizePhone(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	stripped := toDigits(strings.TrimSpace(*raw))
	if stripped == "" {
		empty := ""
		return &empty, nil
	}
	if !egyptianMobilePattern.MatchString(stripped) {
		return nil, apierr.BadRequestError("Invalid Egyptian phone number").WithCode(apierr.CodeInvalidInput)
	}
	return &stripped, nil
}

// validatePostalCode rejects non-Egyptian postal codes when one is provided.
// Empty values are allowed (postal code is optional).
func validatePostalCode(raw *string) error {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	if !postalCodePattern.MatchString(trimmed) {
		return apierr.BadRequestError("Invalid postal code").WithCode(apierr.CodeInvalidInput)
	}
	return nil
}
