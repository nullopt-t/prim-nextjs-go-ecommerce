package job

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/broker/kafka"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
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

// Start begins processing messages until ctx is cancelled.
func (w *KafkaWorker) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	w.logger.Info("starting Kafka worker", log.Meta{
		"topic":   w.topic,
		"workers": w.workers,
	})

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("context cancelled, stopping Kafka worker", log.Meta{"topic": w.topic})
			return
		default:
		}

		msg, err := w.consumer.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
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
			w.logger.Error("failed to process event after retries, routing to DLQ", log.Meta{
				"topic":   w.topic,
				"eventId": event.EventID.String(),
				"error":   err,
			})

			event.LastError = err.Error()
			dlqBytes, _ := json.Marshal(event)
			if w.producer != nil {
				_ = w.producer.Publish(ctx, TopicNotificationsDLQ, string(msg.Key), dlqBytes)
			}
		}

		// Commit offset upon completion or DLQ routing to progress the consumer group
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
