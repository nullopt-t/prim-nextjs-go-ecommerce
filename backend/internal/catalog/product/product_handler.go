package product

import (
	"errors"
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
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
)

type ProductHandler struct {
	service *ProductService
}

func NewHandler(s *ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

type CreateProductRequest struct {
	BrandID     *string `json:"brandId,omitempty" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	CategoryID  string  `json:"categoryId" binding:"required,uuid" example:"356cbaee-4700-4af5-ac9c-61aeeafd541c"`
	Slug        string  `json:"slug" binding:"required" example:"wireless-noise-canceling-headphones"`
	Title       string  `json:"title" binding:"required" example:"Wireless Noise-Canceling Headphones"`
	Description *string `json:"description,omitempty" example:"Premium over-ear Bluetooth headphones with active noise cancellation."`
	ProductType string  `json:"productType" example:"simple"`
}

type UpdateProductRequest struct {
	BrandID     *string  `json:"brandId,omitempty"`
	CategoryID  *string  `json:"categoryId,omitempty"`
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Highlights  []string `json:"highlights,omitempty"`
	ProductType *string  `json:"productType,omitempty" example:"simple"`
}

type ProductBrandSummary struct {
	ID   string  `json:"id,omitempty"`
	Name string  `json:"name"`
	Link *string `json:"link,omitempty"`
}

type ProductCategorySummary struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type ProductTagSummary struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type PublicStockSummary struct {
	AvailableQuantity int  `json:"availableQuantity"`
	IsInStock         bool `json:"isInStock"`
}

type AdminStockSummary struct {
	OnHandQuantity    int  `json:"onHandQuantity"`
	ReservedQuantity  int  `json:"reservedQuantity"`
	AvailableQuantity int  `json:"availableQuantity"`
	IsInStock         bool `json:"isInStock"`
}

type VariantMediaSummary struct {
	ID        string `json:"id"`
	MediaType string `json:"mediaType"`
	SortOrder int    `json:"sortOrder"`
	URL       string `json:"url"`
}

type ProductVariantResponse struct {
	ID                     string                        `json:"id"`
	SKU                    string                        `json:"sku,omitempty"`
	Title                  string                        `json:"title"`
	Price                  *string                       `json:"price,omitempty" example:"$24.99"`
	ExtractedPrice         *float64                      `json:"extractedPrice,omitempty" example:"24.99"`
	OriginalPrice          *string                       `json:"originalPrice,omitempty" example:"$29.99"`
	ExtractedOriginalPrice *float64                      `json:"extractedOriginalPrice,omitempty" example:"29.99"`
	Currency               *string                       `json:"currency,omitempty" example:"USD"`
	Thumbnail              *string                       `json:"thumbnail,omitempty"`
	Media                  []VariantMediaSummary         `json:"media"`
	Attributes             map[string]any                `json:"attributes"`
	IsDefault              bool                          `json:"isDefault"`
	Stock                  *PublicStockSummary           `json:"stock,omitempty"`
}

type AdminProductVariantResponse struct {
	ID                     string                 `json:"id"`
	SKU                    string                 `json:"sku,omitempty"`
	Title                  string                 `json:"title"`
	Price                  *string                `json:"price,omitempty" example:"$24.99"`
	ExtractedPrice         *float64               `json:"extractedPrice,omitempty" example:"24.99"`
	OriginalPrice          *string                `json:"originalPrice,omitempty" example:"$29.99"`
	ExtractedOriginalPrice *float64               `json:"extractedOriginalPrice,omitempty" example:"29.99"`
	Currency               *string                `json:"currency,omitempty" example:"USD"`
	Thumbnail              *StorageObjectResponse `json:"thumbnail,omitempty"`
	Media                  []VariantMediaSummary  `json:"media"`
	Attributes             map[string]any         `json:"attributes"`
	IsDefault              bool                   `json:"isDefault"`
	Stock                  *AdminStockSummary     `json:"stock,omitempty"`
}

type ProductResponse struct {
	ID                     string   `json:"id" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	Slug                   string   `json:"slug" example:"wireless-headphones"`
	Title                  string   `json:"title" example:"Wireless Headphones"`
	Description            *string  `json:"description,omitempty"`
	Highlights             []string `json:"highlights"`
	ProductType            string   `json:"productType"`
	Price                  *string  `json:"price,omitempty" example:"$24.99"`
	ExtractedPrice         *float64 `json:"extractedPrice,omitempty" example:"24.99"`
	OriginalPrice          *string  `json:"originalPrice,omitempty" example:"$29.99"`
	ExtractedOriginalPrice *float64 `json:"extractedOriginalPrice,omitempty" example:"29.99"`
	Currency               *string  `json:"currency,omitempty" example:"USD"`
}

type ProductRatingSummaryResponse struct {
	AverageRating float64       `json:"averageRating" example:"4.5"`
	ReviewCount   int           `json:"reviewCount" example:"24"`
	Distribution  map[int16]int `json:"distribution"`
}

type ProductDetailsResponse struct {
	ProductResponse
	Brand    *ProductBrandSummary          `json:"brand,omitempty"`
	Variants []ProductVariantResponse      `json:"variants"`
	Tags     []ProductTagSummary           `json:"tags"`
	Rating   *ProductRatingSummaryResponse `json:"rating,omitempty"`
}

type AdminProductDetailsResponse struct {
	ID          string                        `json:"id" example:"a1b2c3d4-e5f6-7890-1234-56789abcdef0"`
	Slug        string                        `json:"slug"`
	Title       string                        `json:"title" example:"Wireless Headphones"`
	Description *string                       `json:"description,omitempty"`
	Highlights  []string                      `json:"highlights"`
	Status      model.PublicationStatus       `json:"status" example:"draft"`
	ProductType string                        `json:"productType"`
	Thumbnail   *StorageObjectResponse        `json:"thumbnail,omitempty"`
	BrandID     *string                       `json:"brandId,omitempty"`
	Brand       *ProductBrandSummary          `json:"brand,omitempty"`
	CategoryID  string                        `json:"categoryId"`
	Category    *ProductCategorySummary       `json:"category,omitempty"`
	Variants    []AdminProductVariantResponse `json:"variants"`
	Tags        []ProductTagSummary           `json:"tags"`
	CreatedAt   string                        `json:"createdAt" example:"2026-08-05T19:00:00Z"`
	UpdatedAt   string                        `json:"updatedAt" example:"2026-08-05T19:00:00Z"`
	DeletedAt   *string                       `json:"deletedAt,omitempty" example:"2026-08-05T19:30:00Z"`
}

type ProductReviewResponse struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"productId"`
	UserID      string    `json:"userId"`
	OrderItemID string    `json:"orderItemId"`
	Rating      int16     `json:"rating"`
	Title       *string   `json:"title,omitempty"`
	Body        *string   `json:"body,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
}

