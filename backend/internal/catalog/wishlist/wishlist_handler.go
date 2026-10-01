package wishlist

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
)

type WishlistHandler struct {
	service *WishlistService
}

func NewHandler(s *WishlistService) *WishlistHandler {
	return &WishlistHandler{service: s}
}

type AddToWishlistRequest struct {
	// Product UUID to save to wishlist
	ProductID string `json:"productId" binding:"required,uuid" example:"60000000-0000-0000-0000-000000000001"`
	// Optional specific product variant UUID to save
	VariantID *string `json:"variantId,omitempty" binding:"omitempty,uuid" example:"70000000-0000-0000-0000-000000000001"`
}

type WishlistProductResponse struct {
	// Product database UUID
	ID string `json:"id" example:"60000000-0000-0000-0000-000000000001"`
	// URL-friendly unique slug
	Slug string `json:"slug" example:"macbook-pro-16"`
	// Product display title
	Title string `json:"title" example:"MacBook Pro 16\""`
	// Detailed product description
	Description *string `json:"description,omitempty" example:"Supercharged by M3 Pro or M3 Max."`
	// Product type: simple or variable
	ProductType string `json:"productType" example:"variable"`
	// Brand display name
	BrandName *string `json:"brandName,omitempty" example:"Apple"`
	// Primary category name
	CategoryName *string `json:"categoryName,omitempty" example:"Laptops"`
	// Product display image public URL
	ThumbnailURL *string `json:"thumbnailUrl,omitempty" example:"https://example.com/thumb.jpg"`
	// Formatted active display price
	Price *string `json:"price,omitempty" example:"$2499.00"`
	// Numeric active display price
	ExtractedPrice *float64 `json:"extractedPrice,omitempty" example:"2499.00"`
	// Formatted original crossed-out price
	OriginalPrice *string `json:"originalPrice,omitempty" example:"$2999.00"`
	// Numeric original crossed-out price
	ExtractedOriginalPrice *float64 `json:"extractedOriginalPrice,omitempty" example:"2999.00"`
	// Currency ISO code
	Currency *string `json:"currency,omitempty" example:"USD"`
	// Salable inventory availability
	InStock bool `json:"inStock" example:"true"`
	// Variant specific details
	VariantID *string `json:"variantId,omitempty" example:"70000000-0000-0000-0000-000000000001"`
	VariantTitle *string `json:"variantTitle,omitempty" example:"Space Black"`
	VariantSKU *string `json:"variantSku,omitempty" example:"SKU-MBP-16"`
}

type WishlistItemResponse struct {
	// Wishlist Item record UUID
	ID string `json:"id" example:"030553cd-712a-4950-913f-6c26fdd6a2b5"`
	// Saved Product UUID
	ProductID string `json:"productId" example:"60000000-0000-0000-0000-000000000001"`
	// Saved Variant UUID if specific variant was selected
	VariantID *string `json:"variantId,omitempty" example:"70000000-0000-0000-0000-000000000001"`
	// Customer UUID owning this wishlist item
	UserID string `json:"userId" example:"10000000-0000-0000-0000-000000000002"`
	// Timestamp when item was added to wishlist (RFC3339)
	CreatedAt string `json:"createdAt" example:"2026-08-23T16:26:25Z"`
	// Hydrated Product details
	Product WishlistProductResponse `json:"product"`
}

type WishlistCheckResponse struct {
	// True if the product is currently saved in the user's wishlist
	InWishlist bool `json:"inWishlist" example:"true"`
	// Wishlist item UUID if present
	WishlistItemID *string `json:"wishlistItemId,omitempty" example:"030553cd-712a-4950-913f-6c26fdd6a2b5"`
}

type WishlistCountResponse struct {
	// Total count of items currently in the user's wishlist
	Count int `json:"count" example:"5"`
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, apierr.ErrUnauthorized("unauthorized")
	}
	if id, ok := userIDVal.(uuid.UUID); ok {
		return id, nil
	}
	if idStr, ok := userIDVal.(string); ok {
		id, err := uuid.Parse(idStr)
		if err == nil {
			return id, nil
		}
	}
	return uuid.Nil, apierr.ErrUnauthorized("invalid user identifier in context")
}

