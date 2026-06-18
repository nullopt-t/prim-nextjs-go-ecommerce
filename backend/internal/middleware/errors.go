package middleware

import (
	"errors"

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
		if se, ok := errors.AsType[*security.SecureError](err); ok {
			logger.Error(
				se.UserMsg,
				log.Meta{
					"Error": se.Internal,
					"Stack": se.Stack,
				},
			)
			c.JSON(se.Status, api.BadReqResponse{
				Code:    se.Code,
				Message: se.UserMsg,
				Details: se.Fields,
			})
			c.Abort()
			return
		}

		logger.Error(
			err.Error(),
			log.Meta{},
		)
		c.JSON(500, api.ErrorResponse{
			Code:    "internal_error",
			Message: "Internal server error",
		})
		c.Abort()
	}
}
