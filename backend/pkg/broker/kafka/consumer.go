package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer manages reading messages from a topic as part of a Consumer Group.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a consumer connected to a specific topic and groupID.
func NewConsumer(cfg Config, topic string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          topic,
		MinBytes:       10e3,            // 10KB
		MaxBytes:       10e6,            // 10MB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: 0,               // 0 = manual commits for at-least-once delivery
		StartOffset:    kafka.FirstOffset,
	})

	return &Consumer{
		reader: reader,
	}
}

// FetchMessage retrieves the next uncommitted message without auto-committing.
func (c *Consumer) FetchMessage(ctx context.Context) (kafka.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return kafka.Message{}, fmt.Errorf("kafka fetch message: %w", err)
	}
	return msg, nil
}

// CommitMessage commits the offset of a successfully processed message.
func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	if err := c.reader.CommitMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka commit message (topic: %s, partition: %d, offset: %d): %w",
			msg.Topic, msg.Partition, msg.Offset, err)
	}
	return nil
}

// Close cleanly leaves the consumer group and closes connection.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
