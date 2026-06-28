package notifier

import (
	"encoding/json"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/html"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/job"
)

type Mailer interface {
	SendHTML(to []string, subject, html string) error
}

type EmailHandler struct {
	renderer   *html.Renderer
	dispatcher *job.MessageDispatcher
	mailer     Mailer
}

func NewEmailHandler(
	renderer *html.Renderer,
	dispatcher *job.MessageDispatcher,
	mailer Mailer,
) *EmailHandler {
	return &EmailHandler{
		renderer:   renderer,
		dispatcher: dispatcher,
		mailer:     mailer,
	}
}

func (h *EmailHandler) SendEmailOTP(
	msg *job.JobMessage,
) error {
	var payload OTPPayload
	if err := json.Unmarshal(
		msg.Payload,
		&payload,
	); err != nil {
		return err
	}

	html, err := h.renderer.Render(
		"email-otp",
		map[string]any{
			"Code": payload.Code,
		},
	)
	if err != nil {
		return err
	}

	return h.mailer.SendHTML(
		[]string{payload.Identifier},
		"OTP Code",
		html,
	)
}

func (h *EmailHandler) HandleMessage(
	msg *job.JobMessage,
) error {
	return h.dispatcher.Dispatch(string(msg.Command), msg)
}
