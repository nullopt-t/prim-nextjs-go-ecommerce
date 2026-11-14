package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePublisher struct {
	topic    string
	key      string
	payload  []byte
	err      error
	calls    int
	topics   []string
	keys     []string
	payloads [][]byte
}

func (p *fakePublisher) Publish(_ context.Context, topic, key string, payload []byte) error {
	p.calls++
	p.topic = topic
	p.key = key
	p.payload = payload
	p.topics = append(p.topics, topic)
	p.keys = append(p.keys, key)
	p.payloads = append(p.payloads, payload)
	return p.err
}

type fakeFallbackQueue struct {
	err    error
	events []*job.NotificationEvent
}

func (f *fakeFallbackQueue) Enqueue(_ context.Context, event *job.NotificationEvent) error {
	f.events = append(f.events, event)
	return f.err
}

type testLogger struct{}

func (testLogger) Info(string, ...log.Meta)  {}
func (testLogger) Error(string, ...log.Meta) {}
func (testLogger) Warn(string, ...log.Meta)  {}
func (testLogger) Debug(string, ...log.Meta) {}

func TestEmailPublishUsesKafkaWhenAvailable(t *testing.T) {
	publisher := &fakePublisher{}
	emailFallback := &fakeFallbackQueue{}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)

	err := n.NotifyOTP(context.Background(), "email", "user@example.com", "123456")
	require.NoError(t, err)
	require.Equal(t, 1, publisher.calls)
	assert.Equal(t, job.TopicNotificationsEmail, publisher.topic)
	assert.Equal(t, "user@example.com", publisher.key)
	assert.Empty(t, emailFallback.events)

	var event job.NotificationEvent
	require.NoError(t, json.Unmarshal(publisher.payload, &event))
	assert.Equal(t, "email_otp", event.Template)
	assert.Equal(t, "user@example.com", event.Recipient)
}

func TestNotifyOTPSMSLogsInsteadOfPublishing(t *testing.T) {
	publisher := &fakePublisher{}
	emailFallback := &fakeFallbackQueue{}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)

	err := n.NotifyOTP(context.Background(), "sms", "01012340001", "123456")
	require.NoError(t, err)
	assert.Zero(t, publisher.calls)
	assert.Empty(t, emailFallback.events)
	assert.Empty(t, inAppFallback.events)
}

func TestEmailPublishFallsBackToRedisWhenKafkaFails(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka unavailable")}
	emailFallback := &fakeFallbackQueue{}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)

	err := n.NotifyWelcome(context.Background(), "user@example.com")
	require.NoError(t, err)
	require.Len(t, emailFallback.events, 1)

	event := emailFallback.events[0]
	assert.Equal(t, "welcome", event.Template)
	assert.Equal(t, "email", event.Type)
	assert.Equal(t, "user@example.com", event.Recipient)
	assert.NotEqual(t, uuid.Nil, event.EventID)
}

func TestEmailPublishReturnsErrorWhenKafkaAndRedisFail(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka unavailable")}
	emailFallback := &fakeFallbackQueue{err: errors.New("redis unavailable")}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)

	err := n.NotifyResetPassword(context.Background(), "user@example.com", "reset-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis fallback failed")
}

func TestInAppPublishUsesKafkaWhenAvailable(t *testing.T) {
	publisher := &fakePublisher{}
	emailFallback := &fakeFallbackQueue{}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)

	userID := uuid.New()
	err := n.SendInApp(context.Background(), SendInAppInput{
		UserID:   userID,
		Category: "system",
	})
	require.NoError(t, err)
	require.Equal(t, 1, publisher.calls)
	assert.Equal(t, job.TopicNotificationsInApp, publisher.topic)
	assert.Equal(t, userID.String(), publisher.key)
	assert.Empty(t, inAppFallback.events)
}

func TestInAppPublishFallsBackToRedisWhenKafkaFails(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka unavailable")}
	emailFallback := &fakeFallbackQueue{}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)

	userID := uuid.New()
	err := n.SendInApp(context.Background(), SendInAppInput{
		UserID:   userID,
		Category: "order",
	})
	require.NoError(t, err)
	require.Len(t, inAppFallback.events, 1)
	assert.Equal(t, "in_app", inAppFallback.events[0].Type)
	assert.Equal(t, "order", inAppFallback.events[0].Category)
	assert.Equal(t, userID.String(), inAppFallback.events[0].Recipient)
	assert.Empty(t, emailFallback.events)
}

