package variant

import (
	"net/http"
	"time"

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

type VariantMediaResponse struct {
	ID     uuid.UUID `json:"id,omitzero"`
	Object struct {
		ID          uuid.UUID `json:"id"`
		URL         string    `json:"url"`
		ContentType string    `json:"content_type"`
		Size        int64     `json:"size"`
		Status      string    `json:"status"`
	} `json:"object"`
	Type      string `json:"type,omitempty"`
	SortOrder int    `json:"order"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
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

	var res = make([]*VariantMediaResponse, 0, len(media))
	for _, m := range media {
		vm := &VariantMediaResponse{
			ID:        m.ID,
			Type:      m.Type.String(),
			SortOrder: m.SortOrder,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
			UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
		}

		vm.Object.ID = m.Object.ID
		vm.Object.URL = m.Object.PublicURL
		vm.Object.Size = m.Object.Size
		vm.Object.Status = m.Object.Status.String()
		vm.Object.ContentType = m.Object.ContentType
		res = append(res, vm)
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: res,
		},
	)
}

func (vh *VariantHandler) UploadVariantMedia(c *gin.Context) {
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

	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(err)
		return
	}

	err = vh.vs.UploadVariantMedia(
		c.Request.Context(),
		variantID,
		file,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
		api.SuccessResponse{
			Message: "media uploaded successfully",
		},
	)
}

func (vh *VariantHandler) ReorderVariantMedia(c *gin.Context) {
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

	var body struct {
		IDs []uuid.UUID `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	err = vh.vs.ReorderVariantMedia(
		c.Request.Context(),
		variantID,
		body.IDs,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "media reordered successfully",
		},
	)
}
