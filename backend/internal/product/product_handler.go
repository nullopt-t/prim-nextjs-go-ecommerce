package product

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/product/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/types"
)

type ProductHandler struct {
	service *ProductService
}

func NewHandler(s *ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

// --- DTOs: Products ---

type CreateProductRequest struct {
	BrandID     *string  `json:"brand_id,omitempty" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	CategoryID  string   `json:"category_id" binding:"required,uuid" example:"356cbaee-4700-4af5-ac9c-61aeeafd541c"`
	Title       string   `json:"title" binding:"required" example:"Wireless Noise-Canceling Headphones"`
	Description string   `json:"description" binding:"required" example:"Premium over-ear Bluetooth headphones with active noise cancellation."`
	Highlights  []string `json:"highlights,omitempty"`
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
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductResponse struct {
	ID          string   `json:"id" example:"96c4e462-ed4a-4fec-9115-47cbf12206a7"`
	BrandID     *string  `json:"brand_id,omitempty" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	CategoryID  string   `json:"category_id" example:"356cbaee-4700-4af5-ac9c-61aeeafd541c"`
	PublicID    string   `json:"public_id" example:"prod_abc123"`
	Title       string   `json:"title" example:"Wireless Headphones"`
	Description string   `json:"description"`
	Highlights  []string `json:"highlights,omitempty"`
	Status      string   `json:"status" example:"draft"`
	CreatedAt   string   `json:"created_at" example:"2026-08-02T16:00:00Z"`
	UpdatedAt   string   `json:"updated_at" example:"2026-08-02T16:00:00Z"`
}

type ProductDetailsResponse struct {
	ProductResponse
	Brand *ProductBrandSummary `json:"brand,omitempty"`
}

// --- DTOs: Media ---

type StorageObjectResponse struct {
	ID          string `json:"id"`
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	ContentType string `json:"content_type,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
	PublicURL   string `json:"public_url,omitempty"`
}

type ProductMediaResponse struct {
	ID              string                 `json:"id" example:"8f123456-e89b-12d3-a456-426614174000"`
	ProductID       string                 `json:"product_id" example:"96c4e462-ed4a-4fec-9115-47cbf12206a7"`
	StorageObjectID string                 `json:"storage_object_id" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	MediaType       string                 `json:"media_type" example:"image"`
	SortOrder       int                    `json:"sort_order" example:"0"`
	Object          *StorageObjectResponse `json:"object,omitempty"`
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
	var brandIDStr *string
	if p.BrandID != nil {
		idStr := p.BrandID.String()
		brandIDStr = &idStr
	}

	highlights := p.Highlights
	if highlights == nil {
		highlights = make([]string, 0)
	}

	return ProductResponse{
		ID:          p.ID.String(),
		BrandID:     brandIDStr,
		CategoryID:  p.CategoryID.String(),
		PublicID:    p.PublicID,
		Title:       p.Title,
		Description: p.Description,
		Highlights:  highlights,
		Status:      p.Status.String(),
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func mapProductMediaResponse(m *model.ProductMedia) ProductMediaResponse {
	res := ProductMediaResponse{
		ID:              m.ID.String(),
		ProductID:       m.ProductID.String(),
		StorageObjectID: m.StorageObjectID.String(),
		MediaType:       m.MediaType.String(),
		SortOrder:       m.SortOrder,
	}

	if m.Object != nil {
		res.Object = &StorageObjectResponse{
			ID:          m.Object.ID.String(),
			Bucket:      m.Object.Bucket,
			Key:         m.Object.Key,
			ContentType: m.Object.ContentType,
			FileSize:    m.Object.FileSize,
			PublicURL:   m.Object.PublicURL,
		}
	}

	return res
}

// --- Product Handlers ---

// CreateProductAsDraft godoc
// @Summary create a new draft product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product Data"
// @Failure 400 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=ProductResponse}
// @Router /admin/products [post]
func (h *ProductHandler) CreateProductAsDraft(c *gin.Context) {
	var body CreateProductRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	categoryID, err := uuid.Parse(body.CategoryID)
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "category_id",
			Message: "invalid category UUID format",
		}))
		return
	}

	in := CreateProductInput{
		Title:       body.Title,
		Description: body.Description,
		CategoryID:  categoryID,
		Highlights:  body.Highlights,
	}

	if body.BrandID != nil {
		brandID, err := uuid.Parse(*body.BrandID)
		if err != nil {
			_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "brand_id",
				Message: "invalid brand UUID format",
			}))
			return
		}
		in.BrandID = types.Ptr(brandID)
	}

	product, err := h.service.CreateProductAsDraft(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{
		Data: mapProductResponse(product),
	})
}

// GetAllProducts godoc
// @Summary list active products
// @Tags Products
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]ProductListItem,meta=api.Page}
// @Router /products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(api.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	result, err := h.service.List(c.Request.Context(), q, false)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: result.Items,
		Meta: result.Page,
	})
}

// AdminGetAllProducts godoc
// @Summary list all products including soft-deleted ones (admin)
// @Tags Products
// @Produce json
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]ProductListItem,meta=api.Page}
// @Router /admin/products [get]
func (h *ProductHandler) AdminGetAllProducts(c *gin.Context) {
	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(api.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	result, err := h.service.List(c.Request.Context(), q, true)
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
// @Summary get product by id
// @Tags Products
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=ProductResponse}
// @Router /products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	product, err := h.service.GetByID(c.Request.Context(), productID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapProductResponse(product),
	})
}

