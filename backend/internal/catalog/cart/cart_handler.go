package cart

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
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
	VariantID string `json:"variant_id" binding:"required" example:"nano_id_string"`
	Quantity  int    `json:"quantity" binding:"required,gt=0" example:"2"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0" example:"5"`
}

type CartSummaryResponse struct {
	Subtotal int64 `json:"subtotal"`
	Discount int64 `json:"discount"`
	Shipping int64 `json:"shipping"`
	Tax      int64 `json:"tax"`
	Total    int64 `json:"total"`
}

type CartItemResponse struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	VariantID    string `json:"variant_id"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"`
	Subtotal     int64  `json:"subtotal"`
	InStock      bool   `json:"in_stock"`
}

type CartResponse struct {
	ID        string              `json:"id"`
	Currency  string              `json:"currency"`
	Items     []CartItemResponse  `json:"items"`
	Summary   CartSummaryResponse `json:"summary"`
	ItemCount int                 `json:"item_count"`
	UpdatedAt string              `json:"updated_at"`
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
			InStock:   true, // Assuming in stock for now, can be updated with inventory logic
		}

		if item.Variant != nil {
			itemRes.VariantID = item.Variant.PublicID
			itemRes.Title = item.Variant.Title
		}

		if item.Product != nil {
			itemRes.ProductID = item.Product.PublicID
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
		Shipping: 0, // Placeholder
		Tax:      0, // Placeholder
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
//	@Success		200				{object}	api.DataResponse{data=model.Cart}	"Cart details"
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
//	@Success		200				{object}	api.DataResponse{data=model.Cart}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Validation error"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Variant not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items [post]
func (h *CartHandler) AddItem(c *gin.Context) {
	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.ErrBadRequest("Invalid input payload").WithCode(apierr.CodeValidationFailed).Wrap(err))
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
//	@Success		200				{object}	api.DataResponse{data=model.Cart}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Validation error or invalid UUID"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Cart item not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items/{id} [patch]
func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	variantPublicID := c.Param("id")
	if variantPublicID == "" {
		_ = c.Error(apierr.ErrBadRequest("Variant ID is required").WithCode(apierr.CodeInvalidInput))
		return
	}

	var req UpdateQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.ErrBadRequest("Invalid input payload").WithCode(apierr.CodeValidationFailed).Wrap(err))
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.UpdateCartItemQuantity(c.Request.Context(), userID, sessionID, variantPublicID, req.Quantity)
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
//	@Success		200				{object}	api.DataResponse{data=model.Cart}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Invalid item ID"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Cart item not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items/{id} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	variantPublicID := c.Param("id")
	if variantPublicID == "" {
		_ = c.Error(apierr.ErrBadRequest("Variant ID is required").WithCode(apierr.CodeInvalidInput))
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.RemoveCartItem(c.Request.Context(), userID, sessionID, variantPublicID)
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
