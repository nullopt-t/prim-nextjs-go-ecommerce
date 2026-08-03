package validation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
)

func ValidationError(c *gin.Context, err error) {
	if ve, ok := err.(validator.ValidationErrors); ok && ve != nil {
		fieldErrors := make([]api.FieldError, 0, len(ve))
		for _, e := range ve {
			fieldErrors = append(fieldErrors, api.FieldError{
				Field: e.Field(),
				Tags:  e.Tag(),
			})
		}
		_ = c.Error(security.NewSecureError(
			http.StatusBadRequest,
		).WithFields(fieldErrors...))
		return
	}
	_ = c.Error(
		security.NewSecureError(
			http.StatusBadRequest,
		).WithMessage("bad request data").Wrap(err),
	)
}
