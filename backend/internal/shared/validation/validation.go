package validation

import (
	"errors"
	"strings"
)

const (
	maxEmailLength    = 255
	maxNameLength     = 100
	maxPasswordLength = 100
)

func IsValidEmail(email string) error {
	if len(email) > maxEmailLength {
		return errors.New("too long email")
	}

	if !strings.Contains(email, "@") {
		return errors.New("invalid email")
	}

	return nil
}

func IsValidName(name string) error {
	if len(name) > maxNameLength {
		return errors.New("too long name")
	}

	if len(name) == 0 {
		return errors.New("name is required")
	}

	return nil
}

func IsValidPassword(password string) error {
	if len(password) > maxPasswordLength {
		return errors.New("too long password")
	}

	if len(password) == 0 {
		return errors.New("password is required")
	}

	return nil
}
