package cart

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
)

type CartHandler struct {
	cartService *CartService
}

func NewHandler(cartService *CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

type AddItemRequest struct {
	// Product Variant UUID to add to the cart
	VariantID uuid.UUID `json:"variantId" binding:"required" example:"70000000-0000-0000-0000-000000000001"`
	// Quantity of units to add (must be at least 1)
	Quantity int `json:"quantity" binding:"required,gt=0" example:"2"`
}

type UpdateQuantityRequest struct {
	// New total quantity desired for this cart item (must be at least 1)
	Quantity int `json:"quantity" binding:"required,gt=0" example:"3"`
}

type CartSummaryResponse struct {
	// Sum total of all items before discounts/tax in cents
	Subtotal int64 `json:"subtotal" example:"499800"`
	// Applied discount deductions in cents
	Discount int64 `json:"discount" example:"0"`
	// Estimated shipping cost in cents
	Shipping int64 `json:"shipping" example:"0"`
	// Estimated tax amount in cents
	Tax int64 `json:"tax" example:"0"`
	// Final calculated order total (subtotal - discount + shipping + tax) in cents
	Total int64 `json:"total" example:"499800"`
}

type CartItemResponse struct {
	// Unique Cart Item record UUID
	ID string `json:"id" example:"030553cd-712a-4950-913f-6c26fdd6a2b5"`
	// Parent Product UUID
	ProductID string `json:"productId" example:"60000000-0000-0000-0000-000000000001"`
	// Selected SKU Variant UUID
	VariantID string `json:"variantId" example:"70000000-0000-0000-0000-000000000001"`
	// Formatted line item title (Product Title + Variant Title)
	Title string `json:"title" example:"MacBook Pro 16\" - Space Black, 18GB RAM, 512GB SSD"`
	// Thumbnail preview image URL
	ThumbnailURL string `json:"thumbnailUrl" example:"https://example.com/thumbnail.png"`
	// Selected quantity units in cart
	Quantity int `json:"quantity" example:"2"`
	// Unit price per single item in cents
	UnitPrice int64 `json:"unitPrice" example:"249900"`
	// Line item subtotal (quantity * unitPrice) in cents
	Subtotal int64 `json:"subtotal" example:"499800"`
	// Live inventory availability check: true if inventory >= quantity
	InStock bool `json:"inStock" example:"true"`
}

type CartResponse struct {
	// Cart UUID
	ID string `json:"id" example:"c23c12b7-37ff-49e6-a70f-1dcd122a0b99"`
	// Active store currency code
	Currency string `json:"currency" example:"USD"`
	// List of all items currently in cart
	Items []CartItemResponse `json:"items"`
	// Monetary calculation breakdown
	Summary CartSummaryResponse `json:"summary"`
	// Total count of distinct units in cart
	ItemCount int `json:"itemCount" example:"2"`
	// Timestamp of the most recent cart mutation (RFC3339)
	UpdatedAt string `json:"updatedAt" example:"2026-08-23T16:26:25Z"`
}

func mapCartResponse(cart *model.Cart) CartResponse {
	if cart == nil {
		return CartResponse{Items: []CartItemResponse{}}
	}

	var subtotal int64
	var itemCount int
	cartCurrency := "USD" // Default fallback

	res := CartResponse{
		ID:        cart.ID.String(),
		UpdatedAt: cart.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Items:     make([]CartItemResponse, 0, len(cart.Items)),
	}

	for _, item := range cart.Items {
		if item.Currency != "" {
			cartCurrency = item.Currency
		}

		itemSubtotal := int64(item.Quantity) * item.PriceAtPurchase
		subtotal += itemSubtotal
		itemCount += item.Quantity

		itemRes := CartItemResponse{
			ID:        item.ID.String(),
			Quantity:  item.Quantity,
			UnitPrice: item.PriceAtPurchase,
			Subtotal:  itemSubtotal,
			InStock:   item.InStock,
		}

		if item.Variant != nil {
			itemRes.VariantID = item.Variant.ID.String()
			itemRes.Title = item.Variant.Title
			if item.Variant.Price != nil {
				itemRes.UnitPrice = *item.Variant.Price
				itemSubtotal = int64(item.Quantity) * (*item.Variant.Price)
				itemRes.Subtotal = itemSubtotal
			}
		} else {
			itemRes.VariantID = item.VariantID.String()
		}

		if item.Product != nil {
			itemRes.ProductID = item.Product.ID.String()
			if itemRes.Title == "" {
				itemRes.Title = item.Product.Title
			} else {
				itemRes.Title = item.Product.Title + " - " + itemRes.Title
			}
		}

		itemRes.ThumbnailURL = item.ThumbnailURL

		res.Items = append(res.Items, itemRes)
	}

	res.Currency = cartCurrency
	res.ItemCount = itemCount
	res.Summary = CartSummaryResponse{
		Subtotal: subtotal,
		Discount: 0,
		Shipping: 0,        // Placeholder
		Tax:      0,        // Placeholder
		Total:    subtotal, // Assuming no tax/shipping yet
	}

	return res
}

const GuestSessionCookieMaxAge = 30 * 24 * 3600 // 30 days

// extractUserAndSession attempts to identify the current cart owner.
// It checks the gin context for an authenticated user ID.
// If the user is unauthenticated, it checks for an existing session ID in headers or cookies.
// If neither exists, it generates a new anonymous session ID and sets it as a cookie for future requests.
func (h *CartHandler) extractUserAndSession(c *gin.Context) (*uuid.UUID, *string) {
	var userID *uuid.UUID
	var sessionID *string

	// Check if user is authenticated (set by auth middleware)
	if uVal, exists := c.Get("userID"); exists {
		if uid, ok := uVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Look for existing session ID in header, then fallback to cookie
	sess := c.GetHeader("X-Session-ID")
	if sess == "" {
		if cookieSess, err := c.Cookie("session_id"); err == nil && cookieSess != "" {
			sess = cookieSess
		}
	}

	// Use existing session if found, otherwise create a new guest session if not authenticated
	if sess != "" {
		sessionID = &sess
	} else if userID == nil {
		guestID := fmt.Sprintf("sess_%s", uuid.New().String())
		c.SetCookie("session_id", guestID, GuestSessionCookieMaxAge, "/", "", false, true)
		sessionID = &guestID
	}

	return userID, sessionID
}

// GetCart godoc
//
//	@Summary		Get shopping cart
//	@Description	Fetches the current user's or guest session's shopping cart and item details.
//	@Tags			Cart
//	@Produce		json
//	@Param			X-Session-ID	header		string								false	"Guest Session ID"
//	@Success		200				{object}	api.DataResponse{data=CartResponse}	"Cart details"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Invalid input or missing session"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// AddItem godoc
//
//	@Summary		Add item to cart
//	@Description	Adds a product variant with specified quantity to the shopping cart.
//	@Tags			Cart
//	@Accept			json
//	@Produce		json
//	@Param			X-Session-ID	header		string								false	"Guest Session ID"
//	@Param			body			body		AddItemRequest						true	"Item variant and quantity"
//	@Success		200				{object}	api.DataResponse{data=CartResponse}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Validation error"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Variant not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items [post]
func (h *CartHandler) AddItem(c *gin.Context) {
	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.AddItem(c.Request.Context(), userID, sessionID, req.VariantID, req.Quantity)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// UpdateItemQuantity godoc
//
//	@Summary		Update cart item quantity
//	@Description	Updates the quantity of a specific item in the cart.
//	@Tags			Cart
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string								true	"Cart Item UUID"
//	@Param			X-Session-ID	header		string								false	"Guest Session ID"
//	@Param			body			body		UpdateQuantityRequest				true	"New quantity"
//	@Success		200				{object}	api.DataResponse{data=CartResponse}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Validation error or invalid UUID"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Cart item not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items/{id} [patch]
func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("Invalid item ID format").WithCode(apierr.CodeInvalidInput).Wrap(err))
		return
	}

	var req UpdateQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.UpdateCartItemQuantity(c.Request.Context(), userID, sessionID, itemID, req.Quantity)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// RemoveItem godoc
//
//	@Summary		Remove item from cart
//	@Description	Removes a single item from the cart by item ID.
//	@Tags			Cart
//	@Produce		json
//	@Param			id				path		string								true	"Cart Item UUID"
//	@Param			X-Session-ID	header		string								false	"Guest Session ID"
//	@Success		200				{object}	api.DataResponse{data=CartResponse}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Invalid item ID"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Cart item not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items/{id} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("Invalid item ID format").WithCode(apierr.CodeInvalidInput).Wrap(err))
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.RemoveCartItem(c.Request.Context(), userID, sessionID, itemID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// ClearCart godoc
//
//	@Summary		Clear cart
//	@Description	Removes all items from the current cart.
//	@Tags			Cart
//	@Produce		json
//	@Param			X-Session-ID	header	string	false	"Guest Session ID"
//	@Success		204				"Cart cleared"
//	@Failure		500				{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, sessionID := h.extractUserAndSession(c)

	err := h.cartService.ClearCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
