package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/redis/go-redis/v9"
)

const (
	// RedisInAppFallbackStream receives in-app notifications when Kafka is
	// unavailable. It uses a Redis stream and consumer group so workers can
	// acknowledge messages and recover unacknowledged deliveries after a crash.
	RedisInAppFallbackStream = "notifications.inapp.fallback"

	// RedisInAppFallbackGroup is the consumer group used by fallback workers.
	RedisInAppFallbackGroup = "notifications-inapp-fallback"

	// RedisInAppFallbackDLQStream stores events that exhausted their retries.
	RedisInAppFallbackDLQStream = "notifications.inapp.fallback.dlq"

	redisInAppFallbackProcessedPrefix = "notifications:inapp:fallback:processed:"
	redisInAppFallbackProcessedTTL    = 7 * 24 * time.Hour
)

// RedisInAppFallbackQueue enqueues in-app notification events into Redis. It is
// used by the API when publishing to Kafka fails.
type RedisInAppFallbackQueue struct {
	client *redis.Client
}

// NewRedisInAppFallbackQueue creates a Redis-backed in-app fallback queue.
func NewRedisInAppFallbackQueue(client *redis.Client) *RedisInAppFallbackQueue {
	if client == nil {
		panic("job: redis client is required")
	}
	return &RedisInAppFallbackQueue{client: client}
}

// Enqueue appends an in-app event to the fallback stream.
func (q *RedisInAppFallbackQueue) Enqueue(ctx context.Context, event *NotificationEvent) error {
	if q == nil || q.client == nil {
		return fmt.Errorf("redis in-app fallback queue is not configured")
	}
	if event == nil {
		return fmt.Errorf("redis in-app fallback event is nil")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal redis in-app fallback event: %w", err)
	}

	return q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: RedisInAppFallbackStream,
		MaxLen: 100_000,
		Approx: true,
		Values: map[string]any{"event": payload},
	}).Err()
}

// RedisInAppFallbackWorker consumes in-app events from the Redis fallback
// stream. It mirrors the Kafka worker's retry/DLQ behavior and uses a
// consumer group so unacknowledged events can be claimed after a crash.
type RedisInAppFallbackWorker struct {
	client      *redis.Client
	handler     EventHandlerFunc
	logger      log.Logger
	workers     int
	maxRetries  int
	consumer    string
	claimMinAge time.Duration
}

// NewRedisInAppFallbackWorker creates a worker that processes Redis fallback
// in-app events through handler.
func NewRedisInAppFallbackWorker(
	client *redis.Client,
	handler EventHandlerFunc,
	logger log.Logger,
	workers int,
) *RedisInAppFallbackWorker {
	if client == nil {
		panic("job: redis client is required")
	}
	if handler == nil {
		panic("job: handler is required")
	}
	if logger == nil {
		panic("job: logger is required")
	}
	if workers <= 0 {
		workers = 1
	}

	return &RedisInAppFallbackWorker{
		client:      client,
		handler:     handler,
		logger:      logger,
		workers:     workers,
		maxRetries:  3,
		consumer:    "redis-inapp-fallback-" + uuid.NewString(),
		claimMinAge: time.Minute,
	}
}

// Start begins consuming fallback events until ctx is cancelled.
func (w *RedisInAppFallbackWorker) Start(ctx context.Context, wg *sync.WaitGroup) {
	w.logger.Info("starting Redis in-app fallback worker", log.Meta{
		"stream":  RedisInAppFallbackStream,
		"group":   RedisInAppFallbackGroup,
		"workers": w.workers,
	})

	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go w.run(ctx, wg)
	}

	wg.Add(1)
	go w.claimLoop(ctx, wg)
}

func (w *RedisInAppFallbackWorker) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if ctx.Err() != nil {
			return
		}

		if err := w.ensureGroup(ctx); err != nil {
			w.logger.Error("failed to ensure Redis in-app fallback consumer group", log.Meta{
				"stream": RedisInAppFallbackStream,
				"error":  err,
			})
			if !sleepContext(ctx, time.Second) {
				return
			}
			continue
		}

		streams, err := w.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    RedisInAppFallbackGroup,
			Consumer: w.consumer,
			Streams:  []string{RedisInAppFallbackStream, ">"},
			Count:    1,
			Block:    time.Second,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			w.logger.Error("failed to read Redis in-app fallback stream", log.Meta{
				"stream": RedisInAppFallbackStream,
				"error":  err,
			})
			if !sleepContext(ctx, 500*time.Millisecond) {
				return
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				if !w.processMessage(ctx, msg) {
					return
				}
			}
		}
	}
}

func (w *RedisInAppFallbackWorker) claimLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	w.claimPending(ctx)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.claimPending(ctx)
		}
	}
}

// claimPending recovers messages that were delivered to a worker which exited
// before acknowledging them. MinIdle avoids stealing messages from an active worker.
func (w *RedisInAppFallbackWorker) claimPending(ctx context.Context) {
	start := "0-0"
	for {
		if ctx.Err() != nil {
			return
		}

		messages, next, err := w.client.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   RedisInAppFallbackStream,
			Group:    RedisInAppFallbackGroup,
			Consumer: w.consumer,
			MinIdle:  w.claimMinAge,
			Start:    start,
			Count:    10,
		}).Result()
		if err != nil {
			if ctx.Err() == nil {
				w.logger.Error("failed to claim pending Redis in-app fallback events", log.Meta{
					"stream": RedisInAppFallbackStream,
					"error":  err,
				})
			}
			return
		}

		for _, msg := range messages {
			if !w.processMessage(ctx, msg) {
				return
			}
		}

		if next == "" || next == "0-0" || len(messages) == 0 {
			return
		}
		start = next
	}
}

