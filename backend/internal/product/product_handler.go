package product

import (
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductURIParam struct {
	ID uuid.UUID `uri:"id" binding:"required,uuid"`
}

type ProductSlugParam struct {
	Slug string `uri:"slug" binding:"required"`
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
	BrandID          uuid.UUID `json:"brand_id,omitempty"`
	Title            string    `json:"title,omitempty"`
	ShortDescription string    `json:"short_description,omitempty"`
	Description      string    `json:"description,omitempty"`
	Sku              string    `json:"sku,omitempty"`
	ThumbnailURL     string    `json:"thumbnail_url,omitempty"`
	BrandName        string    `json:"brand_name,omitempty"`
	Slug             string    `json:"slug,omitempty"`
	Status           string    `json:"status,omitempty"`
	Price            int64     `json:"price,omitempty"`
	Currency         string    `json:"currency,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitzero"`
	UpdatedAt        time.Time `json:"updated_at,omitzero"`
}

type CreateProductRequest struct {
	BrandID          string `json:"brand_id" binding:"required,uuid"`
	Title            string `json:"title" binding:"required"`
	ShortDescription string `json:"short_description" binding:"requried"`
	Description      string `json:"description" binding:"requried"`
	Slug             string `json:"slug" binding:"required"`
	SKU              string `json:"sku" binding:"requried"`
	Status           string `json:"status" binding:"required"`
	Price            int64  `json:"price" binding:"required"`
}

type ProductHandler struct {
	service *ProductService
}

func NewHandler(
	s *ProductService,
) *ProductHandler {
	return &ProductHandler{s}
}

// GetAllProducts godoc
// @Summary Get all products
// @Description Get all products
// @Tags Products
// @Accept json
// @Produce json
// @Param query query api.ListQuery true "Pagination query"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	query := &api.ListQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	result, page, err := h.service.List(
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
				ID:               p.ID,
				Title:            p.Title,
				ShortDescription: p.ShortDescription,
				ThumbnailURL:     p.ThumbnailURL,
				Status:           string(p.Status),
				BrandName:        p.BrandName,
				Price:            p.Price,
				CreatedAt:        p.CreatedAt,
				UpdatedAt:        p.UpdatedAt,
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
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	param := &ProductURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	p, err := h.service.GetByID(
		c.Request.Context(),
		param.ID,
	)
	if err != nil {
		return
	}

	res := &ProductResponse{
		ID:               p.ID,
		BrandID:          p.BrandID,
		Title:            p.Title,
		ShortDescription: p.ShortDescription,
		Description:      p.Description,
		Slug:             p.Slug,
		Status:           string(p.Status),
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

// GetProductBySlug godoc
// @Summary Get a product by slug
// @Description Get a product by passing it's slug
// @Tags Products
// @Accept json
// @Produce json
// @Param slug path string true "Product Slug"
// @Success 200 {object} ProductResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 404 {object} api.ErrorResponse
// @Router /products/{slug} [get]
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	param := &ProductSlugParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	p, err := h.service.GetBySlug(
		c.Request.Context(),
		param.Slug,
	)
	if err != nil {
		return
	}

	res := &ProductResponse{
		ID:               p.ID,
		BrandID:          p.BrandID,
		Title:            p.Title,
		ShortDescription: p.ShortDescription,
		Description:      p.Description,
		Slug:             p.Slug,
		Status:           string(p.Status),
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
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	req := &CreateProductRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := CreateProductInput{
		BrandID:          uuid.MustParse(req.BrandID),
		Title:            req.Title,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		Slug:             req.Slug,
		Status:           model.ProductStatus(req.Status),
	}

	product, err := h.service.Create(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := &ProductResponse{
		ID:               product.ID,
		BrandID:          product.BrandID,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Slug:             product.Slug,
		Status:           string(product.Status),
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
	}

	c.JSON(http.StatusCreated, api.SuccessResponse{
		Data: res,
	})
}
