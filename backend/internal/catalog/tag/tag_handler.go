package tag

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
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
//
//	@Summary		Create a product tag
//	@Description	Adds a new product classification tag (e.g. 'black-friday', 'best-seller'). Tag names must be unique.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			data	body		CreateTagRequest					true	"Tag name payload"
//	@Failure		400		{object}	api.BadRequestErrorResponse			"Validation error or missing tag name"
//	@Failure		409		{object}	api.ConflictErrorResponse			"A tag with this name already exists"
//	@Failure		500		{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Success		201		{object}	api.DataResponse{data=TagResponse}	"Created tag details"
//	@Router			/admin/tags [post]
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
//
//	@Summary		Get tag details by ID
//	@Description	Retrieves active product tag details by its UUID.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string								true	"Tag UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse			"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse			"Tag not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=TagResponse}	"Tag details"
//	@Router			/admin/tags/{id} [get]
func (th *TagHandler) GetTagByID(c *gin.Context) {
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			apierr.ErrInvalidUUID().WithFields(
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
//
//	@Summary		Update tag details
//	@Description	Updates the name of an existing product tag. Tag names must be unique.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Tag UUID"	format(uuid)
//	@Param			input	body		UpdateTagRequest				true	"Fields to update (name)"
//	@Failure		400		{object}	api.BadRequestErrorResponse		"Validation error or invalid UUID format"
//	@Failure		404		{object}	api.NotFoundErrorResponse		"Tag not found"
//	@Failure		409		{object}	api.ConflictErrorResponse		"A tag with updated name already exists"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200		{object}	api.MessageResponse				"Update confirmation message"
//	@Router			/admin/tags/{id} [patch]
func (th *TagHandler) UpdateTagByID(c *gin.Context) {
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			apierr.ErrInvalidUUID().WithFields(
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

	err = th.tservice.UpdateTagByID(
		c.Request.Context(),
		tagID,
		&UpdateTagInput{
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
//
//	@Summary		Soft-delete a tag
//	@Description	Marks an active product tag as soft-deleted (`deleted_at = NOW()`), removing it from tag selections.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string							true	"Tag UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Tag not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.MessageResponse				"Deletion confirmation message"
//	@Router			/admin/tags/{id} [delete]
func (th *TagHandler) DeleteTagByID(c *gin.Context) {
	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(
			apierr.ErrInvalidUUID().WithFields(
				api.FieldError{
					Field:   "tag_id",
					Message: "invalid tag UUID format",
				},
			),
		)
		return
	}

	if err := th.tservice.DeleteTagByID(c.Request.Context(), tagID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListTags godoc
//
//	@Summary		List active product tags
//	@Description	Returns a paginated list of active product tags for customer storefront filtering.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			q	query		pagination.ListQuery											true	"Pagination, search query, and sorting parameters"
//	@Failure		400	{object}	api.BadRequestErrorResponse										"Invalid query parameters"
//	@Failure		500	{object}	api.InternalServerErrorResponse									"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]TagResponse,meta=pagination.Page}	"Paginated list of active tags"
//	@Router			/tags [get]
func (th *TagHandler) ListTags(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(pagination.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := th.tservice.ListTags(c.Request.Context(), ListTagsInput{
		Query:          q,
		IncludeDeleted: false,
	})
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
//
//	@Summary		List all tags including soft-deleted ones (Admin)
//	@Description	Returns a paginated list of all product tags including soft-deleted records for administrator management.
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			q	query		pagination.ListQuery												true	"Pagination, search query, and sorting parameters"
//	@Failure		400	{object}	api.BadRequestErrorResponse											"Invalid query parameters"
//	@Failure		500	{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]AdminTagResponse,meta=pagination.Page}	"Paginated list of all tags including deleted"
//	@Router			/admin/tags [get]
func (th *TagHandler) AdminListTags(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(pagination.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := th.tservice.AdminList(c.Request.Context(), q)
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