func (w *RedisInAppFallbackWorker) ensureGroup(ctx context.Context) error {
	err := w.client.XGroupCreateMkStream(ctx, RedisInAppFallbackStream, RedisInAppFallbackGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (w *RedisInAppFallbackWorker) processMessage(ctx context.Context, msg redis.XMessage) bool {
	event, err := decodeRedisFallbackEvent(msg)
	if err != nil {
		w.logger.Error("failed to decode Redis in-app fallback event, routing to DLQ", log.Meta{
			"messageId": msg.ID,
			"error":     err,
		})
		if w.publishDLQ(ctx, msg, nil, err) {
			w.ack(ctx, msg)
		}
		return true
	}

	if w.alreadyProcessed(ctx, event.EventID) {
		w.logger.Info("skipping duplicate Redis in-app fallback event", log.Meta{
			"eventId": event.EventID.String(),
		})
		w.ack(ctx, msg)
		return true
	}

	w.logger.Info("processing Redis in-app fallback event", log.Meta{
		"eventId":   event.EventID.String(),
		"recipient": event.Recipient,
		"template":  event.Template,
		"locale":    event.Locale,
		"messageId": msg.ID,
	})

	err = w.processWithRetry(ctx, event)
	if err != nil {
		if ctx.Err() != nil {
			// Leave unacknowledged so another worker can claim it
			return false
		}

		event.LastError = err.Error()
		w.logger.Error("failed to process Redis in-app fallback event after retries, routing to DLQ", log.Meta{
			"eventId": event.EventID.String(),
			"error":   err,
		})
		if w.publishDLQ(ctx, msg, event, err) {
			w.ack(ctx, msg)
		}
		return true
	}

	w.markProcessed(ctx, event.EventID)
	w.ack(ctx, msg)
	return true
}

func (w *RedisInAppFallbackWorker) processWithRetry(ctx context.Context, event *NotificationEvent) error {
	var lastErr error

	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		event.RetryCount = attempt - 1
		err := w.handler(ctx, event)
		if err == nil {
			return nil
		}

		lastErr = err
		w.logger.Warn("Redis in-app fallback event processing failed, scheduling retry", log.Meta{
			"eventId": event.EventID.String(),
			"attempt": attempt,
			"max":     w.maxRetries,
			"error":   err.Error(),
		})

		if attempt < w.maxRetries {
			if !sleepContext(ctx, time.Duration(1<<uint(attempt-1))*time.Second) {
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("exhausted %d retries: %w", w.maxRetries, lastErr)
}

func (w *RedisInAppFallbackWorker) publishDLQ(ctx context.Context, msg redis.XMessage, event *NotificationEvent, processErr error) bool {
	values := map[string]any{
		"sourceId": msg.ID,
		"error":    processErr.Error(),
	}
	if event != nil {
		payload, err := json.Marshal(event)
		if err != nil {
			w.logger.Error("failed to marshal Redis in-app fallback DLQ event", log.Meta{"error": err})
			return false
		}
		values["event"] = payload
	} else if raw, ok := msg.Values["event"]; ok {
		values["event"] = raw
	}

	dlqCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := w.client.XAdd(dlqCtx, &redis.XAddArgs{
		Stream: RedisInAppFallbackDLQStream,
		MaxLen: 100_000,
		Approx: true,
		Values: values,
	}).Err(); err != nil {
		w.logger.Error("failed to publish Redis in-app fallback event to DLQ; leaving it unacknowledged", log.Meta{
			"messageId": msg.ID,
			"error":     err,
		})
		return false
	}

	return true
}

func (w *RedisInAppFallbackWorker) alreadyProcessed(ctx context.Context, eventID uuid.UUID) bool {
	if eventID == uuid.Nil {
		return false
	}
	exists, err := w.client.Exists(ctx, redisInAppFallbackProcessedPrefix+eventID.String()).Result()
	if err != nil {
		w.logger.Warn("failed to check Redis in-app fallback deduplication key", log.Meta{
			"eventId": eventID.String(),
			"error":   err,
		})
		return false
	}
	return exists > 0
}

func (w *RedisInAppFallbackWorker) markProcessed(ctx context.Context, eventID uuid.UUID) {
	if eventID == uuid.Nil {
		return
	}
	markCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()

	if err := w.client.Set(markCtx, redisInAppFallbackProcessedPrefix+eventID.String(), "1", redisInAppFallbackProcessedTTL).Err(); err != nil {
		w.logger.Warn("failed to mark Redis in-app fallback event as processed", log.Meta{
			"eventId": eventID.String(),
			"error":   err,
		})
	}
}

func (w *RedisInAppFallbackWorker) ack(ctx context.Context, msg redis.XMessage) {
	ackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()

	if err := w.client.XAck(ackCtx, RedisInAppFallbackStream, RedisInAppFallbackGroup, msg.ID).Err(); err != nil {
		w.logger.Error("failed to acknowledge Redis in-app fallback event", log.Meta{
			"messageId": msg.ID,
			"error":     err,
		})
	}
}