// --- DTOs: Media ---

type StorageObjectResponse struct {
	ContentType string `json:"contentType,omitempty"`
	FileSize    int64  `json:"fileSize,omitempty"`
	PublicURL   string `json:"url,omitempty"`
}

type ReorderMediaRequest struct {
	OrderedMediaIDs []string `json:"orderedMediaIds" binding:"required,gt=0,dive,uuid"`
}

// --- DTOs: Variants ---

type CreateProductVariantRequest struct {
	Title           string         `json:"title" binding:"required" example:"Black / XL"`
	Price           *int64         `json:"price,omitempty" example:"2999"`
	CrossedOutPrice *int64         `json:"crossedOutPrice,omitempty" example:"3999"`
	Currency        *string        `json:"currency,omitempty" example:"USD"`
	Attributes      map[string]any `json:"attributes,omitempty"`
	IsDefault       bool           `json:"isDefault" example:"false"`
	InitialStock    *int           `json:"initialStock,omitempty" example:"100"`
}

type SetDefaultVariantRequest struct {
	VariantID string `json:"variantId" binding:"required,uuid"`
}

type PutProductTagsRequest struct {
	TagIDs []string `json:"tagIds" binding:"required,dive,uuid"`
}

type PutProductCategoryRequest struct {
	CategoryID string `json:"categoryId" binding:"required,uuid"`
}

