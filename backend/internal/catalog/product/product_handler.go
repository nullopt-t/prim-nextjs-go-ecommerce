package product

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
)

type ProductHandler struct {
	service *ProductService
}

func NewHandler(s *ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

type CreateProductRequest struct {
	BrandID     *string `json:"brand_id,omitempty" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	CategoryID  string  `json:"category_id" binding:"required,uuid" example:"356cbaee-4700-4af5-ac9c-61aeeafd541c"`
	Title       string  `json:"title" binding:"required" example:"Wireless Noise-Canceling Headphones"`
	Description string  `json:"description" binding:"required" example:"Premium over-ear Bluetooth headphones with active noise cancellation."`
}

type UpdateProductRequest struct {
	BrandID     *string  `json:"brand_id,omitempty"`
	CategoryID  *string  `json:"category_id,omitempty"`
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Highlights  []string `json:"highlights,omitempty"`
	Status      *string  `json:"status,omitempty"`
}

type ProductBrandSummary struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Link *string `json:"link,omitempty"`
}

type ProductCategorySummary struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Slug *string `json:"slug,omitempty"`
}

type ProductTagSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductVariantResponse struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Price           *int64                 `json:"price,omitempty"`
	CrossedOutPrice *int64                 `json:"crossed_out_price,omitempty"`
	Currency        *string                `json:"currency,omitempty"`
	Attributes      map[string]any         `json:"attributes"`
	IsDefault       bool                   `json:"is_default"`
	Media           []ProductMediaResponse `json:"media"`
}

type ProductResponse struct {
	PublicID    string   `json:"id" example:"prod_abc123"`
	Title       string   `json:"title" example:"Wireless Headphones"`
	Description string   `json:"description"`
	Highlights  []string `json:"highlights"`
}

type ProductDetailsResponse struct {
	ProductResponse
	Brand    *ProductBrandSummary     `json:"brand,omitempty"`
	Media    []ProductMediaResponse   `json:"media"`
	Variants []ProductVariantResponse `json:"variants"`
	Tags     []ProductTagSummary      `json:"tags"`
}

type AdminProductDetailsResponse struct {
	ID          string                   `json:"id" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	PublicID    string                   `json:"public_id" example:"prod_abc123"`
	Title       string                   `json:"title" example:"Wireless Headphones"`
	Description string                   `json:"description"`
	Highlights  []string                 `json:"highlights"`
	Status      model.PublicationStatus  `json:"status" example:"draft"`
	BrandID     *string                  `json:"brand_id,omitempty"`
	Brand       *ProductBrandSummary     `json:"brand,omitempty"`
	CategoryID  string                   `json:"category_id"`
	Category    *ProductCategorySummary  `json:"category,omitempty"`
	Media       []ProductMediaResponse   `json:"media"`
	Variants    []ProductVariantResponse `json:"variants"`
	Tags        []ProductTagSummary      `json:"tags"`
	CreatedAt   string                   `json:"created_at" example:"2026-08-05T19:00:00Z"`
	UpdatedAt   string                   `json:"updated_at" example:"2026-08-05T19:00:00Z"`
	DeletedAt   *string                  `json:"deleted_at,omitempty" example:"2026-08-05T19:30:00Z"`
}

// --- DTOs: Media ---

type StorageObjectResponse struct {
	ContentType string `json:"content_type,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
	PublicURL   string `json:"public_url,omitempty"`
}

type ProductMediaResponse struct {
	ID        string                 `json:"id"`
	MediaType string                 `json:"media_type" example:"image"`
	SortOrder int                    `json:"sort_order" example:"0"`
	Object    *StorageObjectResponse `json:"object,omitempty"`
}

type ReorderMediaRequest struct {
	OrderedMediaIDs []string `json:"ordered_media_ids" binding:"required,gt=0,dive,uuid"`
}

// --- DTOs: Variants ---

type CreateProductVariantRequest struct {
	Title           string         `json:"title" binding:"required" example:"Black / XL"`
	Price           *int64         `json:"price,omitempty" example:"2999"`
	CrossedOutPrice *int64         `json:"crossed_out_price,omitempty" example:"3999"`
	Currency        *string        `json:"currency,omitempty" example:"USD"`
	Attributes      map[string]any `json:"attributes,omitempty"`
	IsDefault       bool           `json:"is_default" example:"false"`
}

type SetDefaultVariantRequest struct {
	VariantID string `json:"variant_id" binding:"required,uuid"`
}

type PutProductTagsRequest struct {
	TagIDs []string `json:"tag_ids" binding:"required,dive,uuid"`
}

type PutProductCategoryRequest struct {
	CategoryID string `json:"category_id" binding:"required,uuid"`
}

// --- Mappers ---

func mapProductResponse(p *model.Product) ProductResponse {
	highlights := p.Highlights
	if highlights == nil {
		highlights = make([]string, 0)
	}

	return ProductResponse{
		PublicID:    p.PublicID,
		Title:       p.Title,
		Description: p.Description,
		Highlights:  highlights,
	}
}

func mapAdminProductDetailsResponse(details *ProductDetails) AdminProductDetailsResponse {
	p := details.Product

	highlights := p.Highlights
	if highlights == nil {
		highlights = make([]string, 0)
	}

	res := AdminProductDetailsResponse{
		ID:          p.ID.String(),
		PublicID:    p.PublicID,
		Title:       p.Title,
		Description: p.Description,
		Highlights:  highlights,
		Status:      p.Status,
		CategoryID:  p.CategoryID.String(),
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		Media:       make([]ProductMediaResponse, 0),
		Variants:    make([]ProductVariantResponse, 0),
		Tags:        make([]ProductTagSummary, 0),
	}

	if p.BrandID != nil {
		brandIDStr := p.BrandID.String()
		res.BrandID = &brandIDStr
	}

	if p.DeletedAt != nil {
		delStr := p.DeletedAt.Format(time.RFC3339)
		res.DeletedAt = &delStr
	}

	if details.Brand != nil {
		res.Brand = &ProductBrandSummary{
			ID:   details.Brand.PublicID,
			Name: details.Brand.Name,
			Link: details.Brand.Link,
		}
	}

	if details.Category != nil {
		res.Category = &ProductCategorySummary{
			ID:   details.Category.PublicID,
			Name: details.Category.Name,
			Slug: &details.Category.Name,
		}
	}

	if len(details.Media) > 0 {
		mediaRes := make([]ProductMediaResponse, 0, len(details.Media))
		for _, m := range details.Media {
			mediaRes = append(mediaRes, mapProductMediaResponse(m))
		}
		res.Media = mediaRes
	}

	if len(details.Variants) > 0 {
		variantRes := make([]ProductVariantResponse, 0, len(details.Variants))
		for _, v := range details.Variants {
			attrs := v.Attributes
			if attrs == nil {
				attrs = make(map[string]any)
			}
			variantRes = append(variantRes, ProductVariantResponse{
				ID:              v.PublicID,
				Title:           v.Title,
				Price:           v.Price,
				CrossedOutPrice: v.CrossedOutPrice,
				Currency:        v.Currency,
				Attributes:      attrs,
				IsDefault:       v.IsDefault,
				Media:           make([]ProductMediaResponse, 0),
			})
		}
		res.Variants = variantRes
	}

	if len(details.Tags) > 0 {
		tagRes := make([]ProductTagSummary, 0, len(details.Tags))
		for _, t := range details.Tags {
			tagRes = append(tagRes, ProductTagSummary{
				ID:   t.PublicID,
				Name: t.Name,
			})
		}
		res.Tags = tagRes
	}

	return res
}

func mapProductMediaResponse(m *model.ProductMedia) ProductMediaResponse {
	res := ProductMediaResponse{
		ID:        m.PublicID,
		MediaType: m.MediaType.String(),
		SortOrder: m.SortOrder,
	}

	if m.Object != nil {
		res.Object = &StorageObjectResponse{
			ContentType: m.Object.ContentType,
			FileSize:    m.Object.FileSize,
			PublicURL:   m.Object.PublicURL,
		}
	}

	return res
}

// --- Product Handlers ---

// CreateProductAsDraft godoc
// @Summary Create a new draft product
// @Description Creates a new product in 'draft' status. Products remain in draft state until at least one variant is created and the product is explicitly published.
// @Tags Products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product title, description, category ID, optional brand ID and highlights"
// @Failure 400 {object} api.BadRequestErrorResponse "Validation error or invalid UUID reference"
// @Failure 409 {object} api.ConflictErrorResponse "Product with generated public ID already exists"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 201 {object} api.DataResponse{data=AdminProductDetailsResponse} "Created draft product details"
// @Router /admin/products [post]
func (h *ProductHandler) CreateProductAsDraft(c *gin.Context) {
	var body CreateProductRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	categoryID, err := uuid.Parse(body.CategoryID)
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "category_id",
			Message: "invalid category UUID format",
		}))
		return
	}

	in := CreateProductInput{
		Title:       strings.TrimSpace(body.Title),
		Description: strings.TrimSpace(body.Description),
		CategoryID:  categoryID,
	}

	if body.BrandID != nil {
		brandID, err := uuid.Parse(*body.BrandID)
		if err != nil {
			_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "brand_id",
				Message: "invalid brand UUID format",
			}))
			return
		}
		in.BrandID = &brandID
	}

	product, err := h.service.CreateProductAsDraft(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	details, err := h.service.GetAdminDetailsByID(c.Request.Context(), product.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{
		Data: mapAdminProductDetailsResponse(details),
	})
}

type ProductListItemResponse struct {
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Status      model.PublicationStatus `json:"status"`
	Brand       *ProductBrandSummary    `json:"brand,omitempty"`
	Category    *ProductCategorySummary `json:"category,omitempty"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

func mapProductListItemResponse(item *PublicProductListReadModel) ProductListItemResponse {
	res := ProductListItemResponse{
		ID:          item.PublicID,
		Title:       item.Title,
		Description: item.Description,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}

	if item.Brand != nil {
		res.Brand = &ProductBrandSummary{
			ID:   item.Brand.PublicID,
			Name: item.Brand.Name,
			Link: item.Brand.Link,
		}
	}

	if item.Category != nil {
		res.Category = &ProductCategorySummary{
			ID:   item.Category.PublicID,
			Name: item.Category.Name,
		}
	}

	return res
}

// GetAllProducts godoc
// @Summary List active published products
// @Description Returns a paginated list of active, published products for customer browsing. Soft-deleted and draft products are hidden.
// @Tags Products
// @Produce json
// @Param q query pagination.ListQuery true "Pagination, search query, and sorting parameters"
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid query parameters"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.PaginatedResponse{data=[]ProductListItemResponse,meta=pagination.Page} "Paginated list of active products"
// @Router /products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	result, err := h.service.List(c.Request.Context(), q, false)
	if err != nil {
		_ = c.Error(err)
		return
	}

	itemsRes := make([]ProductListItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		itemsRes = append(itemsRes, mapProductListItemResponse(item))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: itemsRes,
		Meta: result.Page,
	})
}

