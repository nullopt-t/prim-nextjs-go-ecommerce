package variant

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/inventory"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
)

type VariantHandler struct {
	vservice *VariantService
}

func NewHandler(s *VariantService) *VariantHandler {
	return &VariantHandler{
		vservice: s,
	}
}

type CreateVariantRequest struct {
	// Optional custom SKU identifier
	SKU *string `json:"sku,omitempty" example:"MAC-PRO-16-BLK-18"`
	// Variant display title
	Title string `json:"title" binding:"required" example:"Space Black, 18GB RAM, 512GB SSD"`
	// Price in cents (e.g. 249900 = $2499.00)
	Price *int64 `json:"price,omitempty" example:"249900"`
	// Crossed-out/original MSRP price in cents
	CrossedOutPrice *int64 `json:"crossedOutPrice,omitempty" example:"299900"`
	// Currency ISO code
	Currency *string `json:"currency,omitempty" example:"USD"`
	// Key-value attribute specifications (e.g. {"color":"Black","ram":"18GB"})
	Attributes map[string]any `json:"attributes,omitempty"`
	// True if this is the default selected variant on storefronts
	IsDefault bool `json:"isDefault" example:"true"`
	// Storage object UUID for the primary thumbnail
	ThumbnailObjectID *string `json:"thumbnailObjectId,omitempty" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
}

type UpdateVariantRequest struct {
	// Updated SKU identifier
	SKU *string `json:"sku,omitempty" example:"MAC-PRO-16-BLK-18"`
	// Updated variant display title
	Title *string `json:"title,omitempty" example:"Space Black, 18GB RAM, 512GB SSD"`
	// Updated price in cents
	Price *int64 `json:"price,omitempty" example:"249900"`
	// Updated original price in cents
	CrossedOutPrice *int64 `json:"crossedOutPrice,omitempty" example:"299900"`
	// Currency ISO code
	Currency *string `json:"currency,omitempty" example:"USD"`
	// Updated attributes map
	Attributes map[string]any `json:"attributes,omitempty"`
	// Updated default selection flag
	IsDefault *bool `json:"isDefault,omitempty" example:"true"`
	// Updated thumbnail storage object UUID
	ThumbnailObjectID *string `json:"thumbnailObjectId,omitempty" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
}

type VariantMediaSummary struct {
	// Media record UUID
	ID string `json:"id" example:"80000000-0000-0000-0000-000000000001"`
	// Media MIME type or format (e.g. image/jpeg, video/mp4)
	MediaType string `json:"mediaType" example:"image/jpeg"`
	// Display sequence sort order
	SortOrder int `json:"sortOrder" example:"1"`
	// Public CDN or storage URL
	URL string `json:"url" example:"https://example.com/media/headphones.jpg"`
}

type VariantResponse struct {
	// Variant UUID
	ID string `json:"id" example:"70000000-0000-0000-0000-000000000001"`
	// Stock Keeping Unit
	SKU string `json:"sku,omitempty" example:"MAC-PRO-16-BLK-18"`
	// Parent product UUID
	ProductID string `json:"productId" example:"60000000-0000-0000-0000-000000000001"`
	// Variant title
	Title string `json:"title" example:"Space Black, 18GB RAM, 512GB SSD"`
	// Formatted active display price
	Price *string `json:"price,omitempty" example:"$2499.00"`
	// Numeric active price value
	ExtractedPrice *float64 `json:"extractedPrice,omitempty" example:"2499.00"`
	// Formatted original crossed-out price
	OriginalPrice *string `json:"originalPrice,omitempty" example:"$2999.00"`
	// Numeric original price value
	ExtractedOriginalPrice *float64 `json:"extractedOriginalPrice,omitempty" example:"2999.00"`
	// Currency ISO code
	Currency *string `json:"currency,omitempty" example:"USD"`
	// Primary thumbnail image URL
	Thumbnail *string `json:"thumbnail,omitempty" example:"https://example.com/thumb.jpg"`
	// Attached gallery media assets
	Media []VariantMediaSummary `json:"media"`
	// Custom variant attributes
	Attributes map[string]any `json:"attributes"`
	// True if this is the default variant
	IsDefault bool `json:"isDefault" example:"true"`
}

