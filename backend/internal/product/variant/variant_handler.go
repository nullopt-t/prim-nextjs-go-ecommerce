package variant

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

type VariantHandler struct {
	vservice *VariantService
}

func NewHandler(s *VariantService) *VariantHandler {
	return &VariantHandler{
		vservice: s,
	}
}

type CreateVariantRequest struct {
	Title           string         `json:"title" binding:"required" example:"Red / XL"`
	Price           *int64         `json:"price,omitempty" example:"2999"`
	CrossedOutPrice *int64         `json:"crossed_out_price,omitempty" example:"3999"`
	Currency        *string        `json:"currency,omitempty" example:"USD"`
	Attributes      map[string]any `json:"attributes,omitempty"`
	IsDefault       bool           `json:"is_default" example:"false"`
}

type UpdateVariantRequest struct {
	Title           *string        `json:"title,omitempty" example:"Red / XXL"`
	Price           *int64         `json:"price,omitempty" example:"3499"`
	CrossedOutPrice *int64         `json:"crossed_out_price,omitempty" example:"4499"`
	Currency        *string        `json:"currency,omitempty" example:"USD"`
	Attributes      map[string]any `json:"attributes,omitempty"`
	IsDefault       *bool          `json:"is_default,omitempty" example:"true"`
}

type VariantResponse struct {
	ID              string         `json:"id" example:"96c4e462-ed4a-4fec-9115-47cbf12206a7"`
	ProductID       string         `json:"product_id" example:"356cbaee-4700-4af5-ac9c-61aeeafd541c"`
	Title           string         `json:"title" example:"Red / XL"`
	Price           *int64         `json:"price,omitempty" example:"2999"`
	CrossedOutPrice *int64         `json:"crossed_out_price,omitempty" example:"3999"`
	Currency        *string        `json:"currency,omitempty" example:"USD"`
	Attributes      map[string]any `json:"attributes"`
	IsDefault       bool           `json:"is_default" example:"false"`
	CreatedAt       string         `json:"created_at" example:"2026-08-02T16:00:00Z"`
	UpdatedAt       string         `json:"updated_at" example:"2026-08-02T16:00:00Z"`
}

type AdminVariantResponse struct {
	VariantResponse
	DeletedAt *string `json:"deleted_at,omitempty" example:"2026-08-02T16:15:00Z"`
}

type AttachMediaRequest struct {
	StorageObjectID string `json:"storage_object_id" binding:"required,uuid" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	MediaType       string `json:"media_type" binding:"required" example:"image"`
	SortOrder       int    `json:"sort_order" example:"0"`
}

type ReorderMediaRequest struct {
	OrderedMediaIDs []string `json:"ordered_media_ids" binding:"required,gt=0,dive,uuid"`
}

type StorageObjectResponse struct {
	ID          string `json:"id"`
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	ContentType string `json:"content_type,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
	PublicURL   string `json:"public_url,omitempty"`
}

