package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateOTP(length int) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	return fmt.Sprintf("%0*d", length, n.Int64()), nil
}
