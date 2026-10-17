package middleware

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// Package middleware provides token bucket or sliding window rate limiting.
// It protects the application from denial-of-service (DoS) attacks, brute-force
// credential stuffing, and abusive traffic spikes.

const (
	// RateLimitPrefix is the default Redis key prefix for rate limiting keys.
	RateLimitPrefix = "rate_limit"
)

// RateLimiter defines the contract for checking and decrementing rate limit quotas.
// Implementations typically use Redis sliding window logs or token bucket algorithms.
type RateLimiter interface {
	// Allow checks if an operation under key is allowed given the max limit within the duration window.
	// Returns true if allowed, false if quota is exhausted, or an error if the backing store fails.
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
}

// RateLimit returns a reusable Gin middleware for rate limiting incoming requests.
// keyPrefix differentiates rate-limiting namespaces (e.g., "auth_login", "order_create").
// limit specifies the maximum number of allowed requests within the time window.
func RateLimit(
	limiter RateLimiter,
	keyPrefix string,
	limit int64,
	window time.Duration,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no limiter is provided, fail-open to preserve route availability
		if limiter == nil {
			c.Next()
			return
		}

		// Partition key by authenticated userID if present, otherwise by client IP
		identifier := c.ClientIP()
		if uidVal, exists := c.Get("userID"); exists && uidVal != nil {
			if uidStr, ok := uidVal.(string); ok && uidStr != "" {
				identifier = "usr:" + uidStr
			} else if stringer, ok := uidVal.(fmt.Stringer); ok && stringer.String() != "" {
				identifier = "usr:" + stringer.String()
			}
		}

		key := RateLimitPrefix + ":" + keyPrefix + ":" + identifier

		allowed, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			// Fail-open strategy with warning logging to prevent total outage if Redis is unavailable
			c.Next()
			return
		}

		if !allowed {
			retryAfterSec := int(math.Ceil(window.Seconds()))
			if retryAfterSec < 1 {
				retryAfterSec = 1
			}
			resetEpoch := time.Now().Add(window).Unix()

			c.Header("Retry-After", strconv.Itoa(retryAfterSec))
			c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetEpoch, 10))

			_ = c.Error(apierr.TooManyRequestsError())
			c.Abort()
			return
		}

		// Set rate limit headers for allowed request
		c.Header("X-RateLimit-Limit", strconv.FormatInt(limit, 10))

		// Request is within allowed quota; proceed
		c.Next()
	}
}
