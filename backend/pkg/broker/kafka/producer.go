package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer handles publishing messages to Kafka topics with partition-key routing.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer constructs a configured Kafka producer.
// It uses a Hash balancer so messages with the same Key always land on the same partition.
func NewProducer(cfg Config) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{},
		MaxAttempts:  5,
		BatchTimeout: cfg.BatchTimeout,
		RequiredAcks: kafka.RequireOne, // Leader acknowledgement for balance between safety and speed
		Compression:  kafka.Snappy,
	}

	return &Producer{
		writer: writer,
	}
}

// Publish writes a single keyed message to a Kafka topic.
func (p *Producer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka publish to topic %s: %w", topic, err)
	}

	return nil
}

// Close closes the underlying Kafka writer gracefully.
func (p *Producer) Close() error {
	return p.writer.Close()
}
