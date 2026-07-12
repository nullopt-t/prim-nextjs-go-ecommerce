package category

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
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
	ParentID *uuid.UUID `json:"parent_id" binding:"uuid" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
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
	req := &CreateCategoryRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	userID, _ := c.Get("userID")
	in := &CreateCategoryInput{
		Name:      req.Name,
		ParentID:  req.ParentID,
		CreatedBy: userID.(uuid.UUID),
		UpdatedBy: userID.(uuid.UUID),
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
				CreatedBy: category.CreatedBy,
				UpdatedBy: category.UpdatedBy,
				CreatedAt: category.CreatedAt.Format(time.RFC3339),
				UpdatedAt: category.UpdatedAt.Format(time.RFC3339),
			},
		},
	)
}

type CategoryIDURIParam struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

type CategoryResponse struct {
	ID        uuid.UUID  `json:"id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	Name      string     `json:"name,omitempty" example:"Electronic"`
	ParentID  *uuid.UUID `json:"parent_id" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	CreatedBy uuid.UUID  `json:"created_by,omitempty" exmaple:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	UpdatedBy uuid.UUID  `json:"updated_by,omitempty" exmaple:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
	CreatedAt string     `json:"created_at,omitempty" example:"2026-06-30T15:47:19Z"`
	UpdatedAt string     `json:"updated_at,omitempty" example:"2026-06-30T15:47:19Z"`
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
	param := &CategoryIDURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	category, err := ch.cservice.GetCategoryByID(
		ctx,
		param.ID,
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
				ParentID:  category.ParentID,
				CreatedBy: category.CreatedBy,
				UpdatedBy: category.UpdatedBy,
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
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]CategoryResponse,meta=api.Page}
// @Router /categories [get]
func (ch *CategoryHandler) ListCategories(c *gin.Context) {
	q := &api.ListQuery{}
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
		q,
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
// @Router /categories [get]
func (ch *CategoryHandler) ListAdminCategories(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
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
		q,
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
			ParentID:  c.ParentID,
			CreatedBy: c.CreatedBy,
			UpdatedBy: c.UpdatedBy,
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

type UpdateCategoryRequest struct {
	Name     *string    `json:"name" example:"electronic"`
	ParentID *uuid.UUID `json:"parent_id" binding:"uuid" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
}

type UpdateCategoryResponse struct {
	Name     *string    `json:"name,omitempty" example:"electronic"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" example:"c8ccec1c-ded5-4380-9f78-a1d4eb3d4f28"`
}

// UpdateCategory godoc
// @Summary update a category details
// @Description update a category details
// @Tags Category
// @Accept json
// @Produce json
// @Param id path CategoryIDURIParam true "category id"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=Category}
// @Router /admin/categories [get]
func (ch *CategoryHandler) UpdateCategory(c *gin.Context) {
	param := &CategoryIDURIParam{}
	if err := c.ShouldBindJSON(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	body := &UpdateCategoryRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	userID, _ := c.Get("userID")
	in := &UpdateCategoryInput{
		Name:      body.Name,
		ParentID:  body.ParentID,
		UpdatedBy: userID.(uuid.UUID),
	}

	ctx := c.Request.Context()
	err := ch.cservice.UpdateCategory(
		ctx,
		param.ID,
		in,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK,
		api.DataResponse{
			Data: UpdateCategoryResponse{
				Name:     body.Name,
				ParentID: body.ParentID,
			},
		},
	)
}
