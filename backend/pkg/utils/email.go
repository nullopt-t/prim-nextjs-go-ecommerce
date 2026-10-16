// Package utils provides utility helpers for email and phone normalization and validation.
package utils

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// IsValidPhone checks if the given phone number string is valid according to E.164 standards.
// It accepts international numbers with '+' or '00' prefixes, country codes without '+',
// or local phone numbers paired with an optional defaultRegion code (e.g. "US").
// If valid, it returns the normalized E.164 formatted string.
//
// Parameters:
//   - phone: A raw string containing a phone number.
//   - defaultRegion: Optional ISO 3166-1 alpha-2 region code (e.g., "US", "GB").
//
// Returns:
//   - string: The normalized E.164 formatted telephone number (e.g., "+12025550123").
//   - error: Non-nil if the number fails parsing or is an invalid phone number.
func IsValidPhone(phone string, defaultRegion ...string) (string, error) {
	input := strings.TrimSpace(phone)
	if input == "" {
		return "", fmt.Errorf("phone number cannot be empty")
	}

	region := ""
	if len(defaultRegion) > 0 {
		region = strings.ToUpper(strings.TrimSpace(defaultRegion[0]))
	}

	// Replace international call prefix "00" with "+"
	if strings.HasPrefix(input, "00") {
		input = "+" + strings.TrimPrefix(input, "00")
	}

	var num *phonenumbers.PhoneNumber
	var err error

	if strings.HasPrefix(input, "+") {
		num, err = phonenumbers.Parse(input, "")
	} else if region != "" {
		num, err = phonenumbers.Parse(input, region)
	} else {
		// Attempt parsing by auto-prepending '+' for numbers provided with country code but no '+'
		num, err = phonenumbers.Parse("+"+input, "")
	}

	if err == nil && phonenumbers.IsValidNumber(num) {
		return phonenumbers.Format(num, phonenumbers.E164), nil
	}

	return "", fmt.Errorf("invalid phone number")
}

// IsValidEmail verifies whether the provided email string has valid RFC 5322 syntax and returns
// a normalized version with lowercased domain and local-part.
//
// Parameters:
//   - email: The raw email string.
//
// Returns:
//   - string: The normalized email address.
//   - error: Non-nil if the email string cannot be parsed as a valid RFC address.
func IsValidEmail(email string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", err
	}
	return NormalizeEmail(addr.Address), nil
}

// NormalizeEmail normalizes an email address by converting domain and local-part to lowercase,
// and stripping plus-addressing tags only for providers known to support subaddressing (e.g. Gmail).
// For other providers, '+' is preserved per RFC 5322 to prevent mailbox delivery failures.
//
// Parameters:
//   - email: The raw or formatted email address.
//
// Returns:
//   - string: The normalized email address.
func NormalizeEmail(email string) string {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return strings.ToLower(email)
	}

	local := strings.ToLower(parts[0])
	domain := strings.ToLower(parts[1])

	// Subaddressing via '+' is standard for Gmail/Googlemail where it acts as a user alias.
	// For other domains, '+' is preserved per RFC 5322 to prevent routing issues or account lockout.
	if domain == "gmail.com" || domain == "googlemail.com" {
		if idx := strings.IndexByte(local, '+'); idx != -1 {
			local = local[:idx]
		}
	}

	return local + "@" + domain
}