// --- Mappers ---

func mapProductResponse(p *model.Product) ProductResponse {
	highlights := p.Highlights
	if highlights == nil {
		highlights = make([]string, 0)
	}

	res := ProductResponse{
		ID:          p.ID.String(),
		Slug:        p.Slug,
		Title:       p.Title,
		Description: p.Description,
		Highlights:  highlights,
		ProductType: p.ProductType.String(),
	}




	return res
}

func mapAdminProductDetailsResponse(details *ProductDetails) AdminProductDetailsResponse {
	p := details.Product

	highlights := p.Highlights
	if highlights == nil {
		highlights = make([]string, 0)
	}

	res := AdminProductDetailsResponse{
		ID:          p.ID.String(),
		Slug:        p.Slug,
		Title:       p.Title,
		Description: p.Description,
		Highlights:  highlights,
		Status:      p.Status,
		ProductType: p.ProductType.String(),
		CategoryID:  p.CategoryID.String(),
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
		Variants:    make([]AdminProductVariantResponse, 0),
		Tags:        make([]ProductTagSummary, 0),
	}

	if p.Thumbnail != nil {
		res.Thumbnail = &StorageObjectResponse{
			PublicURL:   p.Thumbnail.PublicURL,
			ContentType: p.Thumbnail.ContentType,
			FileSize:    p.Thumbnail.FileSize,
		}
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
			ID:   details.Brand.ID.String(),
			Name: details.Brand.Name,
			Link: details.Brand.Link,
		}
	}

	if details.Category != nil {
		res.Category = &ProductCategorySummary{
			ID:   details.Category.ID.String(),
			Name: details.Category.Name,
		}
	}


	if len(details.Variants) > 0 {
		variantRes := make([]AdminProductVariantResponse, 0, len(details.Variants))
		for _, v := range details.Variants {
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

			vr := AdminProductVariantResponse{
				ID:                     v.ID.String(),
				SKU:                    v.SKU,
				Title:                  v.Title,
				Price:                  priceStr,
				ExtractedPrice:         extPrice,
				OriginalPrice:          origPriceStr,
				ExtractedOriginalPrice: extOrigPrice,
				Currency:               v.Currency,
				Media:                  mediaSummaries,
				Attributes:             attrs,
				IsDefault:              v.IsDefault,
			}
			if v.Thumbnail != nil {
				vr.Thumbnail = &StorageObjectResponse{
					ContentType: v.Thumbnail.ContentType,
					FileSize:    v.Thumbnail.FileSize,
					PublicURL:   v.Thumbnail.PublicURL,
				}
			}
			if details.Stock != nil {
				if s, ok := details.Stock[v.ID]; ok && s != nil {
					vr.Stock = &AdminStockSummary{
						OnHandQuantity:    s.OnHandQuantity,
						ReservedQuantity:  s.ReservedQuantity,
						AvailableQuantity: s.AvailableQuantity,
						IsInStock:         s.IsInStock,
					}
				}
			}
			variantRes = append(variantRes, vr)
		}
		res.Variants = variantRes
	}

	if len(details.Tags) > 0 {
		tagRes := make([]ProductTagSummary, 0, len(details.Tags))
		for _, t := range details.Tags {
			tagRes = append(tagRes, ProductTagSummary{
				ID:   t.ID.String(),
				Name: t.Name,
			})
		}
		res.Tags = tagRes
	}

	return res
}


// --- Product Handlers ---

