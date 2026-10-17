package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// Package middleware provides centralized error handling and translation.
// It intercepts errors appended to gin.Context, determines appropriate HTTP status
// codes and JSON payload formatting, and logs operational events according to severity.

// ErrorHandler intercepts any errors accumulated during request execution and
// formats them into a standardized JSON response matching api.ErrorResponse.
// It ensures non-cacheable responses on failure and writes structured log entries.
func ErrorHandler(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow downstream handlers and endpoints to process the request first
		c.Next()

		// If no errors were registered during processing, exit cleanly
		if len(c.Errors) == 0 {
			return
		}

		// Fallback if logger is nil
		if logger == nil {
			logger = log.NewConsoleLogger()
		}

		// Prevent writing response headers/body if downstream handler already wrote to the client
		if c.Writer.Written() {
			return
		}

		// Prevent caching of error responses
		c.Header("Cache-Control", "no-store, private")

		// Log all accumulated errors in the context for observability
		for i, ginErr := range c.Errors {
			var ae *apierr.APIError
			if errors.As(ginErr.Err, &ae) {
				meta := log.Meta{"error": ae.LogValue(), "index": i}
				if ae.Status >= 500 {
					logger.Error("internal server error", meta)
				} else {
					logger.Warn("client error", meta)
				}
			} else {
				logger.Error("accumulated error", log.Meta{"error": ginErr.Err, "index": i})
			}
		}

		// Extract the most recent error from the context to determine the response
		err := c.Errors.Last().Err

		var ae *apierr.APIError
		if errors.As(err, &ae) {
			// Write formatted structured API error JSON response
			c.JSON(ae.Status, api.ErrorResponse{
				Code:    ae.Code,
				Message: ae.Error(),
				Details: ae.Fields,
			})
			c.Abort()
			return
		}

		// Unhandled internal error fallback: avoid leaking raw internal details to clients
		logger.Error(
			"internal error",
			log.Meta{
				"error": err,
			},
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		})
		c.Abort()
	}
}
