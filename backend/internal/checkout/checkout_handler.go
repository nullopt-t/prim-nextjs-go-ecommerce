package checkout

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
)

type AddressRequest struct {
	Street     string `json:"street" binding:"required"`
	City       string `json:"city" binding:"required"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code" binding:"required"`
	Country    string `json:"country" binding:"required"`
}

type CheckoutRequest struct {
	CustomerEmail   string         `json:"customer_email" binding:"required,email"`
	ShippingAddress AddressRequest `json:"shipping_address" binding:"required"`
	BillingAddress  AddressRequest `json:"billing_address" binding:"required"`
	CouponID        *uuid.UUID     `json:"coupon_id,omitempty"`
}

type CheckoutResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type CheckoutHandler struct {
	checkoutService *CheckoutService
}

func NewHandler(checkoutService *CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{checkoutService: checkoutService}
}

// Checkout godoc
// @Summary Perform checkout
// @Description Creates an order from the current user's cart and clears the cart.
// @Tags Checkout
// @Accept json
// @Produce json
// @Param request body CheckoutRequest true "Checkout Details"
// @Failure 400 {object} api.BadRequestErrorResponse "Validation error or empty cart"
// @Failure 401 {object} api.UnauthorizedErrorResponse "Authentication required"
// @Failure 500 {object} api.InternalServerErrorResponse "Internal server error"
// @Success 200 {object} api.DataResponse{data=CheckoutResponse} "Checkout result"
// @Router /checkout [post]
func (h *CheckoutHandler) Checkout(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		_ = c.Error(apierr.ErrUnauthorized("Authentication required"))
		return
	}

	customerID, ok := userIDVal.(uuid.UUID)
	if !ok {
		_ = c.Error(apierr.ErrUnauthorized("Invalid user identity"))
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	in := &CheckoutInput{
		CustomerID:    customerID,
		CustomerEmail: req.CustomerEmail,
		ShippingAddress: model.Address{
			Street:     req.ShippingAddress.Street,
			City:       req.ShippingAddress.City,
			State:      req.ShippingAddress.State,
			PostalCode: req.ShippingAddress.PostalCode,
			Country:    req.ShippingAddress.Country,
		},
		BillingAddress: model.Address{
			Street:     req.BillingAddress.Street,
			City:       req.BillingAddress.City,
			State:      req.BillingAddress.State,
			PostalCode: req.BillingAddress.PostalCode,
			Country:    req.BillingAddress.Country,
		},
		CouponID: req.CouponID,
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