type AdminVariantResponse struct {
	// Variant database UUID
	ID string `json:"id" example:"70000000-0000-0000-0000-000000000001"`
	// Stock Keeping Unit
	SKU string `json:"sku" example:"MAC-PRO-16-BLK-18"`
	// Parent product UUID
	ProductID string `json:"productId" example:"60000000-0000-0000-0000-000000000001"`
	// Variant title
	Title string `json:"title" example:"Space Black, 18GB RAM, 512GB SSD"`
	// Formatted active display price
	Price *string `json:"price,omitempty" example:"$2499.00"`
	// Numeric active price value
	ExtractedPrice *float64 `json:"extractedPrice,omitempty" example:"2499.00"`
	// Formatted original crossed-out price
	OriginalPrice *string `json:"originalPrice,omitempty" example:"$2999.00"`
	// Numeric original price value
	ExtractedOriginalPrice *float64 `json:"extractedOriginalPrice,omitempty" example:"2999.00"`
	// Currency ISO code
	Currency *string `json:"currency,omitempty" example:"USD"`
	// Thumbnail storage object details
	Thumbnail *StorageObjectResponse `json:"thumbnail,omitempty"`
	// Attached gallery media assets
	Media []VariantMediaSummary `json:"media"`
	// Custom variant attributes
	Attributes map[string]any `json:"attributes"`
	// True if this is the default variant
	IsDefault bool `json:"isDefault" example:"true"`
	// Timestamp when variant was created (RFC3339)
	CreatedAt string `json:"createdAt" example:"2026-08-02T16:00:00Z"`
	// Timestamp when variant was last updated (RFC3339)
	UpdatedAt string `json:"updatedAt" example:"2026-08-02T16:00:00Z"`
	// Timestamp when variant was soft-deleted, if applicable
	DeletedAt *string `json:"deletedAt,omitempty" example:"2026-08-02T16:15:00Z"`
}

type AttachMediaRequest struct {
	// Storage object UUID of the uploaded media file
	StorageObjectID string `json:"storageObjectId" binding:"required,uuid" example:"80000000-0000-0000-0000-000000000001"`
	// Media classification (image or video)
	MediaType string `json:"mediaType" binding:"required" example:"image"`
	// Presentation sequence position index
	SortOrder int `json:"sortOrder" example:"1"`
}

type ReorderMediaRequest struct {
	// Array of media UUIDs in desired presentation sequence
	OrderedMediaIDs []string `json:"orderedMediaIds" binding:"required,gt=0,dive,uuid" example:"['80000000-0000-0000-0000-000000000001']"`
}

type StorageObjectResponse struct {
	// Internal object storage UUID
	ID string `json:"id" example:"80000000-0000-0000-0000-000000000001"`
	// Storage bucket name
	Bucket string `json:"bucket" example:"product-media"`
	// Storage object key / filepath
	Key string `json:"key" example:"variants/70000000/image.webp"`
	// MIME content type
	ContentType string `json:"contentType,omitempty" example:"image/webp"`
	// File size in bytes
	FileSize int64 `json:"fileSize,omitempty" example:"1048576"`
	// Public CDN URL
	PublicURL string `json:"publicUrl,omitempty" example:"https://example.com/media/image.webp"`
}

type VariantMediaResponse struct {
	// Media record UUID
	ID string `json:"id" example:"80000000-0000-0000-0000-000000000001"`
	// Associated variant UUID
	VariantID string `json:"variantId" example:"70000000-0000-0000-0000-000000000001"`
	// Underlying storage object UUID
	ObjectID string `json:"objectId" example:"80000000-0000-0000-0000-000000000001"`
	// Media classification (image or video)
	MediaType string `json:"mediaType" example:"image"`
	// Display sequence sort order
	SortOrder int `json:"sortOrder" example:"1"`
	// Detailed storage file metadata
	Object *StorageObjectResponse `json:"object,omitempty"`
}

