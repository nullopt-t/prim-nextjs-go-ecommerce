package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
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
	kafkaInAppPublishTimeout = 3 * time.Second
	redisFallbackTimeout     = 5 * time.Second
	inAppPublishTimeout      = 5 * time.Second
)

// EventPublisher publishes a serialized notification event to a topic.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload []byte) error
}

// EmailFallbackQueue stores email events when Kafka is unavailable.
type EmailFallbackQueue interface {
	Enqueue(ctx context.Context, event *job.NotificationEvent) error
}

// InAppFallbackQueue stores in-app notification events when Kafka is unavailable.
type InAppFallbackQueue interface {
	Enqueue(ctx context.Context, event *job.NotificationEvent) error
}

// NoopFallbackQueue provides an empty fallback queue for testing or unconfigured fallbacks.
type NoopFallbackQueue struct{}

// Enqueue implements EmailFallbackQueue and InAppFallbackQueue without taking action.
func (NoopFallbackQueue) Enqueue(ctx context.Context, event *job.NotificationEvent) error {
	return nil
}

// KafkaNotifier acts as a domain event producer publishing to Kafka topics.
// Email and in-app events fall back to Redis when Kafka cannot accept them.
type KafkaNotifier struct {
	producer      EventPublisher
	emailFallback EmailFallbackQueue
	inAppFallback InAppFallbackQueue
	logger        log.Logger
}

// NewKafkaNotifier returns a ready-to-use KafkaNotifier with fail-fast constructor invariants.
func NewKafkaNotifier(
	producer EventPublisher,
	logger log.Logger,
	emailFallback EmailFallbackQueue,
	inAppFallback InAppFallbackQueue,
) *KafkaNotifier {
	if producer == nil {
		panic("notifier: producer is required")
	}
	if logger == nil {
		panic("notifier: logger is required")
	}
	if emailFallback == nil {
		panic("notifier: email fallback is required")
	}
	if inAppFallback == nil {
		panic("notifier: in-app fallback is required")
	}
	return &KafkaNotifier{
		producer:      producer,
		emailFallback: emailFallback,
		inAppFallback: inAppFallback,
		logger:        logger,
	}
}

