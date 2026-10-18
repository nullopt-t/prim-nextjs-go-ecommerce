package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// retryAfterSeconds is the cooldown advertised on transient service failures.
// It gives clients and intermediary gateways a concrete, short back-off window
// instead of letting them retry a struggling dependency immediately.
const retryAfterSeconds = "5"

// setRetryAfterHint advertises a Retry-After header on retryable service
// failures (503 Service Unavailable, 504 Gateway Timeout). Any other status is
// left untouched so the header only appears where a retry is actually sensible.
func setRetryAfterHint(c *gin.Context, status int) {
	if status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout {
		c.Header("Retry-After", retryAfterSeconds)
	}
}
