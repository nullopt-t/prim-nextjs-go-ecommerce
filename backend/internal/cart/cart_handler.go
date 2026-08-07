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
	VariantID uuid.UUID `json:"variant_id" binding:"required" example:"96c4e462-ed4a-4fec-9115-47cbf12206a7"`
	Quantity  int       `json:"quantity" binding:"required,gt=0" example:"2"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0" example:"5"`
}

type MergeCartRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"sess_abc123xyz"`
}

type CartItemResponse struct {
	ID              string `json:"id"`
	CartID          string `json:"cart_id"`
	VariantID       string `json:"variant_id"`
	Quantity        int    `json:"quantity"`
	PriceAtPurchase int64  `json:"price_at_purchase"`
	Currency        string `json:"currency"`
	CartedAt        string `json:"carted_at"`
	Variant         any    `json:"variant,omitempty"`
}

type CartResponse struct {
	ID        string             `json:"id"`
	UserID    *string            `json:"user_id,omitempty"`
	SessionID *string            `json:"session_id,omitempty"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt string             `json:"updated_at"`
	Items     []CartItemResponse `json:"items,omitempty"`
}

func mapCartResponse(cart *model.Cart) CartResponse {
	if cart == nil {
		return CartResponse{}
	}

	res := CartResponse{
		ID:        cart.ID.String(),
		CreatedAt: cart.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: cart.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Items:     make([]CartItemResponse, 0, len(cart.Items)),
	}

	if cart.UserID != nil {
		idStr := cart.UserID.String()
		res.UserID = &idStr
	}
	res.SessionID = cart.SessionID

	for _, item := range cart.Items {
		itemRes := CartItemResponse{
			ID:              item.ID.String(),
			CartID:          item.CartID.String(),
			VariantID:       item.VariantID.String(),
			Quantity:        item.Quantity,
			PriceAtPurchase: item.PriceAtPurchase,
			Currency:        item.Currency,
			CartedAt:        item.CartedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if item.Variant != nil {
			itemRes.Variant = item.Variant
		}
		res.Items = append(res.Items, itemRes)
	}

	return res
}

func (h *CartHandler) extractUserAndSession(c *gin.Context) (*uuid.UUID, *string) {
	var userID *uuid.UUID
	var sessionID *string

	if uVal, exists := c.Get("userID"); exists {
		if uid, ok := uVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	sess := c.GetHeader("X-Session-ID")
	if sess == "" {
		if cookieSess, err := c.Cookie("session_id"); err == nil && cookieSess != "" {
			sess = cookieSess
		}
	}

	if sess != "" {
		sessionID = &sess
	} else if userID == nil {
		guestID := fmt.Sprintf("sess_guest_%s", uuid.New().String())
		c.SetCookie("session_id", guestID, 30*24*3600, "/", "", false, true)
		sessionID = &guestID
	}

	return userID, sessionID
}

// GetCart godoc
// @Summary Get shopping cart
// @Description Fetches the current user's or guest session's shopping cart and item details.
// @Tags Cart
// @Produce json
// @Param X-Session-ID header string false "Guest Session ID"
// @Success 200 {object} api.DataResponse{data=model.Cart} "Cart details"
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid input or missing session"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Router /cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// AddItem godoc
// @Summary Add item to cart
// @Description Adds a product variant with specified quantity to the shopping cart.
// @Tags Cart
// @Accept json
// @Produce json
// @Param X-Session-ID header string false "Guest Session ID"
// @Param body body AddItemRequest true "Item variant and quantity"
// @Success 200 {object} api.DataResponse{data=model.Cart} "Updated cart"
// @Failure 400 {object} api.BadRequestErrorResponse "Validation error"
// @Failure 404 {object} api.NotFoundErrorResponse "Variant not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Router /cart/items [post]
func (h *CartHandler) AddItem(c *gin.Context) {
	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apierr.ErrBadRequest("Invalid input payload").WithCode(apierr.CodeValidationFailed).Wrap(err))
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.AddItemToCart(c.Request.Context(), userID, sessionID, req.VariantID, req.Quantity)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// UpdateItemQuantity godoc
// @Summary Update cart item quantity
// @Description Updates the quantity of a specific item in the cart.
// @Tags Cart
// @Accept json
// @Produce json
// @Param id path string true "Cart Item UUID"
// @Param X-Session-ID header string false "Guest Session ID"
// @Param body body UpdateQuantityRequest true "New quantity"
// @Success 200 {object} api.DataResponse{data=model.Cart} "Updated cart"
// @Failure 400 {object} api.BadRequestErrorResponse "Validation error or invalid UUID"
// @Failure 404 {object} api.NotFoundErrorResponse "Cart item not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Router /cart/items/{id} [patch]
func (h *CartHandler) UpdateItemQuantity(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apierr.ErrBadRequest("Invalid item ID format").WithCode(apierr.CodeInvalidInput).Wrap(err))
		return
	}

	var req UpdateQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apierr.ErrBadRequest("Invalid input payload").WithCode(apierr.CodeValidationFailed).Wrap(err))
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.UpdateCartItemQuantity(c.Request.Context(), userID, sessionID, itemID, req.Quantity)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// RemoveItem godoc
// @Summary Remove item from cart
// @Description Removes a single item from the cart by item ID.
// @Tags Cart
// @Produce json
// @Param id path string true "Cart Item UUID"
// @Param X-Session-ID header string false "Guest Session ID"
// @Success 200 {object} api.DataResponse{data=model.Cart} "Updated cart"
// @Failure 400 {object} api.BadRequestErrorResponse "Invalid item ID"
// @Failure 404 {object} api.NotFoundErrorResponse "Cart item not found"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Router /cart/items/{id} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apierr.ErrBadRequest("Invalid item ID format").WithCode(apierr.CodeInvalidInput).Wrap(err))
		return
	}

	userID, sessionID := h.extractUserAndSession(c)

	cart, err := h.cartService.RemoveCartItem(c.Request.Context(), userID, sessionID, itemID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}

// ClearCart godoc
// @Summary Clear cart
// @Description Removes all items from the current cart.
// @Tags Cart
// @Produce json
// @Param X-Session-ID header string false "Guest Session ID"
// @Success 204 "Cart cleared"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Router /cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, sessionID := h.extractUserAndSession(c)

	err := h.cartService.ClearCart(c.Request.Context(), userID, sessionID)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// MergeGuestCart godoc
// @Summary Merge guest cart into user cart
// @Description Merges items from a guest session cart into the authenticated user's cart.
// @Tags Cart
// @Accept json
// @Produce json
// @Param body body MergeCartRequest true "Guest Session ID to merge"
// @Success 200 {object} api.DataResponse{data=model.Cart} "Merged user cart"
// @Failure 400 {object} api.BadRequestErrorResponse "Missing user or session ID"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Router /cart/merge [post]
func (h *CartHandler) MergeGuestCart(c *gin.Context) {
	var req MergeCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apierr.ErrBadRequest("Invalid input payload").WithCode(apierr.CodeValidationFailed).Wrap(err))
		return
	}

	userID, _ := h.extractUserAndSession(c)
	if userID == nil {
		c.Error(apierr.ErrBadRequest("User authentication required to merge cart").WithCode(apierr.CodeInvalidInput))
		return
	}

	cart, err := h.cartService.MergeGuestCart(c.Request.Context(), req.SessionID, *userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: mapCartResponse(cart)})
}