// publishEmail publishes an email event to Kafka and falls back to Redis if
// Kafka is unavailable or rejects the write. The fallback uses a detached
// context so an already-cancelled HTTP request cannot drop the notification.
func (n *KafkaNotifier) publishEmail(ctx context.Context, key string, event *job.NotificationEvent, payload []byte) error {
	pubCtx, cancel := context.WithTimeout(ctx, kafkaEmailPublishTimeout)
	kafkaErr := n.producer.Publish(pubCtx, job.TopicNotificationsEmail, key, payload)
	cancel()
	if kafkaErr == nil {
		return nil
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

// publishInApp publishes an in-app notification event to Kafka and falls back to
// Redis if Kafka is unavailable or rejects the write.
func (n *KafkaNotifier) publishInApp(ctx context.Context, key string, event *job.NotificationEvent, payload []byte) error {
	pubCtx, cancel := context.WithTimeout(ctx, kafkaInAppPublishTimeout)
	kafkaErr := n.producer.Publish(pubCtx, job.TopicNotificationsInApp, key, payload)
	cancel()
	if kafkaErr == nil {
		return nil
	}

	n.logger.Warn("Kafka in-app publish failed; falling back to Redis", log.Meta{
		"recipient": event.Recipient,
		"eventId":   event.EventID.String(),
		"error":     kafkaErr.Error(),
	})

	fallbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), redisFallbackTimeout)
	defer cancel()

	if err := n.inAppFallback.Enqueue(fallbackCtx, event); err != nil {
		return fmt.Errorf("kafka in-app publish failed (%v) and redis fallback failed: %w", kafkaErr, err)
	}

	n.logger.Info("queued in-app notification in Redis fallback", log.Meta{
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

	// shortcut: no SMS provider is wired yet; log the code so phone sign-in
	// still works in development/staging. Replace with a real sender before
	// production traffic.
	if channel == "sms" {
		n.logger.Info("sms OTP delivery is not wired; logging code", log.Meta{
			"recipient": identifier,
			"otp":       otp,
		})
		return nil
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

// NotifyOrderPlaced publishes the order-confirmation email and, for registered
// customers, an in-app alert. Email and in-app delivery both fall back to Redis
// when Kafka is unavailable.
func (n *KafkaNotifier) NotifyOrderPlaced(
	ctx context.Context,
	in OrderPlacedPayload,
) error {
	locale := in.Locale
	if locale != "ar" && locale != "en" {
		locale = middleware.GetLocaleFromContext(ctx)
	}
	if locale != "ar" && locale != "en" {
		locale = "en"
	}

	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}

	event := &job.NotificationEvent{
		EventID:    uuid.New(),
		Type:       "email",
		Category:   "order",
		Recipient:  in.CustomerEmail,
		Locale:     locale,
		Template:   "order_placed",
		Data:       payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	emailErr := n.publishEmail(ctx, in.CustomerEmail, event, bytes)

	var inAppErr error
	if in.CustomerID != nil && *in.CustomerID != uuid.Nil {
		orderRef := in.OrderNumber
		if orderRef == "" {
			orderRef = in.OrderID.String()
		}

		inAppErr = n.notifyOrderInApp(ctx, *in.CustomerID, in.OrderID, map[string]string{
			"en": "Order confirmed",
			"ar": "تم تأكيد طلبك",
		}, map[string]string{
			"en": fmt.Sprintf("Order %s has been placed successfully.", orderRef),
			"ar": fmt.Sprintf("تم إنشاء الطلب %s بنجاح.", orderRef),
		})
	}

	return errors.Join(emailErr, inAppErr)
}

// NotifyOrderShipped publishes a dispatch email and an in-app alert when the
// customer is registered.
func (n *KafkaNotifier) NotifyOrderShipped(
	ctx context.Context,
	in OrderShippedPayload,
) error {
	locale := in.Locale
	if locale != "ar" && locale != "en" {
		locale = middleware.GetLocaleFromContext(ctx)
	}
	if locale != "ar" && locale != "en" {
		locale = "en"
	}

	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}

	event := &job.NotificationEvent{
		EventID:    uuid.New(),
		Type:       "email",
		Category:   "order",
		Recipient:  in.CustomerEmail,
		Locale:     locale,
		Template:   "order_shipped",
		Data:       payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	emailErr := n.publishEmail(ctx, in.CustomerEmail, event, bytes)

	var inAppErr error
	if in.CustomerID != nil && *in.CustomerID != uuid.Nil {
		orderRef := in.OrderNumber
		if orderRef == "" {
			orderRef = in.OrderID.String()
		}

		enMessage := fmt.Sprintf("Order %s has shipped.", orderRef)
		arMessage := fmt.Sprintf("تم شحن الطلب %s.", orderRef)
		if in.Carrier != "" && in.TrackingNumber != "" {
			enMessage = fmt.Sprintf("Order %s has shipped via %s. Tracking number: %s.", orderRef, in.Carrier, in.TrackingNumber)
			arMessage = fmt.Sprintf("تم شحن الطلب %s عبر %s. رقم التتبع: %s.", orderRef, in.Carrier, in.TrackingNumber)
		} else if in.TrackingNumber != "" {
			enMessage = fmt.Sprintf("Order %s has shipped. Tracking number: %s.", orderRef, in.TrackingNumber)
			arMessage = fmt.Sprintf("تم شحن الطلب %s. رقم التتبع: %s.", orderRef, in.TrackingNumber)
		}

		inAppErr = n.notifyOrderInApp(ctx, *in.CustomerID, in.OrderID, map[string]string{
			"en": "Order shipped",
			"ar": "تم شحن طلبك",
		}, map[string]string{
			"en": enMessage,
			"ar": arMessage,
		})
	}

	return errors.Join(emailErr, inAppErr)
}

// NotifyOrderCancelled publishes a cancellation email and an in-app alert when
// the customer is registered.
func (n *KafkaNotifier) NotifyOrderCancelled(
	ctx context.Context,
	in OrderCancelledPayload,
) error {
	locale := in.Locale
	if locale != "ar" && locale != "en" {
		locale = middleware.GetLocaleFromContext(ctx)
	}
	if locale != "ar" && locale != "en" {
		locale = "en"
	}

	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}

	event := &job.NotificationEvent{
		EventID:    uuid.New(),
		Type:       "email",
		Category:   "order",
		Recipient:  in.CustomerEmail,
		Locale:     locale,
		Template:   "order_cancelled",
		Data:       payload,
		Timestamp:  time.Now(),
		RetryCount: 0,
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	emailErr := n.publishEmail(ctx, in.CustomerEmail, event, bytes)

	var inAppErr error
	if in.CustomerID != nil && *in.CustomerID != uuid.Nil {
		orderRef := in.OrderNumber
		if orderRef == "" {
			orderRef = in.OrderID.String()
		}

		enMessage := fmt.Sprintf("Order %s has been cancelled.", orderRef)
		arMessage := fmt.Sprintf("تم إلغاء الطلب %s.", orderRef)
		if in.Reason != "" {
			enMessage = fmt.Sprintf("%s Reason: %s.", enMessage, in.Reason)
			arMessage = fmt.Sprintf("%s السبب: %s.", arMessage, in.Reason)
		}

		inAppErr = n.notifyOrderInApp(ctx, *in.CustomerID, in.OrderID, map[string]string{
			"en": "Order cancelled",
			"ar": "تم إلغاء طلبك",
		}, map[string]string{
			"en": enMessage,
			"ar": arMessage,
		})
	}

	return errors.Join(emailErr, inAppErr)
}

// notifyOrderInApp publishes a category "order" alert with a stable dashboard
// link. The detached timeout keeps the post-commit notification alive if the
// originating request is already being cancelled.
func (n *KafkaNotifier) notifyOrderInApp(
	ctx context.Context,
	userID uuid.UUID,
	orderID uuid.UUID,
	title map[string]string,
	message map[string]string,
) error {
	actionURL := "/user/orders/" + orderID.String()

	inAppCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), inAppPublishTimeout)
	defer cancel()

	return n.SendInApp(inAppCtx, SendInAppInput{
		UserID:    userID,
		Title:     title,
		Message:   message,
		Category:  "order",
		ActionURL: &actionURL,
		Metadata: map[string]any{
			"orderId": orderID.String(),
		},
	})
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

	return n.publishInApp(ctx, in.UserID.String(), event, bytes)
}
