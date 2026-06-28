package notifier

type OTPPayload struct {
	Identifier string `json:"identifier"`
	Code       string `json:"code"`
}

type ResetPasswordPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type WelcomePayload struct {
	Email string `json:"email"`
}