// AdminGetAllProducts godoc
// @Summary List all products for management (Admin)
// @Description Returns a paginated list of all products including draft, published, archived, and soft-deleted records for administrator catalog management.
// @Tags Products
// @Produce json
// @Param q query pagination.ListQuery true "Pagination, search query, and sorting parameters"
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid query parameters"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.PaginatedResponse{data=[]model.Product,meta=pagination.Page} "Paginated list of all products including deleted"
// @Router /admin/products [get]
func (h *ProductHandler) AdminGetAllProducts(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	result, err := h.service.AdminList(c.Request.Context(), q, true)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: result.Items,
		Meta: result.Page,
	})
}

// GetProductByID godoc
// @Summary Get product by internal UUID
// @Description Retrieves full product details by its unique internal UUID.
// @Tags Products
// @Produce json
// @Param id path string true "Internal Product UUID" format(uuid)
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid UUID format"
// @Failure 404 {object} api.NotFoundErrorResponse "Product not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.DataResponse{data=AdminProductDetailsResponse} "Product details"
// @Router /admin/products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	details, err := h.service.GetAdminDetailsByID(c.Request.Context(), productID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapAdminProductDetailsResponse(details),
	})
}

// GetProductByPublicID godoc
// @Summary Get product details by public ID
// @Description Retrieves full public-facing product details including brand information by its human-readable public ID.
// @Tags Products
// @Produce json
// @Param id path string true "Product Public ID (e.g. prod_01h8x9a...)"
// @Failure 400 {object} api.BadRequestErrorResponse "Public ID is required"
// @Failure 404 {object} api.NotFoundErrorResponse "Product not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.DataResponse{data=ProductDetailsResponse} "Product and brand details"
// @Router /products/{id} [get]
func (h *ProductHandler) GetProductByPublicID(c *gin.Context) {
	publicID := strings.TrimSpace(c.Param("id"))
	if publicID == "" {
		_ = c.Error(apierr.ErrBadRequest("Public ID is required").
			WithCode(apierr.CodeInvalidInput))
		return
	}

	details, err := h.service.GetByPublicID(c.Request.Context(), publicID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := ProductDetailsResponse{
		ProductResponse: mapProductResponse(details.Product),
		Media:           make([]ProductMediaResponse, 0),
		Variants:        make([]ProductVariantResponse, 0),
		Tags:            make([]ProductTagSummary, 0),
	}

	if details.Brand != nil {
		res.Brand = &ProductBrandSummary{
			ID:   details.Brand.PublicID,
			Name: details.Brand.Name,
			Link: details.Brand.Link,
		}
	}

	if len(details.Media) > 0 {
		mediaRes := make([]ProductMediaResponse, 0, len(details.Media))
		for _, m := range details.Media {
			mediaRes = append(mediaRes, mapProductMediaResponse(m))
		}
		res.Media = mediaRes
	}

	if len(details.Variants) > 0 {
		variantRes := make([]ProductVariantResponse, 0, len(details.Variants))
		for _, v := range details.Variants {
			attrs := v.Attributes
			if attrs == nil {
				attrs = make(map[string]any)
			}
			variantRes = append(variantRes, ProductVariantResponse{
				ID:              v.PublicID,
				Title:           v.Title,
				Price:           v.Price,
				CrossedOutPrice: v.CrossedOutPrice,
				Currency:        v.Currency,
				Attributes:      attrs,
				IsDefault:       v.IsDefault,
				Media:           make([]ProductMediaResponse, 0),
			})
		}
		res.Variants = variantRes
	}

	if len(details.Tags) > 0 {
		tagRes := make([]ProductTagSummary, 0, len(details.Tags))
		for _, t := range details.Tags {
			tagRes = append(tagRes, ProductTagSummary{
				ID:   t.PublicID,
				Name: t.Name,
			})
		}
		res.Tags = tagRes
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}

// UpdateProduct godoc
// @Summary Update product attributes
// @Description Updates specific fields of an existing product such as title, description, category, brand, or publication status.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product UUID" format(uuid)
// @Param body body UpdateProductRequest true "Fields to update (title, description, brand_id, category_id, status, highlights)"
// @Failure 400 {object} api.BadRequestErrorResponse "Validation error or invalid UUID format"
// @Failure 404 {object} api.NotFoundErrorResponse "Product, brand, or category not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.DataResponse{data=AdminProductDetailsResponse} "Updated product details"
// @Router /admin/products/{id} [patch]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	var body UpdateProductRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	input := UpdateProductInput{
		Highlights: body.Highlights,
	}

	if body.Title != nil {
		cleanTitle := strings.TrimSpace(*body.Title)
		if cleanTitle == "" {
			_ = c.Error(apierr.ErrBadRequest("Title cannot be empty").WithCode(apierr.CodeValidationFailed))
			return
		}
		input.Title = &cleanTitle
	}

	if body.Description != nil {
		cleanDesc := strings.TrimSpace(*body.Description)
		input.Description = &cleanDesc
	}

	if body.BrandID != nil {
		parsedBrand, err := uuid.Parse(*body.BrandID)
		if err != nil {
			_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "brand_id",
				Message: "invalid brand UUID format",
			}))
			return
		}
		input.BrandID = &parsedBrand
	}

	if body.CategoryID != nil {
		parsedCat, err := uuid.Parse(*body.CategoryID)
		if err != nil {
			_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "category_id",
				Message: "invalid category UUID format",
			}))
			return
		}
		input.CategoryID = &parsedCat
	}

	if body.Status != nil {
		ps, err := model.ParsePublicationStatus(*body.Status)
		if err != nil {
			_ = c.Error(apierr.ErrBadRequest("invalid publication status").WithCode(apierr.CodeValidationFailed))
			return
		}
		input.Status = &ps
	}

	if err := h.service.UpdateProduct(c.Request.Context(), productID, input); err != nil {
		_ = c.Error(err)
		return
	}

	details, err := h.service.GetAdminDetailsByID(c.Request.Context(), productID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapAdminProductDetailsResponse(details),
	})
}

