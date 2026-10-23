package job

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	redisContainer "github.com/testcontainers/testcontainers-go/modules/redis"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...log.Meta)  {}
func (noopLogger) Error(string, ...log.Meta) {}
func (noopLogger) Warn(string, ...log.Meta)  {}
func (noopLogger) Debug(string, ...log.Meta) {}

func TestRedisEmailFallbackWorker(t *testing.T) {
	ctx := context.Background()

	container, err := redisContainer.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{Addr: host + ":" + port.Port()})
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Ping(ctx).Err())

	t.Run("processes and deduplicates events", func(t *testing.T) {
		require.NoError(t, client.FlushDB(ctx).Err())

		event := &NotificationEvent{
			EventID:   uuid.New(),
			Type:      "email",
			Category:  "auth",
			Recipient: "user@example.com",
			Locale:    "en",
			Template:  "welcome",
			Data:      json.RawMessage(`{"email":"user@example.com"}`),
			Timestamp: time.Now(),
		}
		queue := NewRedisEmailFallbackQueue(client)
		require.NoError(t, queue.Enqueue(ctx, event))
		require.NoError(t, queue.Enqueue(ctx, event)) // duplicate delivery

		received := make(chan *NotificationEvent, 2)
		worker := NewRedisEmailFallbackWorker(client, func(_ context.Context, got *NotificationEvent) error {
			received <- got
			return nil
		}, noopLogger{}, 1)

		workerCtx, cancel := context.WithCancel(ctx)
		var wg sync.WaitGroup
		worker.Start(workerCtx, &wg)

		select {
		case got := <-received:
			assert.Equal(t, event.EventID, got.EventID)
			assert.Equal(t, event.Recipient, got.Recipient)
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for fallback event")
		}

		// Give the worker time to consume and deduplicate the second entry.
		time.Sleep(500 * time.Millisecond)
		cancel()
		wg.Wait()

		assert.Len(t, received, 0)
		pending, err := client.XPending(ctx, RedisEmailFallbackStream, RedisEmailFallbackGroup).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), pending.Count)
	})

	t.Run("routes exhausted events to DLQ", func(t *testing.T) {
		require.NoError(t, client.FlushDB(ctx).Err())

		event := &NotificationEvent{
			EventID:   uuid.New(),
			Type:      "email",
			Recipient: "fail@example.com",
			Locale:    "en",
			Template:  "welcome",
			Data:      json.RawMessage(`{}`),
			Timestamp: time.Now(),
		}
		queue := NewRedisEmailFallbackQueue(client)
		require.NoError(t, queue.Enqueue(ctx, event))

		worker := NewRedisEmailFallbackWorker(client, func(_ context.Context, _ *NotificationEvent) error {
			return errors.New("smtp down")
		}, noopLogger{}, 1)
		worker.maxRetries = 1

		workerCtx, cancel := context.WithCancel(ctx)
		var wg sync.WaitGroup
		worker.Start(workerCtx, &wg)

		deadline := time.Now().Add(10 * time.Second)
		for {
			n, err := client.XLen(ctx, RedisEmailFallbackDLQStream).Result()
			require.NoError(t, err)
			if n == 1 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("timed out waiting for DLQ event")
			}
			time.Sleep(100 * time.Millisecond)
		}

		cancel()
		wg.Wait()

		pending, err := client.XPending(ctx, RedisEmailFallbackStream, RedisEmailFallbackGroup).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), pending.Count)
	})
}
