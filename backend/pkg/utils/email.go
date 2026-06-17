package utils

import (
	"strings"
)

func IsValidPhone(phone string) bool {
	return len(phone) >= 10 && len(phone) <= 15
}

func IsValidEmail(email string) bool {
	return strings.Contains(email, "@")
}
