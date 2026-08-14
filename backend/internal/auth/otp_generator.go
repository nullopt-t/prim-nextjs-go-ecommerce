package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const OTPLength = 6

// GenerateOTP generates a random numeric string of the specified length.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	return fmt.Sprintf("%0*d", OTPLength, n.Int64()), nil
}