// PublishProduct godoc
// @Summary Publish a draft product
// @Description Transitions a product status to 'published' so it becomes visible to customers. Requires the product to have at least one active variant.
// @Tags Products
// @Produce json
// @Param id path string true "Product UUID" format(uuid)
// @Failure 400 {object} api.BadRequestErrorResponse "Product lacks active variants or invalid UUID"
// @Failure 404 {object} api.NotFoundErrorResponse "Product not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.MessageResponse "Publication confirmation message"
// @Router /admin/products/{id}/publish [post]
func (h *ProductHandler) PublishProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	if err := h.service.PublishProduct(c.Request.Context(), productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "product published successfully",
	})
}

// ArchiveProduct godoc
// @Summary Archive a product
// @Description Transitions a product status to 'archived', hiding it from public store front listings while retaining all historical sales and inventory data.
// @Tags Products
// @Produce json
// @Param id path string true "Product UUID" format(uuid)
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid UUID format"
// @Failure 404 {object} api.NotFoundErrorResponse "Product not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.MessageResponse "Archival confirmation message"
// @Router /admin/products/{id}/archive [post]
func (h *ProductHandler) ArchiveProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	if err := h.service.ArchiveProduct(c.Request.Context(), productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "product archived successfully",
	})
}

// SoftDeleteProduct godoc
// @Summary Soft-delete a product
// @Description Marks a product as soft-deleted (`deleted_at = NOW()`), removing it from active administrative and customer listings.
// @Tags Products
// @Produce json
// @Param id path string true "Product UUID" format(uuid)
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid UUID format"
// @Failure 404 {object} api.NotFoundErrorResponse "Product not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.MessageResponse "Deletion confirmation message"
// @Router /admin/products/{id} [delete]
func (h *ProductHandler) SoftDeleteProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	if err := h.service.SoftDeleteProduct(c.Request.Context(), productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Product Media Handlers ---

// UploadProductMedia godoc
// @Summary upload media file for a product
// @Tags Product Media
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param file formData file true "Media File"
// @Failure 400 {object} api.BadRequestErrorResponse
// @Failure 500 {object} api.InternalServerErrorResponse
// @Success 201 {object} api.DataResponse{data=ProductMediaResponse}
// @Router /admin/products/{id}/media [post]
func (h *ProductHandler) UploadProductMedia(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	mfile, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("file is required").
			WithCode(apierr.CodeFileRequired))
		return
	}

	media, err := h.service.UploadProductMedia(c.Request.Context(), productID, mfile)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{
		Data: mapProductMediaResponse(media),
	})
}