// GetProductByPID godoc
// @Summary get product details by public ID
// @Tags Products
// @Produce json
// @Param pid path string true "Product Public ID"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=ProductDetailsResponse}
// @Router /products/p/:pid [get]
func (h *ProductHandler) GetProductByPID(c *gin.Context) {
	pid := c.Param("pid")

	details, err := h.service.GetByPID(c.Request.Context(), pid)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := ProductDetailsResponse{
		ProductResponse: mapProductResponse(details.Product),
	}

	if details.Brand != nil {
		res.Brand = &ProductBrandSummary{
			ID:   details.Brand.ID.String(),
			Name: details.Brand.Name,
		}
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}

// UpdateProduct godoc
// @Summary update product details
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body UpdateProductRequest true "Update payload"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id} [patch]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
		Title:       body.Title,
		Description: body.Description,
		Highlights:  body.Highlights,
	}

	if body.BrandID != nil {
		parsedBrand, err := uuid.Parse(*body.BrandID)
		if err != nil {
			_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
			_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
			_ = c.Error(security.NewSecureError(http.StatusBadRequest).
				WithCode("INVALID_STATUS").
				WithMessage("invalid publication status"))
			return
		}
		input.Status = &ps
	}

	if err := h.service.UpdateProduct(c.Request.Context(), productID, input); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "updated successfully",
	})
}

// PublishProduct godoc
// @Summary publish a product
// @Tags Products
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/publish [post]
func (h *ProductHandler) PublishProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
// @Summary archive a product
// @Tags Products
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/archive [post]
func (h *ProductHandler) ArchiveProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
// @Summary soft-delete a product
// @Tags Products
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id} [delete]
func (h *ProductHandler) SoftDeleteProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	if err := h.service.SoftDeleteProduct(c.Request.Context(), productID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "deleted successfully",
	})
}

// --- Product Media Handlers ---

// UploadProductMedia godoc
// @Summary upload media file for a product
// @Tags Product Media
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param file formData file true "Media File"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=ProductMediaResponse}
// @Router /admin/products/{id}/media [post]
func (h *ProductHandler) UploadProductMedia(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	mfile, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(security.NewSecureError(http.StatusBadRequest).
			WithCode("FILE_REQUIRED").
			WithMessage("file is required"))
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

// GetProductMedia godoc
// @Summary list all media for a product
// @Tags Product Media
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=[]ProductMediaResponse}
// @Router /products/{id}/media [get]
func (h *ProductHandler) GetProductMedia(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	mediaList, err := h.service.GetProductMedia(c.Request.Context(), productID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]ProductMediaResponse, 0, len(mediaList))
	for _, m := range mediaList {
		res = append(res, mapProductMediaResponse(m))
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}

// DetachMedia godoc
// @Summary remove a media attachment from a product
// @Tags Product Media
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param media_id path string true "Media ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/media/{media_id} [delete]
func (h *ProductHandler) DetachMedia(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	mediaID, err := uuid.Parse(c.Param("media_id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "media_id",
			Message: "invalid media UUID format",
		}))
		return
	}

	if err := h.service.DetachMedia(c.Request.Context(), productID, mediaID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "detached successfully",
	})
}

// ReorderMedia godoc
// @Summary batch reorder media items for a product
// @Tags Product Media
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body ReorderMediaRequest true "ordered media IDs"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/media/reorder [patch]
func (h *ProductHandler) ReorderMedia(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
			_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=variant.VariantResponse}
// @Router /admin/products/{id}/variants [post]
func (h *ProductHandler) CreateProductVariant(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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

	v, err := h.service.CreateProductVariant(c.Request.Context(), productID, CreateProductVariantInput{
		Title:           body.Title,
		Price:           body.Price,
		CrossedOutPrice: body.CrossedOutPrice,
		Currency:        body.Currency,
		Attributes:      body.Attributes,
		IsDefault:       body.IsDefault,
	})
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

// GetProductVariants godoc
// @Summary list public variants for a product
// @Tags Product Variants
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]variant.VariantResponse,meta=api.Page}
// @Router /products/{id}/variants [get]
func (h *ProductHandler) GetProductVariants(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid product UUID format",
		}))
		return
	}

	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(api.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	result, err := h.service.GetProductVariants(c.Request.Context(), productID, q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]variant.VariantResponse, 0, len(result.Items))
	for _, v := range result.Items {
		attrs := v.Attributes
		if attrs == nil {
			attrs = make(map[string]any)
		}
		res = append(res, variant.VariantResponse{
			ID:              v.ID.String(),
			ProductID:       v.ProductID.String(),
			Title:           v.Title,
			Price:           v.Price,
			CrossedOutPrice: v.CrossedOutPrice,
			Currency:        v.Currency,
			Attributes:      attrs,
			IsDefault:       v.IsDefault,
			CreatedAt:       v.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       v.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: res,
		Meta: result.Page,
	})
}

// SetDefaultVariant godoc
// @Summary set a variant as the default for a product
// @Tags Product Variants
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)" format(uuid)
// @Param body body SetDefaultVariantRequest true "Variant to set as default"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/variants/default [post]
func (h *ProductHandler) SetDefaultVariant(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/tags [put]
func (h *ProductHandler) PutProductTags(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
			_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/products/{id}/categories [put]
func (h *ProductHandler) PutProductCategories(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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
		_ = c.Error(security.ErrInvalidUUID().WithFields(api.FieldError{
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

