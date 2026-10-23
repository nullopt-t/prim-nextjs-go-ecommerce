// Package redisclient builds Redis clients from application configuration.
package redisclient

import (
	"fmt"

	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/redis/go-redis/v9"
)

// New returns a Redis client configured from cfg. The client is lazy: callers
// should Ping it if they need to verify connectivity at startup.
func New(cfg config.RedisConfig) (*redis.Client, error) {
	if cfg.URL != "" {
		opts, err := redis.ParseURL(cfg.URL)
		if err != nil {
			return nil, fmt.Errorf("parse redis url: %w", err)
		}
		return redis.NewClient(opts), nil
	}

	if cfg.Host != "" && cfg.Port != 0 {
		return redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		}), nil
	}

	return nil, fmt.Errorf("REDIS_URL (or REDIS) environment variable is required")
}