func mapWishlistItemResponse(item *model.WishlistItem) WishlistItemResponse {
	var (
		priceStr         *string
		extractedPrice   *float64
		origPriceStr     *string
		extractedOrigVal *float64
		currency         = "USD"
		variantIDStr     *string
	)

	if item.Currency != nil && *item.Currency != "" {
		currency = *item.Currency
	}

	if item.Price != nil {
		p := utils.FormatPrice(*item.Price, currency)
		priceStr = &p
		ep := utils.ExtractPrice(*item.Price)
		extractedPrice = &ep
	}

	if item.CrossedOutPrice != nil {
		op := utils.FormatPrice(*item.CrossedOutPrice, currency)
		origPriceStr = &op
		eop := utils.ExtractPrice(*item.CrossedOutPrice)
		extractedOrigVal = &eop
	}

	if item.VariantID != nil {
		vStr := item.VariantID.String()
		variantIDStr = &vStr
	}

	return WishlistItemResponse{
		ID:        item.ID.String(),
		ProductID: item.ProductID.String(),
		VariantID: variantIDStr,
		UserID:    item.UserID.String(),
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		Product: WishlistProductResponse{
			ID:                     item.ProductID.String(),
			Slug:                   item.ProductSlug,
			Title:                  item.ProductTitle,
			Description:            item.ProductDescription,
			ProductType:            item.ProductType.String(),
			BrandName:              item.BrandName,
			CategoryName:           item.CategoryName,
			ThumbnailURL:           item.ThumbnailURL,
			Price:                  priceStr,
			ExtractedPrice:         extractedPrice,
			OriginalPrice:          origPriceStr,
			ExtractedOriginalPrice: extractedOrigVal,
			Currency:               &currency,
			InStock:                item.InStock,
			VariantID:              variantIDStr,
			VariantTitle:           item.VariantTitle,
			VariantSKU:             item.VariantSKU,
		},
	}
}

// AddToWishlist godoc
//
//	@Summary		Add a product to wishlist
//	@Description	Saves a product to the authenticated user's wishlist.
//	@Tags			Wishlist
//	@Accept			json
//	@Produce		json
//	@Param			request	body		AddToWishlistRequest					true	"Product to wishlist payload"
//	@Success		201		{object}	api.DataResponse{data=WishlistItemResponse}	"Product added to wishlist successfully"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Invalid input or product not published"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse			"Unauthorized"
//	@Failure		404		{object}	api.NotFoundErrorResponse				"Product not found"
//	@Failure		409		{object}	api.ConflictErrorResponse				"Product already in wishlist"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Router			/wishlist [post]
func (h *WishlistHandler) AddToWishlist(c *gin.Context) {
	var req AddToWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.ErrInvalidPayload("invalid request payload"))
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid product id"))
		return
	}

	var variantID *uuid.UUID
	if req.VariantID != nil && *req.VariantID != "" {
		vID, err := uuid.Parse(*req.VariantID)
		if err != nil {
			_ = c.Error(apierr.ErrValidationFailed("invalid variant id"))
			return
		}
		variantID = &vID
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	item, err := h.service.AddToWishlist(c.Request.Context(), userID, productID, variantID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, api.DataResponse{Data: mapWishlistItemResponse(item)})
}

// GetMyWishlist godoc
//
//	@Summary		List wishlist items
//	@Description	Returns a paginated list of products saved in the authenticated user's wishlist.
//	@Tags			Wishlist
//	@Produce		json
//	@Param			q	query		pagination.ListQuery														true	"Pagination and sorting parameters"
//	@Success		200	{object}	api.PaginatedResponse{data=[]WishlistItemResponse,meta=pagination.Page}	"Wishlist items retrieved successfully"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse												"Unauthorized"
//	@Failure		500	{object}	api.InternalServerErrorResponse												"Internal server error"
//	@Router			/wishlist [get]
func (h *WishlistHandler) GetMyWishlist(c *gin.Context) {
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	result, err := h.service.GetMyWishlist(c.Request.Context(), userID, q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]WishlistItemResponse, len(result.Items))
	for i, item := range result.Items {
		responses[i] = mapWishlistItemResponse(item)
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

// GetWishlistCount godoc
//
//	@Summary		Get wishlist item count
//	@Description	Returns the total number of items saved in the authenticated user's wishlist.
//	@Tags			Wishlist
//	@Produce		json
//	@Success		200	{object}	api.DataResponse{data=WishlistCountResponse}	"Wishlist count retrieved successfully"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse					"Unauthorized"
//	@Failure		500	{object}	api.InternalServerErrorResponse					"Internal server error"
//	@Router			/wishlist/count [get]
func (h *WishlistHandler) GetWishlistCount(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	count, err := h.service.GetWishlistCount(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: WishlistCountResponse{Count: count}})
}

// CheckInWishlist godoc
//
//	@Summary		Check if product is in wishlist
//	@Description	Checks whether a specific product (or variant) is saved in the authenticated user's wishlist.
//	@Tags			Wishlist
//	@Produce		json
//	@Param			productId	path		string											true	"Product UUID"
//	@Param			variantId	query		string											false	"Optional Variant UUID"
//	@Success		200			{object}	api.DataResponse{data=WishlistCheckResponse}	"Wishlist status checked successfully"
//	@Failure		400			{object}	api.BadRequestErrorResponse						"Invalid product id"
//	@Failure		401			{object}	api.UnauthorizedErrorResponse					"Unauthorized"
//	@Failure		500			{object}	api.InternalServerErrorResponse					"Internal server error"
//	@Router			/wishlist/check/{productId} [get]
func (h *WishlistHandler) CheckInWishlist(c *gin.Context) {
	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("invalid product id").WithCode(apierr.CodeInvalidInput))
		return
	}

	var variantID *uuid.UUID
	if vIDStr := c.Query("variantId"); vIDStr != "" {
		vID, err := uuid.Parse(vIDStr)
		if err != nil {
			_ = c.Error(apierr.ErrBadRequest("invalid variant id").WithCode(apierr.CodeInvalidInput))
			return
		}
		variantID = &vID
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	inWishlist, itemID, err := h.service.CheckInWishlist(c.Request.Context(), userID, productID, variantID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var itemIDStr *string
	if itemID != nil {
		s := itemID.String()
		itemIDStr = &s
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: WishlistCheckResponse{
		InWishlist:     inWishlist,
		WishlistItemID: itemIDStr,
	}})
}

