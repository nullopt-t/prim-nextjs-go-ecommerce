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
	ID                string `json:"id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name              string `json:"name,omitempty" example:"black-friday"`
	PublicationStatus string `json:"publication_status,omitempty" example:"published"`
	CreatedAt         string `json:"created_at,omitempty" example:"2026-06-30T15:47:19Z"`
	UpdatedAt         string `json:"updated_at,omitempty" example:"2026-06-30T15:47:19Z"`
	DeletedAt         string `json:"deleted_at,omitempty" example:"2026-06-30T15:47:19Z"`
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
	body := &CreateTagRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()

	in := CreateTagInput{
		Name: body.Name,
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
				ID:                tag.ID.String(),
				Name:              tag.Name,
				PublicationStatus: tag.PublicationStatus.String(),
				CreatedAt:         tag.CreatedAt.Format(time.RFC3339),
				UpdatedAt:         tag.UpdatedAt.Format(time.RFC3339),
				DeletedAt: func() string {
					if tag.DeletedAt != nil {
						return tag.DeletedAt.Format(time.RFC3339)
					}
					return ""
				}(),
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
		ID:                tag.ID.String(),
		Name:              tag.Name,
		PublicationStatus: tag.PublicationStatus.String(),
		CreatedAt:         tag.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         tag.UpdatedAt.Format(time.RFC3339),
		DeletedAt: func() string {
			if tag.DeletedAt != nil {
				return tag.DeletedAt.Format(time.RFC3339)
			}
			return ""
		}(),
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
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.ApplyDefaults(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	}).Parse()

	ctx := c.Request.Context()
	tags, page, err := th.tservice.ListTags(
		ctx,
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*TagResponse, 0, len(tags))
	for _, tag := range tags {
		res = append(res, &TagResponse{
			ID:   tag.ID.String(),
			Name: tag.Name,
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

// AdminListTags godoc
// @Summary list all tags
// @Description list all tags
// @Tags Tag
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]TagResponse,meta=api.Page}
// @Router /admin/tags [get]
func (th *TagHandler) AdminListTags(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.ApplyDefaults(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	}).Parse()

	ctx := c.Request.Context()
	tags, page, err := th.tservice.AdminListTags(
		ctx,
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*TagResponse, 0, len(tags))
	for _, tag := range tags {
		res = append(res, &TagResponse{
			ID:                tag.ID.String(),
			Name:              tag.Name,
			PublicationStatus: tag.PublicationStatus.String(),
			CreatedAt:         tag.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         tag.UpdatedAt.Format(time.RFC3339),
			DeletedAt: func() string {
				if tag.DeletedAt != nil {
					return tag.DeletedAt.Format(time.RFC3339)
				}
				return ""
			}(),
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
