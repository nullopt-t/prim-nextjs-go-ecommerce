package utils

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

func IsValidPhone(phone string) (string, error) {
	input := strings.TrimSpace(phone)

	if strings.HasPrefix(input, "+") {
		num, err := phonenumbers.Parse(input, "")
		if err != nil {
			return "", fmt.Errorf("phone must be in E.164 format starting with +")
		}

		if phonenumbers.IsValidNumber(num) {
			return phonenumbers.Format(num, phonenumbers.E164), nil
		}
	}

	return "", fmt.Errorf("invalid phone number")
}

func IsValidEmail(email string) (string, error) {
	_, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", err
	}
	return NormalizeEmail(email), nil
}

func NormalizeEmail(email string) string {
	email = strings.ToLower(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local, _, _ := strings.Cut(parts[0], "+")
	return local + "@" + parts[1]
}