func mapVariantResponse(v *model.ProductVariant) VariantResponse {
	attrs := v.Attributes
	if attrs == nil {
		attrs = make(map[string]any)
	}

	mediaSummaries := make([]VariantMediaSummary, 0)
	for _, m := range v.Media {
		if m != nil && m.Object != nil {
			mediaSummaries = append(mediaSummaries, VariantMediaSummary{
				ID:        m.ID.String(),
				MediaType: m.MediaType,
				SortOrder: m.SortOrder,
				URL:       m.Object.PublicURL,
			})
		}
	}

	var thumbnailURL *string
	if v.Thumbnail != nil && v.Thumbnail.PublicURL != "" {
		thumbnailURL = &v.Thumbnail.PublicURL
	}

	curr := "USD"
	if v.Currency != nil && *v.Currency != "" {
		curr = *v.Currency
	}

	var priceStr *string
	var extPrice *float64
	if v.Price != nil {
		ps := utils.FormatPrice(*v.Price, curr)
		ep := utils.ExtractPrice(*v.Price)
		priceStr = &ps
		extPrice = &ep
	}

	var origPriceStr *string
	var extOrigPrice *float64
	if v.CrossedOutPrice != nil {
		ops := utils.FormatPrice(*v.CrossedOutPrice, curr)
		eop := utils.ExtractPrice(*v.CrossedOutPrice)
		origPriceStr = &ops
		extOrigPrice = &eop
	}

	return VariantResponse{
		ID:                     v.ID.String(),
		SKU:                    v.SKU,
		ProductID:              v.ProductID.String(),
		Title:                  v.Title,
		Price:                  priceStr,
		ExtractedPrice:         extPrice,
		OriginalPrice:          origPriceStr,
		ExtractedOriginalPrice: extOrigPrice,
		Currency:               v.Currency,
		Thumbnail:              thumbnailURL,
		Media:                  mediaSummaries,
		Attributes:             attrs,
		IsDefault:              v.IsDefault,
	}
}

func mapAdminVariantResponse(v *model.ProductVariant) AdminVariantResponse {
	attrs := v.Attributes
	if attrs == nil {
		attrs = make(map[string]any)
	}

	mediaSummaries := make([]VariantMediaSummary, 0)
	for _, m := range v.Media {
		if m != nil && m.Object != nil {
			mediaSummaries = append(mediaSummaries, VariantMediaSummary{
				ID:        m.ID.String(),
				MediaType: m.MediaType,
				SortOrder: m.SortOrder,
				URL:       m.Object.PublicURL,
			})
		}
	}

	curr := "USD"
	if v.Currency != nil && *v.Currency != "" {
		curr = *v.Currency
	}

	var priceStr *string
	var extPrice *float64
	if v.Price != nil {
		ps := utils.FormatPrice(*v.Price, curr)
		ep := utils.ExtractPrice(*v.Price)
		priceStr = &ps
		extPrice = &ep
	}

	var origPriceStr *string
	var extOrigPrice *float64
	if v.CrossedOutPrice != nil {
		ops := utils.FormatPrice(*v.CrossedOutPrice, curr)
		eop := utils.ExtractPrice(*v.CrossedOutPrice)
		origPriceStr = &ops
		extOrigPrice = &eop
	}

	res := AdminVariantResponse{
		ID:                     v.ID.String(),
		SKU:                    v.SKU,
		ProductID:              v.ProductID.String(),
		Title:                  v.Title,
		Price:                  priceStr,
		ExtractedPrice:         extPrice,
		OriginalPrice:          origPriceStr,
		ExtractedOriginalPrice: extOrigPrice,
		Currency:               v.Currency,
		Media:                  mediaSummaries,
		Attributes:             attrs,
		IsDefault:              v.IsDefault,
		CreatedAt:              v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              v.UpdatedAt.Format(time.RFC3339),
	}

	if v.Thumbnail != nil {
		res.Thumbnail = &StorageObjectResponse{
			ID:          v.Thumbnail.ID.String(),
			Bucket:      v.Thumbnail.Bucket,
			Key:         v.Thumbnail.Key,
			ContentType: v.Thumbnail.ContentType,
			FileSize:    v.Thumbnail.FileSize,
			PublicURL:   v.Thumbnail.PublicURL,
		}
	}

	if v.DeletedAt != nil {
		del := v.DeletedAt.Format(time.RFC3339)
		res.DeletedAt = &del
	}
	return res
}

