package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nullopt-t/prim-backend/pkg/broker/kafka"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type EventHandlerFunc func(ctx context.Context, event *NotificationEvent) error

// KafkaWorker subscribes to a Kafka topic as part of a consumer group and processes events concurrently.
type KafkaWorker struct {
	consumer   *kafka.Consumer
	producer   *kafka.Producer
	handler    EventHandlerFunc
	logger     log.Logger
	topic      string
	workers    int
	maxRetries int
}

func NewKafkaWorker(
	consumer *kafka.Consumer,
	producer *kafka.Producer,
	handler EventHandlerFunc,
	logger log.Logger,
	topic string,
	workers int,
) *KafkaWorker {
	if workers <= 0 {
		workers = 3
	}

	return &KafkaWorker{
		consumer:   consumer,
		producer:   producer,
		handler:    handler,
		logger:     logger,
		topic:      topic,
		workers:    workers,
		maxRetries: 3,
	}
}

// Start begins processing messages until ctx is cancelled, spawning w.workers
// concurrent consumer goroutines (the workers setting was previously ignored).
func (w *KafkaWorker) Start(ctx context.Context, wg *sync.WaitGroup) {
	w.logger.Info("starting Kafka worker", log.Meta{
		"topic":   w.topic,
		"workers": w.workers,
	})

	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go w.run(ctx, wg)
	}
}

// run is the per-worker message pump. Each worker fetches, processes, and commits
// messages independently until ctx is cancelled.
func (w *KafkaWorker) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		msg, err := w.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				w.logger.Info("context cancelled, stopping Kafka worker", log.Meta{"topic": w.topic})
				return
			}
			w.logger.Error("failed to fetch message from Kafka", log.Meta{
				"topic": w.topic,
				"error": err,
			})
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var event NotificationEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			w.logger.Error("failed to unmarshal Kafka event payload, skipping", log.Meta{
				"topic": w.topic,
				"error": err,
			})
			_ = w.consumer.CommitMessage(ctx, msg)
			continue
		}

		w.logger.Info("processing Kafka event", log.Meta{
			"topic":     w.topic,
			"eventId":   event.EventID.String(),
			"type":      event.Type,
			"recipient": event.Recipient,
			"locale":    event.Locale,
			"offset":    msg.Offset,
			"partition": msg.Partition,
		})

		// Process event with retry logic
		err = w.processWithRetry(ctx, &event)
		if err != nil {
			// During shutdown, do not DLQ or commit: the message must be
			// redelivered after restart instead of being silently dropped.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			w.logger.Error("failed to process event after retries, routing to DLQ", log.Meta{
				"topic":   w.topic,
				"eventId": event.EventID.String(),
				"error":   err,
			})

			event.LastError = err.Error()
			dlqBytes, _ := json.Marshal(event)
			if w.producer != nil {
				// Publish to the DLQ with a detached timeout context so a
				// cancelled request context cannot silence the DLQ write and
				// lose the message.
				dlqCtx := ctx
				if dlqCtx.Err() != nil {
					var cancel context.CancelFunc
					dlqCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
				}
				if err := w.producer.Publish(dlqCtx, TopicNotificationsDLQ, string(msg.Key), dlqBytes); err != nil {
					w.logger.Error("failed to publish event to DLQ", log.Meta{
						"topic":   w.topic,
						"eventId": event.EventID.String(),
						"error":   err,
					})
				}
			}
		}

		// Commit offset upon completion or DLQ routing to progress the consumer group.
		// CommitMessage falls back to a detached timeout context when ctx is cancelled.
		if err := w.consumer.CommitMessage(ctx, msg); err != nil {
			w.logger.Error("failed to commit message offset", log.Meta{
				"topic":     w.topic,
				"partition": msg.Partition,
				"offset":    msg.Offset,
				"error":     err,
			})
		}
	}
}

func (w *KafkaWorker) processWithRetry(ctx context.Context, event *NotificationEvent) error {
	var lastErr error

	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		event.RetryCount = attempt - 1
		err := w.handler(ctx, event)
		if err == nil {
			return nil
		}

		lastErr = err
		w.logger.Warn("event processing failed, scheduling retry", log.Meta{
			"eventId": event.EventID.String(),
			"attempt": attempt,
			"max":     w.maxRetries,
			"error":   err.Error(),
		})

		if attempt < w.maxRetries {
			// Exponential backoff: 1s, 2s, 4s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("exhausted %d retries: %w", w.maxRetries, lastErr)
}