// RemoveItem godoc
//
//	@Summary		Remove item from wishlist
//	@Description	Deletes a specific wishlist item by its record UUID.
//	@Tags			Wishlist
//	@Produce		json
//	@Param			id	path		string					true	"Wishlist Item UUID"
//	@Success		200	{object}	api.MessageResponse		"Item removed from wishlist successfully"
//	@Failure		400	{object}	api.BadRequestErrorResponse	"Invalid item id"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		404	{object}	api.NotFoundErrorResponse	"Wishlist item not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/wishlist/items/{id} [delete]
func (h *WishlistHandler) RemoveItem(c *gin.Context) {
	idStr := c.Param("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("invalid wishlist item id").WithCode(apierr.CodeInvalidInput))
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.service.RemoveItem(c.Request.Context(), userID, itemID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Item removed from wishlist successfully"})
}

// RemoveByProductID godoc
//
//	@Summary		Remove product from wishlist
//	@Description	Deletes a product from the user's wishlist by the product UUID (and optional variantId).
//	@Tags			Wishlist
//	@Produce		json
//	@Param			productId	path		string					true	"Product UUID"
//	@Param			variantId	query		string					false	"Optional Variant UUID"
//	@Success		200			{object}	api.MessageResponse		"Product removed from wishlist successfully"
//	@Failure		400			{object}	api.BadRequestErrorResponse	"Invalid product id"
//	@Failure		401			{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		404			{object}	api.NotFoundErrorResponse	"Product not found in wishlist"
//	@Failure		500			{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/wishlist/products/{productId} [delete]
func (h *WishlistHandler) RemoveByProductID(c *gin.Context) {
	productIDStr := c.Param("productId")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("invalid product id").WithCode(apierr.CodeInvalidInput))
		return
	}

	var variantID *uuid.UUID
	if vIDStr := c.Query("variantId"); vIDStr != "" {
		vID, err := uuid.Parse(vIDStr)
		if err != nil {
			_ = c.Error(apierr.ErrBadRequest("invalid variant id").WithCode(apierr.CodeInvalidInput))
			return
		}
		variantID = &vID
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.service.RemoveByProductID(c.Request.Context(), userID, productID, variantID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Product removed from wishlist successfully"})
}

// ClearWishlist godoc
//
//	@Summary		Clear wishlist or delete specific item/product
//	@Description	Clears all items from the authenticated user's wishlist. If productId or itemId query parameter is provided, removes only that product or item.
//	@Tags			Wishlist
//	@Produce		json
//	@Param			productId	query		string					false	"Optional product UUID to remove"
//	@Param			variantId	query		string					false	"Optional variant UUID to remove"
//	@Param			itemId		query		string					false	"Optional item UUID to remove"
//	@Success		200			{object}	api.MessageResponse		"Wishlist updated successfully"
//	@Failure		400			{object}	api.BadRequestErrorResponse	"Invalid input"
//	@Failure		401			{object}	api.UnauthorizedErrorResponse	"Unauthorized"
//	@Failure		404			{object}	api.NotFoundErrorResponse	"Item not found"
//	@Failure		500			{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/wishlist [delete]
func (h *WishlistHandler) ClearWishlist(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if pIDStr := c.Query("productId"); pIDStr != "" {
		productID, err := uuid.Parse(pIDStr)
		if err != nil {
			_ = c.Error(apierr.ErrBadRequest("invalid product id").WithCode(apierr.CodeInvalidInput))
			return
		}
		var variantID *uuid.UUID
		if vIDStr := c.Query("variantId"); vIDStr != "" {
			vID, err := uuid.Parse(vIDStr)
			if err != nil {
				_ = c.Error(apierr.ErrBadRequest("invalid variant id").WithCode(apierr.CodeInvalidInput))
				return
			}
			variantID = &vID
		}
		if err := h.service.RemoveByProductID(c.Request.Context(), userID, productID, variantID); err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, api.MessageResponse{Message: "Product removed from wishlist successfully"})
		return
	}

	if iIDStr := c.Query("itemId"); iIDStr != "" {
		itemID, err := uuid.Parse(iIDStr)
		if err != nil {
			_ = c.Error(apierr.ErrBadRequest("invalid item id").WithCode(apierr.CodeInvalidInput))
			return
		}
		if err := h.service.RemoveItem(c.Request.Context(), userID, itemID); err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, api.MessageResponse{Message: "Item removed from wishlist successfully"})
		return
	}

	if err := h.service.ClearWishlist(c.Request.Context(), userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Wishlist cleared successfully"})
}
