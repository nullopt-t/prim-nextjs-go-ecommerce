package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
}

// RateLimit returns a reusable middleware for rate limiting
func RateLimit(
	limiter RateLimiter,
	keyPrefix string,
	limit int64,
	window time.Duration,
	logger log.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}
		
		key := keyPrefix + ":" + c.ClientIP()
		allowed, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			logger.Error("rate limit error", log.Meta{"error": err})
			_ = c.Error(apierr.New(http.StatusInternalServerError, "Internal Server Error"))
			c.Abort()
			return
		}
		
		if !allowed {
			_ = c.Error(apierr.New(http.StatusTooManyRequests, "Too many requests").WithCode(apierr.CodeRateLimitExceeded))
			c.Abort()
			return
		}
		
		c.Next()
	}
}
