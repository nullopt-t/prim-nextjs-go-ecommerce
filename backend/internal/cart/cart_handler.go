package cart

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// StoreConfigProvider exposes the effective store configuration (shipping
// threshold and flat rate) so the cart summary can quote shipping with exactly
// the same rule that checkout charges.
type StoreConfigProvider interface {
	GetConfig(ctx context.Context) (setting.Config, error)
}

type sessionCreator interface {
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		userRole string,
		sessionType model.SessionType,
		userAgent string,
		ipAddress string,
	) (*model.UserSession, error)
}

type CartHandler struct {
	cartService    *CartService
	settings       StoreConfigProvider
	sessionCreator sessionCreator
}

func NewHandler(cartService *CartService, settings ...StoreConfigProvider) *CartHandler {
	if cartService == nil {
		panic("cart: service is required")
	}
	h := &CartHandler{
		cartService: cartService,
	}
	if len(settings) > 0 {
		h.settings = settings[0]
	}
	return h
}

// SetSessionCreator wires the session service used to mint guest sessions.
func (h *CartHandler) SetSessionCreator(creator sessionCreator) {
	if creator == nil {
		panic("cart: session creator is required")
	}
	h.sessionCreator = creator
}

// resolveStoreConfig returns the effective store configuration, falling back to
// the built-in defaults when the provider is absent or errors. This mirrors
// CheckoutService.resolveConfig so the cart quote and the charged amount cannot
// diverge on a transient settings read failure.
func (h *CartHandler) resolveStoreConfig(ctx context.Context) setting.Config {
	if h.settings == nil {
		return setting.DefaultConfig()
	}
	cfg, err := h.settings.GetConfig(ctx)
	if err != nil {
		return setting.DefaultConfig()
	}
	return cfg
}

// AddItemRequest defines the request body payload for adding a product variant to the cart.
type AddItemRequest struct {
	// Product Variant UUID to add to the cart
	VariantID uuid.UUID `json:"variantId" binding:"required" example:"70000000-0000-0000-0000-000000000001"`
	// Quantity of units to add (must be between 1 and 999)
	Quantity int `json:"quantity" binding:"required,gt=0,lte=999" example:"2"`
}

// UpdateQuantityRequest defines the request body payload for modifying an item's quantity in the cart.
type UpdateQuantityRequest struct {
	// New total quantity desired for this cart item (must be between 1 and 999)
	Quantity int `json:"quantity" binding:"required,gt=0,lte=999" example:"3"`
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
	// Product Slug for storefront routing
	ProductSlug string `json:"productSlug,omitempty" example:"macbook-air-15-m3"`
	// Selected SKU Variant UUID
	VariantID string `json:"variantId" example:"70000000-0000-0000-0000-000000000001"`
	// Selected SKU Variant Color
	Color string `json:"color,omitempty" example:"Midnight"`
	// Variant attributes
	Attributes map[string]any `json:"attributes,omitempty"`
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
	// Live salable quantity available in inventory
	AvailableStock int `json:"availableStock" example:"15"`
}

type CartResponse struct {
	// Cart UUID
	ID string `json:"id" example:"c23c12b7-37ff-49e6-a70f-1dcd122a0b99"`
	// Active store currency code
	Currency string `json:"currency" example:"EGP"`
	// List of all items currently in cart
	Items []CartItemResponse `json:"items"`
	// Monetary calculation breakdown
	Summary CartSummaryResponse `json:"summary"`
	// Total count of distinct units in cart
	ItemCount int `json:"itemCount" example:"2"`
	// Timestamp of the most recent cart mutation (RFC3339)
	UpdatedAt string `json:"updatedAt" example:"2026-08-23T16:26:25Z"`
}