func mapVariantMediaResponse(m *model.VariantMedia) VariantMediaResponse {
	res := VariantMediaResponse{
		ID:        m.ID.String(),
		VariantID: m.VariantID.String(),
		ObjectID:  m.ObjectID.String(),
		MediaType: m.MediaType,
		SortOrder: m.SortOrder,
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

// CreateVariant godoc
//
//	@Summary		Create a product variant
//	@Description	Adds a new SKU/variant to an existing product (e.g., specific color, size, price, or custom attributes).
//	@Tags			Admin Product Variants
//	@Accept			json
//	@Produce		json
//	@Param			product_id	path		string									true	"Product UUID"	format(uuid)
//	@Param			data		body		CreateVariantRequest					true	"Variant title, price, crossed-out price, currency, attributes map, and default flag"
//	@Failure		400			{object}	api.BadRequestErrorResponse				"Validation error or missing required fields"
//	@Failure		404			{object}	api.NotFoundErrorResponse				"Parent product not found"
//	@Failure		500			{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		201			{object}	api.DataResponse{data=VariantResponse}	"Created variant details"
//	@Router			/admin/products/{product_id}/variants [post]
func (vh *VariantHandler) CreateVariant(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "product_id",
			Message: "invalid product UUID format",
		}))
		return
	}

	var body CreateVariantRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	var thumbID *uuid.UUID
	if body.ThumbnailObjectID != nil {
		id, err := uuid.Parse(*body.ThumbnailObjectID)
		if err != nil {
			_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "thumbnail_object_id",
				Message: "invalid thumbnail UUID format",
			}))
			return
		}
		thumbID = &id
	}

	in := &CreateVariantInput{
		ProductID:         productID,
		SKU:               body.SKU,
		Title:             body.Title,
		Price:             body.Price,
		CrossedOutPrice:   body.CrossedOutPrice,
		Currency:          body.Currency,
		Attributes:        body.Attributes,
		IsDefault:         body.IsDefault,
		ThumbnailObjectID: thumbID,
	}

	variant, err := vh.vservice.CreateVariant(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{
		Data: mapVariantResponse(variant),
	})
}

// GetVariantByID godoc
//
//	@Summary		Get variant details by ID
//	@Description	Retrieves specific product variant (SKU) details by its UUID.
//	@Tags			Product Variants
//	@Produce		json
//	@Param			id	path		string									true	"Variant UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse				"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse				"Variant not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=VariantResponse}	"Variant details"
//	@Router			/variants/{id} [get]
func (vh *VariantHandler) GetVariantByID(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	variant, err := vh.vservice.GetVariantByID(c.Request.Context(), variantID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapVariantResponse(variant),
	})
}

// GetVariantBySKU godoc
//
//	@Summary		Get variant details by SKU
//	@Description	Retrieves specific product variant details by its SKU string.
//	@Tags			Product Variants
//	@Produce		json
//	@Param			sku	path		string									true	"Variant SKU"
//	@Failure		400	{object}	api.BadRequestErrorResponse				"SKU is required"
//	@Failure		404	{object}	api.NotFoundErrorResponse				"Variant not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=VariantResponse}	"Variant details"
//	@Router			/variants/sku/{sku} [get]
func (vh *VariantHandler) GetVariantBySKU(c *gin.Context) {
	sku := c.Param("sku")
	variant, err := vh.vservice.GetVariantBySKU(c.Request.Context(), sku)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapVariantResponse(variant),
	})
}

