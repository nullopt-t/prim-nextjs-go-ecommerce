package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

const RateLimitPrefix = "rate_limit"

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
}

// RateLimit returns a reusable middleware for rate limiting
func RateLimit(
	limiter RateLimiter,
	keyPrefix string,
	limit int64,
	window time.Duration,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := RateLimitPrefix + ":" + keyPrefix + ":" + c.ClientIP()
		allowed, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			_ = c.Error(apierr.InternalServerError().Wrap(err))
			c.Abort()
			return
		}
		
		if !allowed {
			_ = c.Error(apierr.TooManyRequestsError())
			c.Abort()
			return
		}
		
		c.Next()
	}
}
