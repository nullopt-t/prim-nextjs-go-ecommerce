package tag

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
)

type TagHandler struct {
	tservice *TagService
}

func NewHandler(s *TagService) *TagHandler {
	return &TagHandler{
		tservice: s,
	}
}

type CreateTagRequest struct {
	Name string `json:"name" binding:"required" example:"black-friday"`
}

type UpdateTagRequest struct {
	Name *string `json:"name,omitempty" example:"best-seller"`
}

type TagResponse struct {
	ID        string `json:"id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name      string `json:"name" example:"black-friday"`
	CreatedAt string `json:"created_at" example:"2026-06-30T15:47:19Z"`
	UpdatedAt string `json:"updated_at" example:"2026-06-30T15:47:19Z"`
}

type AdminTagResponse struct {
	TagResponse
	DeletedAt *string `json:"deleted_at,omitempty" example:"2026-06-30T15:47:19Z"`
}

// CreateTag godoc
// @Summary create a new product tag
// @Description create a new product tag
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body CreateTagRequest true "tag data"
// @Failure 400 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=TagResponse}
// @Router /admin/tags [post]
func (th *TagHandler) CreateTag(c *gin.Context) {
	var body CreateTagRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CreateTagInput{
		Name: body.Name,
	}

	tag, err := th.tservice.CreateTag(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
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

// GetTagByID godoc
// @Summary get tag by id
// @Description get tag by id
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path string true "Tag ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=TagResponse}
// @Router /admin/tags/{id} [get]
func (th *TagHandler) GetTagByID(c *gin.Context) {
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			security.ErrInvalidUUID().WithFields(
				api.FieldError{
					Field:   "tag_id",
					Message: "invalid tag UUID format",
				},
			),
		)
		return
	}

	tag, err := th.tservice.GetTagByID(c.Request.Context(), tagID)
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

// UpdateTagByID godoc
// @Summary update a tag
// @Description update a tag details by id
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path string true "Tag ID (UUID)" format(uuid)
// @Param input body UpdateTagRequest true "tag update payload"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/tags/{id} [patch]
func (th *TagHandler) UpdateTagByID(c *gin.Context) {
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			security.ErrInvalidUUID().WithFields(
				api.FieldError{
					Field:   "tag_id",
					Message: "invalid tag UUID format",
				},
			),
		)
		return
	}

	var body UpdateTagRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	err = th.tservice.UpdateByID(
		c.Request.Context(),
		tagID,
		UpdateTagInput{
			Name: body.Name,
		},
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "updated successfully",
	})
}

// DeleteTagByID godoc
// @Summary delete a tag
// @Description soft-delete a tag by id
// @Tags Tag
// @Accept json
// @Produce json
// @Param id path string true "Tag ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/tags/{id} [delete]
func (th *TagHandler) DeleteTagByID(c *gin.Context) {
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			security.ErrInvalidUUID().WithFields(
				api.FieldError{
					Field:   "tag_id",
					Message: "invalid tag UUID format",
				},
			),
		)
		return
	}

	if err := th.tservice.DeleteByID(c.Request.Context(), tagID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "deleted successfully",
	})
}

// ListTags godoc
// @Summary list public active tags
// @Description list public active tags in pages
// @Tags Tag
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]TagResponse,meta=api.Page}
// @Router /tags [get]
func (th *TagHandler) ListTags(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := th.tservice.ListTags(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]*TagResponse, 0, len(result.Items))
	for _, tag := range result.Items {
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
			Meta: result.Page,
		},
	)
}

// AdminListTags godoc
// @Summary list all tags (admin)
// @Description list all tags including soft-deleted ones for administration
// @Tags Tag
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]AdminTagResponse,meta=api.Page}
// @Router /admin/tags [get]
func (th *TagHandler) AdminListTags(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := th.tservice.AdminListTags(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]*AdminTagResponse, 0, len(result.Items))
	for _, tag := range result.Items {
		item := &AdminTagResponse{
			TagResponse: TagResponse{
				ID:        tag.ID.String(),
				Name:      tag.Name,
				CreatedAt: tag.CreatedAt.Format(time.RFC3339),
				UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
			},
		}
		if tag.DeletedAt != nil {
			deletedStr := tag.DeletedAt.Format(time.RFC3339)
			item.DeletedAt = &deletedStr
		}
		res = append(res, item)
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: result.Page,
		},
	)
}