// UpdateVariantByID godoc
//
//	@Summary		Update product variant attributes
//	@Description	Updates specific fields of an existing variant such as SKU, title, price, crossed-out price, currency, attributes, or default status.
//	@Tags			Admin Product Variants
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Variant UUID"	format(uuid)
//	@Param			input	body		UpdateVariantRequest			true	"Fields to update (sku, title, price, crossed_out_price, currency, attributes, is_default)"
//	@Failure		400		{object}	api.BadRequestErrorResponse		"Validation error or invalid UUID format"
//	@Failure		404		{object}	api.NotFoundErrorResponse		"Variant not found"
//	@Failure		409		{object}	api.ConflictErrorResponse		"SKU already in use"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200		{object}	api.MessageResponse				"Update confirmation message"
//	@Router			/admin/variants/{id} [patch]
func (vh *VariantHandler) UpdateVariantByID(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	var body UpdateVariantRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	var thumbID *uuid.UUID
	if body.ThumbnailObjectID != nil {
		id, err := uuid.Parse(*body.ThumbnailObjectID)
		if err != nil {
			_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
				Field:   "thumbnail_object_id",
				Message: "invalid thumbnail UUID format",
			}))
			return
		}
		thumbID = &id
	}

	err = vh.vservice.UpdateVariant(c.Request.Context(), variantID, UpdateVariantInput{
		SKU:               body.SKU,
		Title:             body.Title,
		Price:             body.Price,
		CrossedOutPrice:   body.CrossedOutPrice,
		Currency:          body.Currency,
		Attributes:        body.Attributes,
		IsDefault:         body.IsDefault,
		ThumbnailObjectID: thumbID,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "updated successfully",
	})
}

// DeleteVariantByID godoc
//
//	@Summary		Soft-delete a product variant
//	@Description	Marks an active product variant as soft-deleted (`deleted_at = NOW()`), removing it from active product options.
//	@Tags			Admin Product Variants
//	@Produce		json
//	@Param			id	path		string							true	"Variant UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Variant not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		204													"Deletion confirmation"
//	@Router			/admin/variants/{id} [delete]
func (vh *VariantHandler) DeleteVariantByID(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	if err := vh.vservice.DeleteVariantByID(c.Request.Context(), variantID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RestoreVariantByID godoc
//
//	@Summary		Restore a soft-deleted product variant
//	@Description	Restores a soft-deleted product variant back to active status (`deleted_at = NULL`).
//	@Tags			Admin Product Variants
//	@Produce		json
//	@Param			id	path		string							true	"Variant UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Variant not found or not deleted"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.MessageResponse				"Restore confirmation message"
//	@Router			/admin/variants/{id}/restore [post]
func (vh *VariantHandler) RestoreVariantByID(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	if err := vh.vservice.RestoreVariantByID(c.Request.Context(), variantID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "restored successfully",
	})
}

// ListVariantsByProductID godoc
//
//	@Summary		List active variants for a product
//	@Description	Returns a paginated list of active variants associated with a specific product for storefront selection.
//	@Tags			Product Variants
//	@Produce		json
//	@Param			product_id	path		string																true	"Product UUID"	format(uuid)
//	@Param			q			query		pagination.ListQuery												true	"Pagination, search query, and sorting parameters"
//	@Failure		400			{object}	api.BadRequestErrorResponse											"Invalid query parameters or UUID format"
//	@Failure		500			{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Success		200			{object}	api.PaginatedResponse{data=[]VariantResponse,meta=pagination.Page}	"Paginated list of active product variants"
//	@Router			/products/{product_id}/variants [get]
func (vh *VariantHandler) ListVariantsByProductID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "product_id",
			Message: "invalid product UUID format",
		}))
		return
	}

	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	result, err := vh.vservice.ListVariantsByProductID(c.Request.Context(), productID, q, false)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]VariantResponse, 0, len(result.Items))
	for _, v := range result.Items {
		res = append(res, mapVariantResponse(v))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: res,
		Meta: result.Page,
	})
}