// CreateProductAsDraft godoc
//
//	@Summary		Create a new draft product
//	@Description	Creates a new product in 'draft' status. Products remain in draft state until at least one variant is created and the product is explicitly published.
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			product	body		CreateProductRequest								true	"Product title, description, category ID, optional brand ID and highlights"
//	@Failure		400		{object}	api.BadRequestErrorResponse							"Validation error or invalid UUID reference"
//	@Failure		409		{object}	api.ConflictErrorResponse							"Product with generated public ID already exists"
//	@Failure		500		{object}	api.InternalServerErrorResponse						"Internal server error"
//	@Success		201		{object}	api.DataResponse{data=AdminProductDetailsResponse}	"Created draft product details"
//	@Router			/admin/products [post]
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
		Description: body.Description,
		CategoryID:  categoryID,
		ProductType: body.ProductType,
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
	Slug                   string                        `json:"slug"`
	Title                  string                        `json:"title"`
	Brand                  *string                       `json:"brand,omitempty"`
	Thumbnail              *string                       `json:"thumbnail,omitempty"`
	Price                  *string                       `json:"price,omitempty" example:"$24.99"`
	ExtractedPrice         *float64                      `json:"extractedPrice,omitempty" example:"24.99"`
	OriginalPrice          *string                       `json:"originalPrice,omitempty" example:"$29.99"`
	ExtractedOriginalPrice *float64                      `json:"extractedOriginalPrice,omitempty" example:"29.99"`
	Currency               *string                       `json:"currency,omitempty" example:"USD"`
	Rating                 *ProductRatingSummaryResponse `json:"rating,omitempty"`
}

func mapProductListItemResponse(item *ProductCardReadModel) ProductListItemResponse {
	res := ProductListItemResponse{
		Slug:     item.Slug,
		Title:    item.Title,
		Currency: item.Currency,
	}

	curr := "USD"
	if item.Currency != nil && *item.Currency != "" {
		curr = *item.Currency
	}

	if item.Price != nil {
		priceStr := utils.FormatPrice(*item.Price, curr)
		extPrice := utils.ExtractPrice(*item.Price)
		res.Price = &priceStr
		res.ExtractedPrice = &extPrice
	}

	if item.CrossedOutPrice != nil {
		origPriceStr := utils.FormatPrice(*item.CrossedOutPrice, curr)
		extOrigPrice := utils.ExtractPrice(*item.CrossedOutPrice)
		res.OriginalPrice = &origPriceStr
		res.ExtractedOriginalPrice = &extOrigPrice
	}

	if item.Brand != nil {
		res.Brand = &item.Brand.Name
	}

	if item.Thumbnail != nil {
		res.Thumbnail = &item.Thumbnail.PublicURL
	}

	if item.Rating != nil {
		res.Rating = &ProductRatingSummaryResponse{
			AverageRating: item.Rating.AverageRating,
			ReviewCount:   item.Rating.ReviewCount,
			Distribution:  item.Rating.Distribution,
		}
	}

	return res
}

// GetAllProducts godoc
//
//	@Summary		List active published products
//	@Description	Returns a paginated list of active, published products for customer browsing. Soft-deleted and draft products are hidden.
//	@Tags			Products
//	@Produce		json
//	@Param			q	query		pagination.ListQuery														true	"Pagination, search query, and sorting parameters"
//	@Failure		400	{object}	api.BadRequestErrorResponse													"Invalid query parameters"
//	@Failure		500	{object}	api.InternalServerErrorResponse												"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]ProductListItemResponse,meta=pagination.Page}	"Paginated list of active products"
//	@Router			/products [get]
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
//
//	@Summary		List all products for management (Admin)
//	@Description	Returns a paginated list of all products including draft, published, archived, and soft-deleted records for administrator catalog management.
//	@Tags			Admin Products
//	@Produce		json
//	@Param			q	query		pagination.ListQuery												true	"Pagination, search query, and sorting parameters"
//	@Failure		400	{object}	api.BadRequestErrorResponse											"Invalid query parameters"
//	@Failure		500	{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]model.Product,meta=pagination.Page}	"Paginated list of all products including deleted"
//	@Router			/admin/products [get]
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

