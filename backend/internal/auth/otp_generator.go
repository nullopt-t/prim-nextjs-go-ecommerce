package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const OTPLength = 6

func RandomDigit(length int) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}
	return fmt.Sprintf("%0*d", length, n.Int64()), nil
}

type OTPGenerator struct{}

func NewOTPGenerator() *OTPGenerator {
	return &OTPGenerator{}
}

func (g *OTPGenerator) GenerateOTP() (string, error) {
	otp, err := RandomDigit(OTPLength)
	if err != nil {
		return "", fmt.Errorf(
			"failed to generate OTP: %w",
			err,
		)
	}

	return otp, nil
}