// AdminListVariantsByProductID godoc
//
//	@Summary		List all variants for a product including soft-deleted ones (Admin)
//	@Description	Returns a paginated list of all variants associated with a specific product including soft-deleted records for administrator management.
//	@Tags			Admin Product Variants
//	@Produce		json
//	@Param			product_id	path		string																	true	"Product UUID"	format(uuid)
//	@Param			q			query		pagination.ListQuery													true	"Pagination, search query, and sorting parameters"
//	@Failure		400			{object}	api.BadRequestErrorResponse												"Invalid query parameters or UUID format"
//	@Failure		500			{object}	api.InternalServerErrorResponse											"Internal server error"
//	@Success		200			{object}	api.PaginatedResponse{data=[]AdminVariantResponse,meta=pagination.Page}	"Paginated list of all product variants including deleted"
//	@Router			/admin/products/{product_id}/variants [get]
func (vh *VariantHandler) AdminListVariantsByProductID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "product_id",
			Message: "invalid product UUID format",
		}))
		return
	}

	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	result, err := vh.vservice.ListVariantsByProductID(c.Request.Context(), productID, q, true)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]AdminVariantResponse, 0, len(result.Items))
	for _, v := range result.Items {
		res = append(res, mapAdminVariantResponse(v))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: res,
		Meta: result.Page,
	})
}

// AttachMedia godoc
//
//	@Summary		Attach a storage object to a variant
//	@Description	Links an uploaded storage object (image/video) to a specific product variant with media type and sort order.
//	@Tags			Admin Variant Media
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string										true	"Variant UUID"	format(uuid)
//	@Param			body	body		AttachMediaRequest							true	"Storage object ID, media type, and sort order"
//	@Failure		400		{object}	api.BadRequestErrorResponse					"Validation error or invalid UUID reference"
//	@Failure		409		{object}	api.ConflictErrorResponse					"Storage object is already attached to this variant"
//	@Failure		500		{object}	api.InternalServerErrorResponse				"Internal server error"
//	@Success		201		{object}	api.DataResponse{data=VariantMediaResponse}	"Attached variant media details"
//	@Router			/admin/variants/{id}/media [post]
func (vh *VariantHandler) AttachMedia(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	var body AttachMediaRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	objectID, err := uuid.Parse(body.StorageObjectID)
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "storage_object_id",
			Message: "invalid storage object UUID format",
		}))
		return
	}

	media, err := vh.vservice.AttachMedia(c.Request.Context(), AttachMediaInput{
		VariantID:       variantID,
		StorageObjectID: objectID,
		MediaType:       body.MediaType,
		SortOrder:       body.SortOrder,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{
		Data: mapVariantMediaResponse(media),
	})
}

// ListVariantMedia godoc
//
//	@Summary		List all media items for a variant
//	@Description	Returns all attached media items for a specific variant with presigned object URLs.
//	@Tags			Variant Media
//	@Produce		json
//	@Param			id	path		string											true	"Variant UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse						"Invalid UUID format"
//	@Failure		500	{object}	api.InternalServerErrorResponse					"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=[]VariantMediaResponse}	"List of variant media items"
//	@Router			/variants/{id}/media [get]
func (vh *VariantHandler) ListVariantMedia(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	mediaList, err := vh.vservice.ListVariantMedia(c.Request.Context(), variantID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]VariantMediaResponse, 0, len(mediaList))
	for _, m := range mediaList {
		res = append(res, mapVariantMediaResponse(m))
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}

