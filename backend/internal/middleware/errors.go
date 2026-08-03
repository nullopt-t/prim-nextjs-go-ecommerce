package middleware

import (
	"net/http"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		if se, ok := err.(*security.SecureError); ok {
			logger.Error(
				"something went wrong",
				log.Meta{
					"error": se.LogValue(),
				},
			)
			// TODO: map the error code to the status
			// instead of writing it every where
			c.JSON(se.Status, api.BadReqResponse{
				Code:    se.Code,
				Message: se.Message,
				Details: se.Fields,
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
