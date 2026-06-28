package auth

type OTPService struct {
}

type OTPPayload struct {
	Identifier string `json:"identifier"`
	Code       string `json:"code"`
}

func (os *OTPService) SendOTP(
	channel,
	identifier,
	otpCode string,
) error {
	return nil
}