// DetachMedia godoc
//
//	@Summary		Remove a media attachment from a variant
//	@Description	Removes a media attachment relationship from a variant.
//	@Tags			Admin Variant Media
//	@Produce		json
//	@Param			id			path		string							true	"Variant UUID"			format(uuid)
//	@Param			media_id	path		string							true	"Media Attachment UUID"	format(uuid)
//	@Failure		400			{object}	api.BadRequestErrorResponse		"Invalid UUID format"
//	@Failure		404			{object}	api.NotFoundErrorResponse		"Media relationship not found"
//	@Failure		500			{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		204													"Media detached successfully"
//	@Router			/admin/variants/{id}/media/{media_id} [delete]
func (vh *VariantHandler) DetachMedia(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
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

	if err := vh.vservice.DetachMedia(c.Request.Context(), variantID, mediaID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderMedia godoc
//
//	@Summary		Batch reorder media items for a variant
//	@Description	Reorders attached media items for a variant according to the specified array of media IDs.
//	@Tags			Admin Variant Media
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Variant UUID"	format(uuid)
//	@Param			body	body		ReorderMediaRequest				true	"Ordered list of media UUIDs"
//	@Failure		400		{object}	api.BadRequestErrorResponse		"Invalid UUID format or empty list"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200		{object}	api.MessageResponse				"Reorder confirmation message"
//	@Router			/admin/variants/{id}/media/reorder [patch]
func (vh *VariantHandler) ReorderMedia(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
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

	if err := vh.vservice.ReorderMedia(c.Request.Context(), variantID, orderedIDs); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "reordered successfully",
	})
}

type AdjustStockRequest struct {
	// Adjustment quantity delta (positive to increase, negative to decrease)
	Quantity int `json:"quantity" binding:"required" example:"50"`
	// Business audit reason (restock, sale, return, adjustment, reservation_release)
	Reason string `json:"reason" binding:"required,oneof=restock sale return adjustment reservation_release" example:"restock"`
}

type PublicStockResponse struct {
	// Units available for customer checkout
	AvailableQuantity int `json:"availableQuantity" example:"50"`
	// True if availableQuantity > 0
	IsInStock bool `json:"isInStock" example:"true"`
}

type InventoryStockResponse struct {
	// Variant UUID
	VariantID string `json:"variantId" example:"70000000-0000-0000-0000-000000000001"`
	// Physical units counted on warehouse shelves
	OnHandQuantity int `json:"onHandQuantity" example:"60"`
	// Units currently reserved in open orders / pending checkouts
	ReservedQuantity int `json:"reservedQuantity" example:"10"`
	// Salable stock (onHand - reserved)
	AvailableQuantity int `json:"availableQuantity" example:"50"`
	// True if availableQuantity > 0
	IsInStock bool `json:"isInStock" example:"true"`
}

type InventoryLedgerResponse struct {
	// Ledger entry record UUID
	ID string `json:"id" example:"80000000-0000-0000-0000-000000000001"`
	// Associated variant UUID
	VariantID string `json:"variantId" example:"70000000-0000-0000-0000-000000000001"`
	// Signed quantity adjustment (+/- units)
	Quantity int `json:"quantity" example:"50"`
	// Adjustment reason classification
	Reason string `json:"reason" example:"restock"`
	// Transaction timestamp (RFC3339)
	CreatedAt string `json:"createdAt" example:"2026-08-15T12:00:00Z"`
}

func mapStockResponse(stock *model.InventoryStock) InventoryStockResponse {
	return InventoryStockResponse{
		VariantID:         stock.VariantID.String(),
		OnHandQuantity:    stock.OnHandQuantity,
		ReservedQuantity:  stock.ReservedQuantity,
		AvailableQuantity: stock.AvailableQuantity,
		IsInStock:         stock.IsInStock,
	}
}

