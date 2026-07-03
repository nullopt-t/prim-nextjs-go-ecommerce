package tag

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
)

type TagHandler struct {
	tservice *TagService
}

func NewHandler(
	s *TagService,
) *TagHandler {
	return &TagHandler{
		tservice: s,
	}
}

type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

type TagResponse struct {
	ID        string `json:"id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name      string `json:"name" example:"black-friday"`
	CreatedAt string `json:"created_at" example:"2026-06-30T15:47:19Z"`
	UpdatedAt string `json:"updated_at" example:"2026-06-30T15:47:19Z"`
}

// CreateTag godoc
// @Summary create a new products tag
// @Description create a new products tag
// @Tags Tag
// @Accept json
// @Produce json
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=TagResponse}
func (th *TagHandler) CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()

	in := CreateTagInput{
		Name: req.Name,
	}

	tag, err := th.tservice.CreateTag(
		ctx,
		in,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: TagResponse{
				ID:        tag.ID.String(),
				Name:      tag.Name,
				CreatedAt: tag.CreatedAt.Format(time.RFC3339),
				UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}
