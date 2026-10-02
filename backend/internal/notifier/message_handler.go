package notifier

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/html"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/job"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

type Mailer interface {
	SendHTML(to []string, subject, html string) error
}

type NotificationStore interface {
	CreateNotification(ctx context.Context, n *model.UserNotification) error
}

var emailSubjects = map[string]map[string]string{
	"email_otp": {
		"en": "Verify Your Email - Prim",
		"ar": "تأكيد بريدك الإلكتروني - متجر بريم",
	},
	"welcome": {
		"en": "Welcome to Prim!",
		"ar": "مرحبًا بك في بريم!",
	},
	"reset_password": {
		"en": "Reset Your Password - Prim",
		"ar": "إعادة تعيين كلمة المرور - متجر بريم",
	},
}

// EventHandler processes Kafka events and routes them to SMTP or PostgreSQL.
type EventHandler struct {
	renderer  *html.Renderer
	mailer    Mailer
	store     NotificationStore
	logger    log.Logger
	clientURL string
}

func NewEventHandler(
	renderer *html.Renderer,
	mailer Mailer,
	store NotificationStore,
	logger log.Logger,
	clientURL string,
) *EventHandler {
	return &EventHandler{
		renderer:  renderer,
		mailer:    mailer,
		store:     store,
		logger:    logger,
		clientURL: clientURL,
	}
}

// HandleEmailEvent renders bilingual templates and delivers via SMTP.
func (h *EventHandler) HandleEmailEvent(ctx context.Context, event *job.NotificationEvent) error {
	locale := event.Locale
	if locale != "ar" && locale != "en" {
		locale = "en"
	}

	templateName := fmt.Sprintf("%s.%s", event.Template, locale)
	subjectMap := emailSubjects[event.Template]
	subject := "Notification - Prim"
	if subjectMap != nil {
		if s, ok := subjectMap[locale]; ok && s != "" {
			subject = s
		} else if s, ok := subjectMap["en"]; ok && s != "" {
			subject = s
		}
	}

	var templateData map[string]any

	switch event.Template {
	case "email_otp":
		var payload OTPPayload
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return fmt.Errorf("unmarshal OTP payload: %w", err)
		}
		templateName = fmt.Sprintf("email-otp.%s", locale)
		templateData = map[string]any{
			"Code": payload.Code,
		}

	case "welcome":
		var payload WelcomePayload
		_ = json.Unmarshal(event.Data, &payload)
		templateName = fmt.Sprintf("welcome.%s", locale)
		templateData = map[string]any{
			"ClientURL": h.clientURL,
		}

	case "reset_password":
		var payload ResetPasswordPayload
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return fmt.Errorf("unmarshal reset password payload: %w", err)
		}
		templateName = fmt.Sprintf("reset-password.%s", locale)
		templateData = map[string]any{
			"ResetURL": fmt.Sprintf("%s/auth/reset-password?token=%s", h.clientURL, payload.Token),
		}

	default:
		// Generic or custom template
		_ = json.Unmarshal(event.Data, &templateData)
	}

	renderedHTML, err := h.renderer.Render(templateName, templateData)
	if err != nil {
		return fmt.Errorf("render template %s: %w", templateName, err)
	}

	h.logger.Info("sending email notification", log.Meta{
		"recipient": event.Recipient,
		"template":  templateName,
		"subject":   subject,
	})

	return h.mailer.SendHTML(
		[]string{event.Recipient},
		subject,
		renderedHTML,
	)
}

// HandleInAppEvent stores localized alerts into PostgreSQL for the notification center.
func (h *EventHandler) HandleInAppEvent(ctx context.Context, event *job.NotificationEvent) error {
	if h.store == nil {
		return fmt.Errorf("notification store is nil")
	}

	var in SendInAppInput
	if err := json.Unmarshal(event.Data, &in); err != nil {
		return fmt.Errorf("unmarshal In-App payload: %w", err)
	}

	titleBytes, err := json.Marshal(in.Title)
	if err != nil {
		return fmt.Errorf("marshal title: %w", err)
	}

	messageBytes, err := json.Marshal(in.Message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	var metaBytes []byte
	if in.Metadata != nil {
		metaBytes, _ = json.Marshal(in.Metadata)
	}

	category := in.Category
	if category == "" {
		category = "system"
	}

	notification := &model.UserNotification{
		ID:        uuid.New(),
		UserID:    in.UserID,
		Title:     titleBytes,
		Message:   messageBytes,
		Category:  category,
		ActionURL: in.ActionURL,
		Metadata:  metaBytes,
	}

	h.logger.Info("persisting in-app notification", log.Meta{
		"userId":   in.UserID.String(),
		"category": category,
	})

	return h.store.CreateNotification(ctx, notification)
}
