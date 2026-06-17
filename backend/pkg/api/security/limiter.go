package security

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (r *RateLimiter) Allow(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {
	now := time.Now().Unix()
	cutoff := now - int64(window.Seconds())

	// Remove expired entries
	if err := r.client.ZRemRangeByScore(
		ctx,
		key,
		"0",
		strconv.FormatInt(cutoff, 10),
	).Err(); err != nil {
		return false, err
	}

	count, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count >= limit {
		return false, nil
	}

	if err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: uuid.NewString(), // any unique value works
	}).Err(); err != nil {
		return false, err
	}

	if err := r.client.Expire(ctx, key, window).Err(); err != nil {
		return false, err
	}

	return true, nil
}
