package notifier

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/job"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/broker/kafka"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

// SendInAppInput contains all parameters needed to trigger a localized in-app alert.
type SendInAppInput struct {
	UserID    uuid.UUID         `json:"userId"`
	Title     map[string]string `json:"title"`   // {"en": "...", "ar": "..."}
	Message   map[string]string `json:"message"` // {"en": "...", "ar": "..."}
	Category  string            `json:"category"`
	ActionURL *string           `json:"actionUrl,omitempty"`
	Metadata  map[string]any    `json:"metadata,omitempty"`
}

// KafkaNotifier acts as a domain event producer publishing to Kafka topics.
type KafkaNotifier struct {
	producer *kafka.Producer
	logger   log.Logger
}

// NewKafkaNotifier returns a ready-to-use KafkaNotifier.
func NewKafkaNotifier(producer *kafka.Producer, logger log.Logger) *KafkaNotifier {
	return &KafkaNotifier{
		producer: producer,
		logger:   logger,
	}
}

// NotifyOTP publishes an OTP authentication event to Kafka with the active request locale.
func (n *KafkaNotifier) NotifyOTP(
	ctx context.Context,
	channel,
	identifier,
	otp string,
) error {
	locale := middleware.GetLocaleFromContext(ctx)

	payload, err := json.Marshal(OTPPayload{
		Identifier: identifier,
		Code:       otp,
	})
	if err != nil {
		return err
	}

	event := &job.NotificationEvent{
		EventID:    uuid.New(),
		Type:       channel,
		Category:   "auth",
		Recipient:  identifier,
		Locale:     locale,
		Template:   "email_otp",
		Data:       payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	n.logger.Info("publishing OTP notification to Kafka", log.Meta{
		"recipient": identifier,
		"locale":    locale,
		"eventId":   event.EventID.String(),
	})

	return n.producer.Publish(ctx, job.TopicNotificationsEmail, identifier, bytes)
}

// NotifyWelcome publishes a welcome email event to Kafka.
func (n *KafkaNotifier) NotifyWelcome(
	ctx context.Context,
	email string,
) error {
	locale := middleware.GetLocaleFromContext(ctx)

	payload, err := json.Marshal(WelcomePayload{
		Email: email,
	})
	if err != nil {
		return err
	}

	event := &job.NotificationEvent{
		EventID:    uuid.New(),
		Type:       "email",
		Category:   "auth",
		Recipient:  email,
		Locale:     locale,
		Template:   "welcome",
		Data:       payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return n.producer.Publish(ctx, job.TopicNotificationsEmail, email, bytes)
}

// NotifyResetPassword publishes a password recovery email event to Kafka.
func (n *KafkaNotifier) NotifyResetPassword(
	ctx context.Context,
	email, token string,
) error {
	locale := middleware.GetLocaleFromContext(ctx)

	payload, err := json.Marshal(ResetPasswordPayload{
		Email: email,
		Token: token,
	})
	if err != nil {
		return err
	}

	event := &job.NotificationEvent{
		EventID:    uuid.New(),
		Type:       "email",
		Category:   "auth",
		Recipient:  email,
		Locale:     locale,
		Template:   "reset_password",
		Data:       payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return n.producer.Publish(ctx, job.TopicNotificationsEmail, email, bytes)
}

// SendInApp publishes an in-app notification event to Kafka.
func (n *KafkaNotifier) SendInApp(
	ctx context.Context,
	in SendInAppInput,
) error {
	locale := middleware.GetLocaleFromContext(ctx)

	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}

	event := &job.NotificationEvent{
		EventID:    uuid.New(),
		Type:       "in_app",
		Category:   in.Category,
		UserID:     &in.UserID,
		Recipient:  in.UserID.String(),
		Locale:     locale,
		Data:       payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	n.logger.Info("publishing In-App notification to Kafka", log.Meta{
		"userId":  in.UserID.String(),
		"eventId": event.EventID.String(),
	})

	return n.producer.Publish(ctx, job.TopicNotificationsInApp, in.UserID.String(), bytes)
}