// DetachMedia godoc
// @Summary remove a media attachment from a product
// @Tags Product Media
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param media_id path string true "Media ID (UUID)" format(uuid)
// @Failure 400 {object} api.BadRequestErrorResponse
// @Failure 404 {object} api.NotFoundErrorResponse
// @Failure 500 {object} api.InternalServerErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/media/{media_id} [delete]
func (h *ProductHandler) DetachMedia(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	mediaID, err := uuid.Parse(c.Param("media_id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "media_id",
			Message: "invalid media UUID format",
		}))
		return
	}

	if err := h.service.DetachMedia(c.Request.Context(), productID, mediaID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderMedia godoc
// @Summary batch reorder media items for a product
// @Tags Product Media
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body ReorderMediaRequest true "ordered media IDs"
// @Failure 400 {object} api.BadRequestErrorResponse
// @Failure 500 {object} api.InternalServerErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/media/reorder [patch]
func (h *ProductHandler) ReorderMedia(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	var body ReorderMediaRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	orderedIDs := make([]uuid.UUID, 0, len(body.OrderedMediaIDs))
	for _, idStr := range body.OrderedMediaIDs {
		parsed, err := uuid.Parse(idStr)
		if err != nil {
			_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "ordered_media_ids",
				Message: "invalid UUID in ordered list",
			}))
			return
		}
		orderedIDs = append(orderedIDs, parsed)
	}

	if err := h.service.ReorderMedia(c.Request.Context(), productID, orderedIDs); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "reordered successfully",
	})
}

