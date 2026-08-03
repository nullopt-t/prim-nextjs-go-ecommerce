package category

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
)

type CategoryHandler struct {
	cservice *CategoryService
}

func NewHandler(
	s *CategoryService,
) *CategoryHandler {
	return &CategoryHandler{
		cservice: s,
	}
}

type PublicCategoryListResponse struct {
	ID   uuid.UUID `json:"id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name string    `json:"name,omitempty" example:"Electronic"`
	Slug string    `json:"slug,omitempty" example:"electronic"`
}

type PublicCategoryResponse struct {
	ID   uuid.UUID `json:"id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name string    `json:"name" example:"Electronics"`
}

type AdminCategoryResponse struct {
	ID        uuid.UUID  `json:"id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name      string     `json:"name" example:"Electronics"`
	CreatedAt string     `json:"created_at" example:"2026-06-30T15:47:19Z"`
	UpdatedAt string     `json:"updated_at" example:"2026-06-30T15:47:19Z"`
	DeletedAt *string    `json:"deleted_at,omitempty" example:"2026-07-01T10:00:00Z"`
}

func ToPublicCategoryResponse(c *model.ProductCategory) PublicCategoryResponse {
	return PublicCategoryResponse{
		ID:   c.ID,
		Name: c.Name,
	}
}

func ToAdminCategoryResponse(c *model.ProductCategory) AdminCategoryResponse {
	res := AdminCategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
	if c.DeletedAt != nil {
		deletedStr := c.DeletedAt.Format(time.RFC3339)
		res.DeletedAt = &deletedStr
	}
	return res
}

type CreateCategoryRequest struct {
	Name     string     `json:"name" binding:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// CreateCategory godoc
// @Summary create a new category
// @Description create a new category
// @Tags Category
// @Accept json
// @Produce json
// @Param body body CreateCategoryRequest true "Category Data"
// @Failure 400 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=AdminCategoryResponse}
// @Router /admin/categories [post]
func (ch *CategoryHandler) CreateCategory(c *gin.Context) {
	var body CreateCategoryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CreateCategoryInput{
		Name:     body.Name,
		ParentID: body.ParentID,
	}

	ctx := c.Request.Context()
	category, err := ch.cservice.CreateCategory(ctx, in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
		api.DataResponse{
			Data: ToAdminCategoryResponse(category),
		},
	)
}

// GetCategoryByID godoc
// @Summary get a category by id
// @Description fetch a category by id
// @Tags Category
// @Accept json
// @Produce json
// @Param id path string true "Category ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=AdminCategoryResponse}
// @Router /admin/categories/{id} [get]
func (ch *CategoryHandler) GetCategoryByID(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().
			WithFields(api.FieldError{
				Field:   "id",
				Message: "invalid category UUID format",
			}),
		)
		return
	}

	category, err := ch.cservice.GetCategoryByID(
		c.Request.Context(),
		categoryID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: ToAdminCategoryResponse(category),
		},
	)
}

// ListCategories godoc
// @Summary list all categories
// @Description list all public active categories
// @Tags Category
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]PublicCategoryListResponse,meta=api.Page}
// @Router /categories [get]
func (ch *CategoryHandler) ListCategories(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Consolidate defaults, bounds checks, and sort parsing
	q.Process(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := ch.cservice.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]PublicCategoryResponse, 0, len(result.Items))
	for _, cat := range result.Items {
		res = append(res, ToPublicCategoryResponse(cat))
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: result.Page,
		},
	)
}

// ListAdminCategories godoc
// @Summary list all categories (admin)
// @Description list all categories including soft-deleted ones for administration
// @Tags Category
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]AdminCategoryResponse,meta=api.Page}
// @Router /admin/categories [get]
func (ch *CategoryHandler) ListAdminCategories(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Consolidate defaults, bounds checks, and sort parsing
	q.Process(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	})

	result, err := ch.cservice.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]AdminCategoryResponse, 0, len(result.Items))
	for _, cat := range result.Items {
		res = append(res, ToAdminCategoryResponse(cat))
	}

	c.JSON(
		http.StatusOK,
		api.PaginatedResponse{
			Data: res,
			Meta: result.Page,
		},
	)
}

type UpdateCategoryRequest struct {
	Name     *string    `json:"name,omitempty" example:"Electronics"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
}

// UpdateCategory godoc
// @Summary update category details
// @Description update category details by id
// @Tags Category
// @Accept json
// @Produce json
// @Param id path string true "Category ID (UUID)" format(uuid)
// @Param body body UpdateCategoryRequest true "Category details"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=AdminCategoryResponse}
// @Router /admin/categories/{id} [patch]
func (ch *CategoryHandler) UpdateCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().
			WithFields(api.FieldError{
				Field:   "id",
				Message: "invalid category UUID format",
			}),
		)
		return
	}

	var body UpdateCategoryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &UpdateCategoryInput{
		Name:     body.Name,
		ParentID: body.ParentID,
	}

	ctx := c.Request.Context()
	if err := ch.cservice.UpdateCategory(ctx, categoryID, in); err != nil {
		_ = c.Error(err)
		return
	}

	// Retrieve fresh state from DB to guarantee accurate updated_at and sanitized values
	updatedCategory, err := ch.cservice.GetCategoryByID(ctx, categoryID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: ToAdminCategoryResponse(updatedCategory),
	})
}