// mapCartResponse transforms the internal domain model.Cart into a client-facing CartResponse.
// It computes line item subtotals, aggregated cart totals, item counts, and resolves fallback titles and thumbnails.
func (h *CartHandler) mapCartResponse(ctx context.Context, cart *model.Cart) CartResponse {
	if cart == nil {
		return CartResponse{Items: []CartItemResponse{}}
	}

	var subtotal int64
	var itemCount int
	cartCurrency := "EGP" // Default fallback

	res := CartResponse{
		ID:        cart.ID.String(),
		UpdatedAt: cart.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Items:     make([]CartItemResponse, 0, len(cart.Items)),
	}

	for _, item := range cart.Items {
		if item.Currency != "" {
			cartCurrency = item.Currency
		}

		unitPrice := item.PriceAtPurchase
		if item.Variant != nil && item.Variant.Price != nil {
			unitPrice = *item.Variant.Price
		}
		itemSubtotal := int64(item.Quantity) * unitPrice
		subtotal += itemSubtotal
		itemCount += item.Quantity

		itemRes := CartItemResponse{
			ID:             strconv.FormatInt(item.ID, 10),
			Quantity:       item.Quantity,
			UnitPrice:      unitPrice,
			Subtotal:       itemSubtotal,
			InStock:        item.InStock,
			AvailableStock: item.AvailableStock,
		}

		if item.Variant != nil {
			itemRes.VariantID = item.Variant.ID.String()
			itemRes.Title = item.Variant.Title
			itemRes.Attributes = item.Variant.Attributes
			if colorVal, ok := item.Variant.Attributes["color"].(string); ok {
				itemRes.Color = colorVal
			}
		} else {
			itemRes.VariantID = item.VariantID.String()
		}

		if item.Product != nil {
			itemRes.ProductID = item.Product.ID.String()
			itemRes.ProductSlug = item.Product.Slug
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

	// Quote shipping with the same admin-configured rule checkout charges:
	// the configured flat rate below the free-shipping threshold, free at/above
	// it. Reading it here (instead of a hardcoded 9.99/150) keeps the cart total
	// and the final order total in agreement.
	cfg := h.resolveStoreConfig(ctx)
	var shippingCost int64
	if threshold := cfg.FreeShippingThresholdMinorUnits(); subtotal > 0 && subtotal < threshold {
		shippingCost = cfg.ShippingMinorUnits()
	}

	res.Summary = CartSummaryResponse{
		Subtotal: subtotal,
		Discount: 0,
		Shipping: shippingCost,
		Tax:      0,
		Total:    subtotal + shippingCost,
	}

	return res
}

// parseCartItemID resolves the database-owned numeric cart item id.
func parseCartItemID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}

// resolveSession returns the active server-side session. When create is true
// and no session exists yet, it mints a guest session and writes the canonical
// session cookie so the next request reuses the same cart owner.
func (h *CartHandler) resolveSession(c *gin.Context, create bool) (*model.UserSession, error) {
	if session, ok := middleware.GetSession(c); ok {
		return session, nil
	}
	if !create {
		return nil, nil
	}
	if h.sessionCreator == nil {
		return nil, apierr.InternalServerError("Session service is unavailable")
	}

	session, err := h.sessionCreator.CreateSession(
		c.Request.Context(),
		uuid.Nil,
		"",
		model.SessionTypeGuest,
		c.GetHeader("User-Agent"),
		c.ClientIP(),
	)
	if err != nil {
		return nil, apierr.InternalServerError("Failed to create guest session").
			Wrap(err)
	}

	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	middleware.SetSessionCookie(c, session.ID, secure)
	return session, nil
}

// cartOwner maps a session to the cart ownership columns: user_id for signed-in
// customers, session_id for guests.
func cartOwner(session *model.UserSession) (*uuid.UUID, *string) {
	if session.UserID != uuid.Nil {
		userID := session.UserID
		return &userID, nil
	}
	sessionID := session.ID
	return nil, &sessionID
}

func emptyCartResponse() CartResponse {
	return CartResponse{
		Currency: "EGP",
		Items:    []CartItemResponse{},
		Summary: CartSummaryResponse{
			Subtotal: 0,
			Discount: 0,
			Shipping: 0,
			Tax:      0,
			Total:    0,
		},
		ItemCount: 0,
	}
}

// GetCart godoc
//
//	@Summary		Get shopping cart
//	@Description	Fetches the current user's or guest session's shopping cart and item details.
//	@Tags			Cart
//	@Produce		json
//	@Success		200				{object}	api.DataResponse{data=CartResponse}	"Cart details"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Invalid input or missing session"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	session, err := h.resolveSession(c, false)
	if err != nil {
		_ = c.Error(err)
		return
	}
	// Visitors without a session have no cart to hydrate. Avoid creating an
	// empty cart or a session just to return an empty response.
	if session == nil {
		c.JSON(http.StatusOK, api.DataResponse{Data: emptyCartResponse()})
		return
	}

	userID, sessionID := cartOwner(session)

	cart, err := h.cartService.GetOrCreateCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: h.mapCartResponse(c.Request.Context(), cart)})
}

