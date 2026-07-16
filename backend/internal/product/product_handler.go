package product

import (
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductURIParam struct {
	ID string `uri:"id" binding:"required,uuid"`
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

type CreateProductRequest struct {
	BrandID          *uuid.UUID `json:"brand_id" binding:"required,uuid"`
	Title            string     `json:"title" binding:"required"`
	ShortDescription string     `json:"short_description" binding:"requried"`
	Description      string     `json:"description" binding:"requried"`
	Slug             string     `json:"slug" binding:"required"`
	SKU              string     `json:"sku" binding:"requried"`
	Status           string     `json:"status" binding:"required"`
	Price            int64      `json:"price" binding:"required"`
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

	query.SetDefaults(nil)

	result, err := h.service.List(
		c.Request.Context(),
		query,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: result.Items,
			Meta: result.Page,
		},
	)
}

type ProductBrandResponse struct {
	ID   uuid.UUID `json:"id,omitempty"`
	Name string    `json:"name,omitempty"`
}

type ProductResponse struct {
	ID               uuid.UUID             `json:"id,omitempty"`
	BrandID          *uuid.UUID            `json:"brand_id,omitzero"`
	Title            string                `json:"title,omitempty"`
	ShortDescription string                `json:"short_description,omitempty"`
	Description      string                `json:"description,omitempty"`
	Slug             string                `json:"slug,omitempty"`
	Status           string                `json:"status,omitempty"`
	CreatedBy        uuid.UUID             `json:"created_by,omitzero"`
	UpdatedBy        uuid.UUID             `json:"updated_by,omitzero"`
	CreatedAt        string                `json:"created_at,omitempty"`
	UpdatedAt        string                `json:"updated_at,omitempty"`
	Brand            *ProductBrandResponse `json:"brand,omitempty"`
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

	productID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(
			security.NewSecureError(
				http.StatusBadRequest,
				security.CodeValidation,
				err.Error(),
				err,
			),
		)
		return
	}

	product, err := h.service.GetByID(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		return
	}

	res := &ProductResponse{
		ID:               product.ID,
		BrandID:          product.BrandID,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Slug:             product.Slug,
		Status:           product.Status.String(),
		CreatedAt:        product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        product.UpdatedAt.Format(time.RFC3339),
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

	productDetails, err := h.service.GetBySlug(
		c.Request.Context(),
		param.Slug,
	)
	if err != nil {
		return
	}

	var brand *ProductBrandResponse
	if productDetails.Brand != nil {
		brand = &ProductBrandResponse{
			ID:   productDetails.Brand.ID,
			Name: productDetails.Brand.Name,
		}
	}

	res := &ProductResponse{
		ID:               productDetails.Product.ID,
		Title:            productDetails.Product.Title,
		ShortDescription: productDetails.Product.ShortDescription,
		Description:      productDetails.Product.Description,
		Slug:             productDetails.Product.Slug,
		Status:           productDetails.Product.Status.String(),
		CreatedBy:        productDetails.Product.CreatedBy,
		UpdatedBy:        productDetails.Product.UpdatedBy,
		CreatedAt:        productDetails.Product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        productDetails.Product.UpdatedAt.Format(time.RFC3339),
		Brand:            brand,
	}

	c.JSON(
		http.StatusFound,
		api.SuccessResponse{
			Data: res,
		},
	)
}

// CreateProductAsDraft godoc
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
func (h *ProductHandler) CreateProductAsDraft(c *gin.Context) {
	req := &CreateProductRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := CreateProductInput{
		BrandID:          req.BrandID,
		Title:            req.Title,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		Slug:             req.Slug,
		Status:           model.ProductStatus(req.Status),
	}

	product, err := h.service.CreateProductAsDraft(
		c.Request.Context(),
		in,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	userID, _ := c.Get("userID")
	res := &ProductResponse{
		ID:               product.ID,
		BrandID:          product.BrandID,
		Title:            product.Title,
		ShortDescription: product.ShortDescription,
		Description:      product.Description,
		Slug:             product.Slug,
		Status:           product.Status.String(),
		CreatedBy:        userID.(uuid.UUID),
		UpdatedBy:        userID.(uuid.UUID),
		CreatedAt:        product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        product.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, api.SuccessResponse{
		Data: res,
	})
}

type ProductVariantResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	ProductID uuid.UUID `json:"product_id,omitempty"`
	SKU       string    `json:"sku,omitempty"`
	Price     int64     `json:"price,omitempty"`
	Currency  string    `json:"currency,omitempty"`
	CreatedAt string    `json:"created_at,omitempty"`
	UpdatedAt string    `json:"updated_at,omitempty"`
}

func (h *ProductHandler) GetProductVariants(c *gin.Context) {
	param := &ProductURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		_ = c.Error(err)
		return
	}

	productID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	variants, err := h.service.GetProductVariants(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var variantResponses []ProductVariantResponse
	for _, variant := range variants {
		variantResponses = append(variantResponses, ProductVariantResponse{
			ID:        variant.ID,
			ProductID: variant.ProductID,
			SKU:       variant.SKU,
			Price:     variant.Price,
			Currency:  variant.Currency,
			CreatedAt: variant.CreatedAt.Format(time.RFC3339),
			UpdatedAt: variant.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Data: variantResponses,
	})
}

type ProductCategoryResponse struct {
	ID        uuid.UUID  `json:"id,omitzero"`
	Name      string     `json:"name,omitempty"`
	ParentID  *uuid.UUID `json:"parent_id,omitzero"`
	CreatedAt string     `json:"created_at,omitempty"`
	UpdatedAt string     `json:"updated_at,omitempty"`
}

func (h *ProductHandler) GetProductCategories(c *gin.Context) {
	param := &ProductURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		_ = c.Error(err)
		return
	}

	productID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	categories, err := h.service.GetProductCategories(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res []ProductCategoryResponse
	for _, category := range categories {
		res = append(res, ProductCategoryResponse{
			ID:        category.ID,
			Name:      category.Name,
			ParentID:  category.ParentID,
			CreatedAt: category.CreatedAt.Format(time.RFC3339),
			UpdatedAt: category.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Data: res,
	})
}

type ProductTagResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt string    `json:"created_at,omitempty"`
	UpdatedAt string    `json:"updated_at,omitempty"`
}

func (h *ProductHandler) GetProductTags(c *gin.Context) {
	param := &ProductURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		_ = c.Error(err)
		return
	}

	productID, err := uuid.Parse(param.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	tags, err := h.service.GetProductTags(
		c.Request.Context(),
		productID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var res []ProductTagResponse
	for _, tag := range tags {
		res = append(res, ProductTagResponse{
			ID:        tag.ID,
			Name:      tag.Name,
			CreatedAt: tag.CreatedAt.Format(time.RFC3339),
			UpdatedAt: tag.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Data: res,
	})
}

func (h *ProductHandler) SetDefaultVariant(c *gin.Context) {
	param := &ProductURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	productID, err := uuid.Parse(param.ID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	var body struct {
		VariantID uuid.UUID `json:"variant_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if err := h.service.SetDefaultVariant(
		c.Request.Context(),
		productID,
		body.VariantID,
	); err != nil {
		validation.ValidationError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Message: "default variant set successfully",
	})
}

func (h *ProductHandler) PublishProduct(c *gin.Context) {
	param := &ProductURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	productID, err := uuid.Parse(param.ID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	if err := h.service.PublishProduct(
		c.Request.Context(),
		productID,
	); err != nil {
		validation.ValidationError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Message: "product published successfully",
	})
}

func (h *ProductHandler) ArchiveProduct(c *gin.Context) {
	param := &ProductURIParam{}
	if err := c.ShouldBindUri(param); err != nil {
		validation.ValidationError(c, err)
		return
	}

	productID, err := uuid.Parse(param.ID)
	if err != nil {
		validation.ValidationError(c, err)
		return
	}

	if err := h.service.ArchiveProduct(
		c.Request.Context(),
		productID,
	); err != nil {
		validation.ValidationError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Message: "product archived successfully",
	})
}