func mapLedgerResponse(l *model.InventoryLedger) InventoryLedgerResponse {
	return InventoryLedgerResponse{
		ID:        l.ID.String(),
		VariantID: l.VariantID.String(),
		Quantity:  l.Quantity,
		Reason:    l.Reason.String(),
		CreatedAt: l.CreatedAt.Format(time.RFC3339),
	}
}

// AdjustStock godoc
//
//	@Summary		Adjust variant inventory stock
//	@Description	Records a new inventory ledger transaction (restock, adjustment, sale, return) to increment or decrement the variant's stock.
//	@Tags			Admin Variant Inventory
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Variant UUID"	format(uuid)
//	@Param			body	body		AdjustStockRequest					true	"Stock adjustment details"
//	@Failure		400		{object}	api.BadRequestErrorResponse			"Validation error or insufficient inventory"
//	@Failure		404		{object}	api.NotFoundErrorResponse			"Variant not found"
//	@Failure		500		{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Success		200		{object}	api.DataResponse{data=InventoryStockResponse}	"Updated stock levels"
//	@Router			/admin/variants/{id}/inventory/adjust [post]
func (vh *VariantHandler) AdjustStock(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	var body AdjustStockRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	stock, err := vh.vservice.AdjustStock(c.Request.Context(), inventory.AdjustStockInput{
		VariantID: variantID,
		Quantity:  body.Quantity,
		Reason:    body.Reason,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapStockResponse(stock),
	})
}

// GetVariantStockPublic godoc
//
//	@Summary		Get current available stock for a variant (Public)
//	@Description	Retrieves available quantity and in-stock status for a variant for customer storefront.
//	@Tags			Variant Inventory
//	@Produce		json
//	@Param			id	path		string									true	"Variant UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse				"Invalid UUID format"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=PublicStockResponse}	"Public stock availability"
//	@Router			/variants/{id}/inventory [get]
func (vh *VariantHandler) GetVariantStockPublic(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	stock, err := vh.vservice.GetStockByVariantID(c.Request.Context(), variantID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: PublicStockResponse{
			AvailableQuantity: stock.AvailableQuantity,
			IsInStock:         stock.IsInStock,
		},
	})
}

// GetVariantStockAdmin godoc
//
//	@Summary		Get detailed stock levels for a variant (Admin)
//	@Description	Retrieves on-hand, reserved, available quantities and in-stock status for a variant.
//	@Tags			Admin Variant Inventory
//	@Produce		json
//	@Param			id	path		string								true	"Variant UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse			"Invalid UUID format"
//	@Failure		500	{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=InventoryStockResponse}	"Variant stock summary"
//	@Router			/admin/variants/{id}/inventory [get]
func (vh *VariantHandler) GetVariantStockAdmin(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	stock, err := vh.vservice.GetStockByVariantID(c.Request.Context(), variantID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapStockResponse(stock),
	})
}

// ListVariantLedgers godoc
//
//	@Summary		List inventory audit ledgers for a variant
//	@Description	Returns a paginated list of inventory audit ledger transactions for a specific variant.
//	@Tags			Admin Variant Inventory
//	@Produce		json
//	@Param			id	path		string													true	"Variant UUID"	format(uuid)
//	@Param			q	query		pagination.ListQuery									true	"Pagination query"
//	@Failure		400	{object}	api.BadRequestErrorResponse								"Invalid query parameters or UUID format"
//	@Failure		500	{object}	api.InternalServerErrorResponse							"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]InventoryLedgerResponse,meta=pagination.Page}	"Paginated inventory audit history"
//	@Router			/admin/variants/{id}/inventory/ledgers [get]
func (vh *VariantHandler) ListVariantLedgers(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid variant UUID format",
		}))
		return
	}

	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	result, err := vh.vservice.ListLedgers(c.Request.Context(), variantID, q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := make([]InventoryLedgerResponse, 0, len(result.Items))
	for _, l := range result.Items {
		res = append(res, mapLedgerResponse(l))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: res,
		Meta: result.Page,
	})
}
