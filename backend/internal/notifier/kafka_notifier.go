package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/log"
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

const (
	kafkaEmailPublishTimeout = 3 * time.Second
	redisFallbackTimeout     = 5 * time.Second
)

// EventPublisher publishes a serialized notification event to a topic.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
}

// EmailFallbackQueue stores email events when Kafka is unavailable.
type EmailFallbackQueue interface {
	Enqueue(ctx context.Context, event *job.NotificationEvent) error
}

// KafkaNotifier acts as a domain event producer publishing to Kafka topics.
// Email events fall back to Redis when Kafka cannot accept them, so SMTP
// delivery continues during a Kafka outage.
type KafkaNotifier struct {
	producer      EventPublisher
	emailFallback EmailFallbackQueue
	logger        log.Logger
}

// NewKafkaNotifier returns a ready-to-use KafkaNotifier. The optional
// emailFallback is used only when publishing an email event to Kafka fails.
func NewKafkaNotifier(producer EventPublisher, logger log.Logger, emailFallback ...EmailFallbackQueue) *KafkaNotifier {
	n := &KafkaNotifier{
		producer: producer,
		logger:   logger,
	}
	if len(emailFallback) > 0 {
		n.emailFallback = emailFallback[0]
	}
	return n
}

// publishEmail publishes an email event to Kafka and falls back to Redis if
// Kafka is unavailable or rejects the write. The fallback uses a detached
// context so an already-cancelled HTTP request cannot drop the notification.
func (n *KafkaNotifier) publishEmail(ctx context.Context, key string, event *job.NotificationEvent, payload []byte) error {
	kafkaErr := fmt.Errorf("kafka producer is not configured")
	if n.producer != nil {
		pubCtx, cancel := context.WithTimeout(ctx, kafkaEmailPublishTimeout)
		kafkaErr = n.producer.Publish(pubCtx, job.TopicNotificationsEmail, key, payload)
		cancel()
		if kafkaErr == nil {
			return nil
		}
	}

	if n.emailFallback == nil {
		return kafkaErr
	}

	n.logger.Warn("Kafka email publish failed; falling back to Redis", log.Meta{
		"recipient": event.Recipient,
		"eventId":   event.EventID.String(),
		"error":     kafkaErr.Error(),
	})

	fallbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), redisFallbackTimeout)
	defer cancel()

	if err := n.emailFallback.Enqueue(fallbackCtx, event); err != nil {
		return fmt.Errorf("kafka email publish failed (%v) and redis fallback failed: %w", kafkaErr, err)
	}

	n.logger.Info("queued email notification in Redis fallback", log.Meta{
		"recipient": event.Recipient,
		"eventId":   event.EventID.String(),
	})

	return nil
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

	// SMS is not wired to a delivery consumer; reject explicitly instead of
	// publishing an email-templated OTP to a topic no SMS worker consumes.
	if channel == "sms" {
		return apierr.BadRequestError("SMS OTP delivery is not supported").
			WithCode(apierr.CodeInvalidInput)
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

	return n.publishEmail(ctx, identifier, event, bytes)
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

	return n.publishEmail(ctx, email, event, bytes)
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

	return n.publishEmail(ctx, email, event, bytes)
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
