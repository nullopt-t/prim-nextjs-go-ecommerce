package tag

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
// @Param data body CreateTagRequest true "tag data"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=TagResponse}
// @Router /tags [post]
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

type TagURIParam struct {
	ID string `uri:"id" binding:"uuid"`
}

// GetTagByID godoc
// @Summary get tag by id
// @Description get tag by id
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path TagURIParam true "tag id"
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse
// @Router /tags/{id} [get]
func (th *TagHandler) GetTagByID(c *gin.Context) {
	var param TagURIParam
	if err := c.ShouldBindUri(&param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	tagID, err := uuid.Parse(param.ID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	tag, err := th.tservice.GetTagByID(
		ctx,
		tagID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &TagResponse{
		ID:        tag.ID.String(),
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt.Format(time.RFC3339),
		UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: res,
		},
	)

}

// ListTags godoc
// @Summary list all tags
// @Description list all tags
// @Tags Tag
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]TagResponse,meta=api.Page}
// @Router /tags [get]
func (th *TagHandler) ListTags(c *gin.Context) {
	var q api.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if q.Page == 0 {
		q.Page = 1
	}

	if q.PageSize == 0 {
		q.PageSize = 10
	}

	ctx := c.Request.Context()
	tags, page, err := th.tservice.ListTags(
		ctx,
		&q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*TagResponse, 0, len(tags))
	for _, tag := range tags {
		res = append(res, &TagResponse{
			ID:        tag.ID.String(),
			Name:      tag.Name,
			CreatedAt: tag.CreatedAt.Format(time.RFC3339),
			UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: page,
		},
	)
}
