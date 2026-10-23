package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/nullopt-t/prim-backend/internal/shared/job"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePublisher struct {
	err     error
	calls   int
	topic   string
	key     string
	payload []byte
}

func (p *fakePublisher) Publish(_ context.Context, topic, key string, payload []byte) error {
	p.calls++
	p.topic = topic
	p.key = key
	p.payload = payload
	return p.err
}

type fakeEmailFallback struct {
	err    error
	events []*job.NotificationEvent
}

func (f *fakeEmailFallback) Enqueue(_ context.Context, event *job.NotificationEvent) error {
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
	fallback := &fakeEmailFallback{}
	n := NewKafkaNotifier(publisher, testLogger{}, fallback)

	err := n.NotifyOTP(context.Background(), "email", "user@example.com", "123456")
	require.NoError(t, err)
	require.Equal(t, 1, publisher.calls)
	assert.Equal(t, job.TopicNotificationsEmail, publisher.topic)
	assert.Equal(t, "user@example.com", publisher.key)
	assert.Empty(t, fallback.events)

	var event job.NotificationEvent
	require.NoError(t, json.Unmarshal(publisher.payload, &event))
	assert.Equal(t, "email_otp", event.Template)
	assert.Equal(t, "user@example.com", event.Recipient)
}

func TestEmailPublishFallsBackToRedisWhenKafkaFails(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka unavailable")}
	fallback := &fakeEmailFallback{}
	n := NewKafkaNotifier(publisher, testLogger{}, fallback)

	err := n.NotifyWelcome(context.Background(), "user@example.com")
	require.NoError(t, err)
	require.Len(t, fallback.events, 1)

	event := fallback.events[0]
	assert.Equal(t, "welcome", event.Template)
	assert.Equal(t, "email", event.Type)
	assert.Equal(t, "user@example.com", event.Recipient)
	assert.NotEqual(t, [16]byte{}, event.EventID)
}

func TestEmailPublishReturnsErrorWhenKafkaAndRedisFail(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka unavailable")}
	fallback := &fakeEmailFallback{err: errors.New("redis unavailable")}
	n := NewKafkaNotifier(publisher, testLogger{}, fallback)

	err := n.NotifyResetPassword(context.Background(), "user@example.com", "reset-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis fallback failed")
}

func TestInAppPublishDoesNotUseEmailFallback(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka unavailable")}
	fallback := &fakeEmailFallback{}
	n := NewKafkaNotifier(publisher, testLogger{}, fallback)

	err := n.SendInApp(context.Background(), SendInAppInput{
		UserID:   [16]byte{1},
		Category: "system",
	})
	require.Error(t, err)
	assert.Empty(t, fallback.events)
}