func TestNotifyOrderPlacedPublishesEmailAndInApp(t *testing.T) {
	publisher := &fakePublisher{}
	emailFallback := &fakeFallbackQueue{}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)
	customerID := uuid.New()
	orderID := uuid.New()

	err := n.NotifyOrderPlaced(context.Background(), OrderPlacedPayload{
		OrderID:       orderID,
		OrderNumber:   "PRIM-1234",
		CustomerID:    &customerID,
		CustomerEmail: "buyer@example.com",
		CustomerName:  "Buyer",
		TotalAmount:   1999,
		Currency:      "EGP",
		PaymentMethod: "cash_on_delivery",
		Items: []OrderItemSummary{{
			Title:     "Widget",
			Quantity:  1,
			UnitPrice: 1999,
			LineTotal: 1999,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, 2, publisher.calls)
	assert.Empty(t, emailFallback.events)
	assert.Empty(t, inAppFallback.events)

	var emailEvent, inAppEvent job.NotificationEvent
	for i, payload := range publisher.payloads {
		var event job.NotificationEvent
		require.NoError(t, json.Unmarshal(payload, &event))
		switch publisher.topics[i] {
		case job.TopicNotificationsEmail:
			emailEvent = event
		case job.TopicNotificationsInApp:
			inAppEvent = event
		}
	}

	assert.Equal(t, job.TopicNotificationsEmail, publisher.topics[0])
	assert.Equal(t, "buyer@example.com", publisher.keys[0])
	assert.Equal(t, "order_placed", emailEvent.Template)
	assert.Equal(t, "order", emailEvent.Category)
	assert.Equal(t, "buyer@example.com", emailEvent.Recipient)

	assert.Equal(t, job.TopicNotificationsInApp, publisher.topics[1])
	assert.Equal(t, customerID.String(), publisher.keys[1])
	assert.Equal(t, "in_app", inAppEvent.Type)
	assert.Equal(t, "order", inAppEvent.Category)
	require.NotNil(t, inAppEvent.UserID)
	assert.Equal(t, customerID, *inAppEvent.UserID)

	var inApp SendInAppInput
	require.NoError(t, json.Unmarshal(inAppEvent.Data, &inApp))
	require.NotNil(t, inApp.ActionURL)
	assert.Equal(t, "/user/orders/"+orderID.String(), *inApp.ActionURL)
	assert.NotEmpty(t, inApp.Title["en"])
	assert.NotEmpty(t, inApp.Title["ar"])
}

func TestNotifyOrderPlacedSkipsInAppForGuest(t *testing.T) {
	publisher := &fakePublisher{}
	n := NewKafkaNotifier(publisher, testLogger{}, &fakeFallbackQueue{}, &fakeFallbackQueue{})

	err := n.NotifyOrderPlaced(context.Background(), OrderPlacedPayload{
		OrderID:       uuid.New(),
		CustomerEmail: "guest@example.com",
		TotalAmount:   500,
		Currency:      "EGP",
	})
	require.NoError(t, err)
	require.Equal(t, 1, publisher.calls)
	assert.Equal(t, job.TopicNotificationsEmail, publisher.topic)
}

func TestNotifyOrderPlacedFallsBackToRedisOnKafkaFailure(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka unavailable")}
	emailFallback := &fakeFallbackQueue{}
	inAppFallback := &fakeFallbackQueue{}
	n := NewKafkaNotifier(publisher, testLogger{}, emailFallback, inAppFallback)

	err := n.NotifyOrderPlaced(context.Background(), OrderPlacedPayload{
		OrderID:       uuid.New(),
		CustomerEmail: "guest@example.com",
		TotalAmount:   500,
		Currency:      "EGP",
	})
	require.NoError(t, err)
	require.Len(t, emailFallback.events, 1)
	assert.Equal(t, "order_placed", emailFallback.events[0].Template)
	assert.Equal(t, "guest@example.com", emailFallback.events[0].Recipient)
}

func TestNotifyOrderShippedPublishesEmailAndInApp(t *testing.T) {
	publisher := &fakePublisher{}
	n := NewKafkaNotifier(publisher, testLogger{}, &fakeFallbackQueue{}, &fakeFallbackQueue{})
	customerID := uuid.New()
	orderID := uuid.New()

	err := n.NotifyOrderShipped(context.Background(), OrderShippedPayload{
		OrderID:        orderID,
		OrderNumber:    "PRIM-5678",
		CustomerID:     &customerID,
		CustomerEmail:  "buyer@example.com",
		CustomerName:   "Buyer",
		Carrier:        "Aramex",
		TrackingNumber: "TRACK-999",
		TrackingURL:    "https://track.aramex.com/TRACK-999",
	})
	require.NoError(t, err)
	require.Equal(t, 2, publisher.calls)
	assert.Equal(t, job.TopicNotificationsEmail, publisher.topics[0])
	assert.Equal(t, job.TopicNotificationsInApp, publisher.topics[1])

	var emailEvent job.NotificationEvent
	require.NoError(t, json.Unmarshal(publisher.payloads[0], &emailEvent))
	assert.Equal(t, "order_shipped", emailEvent.Template)
	assert.Equal(t, "order", emailEvent.Category)

	var inAppEvent job.NotificationEvent
	require.NoError(t, json.Unmarshal(publisher.payloads[1], &inAppEvent))
	var inApp SendInAppInput
	require.NoError(t, json.Unmarshal(inAppEvent.Data, &inApp))
	assert.Equal(t, "order", inApp.Category)
	require.NotNil(t, inApp.ActionURL)
	assert.Equal(t, "/user/orders/"+orderID.String(), *inApp.ActionURL)
	assert.Contains(t, inApp.Message["en"], "Aramex")
	assert.Contains(t, inApp.Message["en"], "TRACK-999")
}

func TestNotifyOrderCancelledPublishesEmailAndInApp(t *testing.T) {
	publisher := &fakePublisher{}
	n := NewKafkaNotifier(publisher, testLogger{}, &fakeFallbackQueue{}, &fakeFallbackQueue{})
	customerID := uuid.New()
	orderID := uuid.New()

	err := n.NotifyOrderCancelled(context.Background(), OrderCancelledPayload{
		OrderID:       orderID,
		OrderNumber:   "PRIM-9012",
		CustomerID:    &customerID,
		CustomerEmail: "buyer@example.com",
		CustomerName:  "Buyer",
		Reason:        "Out of stock",
	})
	require.NoError(t, err)
	require.Equal(t, 2, publisher.calls)

	var emailEvent job.NotificationEvent
	require.NoError(t, json.Unmarshal(publisher.payloads[0], &emailEvent))
	assert.Equal(t, "order_cancelled", emailEvent.Template)

	var inAppEvent job.NotificationEvent
	require.NoError(t, json.Unmarshal(publisher.payloads[1], &inAppEvent))
	var inApp SendInAppInput
	require.NoError(t, json.Unmarshal(inAppEvent.Data, &inApp))
	assert.Contains(t, inApp.Message["en"], "Out of stock")
	assert.Contains(t, inApp.Message["ar"], "Out of stock")
}

func TestNotifyOrder_UsesPayloadLocaleOverContextLocale(t *testing.T) {
	publisher := &fakePublisher{}
	n := NewKafkaNotifier(publisher, testLogger{}, &fakeFallbackQueue{}, &fakeFallbackQueue{})

	// Context locale is "en", but order customer preference is "ar"
	ctx := middleware.WithLocale(context.Background(), "en")

	err := n.NotifyOrderShipped(ctx, OrderShippedPayload{
		OrderID:       uuid.New(),
		CustomerEmail: "buyer@example.com",
		Locale:        "ar",
	})
	require.NoError(t, err)

	var emailEvent job.NotificationEvent
	require.NoError(t, json.Unmarshal(publisher.payloads[0], &emailEvent))
	assert.Equal(t, "ar", emailEvent.Locale)
}
