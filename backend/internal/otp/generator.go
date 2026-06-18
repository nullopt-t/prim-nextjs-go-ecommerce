package otp

import (
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
	"fmt"
)

var OTPLength map[string]int = map[string]int{
	"login": 6,
	"reset": 6,
}

type OTPGenerator struct{}

func NewOTPGenerator() *OTPGenerator {
	return &OTPGenerator{}
}

func (g *OTPGenerator) GenerateOTP(otpType string) (string, error) {
	if _, ok := OTPLength[otpType]; !ok {
		otpType = "login"
	}

	otp, err := utils.GenerateOTP(OTPLength[otpType])
	if err != nil {
		return "", fmt.Errorf(
			"failed to generate OTP: %w",
			err,
		)
	}

	return otp, nil
}