// --- Cross-Domain Delegate Handlers ---

// CreateProductVariant godoc
// @Summary create a variant under a product
// @Tags Product Variants
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body CreateProductVariantRequest true "Variant Payload"
// @Failure 400 {object} api.BadRequestErrorResponse
// @Failure 500 {object} api.InternalServerErrorResponse
// @Success 201 {object} api.DataResponse{data=variant.VariantResponse}
// @Router /admin/products/{id}/variants [post]
func (h *ProductHandler) CreateProductVariant(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	var body CreateProductVariantRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	v, err := h.service.CreateProductVariant(c.Request.Context(), productID, CreateProductVariantInput(body))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{
		Data: variant.VariantResponse{
			ID:              v.ID.String(),
			ProductID:       v.ProductID.String(),
			Title:           v.Title,
			Price:           v.Price,
			CrossedOutPrice: v.CrossedOutPrice,
			Currency:        v.Currency,
			Attributes:      v.Attributes,
			IsDefault:       v.IsDefault,
			CreatedAt:       v.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       v.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// SetDefaultVariant godoc
// @Summary set a variant as the default for a product
// @Tags Product Variants
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body SetDefaultVariantRequest true "Variant to set as default"
// @Failure 400 {object} api.BadRequestErrorResponse
// @Failure 404 {object} api.NotFoundErrorResponse
// @Failure 500 {object} api.InternalServerErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/variants/default [post]
func (h *ProductHandler) SetDefaultVariant(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	var body SetDefaultVariantRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	variantID, err := uuid.Parse(body.VariantID)
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "variant_id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	if err := h.service.SetDefaultVariant(c.Request.Context(), productID, variantID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "default variant updated successfully",
	})
}

// PutProductTags godoc
// @Summary replace all tag assignments for a product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body PutProductTagsRequest true "Tag IDs to assign"
// @Failure 400 {object} api.BadRequestErrorResponse
// @Failure 500 {object} api.InternalServerErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/tags [put]
func (h *ProductHandler) PutProductTags(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	var body PutProductTagsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	tagIDs := make([]uuid.UUID, 0, len(body.TagIDs))
	for _, idStr := range body.TagIDs {
		parsed, err := uuid.Parse(idStr)
		if err != nil {
			_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "tag_ids",
				Message: "invalid UUID in tag list",
			}))
			return
		}
		tagIDs = append(tagIDs, parsed)
	}

	if err := h.service.ReplaceProductTags(c.Request.Context(), productID, tagIDs); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "product tags updated successfully",
	})
}

// PutProductCategories godoc
// @Summary update the category of a product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body PutProductCategoryRequest true "Category ID to assign"
// @Failure 400 {object} api.BadRequestErrorResponse
// @Failure 404 {object} api.NotFoundErrorResponse
// @Failure 500 {object} api.InternalServerErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/categories [put]
func (h *ProductHandler) PutProductCategories(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	var body PutProductCategoryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	categoryID, err := uuid.Parse(body.CategoryID)
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "category_id",
			Message: "invalid category UUID format",
		}))
		return
	}

	input := UpdateProductInput{CategoryID: &categoryID}
	if err := h.service.UpdateProduct(c.Request.Context(), productID, input); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "product category updated successfully",
	})
}
