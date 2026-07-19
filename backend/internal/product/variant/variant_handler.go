package variant

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
)

type VariantHandler struct {
	vs *VariantService
}

func NewHandler(vs *VariantService) *VariantHandler {
	return &VariantHandler{
		vs: vs,
	}
}

type VariantIdURIParma struct {
	ID string `uri:"id"`
}

func (vh *VariantHandler) GetVariantMedia(c *gin.Context) {
	param := &VariantIdURIParma{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	variantID, err := uuid.Parse(param.ID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	media, err := vh.vs.GetVariantMedia(
		c.Request.Context(),
		variantID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: media,
		},
	)
}
