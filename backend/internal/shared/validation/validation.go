package validation

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	maxEmailLength    = 255
	maxNameLength     = 100
	minPasswordLength = 8
	maxPasswordLength = 100
)

// IsValidEmail verifies that the email does not exceed maximum length and is a syntactically valid email address.
func IsValidEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if len(trimmed) == 0 {
		return errors.New("email is required")
	}

	if len(trimmed) > maxEmailLength {
		return errors.New("too long email")
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return errors.New("invalid email")
	}

	// Ensure the parsed address matches the trimmed input (no display names like "Name <email>")
	// and contains non-empty local-part and domain separated by '@'
	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return errors.New("invalid email")
	}

	return nil
}

// IsValidName checks if the name string is non-empty (excluding whitespace) and does not exceed maximum rune length.
func IsValidName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) == 0 {
		return errors.New("name is required")
	}

	if utf8.RuneCountInString(trimmed) > maxNameLength {
		return errors.New("too long name")
	}

	return nil
}

// IsValidPassword checks that the password satisfies minimum and maximum length constraints.
func IsValidPassword(password string) error {
	if len(password) == 0 {
		return errors.New("password is required")
	}

	if len(password) < minPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	if len(password) > maxPasswordLength {
		return errors.New("too long password")
	}

	return nil
}

