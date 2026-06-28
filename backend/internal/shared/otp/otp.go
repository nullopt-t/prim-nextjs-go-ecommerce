package otp

type OTPPurpose string

const (
	OTPTypeRegister OTPPurpose = "register"
	OTPTypeEmail    OTPPurpose = "email"
	OTPTypeLogin    OTPPurpose = "login"
	OTPTypeReset    OTPPurpose = "reset"
)

type OTPChannel string

const (
	OTPEmailChannel OTPChannel = "email"
)