// ListProductReviews godoc
//
//	@Summary		List reviews for a specific product
//	@Description	Returns a paginated list of approved reviews for a given product.
//	@Tags			Products
//	@Produce		json
//	@Param			slug	path		string	true	"Product Slug"
//	@Param			q	query		pagination.ListQuery												true	"Pagination, search query, and sorting parameters"
//	@Failure		400	{object}	api.BadRequestErrorResponse											"Invalid query parameters"
//	@Failure		500	{object}	api.InternalServerErrorResponse										"Internal server error"
//	@Success		200	{object}	api.PaginatedResponse{data=[]ProductReviewResponse,meta=pagination.Page}	"Paginated list of product reviews"
//	@Router			/products/{slug}/reviews [get]
func (h *ProductHandler) ListProductReviews(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		_ = c.Error(apierr.ErrValidationFailed("slug is required"))
		return
	}

	details, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		_ = c.Error(err)
		return
	}
	productID := details.Product.ID

	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	status := model.ReviewStatusApproved

	result, err := h.service.reviewService.ListReviews(c.Request.Context(), q, &productID, nil, &status)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var responses []ProductReviewResponse
	for _, rv := range result.Items {
		responses = append(responses, ProductReviewResponse{
			ID:          rv.ID.String(),
			ProductID:   rv.ProductID.String(),
			UserID:      rv.UserID.String(),
			OrderItemID: rv.OrderItemID.String(),
			Rating:      rv.Rating,
			Title:       rv.Title,
			Body:        rv.Body,
			Status:      rv.Status.String(),
			CreatedAt:   rv.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   rv.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

// GetProductRatingSummary godoc
//
//	@Summary		Get rating summary for a product
//	@Description	Returns average rating, total review count, and star distribution for approved reviews.
//	@Tags			Products
//	@Produce		json
//	@Param			slug	path		string													true	"Product Slug"
//	@Success		200		{object}	api.DataResponse{data=ProductRatingSummaryResponse}	"Product rating summary"
//	@Failure		400		{object}	api.BadRequestErrorResponse								"Invalid slug"
//	@Failure		404		{object}	api.NotFoundErrorResponse								"Product not found"
//	@Failure		500		{object}	api.InternalServerErrorResponse							"Internal server error"
//	@Router			/products/{slug}/reviews/summary [get]
func (h *ProductHandler) GetProductRatingSummary(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		_ = c.Error(apierr.ErrValidationFailed("slug is required"))
		return
	}

	details, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		_ = c.Error(err)
		return
	}

	summary, err := h.service.reviewService.GetRatingSummary(c.Request.Context(), details.Product.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := ProductRatingSummaryResponse{
		AverageRating: summary.AverageRating,
		ReviewCount:   summary.ReviewCount,
		Distribution:  summary.Distribution,
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: res})
}

// GetProductByID godoc
//
//	@Summary		Get product by internal UUID
//	@Description	Retrieves full product details by its unique internal UUID.
//	@Tags			Admin Products
//	@Produce		json
//	@Param			id	path		string												true	"Internal Product UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse							"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse							"Product not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse						"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=AdminProductDetailsResponse}	"Product details"
//	@Router			/admin/products/{id} [get]
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

// GetProductBySlug godoc
//
//	@Summary		Get product by slug
//	@Description	Retrieve a single product by its public slug (for storefronts)
//	@Tags			Products
//	@Produce		json
//	@Param			slug	path		string													true	"Product Slug"
//	@Success		200		{object}	api.DataResponse{data=ProductDetailsResponse}			"Product details"
//	@Failure		400		{object}	api.BadRequestErrorResponse								"Invalid slug"
//	@Failure		404		{object}	api.NotFoundErrorResponse								"Product not found"
//	@Failure		500		{object}	api.InternalServerErrorResponse							"Internal server error"
//	@Router			/products/{slug} [get]
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		validation.ValidationError(c, errors.New("slug is required"))
		return
	}

	details, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		_ = c.Error(err)
		return
	}

	res := ProductDetailsResponse{
		ProductResponse: mapProductResponse(details.Product),
		Variants:        make([]ProductVariantResponse, 0),
		Tags:            make([]ProductTagSummary, 0),
	}

	if details.Brand != nil {
		res.Brand = &ProductBrandSummary{
			ID:   details.Brand.ID.String(),
			Name: details.Brand.Name,
			Link: details.Brand.Link,
		}
	}


	if len(details.Variants) > 0 {
		variantRes := make([]ProductVariantResponse, 0, len(details.Variants))
		for _, v := range details.Variants {
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

			vr := ProductVariantResponse{
				ID:                     v.ID.String(),
				SKU:                    v.SKU,
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
			if details.Stock != nil {
				if s, ok := details.Stock[v.ID]; ok && s != nil {
					vr.Stock = &PublicStockSummary{
						AvailableQuantity: s.AvailableQuantity,
						IsInStock:         s.IsInStock,
					}
				}
			}
			variantRes = append(variantRes, vr)

			if v.IsDefault || res.Price == nil {
				res.Price = priceStr
				res.ExtractedPrice = extPrice
				res.OriginalPrice = origPriceStr
				res.ExtractedOriginalPrice = extOrigPrice
				res.Currency = v.Currency
			}
		}
		res.Variants = variantRes
	}

	if len(details.Tags) > 0 {
		tagRes := make([]ProductTagSummary, 0, len(details.Tags))
		for _, t := range details.Tags {
			tagRes = append(tagRes, ProductTagSummary{
				ID:   t.ID.String(),
				Name: t.Name,
			})
		}
		res.Tags = tagRes
	}

	summary, err := h.service.reviewService.GetRatingSummary(c.Request.Context(), details.Product.ID)
	if err == nil && summary != nil {
		res.Rating = &ProductRatingSummaryResponse{
			AverageRating: summary.AverageRating,
			ReviewCount:   summary.ReviewCount,
			Distribution:  summary.Distribution,
		}
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}

// UpdateProduct godoc
//
//	@Summary		Update product attributes
//	@Description	Updates specific fields of an existing product such as title, description, category, brand, or publication status.
//	@Tags			Admin Products
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string												true	"Product UUID"	format(uuid)
//	@Param			body	body		UpdateProductRequest								true	"Fields to update (title, description, brand_id, category_id, status, highlights)"
//	@Failure		400		{object}	api.BadRequestErrorResponse							"Validation error or invalid UUID format"
//	@Failure		404		{object}	api.NotFoundErrorResponse							"Product, brand, or category not found"
//	@Failure		500		{object}	api.InternalServerErrorResponse						"Internal server error"
//	@Success		200		{object}	api.DataResponse{data=AdminProductDetailsResponse}	"Updated product details"
//	@Router			/admin/products/{id} [patch]
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
		Highlights:  body.Highlights,
		ProductType: body.ProductType,
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
//
//	@Summary		Publish a draft product
//	@Description	Transitions a product status to 'published' so it becomes visible to customers. Requires the product to have at least one active variant.
//	@Tags			Admin Products
//	@Produce		json
//	@Param			id	path		string							true	"Product UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Product lacks active variants or invalid UUID"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Product not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.MessageResponse				"Publication confirmation message"
//	@Router			/admin/products/{id}/publish [post]
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
//
//	@Summary		Archive a product
//	@Description	Transitions a product status to 'archived', hiding it from public store front listings while retaining all historical sales and inventory data.
//	@Tags			Admin Products
//	@Produce		json
//	@Param			id	path		string							true	"Product UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Product not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.MessageResponse				"Archival confirmation message"
//	@Router			/admin/products/{id}/archive [post]
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
//
//	@Summary		Soft-delete a product
//	@Description	Marks a product as soft-deleted (`deleted_at = NOW()`), removing it from active administrative and customer listings.
//	@Tags			Admin Products
//	@Produce		json
//	@Param			id	path		string							true	"Product UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Product not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		204													"Product soft-deleted successfully"
//	@Router			/admin/products/{id} [delete]
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


// UploadProductThumbnail godoc
//
//	@Summary	upload thumbnail image for a product
//	@Tags		Admin Product Media
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id		path		string	true	"Product ID (UUID)"	format(uuid)
//	@Param		file	formData	file	true	"Thumbnail File"
//	@Failure	400		{object}	api.BadRequestErrorResponse
//	@Failure	500		{object}	api.InternalServerErrorResponse
//	@Success	200		{object}	api.DataResponse{data=StorageObjectResponse}
//	@Router		/admin/products/{id}/thumbnail [post]
func (h *ProductHandler) UploadProductThumbnail(c *gin.Context) {
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

	obj, err := h.service.UploadProductThumbnail(c.Request.Context(), productID, mfile)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: StorageObjectResponse{
			PublicURL:   obj.PublicURL,
			ContentType: obj.ContentType,
			FileSize:    obj.FileSize,
		},
	})
}