// AddItem godoc
//
//	@Summary		Add item to cart
//	@Description	Adds a product variant with specified quantity to the shopping cart.
//	@Tags			Cart
//	@Accept			json
//	@Produce		json
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

	session, err := h.resolveSession(c, true)
	if err != nil {
		_ = c.Error(err)
		return
	}
	userID, sessionID := cartOwner(session)

	cart, err := h.cartService.GetOrCreateCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	cart, err = h.cartService.AddItem(c.Request.Context(), cart.ID, req.VariantID, req.Quantity)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: h.mapCartResponse(c.Request.Context(), cart)})
}

// UpdateItemQuantity godoc
//
//	@Summary		Update cart item quantity
//	@Description	Updates the quantity of a specific item in the cart.
//	@Tags			Cart
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string								true	"Cart item id"
//	@Param			body			body		UpdateQuantityRequest				true	"New quantity"
//	@Success		200				{object}	api.DataResponse{data=CartResponse}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Validation error or invalid item ID"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Cart item not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items/{id} [patch]
func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	itemID, err := parseCartItemID(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid item ID format").WithCode(apierr.CodeInvalidInput).Wrap(err))
		return
	}

	var req UpdateQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	session, err := h.resolveSession(c, true)
	if err != nil {
		_ = c.Error(err)
		return
	}
	userID, sessionID := cartOwner(session)

	cart, err := h.cartService.GetOrCreateCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	cart, err = h.cartService.UpdateCartItemQuantity(c.Request.Context(), cart.ID, itemID, req.Quantity)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: h.mapCartResponse(c.Request.Context(), cart)})
}

// RemoveItem godoc
//
//	@Summary		Remove item from cart
//	@Description	Removes a single item from the cart by item ID.
//	@Tags			Cart
//	@Produce		json
//	@Param			id				path		string								true	"Cart item id"
//	@Success		200				{object}	api.DataResponse{data=CartResponse}	"Updated cart"
//	@Failure		400				{object}	api.BadRequestErrorResponse			"Invalid item ID"
//	@Failure		404				{object}	api.NotFoundErrorResponse			"Cart item not found"
//	@Failure		500				{object}	api.InternalServerErrorResponse		"Internal server error"
//	@Router			/cart/items/{id} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	itemID, err := parseCartItemID(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid item ID format").WithCode(apierr.CodeInvalidInput).Wrap(err))
		return
	}

	session, err := h.resolveSession(c, true)
	if err != nil {
		_ = c.Error(err)
		return
	}
	userID, sessionID := cartOwner(session)

	cart, err := h.cartService.GetOrCreateCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	cart, err = h.cartService.RemoveCartItem(c.Request.Context(), cart.ID, itemID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: h.mapCartResponse(c.Request.Context(), cart)})
}

// ClearCart godoc
//
//	@Summary		Clear cart
//	@Description	Removes all items from the current cart.
//	@Tags			Cart
//	@Produce		json
//	@Success		204				"Cart cleared"
//	@Failure		500				{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Router			/cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	session, err := h.resolveSession(c, true)
	if err != nil {
		_ = c.Error(err)
		return
	}
	userID, sessionID := cartOwner(session)

	cart, err := h.cartService.GetOrCreateCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	err = h.cartService.ClearCart(c.Request.Context(), cart.ID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
