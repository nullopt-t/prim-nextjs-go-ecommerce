package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

func ErrorHandler(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		if ae, ok := err.(*apierr.APIError); ok {
			meta := log.Meta{"error": ae.LogValue()}
			if ae.Status >= 500 {
				logger.Error("internal server error", meta)
			} else {
				logger.Warn("client error", meta)
			}
			c.JSON(ae.Status, api.ErrorResponse{
				Code:    ae.Code,
				Message: ae.Error(),
				Details: ae.Fields,
			})
			c.Abort()
			return
		}

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
