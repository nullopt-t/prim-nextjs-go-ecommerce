package notifier

type OTPPayload struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ResetPasswordPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type WelcomePayload struct {
	Email string `json:"email"`
}
