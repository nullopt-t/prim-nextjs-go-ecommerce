package category

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
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

type CreateCategoryRequest struct {
	Name     string     `json:"name" binding:"required" example:"electronic"`
	Slug     *string    `json:"slug" example:"electronic"`
	ParentID *uuid.UUID `json:"parent_id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
}

type PublicCategoryListResponse struct {
	ID   uuid.UUID `json:"id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name string    `json:"name,omitempty" example:"Electronic"`
	Slug string    `json:"slug,omitempty" example:"electronic"`
}

type CategoryResponse struct {
	ID                uuid.UUID  `json:"id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name              string     `json:"name,omitempty" example:"Electronic"`
	Slug              string     `json:"slug,omitempty" example:"electronic"`
	ParentID          *uuid.UUID `json:"parent_id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	PublicationStatus string     `json:"publication_status,omitempty" example:"published"`
	CreatedAt         string     `json:"created_at,omitempty" example:"2026-06-30T15:47:19Z"`
	UpdatedAt         string     `json:"updated_at,omitempty" example:"2026-06-30T15:47:19Z"`
}

// CreateCategory godoc
// @Summary create a new category
// @Description create a new category
// @Tags Category
// @Accept json
// @Produce json
// @Param body body CreateCategoryRequest true "Category Data "
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=CategoryResponse}
// @Router /admin/categories [post]
func (ch *CategoryHandler) CreateCategory(c *gin.Context) {
	body := &CreateCategoryRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CreateCategoryInput{
		Name:     body.Name,
		Slug:     body.Slug,
		ParentID: body.ParentID,
	}

	ctx := c.Request.Context()
	category, err := ch.cservice.CreateCategory(
		ctx,
		in,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusCreated,
		api.DataResponse{
			Data: CategoryResponse{
				ID:        category.ID,
				Name:      category.Name,
				ParentID:  category.ParentID,
				CreatedAt: category.CreatedAt.Format(time.RFC3339),
				UpdatedAt: category.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

type CategoryIDURIParam struct {
	ID string `uri:"id"`
}

// GetCategoryById godoc
// @Summary get a category by id
// @Description fetch a category by id
// @Tags Category
// @Accept json
// @Produce json
// @Param id path string true "Category ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=CategoryResponse}
// @Router /admin/categories/{id} [get]
func (ch *CategoryHandler) GetCategoryByID(c *gin.Context) {
	param := &CategoryIDURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	categoryID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(security.SecureErrInvalidUUID(err))
		return
	}

	ctx := c.Request.Context()
	category, err := ch.cservice.GetCategoryByID(
		ctx,
		categoryID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: CategoryResponse{
				ID:                category.ID,
				Name:              category.Name,
				Slug:              category.Slug,
				ParentID:          category.ParentID,
				PublicationStatus: category.PublicationStatus.String(),
				CreatedAt:         category.CreatedAt.Format(time.RFC3339),
				UpdatedAt:         category.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

// ListCategories godco
// @Summary list all categories
// @Description list all categories
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

	q.ApplyDefaults(api.QueryOptions{
		DefaultPageSize: 10,
		MaxPageSize:     100,
	}).Parse()

	ctx := c.Request.Context()
	categories, page, err := ch.cservice.List(
		ctx,
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*PublicCategoryListResponse, 0, len(categories))
	for _, c := range categories {
		res = append(res, &PublicCategoryListResponse{
			ID:   c.ID,
			Name: c.Name,
			Slug: c.Slug,
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

// ListAdminCategories godco
// @Summary list all categories
// @Description list all categories
// @Tags Category
// @Accept json
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]CategoryResponse,meta=api.Page}
// @Router /admin/categories [get]
func (ch *CategoryHandler) ListAdminCategories(c *gin.Context) {
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
	categories, page, err := ch.cservice.AdminList(
		ctx,
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*CategoryResponse, 0, len(categories))
	for _, c := range categories {
		res = append(res, &CategoryResponse{
			ID:                c.ID,
			Name:              c.Name,
			Slug:              c.Slug,
			ParentID:          c.ParentID,
			PublicationStatus: c.PublicationStatus.String(),
			CreatedAt:         c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         c.UpdatedAt.Format(time.RFC3339),
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

type UpdateCategoryRequest struct {
	Name              *string    `json:"name" example:"electronic"`
	ParentID          *uuid.UUID `json:"parent_id"  example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	PublicationStatus *string    `json:"publication_status" example:"published"`
}

type UpdateCategoryResponse struct {
	Name              *string    `json:"name,omitempty" example:"electronic"`
	ParentID          *uuid.UUID `json:"parent_id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	PublicationStatus *string    `json:"publication_status,omitempty" example:"published"`
}

// UpdateCategory godoc
// @Summary update a category details
// @Description update a category details
// @Tags Category
// @Accept json
// @Produce json
// @Param id path CategoryIDURIParam true "category id"
// @Param body body UpdateCategoryRequest true "category details"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=[]CategoryResponse,meta=api.Page}
// @Router /admin/categories/{id} [patch]
func (ch *CategoryHandler) UpdateCategory(c *gin.Context) {
	param := &CategoryIDURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	categoryID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(security.SecureErrInvalidUUID(err))
		return
	}

	body := &UpdateCategoryRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if body.Name == nil && body.ParentID == nil && body.PublicationStatus == nil {
		c.JSON(http.StatusOK,
			api.DataResponse{
				Data: UpdateCategoryResponse{},
			},
		)
		return
	}

	in := &UpdateCategoryInput{
		Name:              body.Name,
		ParentID:          body.ParentID,
		PublicationStatus: body.PublicationStatus,
	}

	ctx := c.Request.Context()
	err = ch.cservice.UpdateCategory(
		ctx,
		categoryID,
		in,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK,
		api.DataResponse{
			Data: UpdateCategoryResponse{
				Name:              body.Name,
				ParentID:          body.ParentID,
				PublicationStatus: body.PublicationStatus,
			},
		},
	)
}
