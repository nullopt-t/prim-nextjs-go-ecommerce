// Package otp defines enum types and constants for one-time password purposes and communication channels.
package otp

// OTPPurpose identifies the domain workflow requiring user verification.
type OTPPurpose string

const (
	// OTPPurposeRegister designates an OTP sent during initial customer account registration.
	OTPPurposeRegister OTPPurpose = "register"
	// OTPPurposeEmailVerification designates an OTP sent to verify or change an email address.
	OTPPurposeEmailVerification OTPPurpose = "email"
	// OTPPurposeLogin designates an OTP sent as a second-factor or passwordless sign-in challenge.
	OTPPurposeLogin OTPPurpose = "login"
	// OTPPurposeReset designates an OTP sent for self-service password reset.
	OTPPurposeReset OTPPurpose = "reset"

	// Backward compatibility aliases
	OTPTypeRegister = OTPPurposeRegister
	OTPTypeEmail    = OTPPurposeEmailVerification
	OTPTypeLogin    = OTPPurposeLogin
	OTPTypeReset    = OTPPurposeReset
)

// IsValid validates whether the OTPPurpose corresponds to a known, supported workflow.
func (p OTPPurpose) IsValid() bool {
	switch p {
	case OTPPurposeRegister, OTPPurposeEmailVerification, OTPPurposeLogin, OTPPurposeReset:
		return true
	default:
		return false
	}
}

// OTPChannel designates the communication medium used to deliver the one-time code.
type OTPChannel string

const (
	// OTPEmailChannel delivers the one-time passcode via SMTP email.
	OTPEmailChannel OTPChannel = "email"
	// OTPSMSChannel delivers the one-time passcode via SMS.
	OTPSMSChannel OTPChannel = "sms"
)

// IsValid validates whether the OTPChannel corresponds to a supported delivery channel.
func (c OTPChannel) IsValid() bool {
	switch c {
	case OTPEmailChannel, OTPSMSChannel:
		return true
	default:
		return false
	}
}

