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
// @Summary Create a product category
// @Description Creates a new product category in the taxonomy tree. Can optionally be assigned a parent category ID for hierarchical nesting.
// @Tags Categories
// @Accept json
// @Produce json
// @Param body body CreateCategoryRequest true "Category name and optional parent category UUID"
// @Failure 400 {object} api.BadRequestErrorResponse "Validation error or referenced parent category does not exist"
// @Failure 409 {object} api.ConflictErrorResponse "A category with this name already exists"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 201 {object} api.DataResponse{data=AdminCategoryResponse} "Created category details"
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
// @Summary Get category details by ID
// @Description Retrieves product category details by its UUID.
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category UUID" format(uuid)
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid UUID format"
// @Failure 404 {object} api.NotFoundErrorResponse "Category not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.DataResponse{data=AdminCategoryResponse} "Category details"
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
// @Summary List active product categories
// @Description Returns a paginated list of active product categories for customer browsing and navigation.
// @Tags Categories
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "Pagination, search query, and sorting parameters"
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid query parameters"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.PaginatedResponse{data=[]PublicCategoryResponse,meta=api.Page} "Paginated list of active categories"
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
// @Summary List all categories including soft-deleted ones (Admin)
// @Description Returns a paginated list of all product categories including soft-deleted records for administrator taxonomy management.
// @Tags Categories
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "Pagination, search query, and sorting parameters"
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid query parameters"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.PaginatedResponse{data=[]AdminCategoryResponse,meta=api.Page} "Paginated list of all categories including deleted"
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
// @Summary Update category details
// @Description Updates specific fields of an existing category (name or parent category ID). Guards against self-referencing and circular tree dependencies.
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category UUID" format(uuid)
// @Param body body UpdateCategoryRequest true "Fields to update (name, parent_id)"
// @Failure 400 {object} api.BadRequestErrorResponse "Validation error, circular hierarchy, or parent not found"
// @Failure 404 {object} api.NotFoundErrorResponse "Category not found"
// @Failure 409 {object} api.ConflictErrorResponse "A category with updated name already exists"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.MessageResponse "Update confirmation message"
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
