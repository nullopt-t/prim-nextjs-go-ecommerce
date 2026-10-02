package kafka

import (
	"time"
)

// Config encapsulates connection and tuning options for Kafka brokers.
type Config struct {
	Brokers       []string
	GroupID       string
	ClientID      string
	DialTimeout   time.Duration
	BatchTimeout  time.Duration
	MaxBatchBytes int
}

// DefaultConfig returns recommended production defaults.
func DefaultConfig(brokers []string, groupID string) Config {
	return Config{
		Brokers:       brokers,
		GroupID:       groupID,
		ClientID:      "prim-backend",
		DialTimeout:   10 * time.Second,
		BatchTimeout:  10 * time.Millisecond,
		MaxBatchBytes: 1024 * 1024, // 1MB
	}
}
