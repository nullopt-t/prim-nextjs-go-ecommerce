// Package security provides distributed rate limiting and token bucket / sliding window
// primitives backed by Redis for mitigating brute-force and denial-of-service traffic.
package security

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// slidingWindowScript executes the sliding window rate limit check and record atomically in Redis.
// KEYS[1]: rate limit key
// ARGV[1]: current timestamp in milliseconds (number)
// ARGV[2]: window duration in milliseconds (number)
// ARGV[3]: max allowed requests (number)
// ARGV[4]: unique member ID (string)
// ARGV[5]: TTL in seconds (integer ceiling)
// Returns 1 if allowed, 0 if limit exceeded.
var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]
local ttl = tonumber(ARGV[5])

local clearBefore = now - window
redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)

local currentRequests = redis.call('ZCARD', key)
if currentRequests < limit then
    redis.call('ZADD', key, now, member)
    redis.call('EXPIRE', key, ttl)
    return 1
else
    return 0
end
`)

// RateLimiter enforces sliding window rate limits across distributed service instances
// by tracking timestamped request identifiers in Redis Sorted Sets (ZSet).
type RateLimiter struct {
	// client is the go-redis client connection.
	client *redis.Client
}

// NewRateLimiter initializes a RateLimiter backed by the supplied Redis client.
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow checks whether an action identified by key is permitted under limit within window.
// If the request is allowed, it records the request timestamp and refreshes key expiry.
// It executes atomically via a Redis Lua script to avoid concurrency race conditions.
func (r *RateLimiter) Allow(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (bool, error) {
	if r == nil || r.client == nil {
		return false, errors.New("rate limiter client is not initialized")
	}

	if limit <= 0 || window <= 0 {
		return false, nil
	}

	nowMs := time.Now().UnixMilli()
	windowMs := window.Milliseconds()
	if windowMs <= 0 {
		windowMs = 1
	}
	ttlSec := int64(math.Ceil(window.Seconds()))
	if ttlSec <= 0 {
		ttlSec = 1
	}

	member := uuid.NewString()

	res, err := slidingWindowScript.Run(
		ctx,
		r.client,
		[]string{key},
		nowMs,
		windowMs,
		limit,
		member,
		ttlSec,
	).Int()

	if err != nil {
		return false, err
	}

	return res == 1, nil
}