// --- Cross-Domain Delegate Handlers ---

// CreateProductVariant godoc
//
//	@Summary	create a variant under a product
//	@Tags		Admin Product Variants
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string						true	"Product ID (UUID)"	format(uuid)
//	@Param		body	body		CreateProductVariantRequest	true	"Variant Payload"
//	@Failure	400		{object}	api.BadRequestErrorResponse
//	@Failure	500		{object}	api.InternalServerErrorResponse
//	@Success	201		{object}	api.DataResponse{data=variant.AdminVariantResponse}
//	@Router		/admin/products/{id}/variants [post]
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

	c.JSON(http.StatusCreated, api.DataResponse{
		Data: variant.AdminVariantResponse{
			ID:                     v.ID.String(),
			SKU:                    v.SKU,
			ProductID:              v.ProductID.String(),
			Title:                  v.Title,
			Price:                  priceStr,
			ExtractedPrice:         extPrice,
			OriginalPrice:          origPriceStr,
			ExtractedOriginalPrice: extOrigPrice,
			Currency:               v.Currency,
			Attributes:             v.Attributes,
			IsDefault:              v.IsDefault,
			CreatedAt:              v.CreatedAt.Format(time.RFC3339),
			UpdatedAt:              v.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// SetDefaultVariant godoc
//
//	@Summary	set a variant as the default for a product
//	@Tags		Admin Product Variants
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string						true	"Product ID (UUID)"	format(uuid)
//	@Param		body	body		SetDefaultVariantRequest	true	"Variant to set as default"
//	@Failure	400		{object}	api.BadRequestErrorResponse
//	@Failure	404		{object}	api.NotFoundErrorResponse
//	@Failure	500		{object}	api.InternalServerErrorResponse
//	@Success	200		{object}	api.MessageResponse
//	@Router		/admin/products/{id}/variants/default [post]
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
//
//	@Summary	replace all tag assignments for a product
//	@Tags		Admin Products
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string					true	"Product ID (UUID)"	format(uuid)
//	@Param		body	body		PutProductTagsRequest	true	"Tag IDs to assign"
//	@Failure	400		{object}	api.BadRequestErrorResponse
//	@Failure	500		{object}	api.InternalServerErrorResponse
//	@Success	200		{object}	api.MessageResponse
//	@Router		/admin/products/{id}/tags [put]
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
//
//	@Summary	update the category of a product
//	@Tags		Admin Products
//	@Accept		json
//	@Produce	json
//	@Param		id		path		string						true	"Product ID (UUID)"	format(uuid)
//	@Param		body	body		PutProductCategoryRequest	true	"Category ID to assign"
//	@Failure	400		{object}	api.BadRequestErrorResponse
//	@Failure	404		{object}	api.NotFoundErrorResponse
//	@Failure	500		{object}	api.InternalServerErrorResponse
//	@Success	200		{object}	api.MessageResponse
//	@Router		/admin/products/{id}/categories [put]
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