type VariantMediaResponse struct {
	ID              string                 `json:"id" example:"8f123456-e89b-12d3-a456-426614174000"`
	VariantID       string                 `json:"variant_id" example:"96c4e462-ed4a-4fec-9115-47cbf12206a7"`
	StorageObjectID string                 `json:"storage_object_id" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	MediaType       string                 `json:"media_type" example:"image"`
	SortOrder       int                    `json:"sort_order" example:"0"`
	Object          *StorageObjectResponse `json:"object,omitempty"`
}

func mapVariantResponse(v *model.ProductVariant) VariantResponse {
	attrs := v.Attributes
	if attrs == nil {
		attrs = make(map[string]any)
	}

	return VariantResponse{
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
	}
}

func mapAdminVariantResponse(v *model.ProductVariant) AdminVariantResponse {
	res := AdminVariantResponse{
		VariantResponse: mapVariantResponse(v),
	}
	if v.DeletedAt != nil {
		del := v.DeletedAt.Format(time.RFC3339)
		res.DeletedAt = &del
	}
	return res
}

func mapVariantMediaResponse(m *model.VariantMedia) VariantMediaResponse {
	res := VariantMediaResponse{
		ID:              m.ID.String(),
		VariantID:       m.VariantID.String(),
		StorageObjectID: m.StorageObjectID.String(),
		MediaType:       m.MediaType,
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

// CreateVariant godoc
// @Summary create a product variant
// @Tags Variant
// @Accept json
// @Produce json
// @Param product_id path string true "Product ID (UUID)" format(uuid)
// @Param data body CreateVariantRequest true "variant payload"
// @Failure 400 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=VariantResponse}
// @Router /admin/products/{product_id}/variants [post]
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

	in := &CreateVariantInput{
		ProductID:       productID,
		Title:           body.Title,
		Price:           body.Price,
		CrossedOutPrice: body.CrossedOutPrice,
		Currency:        body.Currency,
		Attributes:      body.Attributes,
		IsDefault:       body.IsDefault,
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
// @Summary get variant by id
// @Tags Variant
// @Produce json
// @Param id path string true "Variant ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=VariantResponse}
// @Router /variants/{id} [get]
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

// UpdateVariantByID godoc
// @Summary update a product variant
// @Tags Variant
// @Accept json
// @Produce json
// @Param id path string true "Variant ID (UUID)" format(uuid)
// @Param input body UpdateVariantRequest true "variant update payload"
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/variants/{id} [patch]
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

	err = vh.vservice.UpdateVariant(c.Request.Context(), variantID, UpdateVariantInput{
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

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "updated successfully",
	})
}

// DeleteVariantByID godoc
// @Summary soft-delete a product variant
// @Tags Variant
// @Produce json
// @Param id path string true "Variant ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/variants/{id} [delete]
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

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "deleted successfully",
	})
}

// ListVariantsByProductID godoc
// @Summary list public variants for a product
// @Tags Variant
// @Produce json
// @Param product_id path string true "Product ID (UUID)" format(uuid)
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]VariantResponse,meta=api.Page}
// @Router /products/{product_id}/variants [get]
func (vh *VariantHandler) ListVariantsByProductID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "product_id",
			Message: "invalid product UUID format",
		}))
		return
	}

	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(api.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

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
// @Summary list all variants for a product including soft-deleted ones (admin)
// @Tags Variant
// @Produce json
// @Param product_id path string true "Product ID (UUID)" format(uuid)
// @Param q query api.ListQuery true "url query"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.PaginatedResponse{data=[]AdminVariantResponse,meta=api.Page}
// @Router /admin/products/{product_id}/variants [get]
func (vh *VariantHandler) AdminListVariantsByProductID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		_ = c.Error(apierr.ErrInvalidUUID().WithFields(api.FieldError{
			Field:   "product_id",
			Message: "invalid product UUID format",
		}))
		return
	}

	q := &api.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(api.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

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
// @Summary attach a storage object to a variant
// @Tags Variant Media
// @Accept json
// @Produce json
// @Param id path string true "Variant ID (UUID)" format(uuid)
// @Param body body AttachMediaRequest true "media attachment details"
// @Failure 400 {object} api.ErrorResponse
// @Failure 409 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 201 {object} api.DataResponse{data=VariantMediaResponse}
// @Router /admin/variants/{id}/media [post]
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
// @Summary list all media for a variant
// @Tags Variant Media
// @Produce json
// @Param id path string true "Variant ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.DataResponse{data=[]VariantMediaResponse}
// @Router /variants/{id}/media [get]
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
// @Summary remove a media attachment from a variant
// @Tags Variant Media
// @Produce json
// @Param id path string true "Variant ID (UUID)" format(uuid)
// @Param media_id path string true "Media ID (UUID)" format(uuid)
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/variants/{id}/media/{media_id} [delete]
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

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "detached successfully",
	})
}

// ReorderMedia godoc
// @Summary batch reorder media items for a variant
// @Tags Variant Media
// @Accept json
// @Produce json
// @Param id path string true "Variant ID (UUID)" format(uuid)
// @Param body body ReorderMediaRequest true "ordered media IDs"
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Success 200 {object} api.MessageResponse
// @Router /admin/variants/{id}/media/reorder [patch]
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
