// Package kafka provides producer implementations for publishing partition-keyed messages.
package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer handles publishing messages to Kafka topics with partition-key routing.
// It wraps a kafka.Writer configured with hash-based partitioning and Snappy compression.
type Producer struct {
	// writer is the underlying segmentio/kafka-go writer instance.
	writer    *kafka.Writer
	closeOnce sync.Once
	closeErr  error
}

// NewProducer constructs a configured Kafka producer.
// It uses a Hash balancer so messages with the same Key always land on the same partition,
// ensuring strict ordering of events per key (e.g., per user or per order).
func NewProducer(cfg Config) *Producer {
	batchBytes := int64(cfg.MaxBatchBytes)
	if batchBytes <= 0 {
		batchBytes = 1024 * 1024 // 1MB default
	}

	timeout := cfg.DialTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.Hash{},
		MaxAttempts:            5,
		BatchTimeout:           cfg.BatchTimeout,
		BatchBytes:             batchBytes,
		ReadTimeout:            timeout,
		WriteTimeout:           timeout,
		RequiredAcks:           kafka.RequireAll, // Require all in-sync replicas for durability
		Compression:            kafka.Snappy,
		AllowAutoTopicCreation: true,
	}

	return &Producer{
		writer: writer,
	}
}

// Publish writes a single keyed message to a Kafka topic.
// If the key is non-empty, the message is routed deterministically based on Hash(key).
func (p *Producer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("producer or writer is nil")
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	// WriteMessages evaluates context deadline; if ctx expires before batch flushes, an error is returned.
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka publish to topic %s: %w", topic, err)
	}

	return nil
}

// Close closes the underlying Kafka writer gracefully, flushing any buffered messages.
// Protected with sync.Once to avoid repeated closing errors.
func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	p.closeOnce.Do(func() {
		p.closeErr = p.writer.Close()
	})
	return p.closeErr
}
