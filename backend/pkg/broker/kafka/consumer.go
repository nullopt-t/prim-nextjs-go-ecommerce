// Package kafka provides consumer group abstractions for reading and processing Kafka messages.
package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer manages reading messages from a topic as part of a Kafka Consumer Group.
// It wraps a kafka.Reader and provides manual offset commit operations.
type Consumer struct {
	// reader is the underlying segmentio/kafka-go reader instance.
	reader    *kafka.Reader
	closeOnce sync.Once
	closeErr  error
}

// NewConsumer creates a consumer connected to a specific topic and groupID.
func NewConsumer(cfg Config, topic string) *Consumer {
	startOffset := cfg.StartOffset
	if startOffset == 0 {
		startOffset = kafka.LastOffset
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          topic,
		MinBytes:       1,                      // 1 byte for instant processing
		MaxBytes:       10e6,                   // 10MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: 0,                      // 0 = manual commits for at-least-once delivery
		StartOffset:    startOffset,
	})

	return &Consumer{
		reader: reader,
	}
}

// FetchMessage retrieves the next uncommitted message without auto-committing.
// The caller is responsible for calling CommitMessage after successful processing.
func (c *Consumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if c == nil || c.reader == nil {
		return kafka.Message{}, fmt.Errorf("consumer or reader is nil")
	}

	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("kafka fetch message: %w", err)
	}
	return msg, nil
}

// CommitMessage commits the offset of a successfully processed message.
// If the caller's context is canceled or expired, a fresh detached context with a 5-second timeout
// is used to prevent offset commit failures and duplicate message re-delivery.
func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	if c == nil || c.reader == nil {
		return fmt.Errorf("consumer or reader is nil")
	}

	commitCtx := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		commitCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	if err := c.reader.CommitMessages(commitCtx, msg); err != nil {
		return fmt.Errorf("kafka commit message (topic: %s, partition: %d, offset: %d): %w",
			msg.Topic, msg.Partition, msg.Offset, err)
	}
	return nil
}

// Close cleanly leaves the consumer group and closes the connection.
// It is protected by sync.Once to safely handle concurrent or duplicate Close calls.
func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	c.closeOnce.Do(func() {
		c.closeErr = c.reader.Close()
	})
	return c.closeErr
}
