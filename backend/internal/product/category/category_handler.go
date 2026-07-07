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
	Name     string `json:"name" binding:"required" example:"electronic"`
	ParentID string `json:"parent_id" binding:"omitempty,uuid" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
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
// @Router /categories [post]
func (ch *CategoryHandler) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ppID, err := uuid.Parse(req.ParentID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := CreateCategoryInput{
		Name:     req.Name,
		ParentID: &ppID,
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
				Slug:      category.Slug,
				ParentID:  category.ParentID,
				CreatedAt: category.CreatedAt.Format(time.RFC3339),
				UpdatedAt: category.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

type CategoryByIDURIParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type CategoryResponse struct {
	ID        uuid.UUID  `json:"id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name      string     `json:"name" example:"Electronic"`
	Slug      string     `json:"slug" example:"electronic"`
	ParentID  *uuid.UUID `json:"parent_id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	CreatedAt string     `json:"created_at" example:"2026-06-30T15:47:19Z"`
	UpdatedAt string     `json:"updated_at" example:"2026-06-30T15:47:19Z"`
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
// @Router /categories/{id} [get]
func (ch *CategoryHandler) GetCategoryByID(c *gin.Context) {
	var params CategoryByIDURIParam
	if err := c.ShouldBindUri(&params); err != nil {
		validation.ValidationError(c, err)
		return
	}

	cuuid, err := uuid.Parse(params.ID)
	if err != nil {
		_ = c.Error(
			security.NewSecureError(
				http.StatusBadRequest,
				security.CodeValidation,
				"invalid id format",
				err,
			),
		)
		return
	}

	ctx := c.Request.Context()
	category, err := ch.cservice.GetCategoryByID(
		ctx,
		cuuid,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: CategoryResponse{
				ID:        category.ID,
				Name:      category.Name,
				Slug:      category.Slug,
				ParentID:  category.ParentID,
				CreatedAt: category.CreatedAt.Format(time.RFC3339),
				UpdatedAt: category.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

type CategorySlugURIParam struct {
	Slug string `uri:"slug"`
}

// GetCategoryBySlug godoc
// @Summary get a category by slug
// @Description fetch a category by slug
// @Tags Category
// @Accept json
// @Produce json
// @Param slug path string true "Category slug(string)" format(string)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=CategoryResponse}
// @Router /categories/{slug} [get]
func (ch *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	var params CategorySlugURIParam
	if err := c.ShouldBindUri(&params); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	category, err := ch.cservice.GetCategoryBySlug(
		ctx,
		params.Slug,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: CategoryResponse{
				ID:        category.ID,
				Name:      category.Name,
				Slug:      category.Slug,
				ParentID:  category.ParentID,
				CreatedAt: category.CreatedAt.Format(time.RFC3339),
				UpdatedAt: category.UpdatedAt.Format(time.RFC3339),
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
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]CategoryResponse,meta=api.Page}
// @Router /categories [get]
func (ch *CategoryHandler) ListCategories(c *gin.Context) {
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
	categories, page, err := ch.cservice.List(
		ctx,
		&q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res = make([]*CategoryResponse, 0, len(categories))
	for _, c := range categories {
		res = append(res, &CategoryResponse{
			ID:        c.ID,
			Name:      c.Name,
			Slug:      c.Slug,
			ParentID:  c.ParentID,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
			UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
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
