package user

import (
	"regexp"
	"strings"

	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/utils"
)

// postalCodePattern matches Egyptian postal codes (five digits).
var postalCodePattern = regexp.MustCompile(`^\d{5}$`)

// normalizePhone validates an optional phone number and returns it in E.164
// form. Local numbers are parsed as Egyptian, international ones use their
// own country code. An empty value is allowed (phone is optional) and
// returned as the empty string.
func normalizePhone(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		empty := ""
		return &empty, nil
	}
	e164, err := utils.IsValidPhone(trimmed, "EG")
	if err != nil {
		return nil, apierr.BadRequestError("Invalid phone number").WithCode(apierr.CodeInvalidInput)
	}
	return &e164, nil
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
