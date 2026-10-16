package checkout

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

type AddressRequest struct {
	Street     string `json:"street" binding:"required"`
	City       string `json:"city" binding:"required"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country" binding:"required"`
}

type CheckoutRequest struct {
	ShippingAddressID uuid.UUID  `json:"shippingAddressId" binding:"required"`
	CouponID          *uuid.UUID `json:"couponId,omitempty"`
	PaymentMethod     string     `json:"paymentMethod,omitempty"`
}

type CheckoutResponse struct {
	OrderID       string `json:"orderId"`
	Status        string `json:"status"`
	PaymentStatus string `json:"paymentStatus,omitempty"`
}

type CheckoutHandler struct {
	checkoutService *CheckoutService
}

func NewHandler(checkoutService *CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{checkoutService: checkoutService}
}

// Checkout godoc
//
//	@Summary		Perform checkout
//	@Description	Creates an order from the current user's cart and clears the cart.
//	@Tags			Checkout
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CheckoutRequest							true	"Checkout Details"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Validation error or empty cart"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse			"Authentication required"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200		{object}	api.DataResponse{data=CheckoutResponse}	"Checkout result"
//	@Router			/checkout [post]
func (h *CheckoutHandler) Checkout(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	customerID, ok := userIDVal.(uuid.UUID)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Invalid user identity"))
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CheckoutInput{
		CustomerID:        customerID,
		ShippingAddressID: req.ShippingAddressID,
		CouponID:          req.CouponID,
		PaymentMethod:     req.PaymentMethod,
	}

	createdOrder, err := h.checkoutService.Checkout(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: CheckoutResponse{
			OrderID: createdOrder.ID.String(),
			Status:  string(createdOrder.Status),
		},
	})
}

type ApplyCouponRequest struct {
	Code         string `json:"code" binding:"required"`
	CartSubtotal *int64 `json:"cartSubtotal,omitempty"`
}

// ApplyCoupon godoc
//
//	@Summary		Apply coupon
//	@Description	Validates and applies a coupon code, returning discount details.
//	@Tags			Coupons
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ApplyCouponRequest						true	"Coupon Code"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Invalid or expired coupon"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200		{object}	api.DataResponse{data=CouponValidationResult}	"Coupon details"
//	@Router			/coupons/apply [post]
func (h *CheckoutHandler) ApplyCoupon(c *gin.Context) {
	var req ApplyCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	var userID *uuid.UUID
	if userIDVal, exists := c.Get("userID"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	var sessionID *string
	if cookieSess, err := c.Cookie("session_id"); err == nil && cookieSess != "" {
		sessionID = &cookieSess
	}

	res, err := h.checkoutService.ValidateCoupon(c.Request.Context(), req.Code, userID, sessionID, req.CartSubtotal)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}
