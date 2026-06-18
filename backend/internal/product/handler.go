package product

import (
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ProductURIParam struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type CategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	Total      int                `json:"total"`
}

type ProductResponse struct {
	ID               uuid.UUID `json:"id,omitempty"`
	Title            string    `json:"title,omitempty"`
	ShortDescription string    `json:"short_description,omitempty"`
	Description      string    `json:"description,omitempty"`
	Sku              string    `json:"sku,omitempty"`
	Slug             string    `json:"slug,omitempty"`
	Status           string    `json:"status,omitempty"`
	Price            int64     `json:"price,omitempty"`
	Currency         string    `json:"currency,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitzero"`
	UpdatedAt        time.Time `json:"updated_at,omitzero"`
}

type CreateProductRequest struct {
	Title            string `json:"title" binding:"required"`
	ShortDescription string `json:"short_description" binding:"requried"`
	Description      string `json:"description" binding:"requried"`
	Slug             string `json:"slug" binding:"required"`
	SKU              string `json:"sku" binding:"requried"`
	Status           string `json:"status" binding:"required"`
	Price            int64  `json:"price" binding:"required"`
	Currency         string `json:"currency" binding:"requried"`
}

type Handler struct {
	service *ProductService
}

func NewHandler(s *ProductService) ProductHandler {
	return &Handler{s}
}

func (h *Handler) handleValidationError(c *gin.Context, err error) {
	if ve, ok := err.(validator.ValidationErrors); ok && ve != nil {
		fieldErrors := make([]api.FieldError, 0, len(ve))
		for _, e := range ve {
			fieldErrors = append(fieldErrors, api.FieldError{
				Field: e.Field(),
				Tags:  e.Tag(),
			})
		}
		_ = c.Error(security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"bad request data",
			err,
		).WithFields(fieldErrors))
		return
	}
	_ = c.Error(
		security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"bad request data",
			err,
		),
	)
}

// GetAllProducts godoc
// @Summary Get all products
// @Description Get all products
// @Tags Products
// @Accept json
// @Produce json
// @Param query query api.PageQuery true "Pagination query"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /products [get]
func (h *Handler) GetAllProducts(c *gin.Context) {
	var query api.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	result, page, err := h.service.GetAll(
		c.Request.Context(),
		query,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var products []*ProductResponse
	for _, p := range result {
		products = append(
			products,
			&ProductResponse{
				ID:        p.ID,
				Title:     p.Title,
				Price:     p.Price,
				CreatedAt: p.CreatedAt,
				UpdatedAt: p.UpdatedAt,
			},
		)
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: products,
			Meta: page,
		},
	)
}

// GetProductByID godoc
// @Summary Get a product by id
// @Description Get a product by passing it's id
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 404 {object} api.ErrorResponse
// @Router /products/{id} [get]
func (h *Handler) GetProductByID(c *gin.Context) {
	var params ProductURIParam
	if err := c.ShouldBindUri(&params); err != nil {
		h.handleValidationError(c, err)
		return
	}

	p, err := h.service.GetByID(
		c.Request.Context(),
		params.ID,
	)
	if err != nil {
		return
	}

	res := &ProductResponse{
		ID:               p.ID,
		Title:            p.Title,
		ShortDescription: p.ShortDescription,
		Description:      p.Description,
		Sku:              p.SKU,
		Slug:             p.Slug,
		Status:           p.Status,
		Price:            p.Price,
		Currency:         p.Currency,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}

	c.JSON(
		http.StatusFound,
		api.SuccessResponse{
			Data: res,
		},
	)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product Data"
// @Success 201 {object} api.SuccessResponse{data=ProductResponse}
// @Failure 400 {object} api.BadReqResponse
// @Failure 404 {object} api.ErrorResponse
// @Router /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var p CreateProductRequest
	if err := c.ShouldBindJSON(&p); err != nil {
		h.handleValidationError(c, err)
		return
	}

	id, err := h.service.Create(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &ProductResponse{
		ID: id,
	}

	c.JSON(http.StatusCreated, api.SuccessResponse{
		Data: res,
	})
}
