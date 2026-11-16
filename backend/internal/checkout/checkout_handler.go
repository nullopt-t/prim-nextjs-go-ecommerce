package checkout

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
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
	CouponCode        *string    `json:"couponCode,omitempty"`
	CouponID          *uuid.UUID `json:"couponId,omitempty"`
	PaymentMethod     string     `json:"paymentMethod,omitempty"`
}

type CheckoutResponse struct {
	OrderID       string `json:"orderId"`
	Status        string `json:"status"`
	PaymentStatus string `json:"paymentStatus,omitempty"`
	ClientSecret  string `json:"clientSecret,omitempty"`
}

type CheckoutHandler struct {
	checkoutService *CheckoutService
}

func NewHandler(checkoutService *CheckoutService) *CheckoutHandler {
	if checkoutService == nil {
		panic("checkout: service is required")
	}
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
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
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
		CouponCode:        req.CouponCode,
		CouponID:          req.CouponID,
		PaymentMethod:     req.PaymentMethod,
	}

	out, err := h.checkoutService.Checkout(c.Request.Context(), in)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: CheckoutResponse{
			OrderID:      out.Order.ID.String(),
			Status:       string(out.Order.Status),
			ClientSecret: out.ClientSecret,
		},
	})
}

type ConfirmPaymentRequest struct {
	OrderID         uuid.UUID `json:"orderId" binding:"required"`
	PaymentIntentID string    `json:"paymentIntentId"`
}

// ConfirmPayment godoc
//
//	@Summary		Confirm payment
//	@Description	Confirms an order payment after external card confirmation, transitioning order to processing and committing stock.
//	@Tags			Checkout
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ConfirmPaymentRequest						true	"Payment Confirmation Details"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Validation error"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse			"Authentication required"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200		{object}	api.DataResponse{data=PaymentWebhookResult}	"Confirmation result"
//	@Router			/checkout/confirm [post]
func (h *CheckoutHandler) ConfirmPayment(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var req ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	res, err := h.checkoutService.ConfirmPayment(c.Request.Context(), &ConfirmPaymentInput{
		CustomerID:      customerID,
		OrderID:         req.OrderID,
		PaymentIntentID: req.PaymentIntentID,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}

type ApplyCouponRequest struct {
	Code string `json:"code" binding:"required"`
}

// PaymentWebhookRequest is the provider payment-event payload. The event
// travels in `event` (Stripe-style `type` is accepted as an alias). The
// PaymentWebhookRequest represents a webhook payload from a payment provider (such as Stripe)
// or an internal HMAC test call.
type PaymentWebhookRequest struct {
	Event         string `json:"event,omitempty"`
	Type          string `json:"type,omitempty"`
	OrderID       string `json:"order_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	Signature     string `json:"signature,omitempty"`
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
	var sessionID *string
	if session, ok := middleware.GetSession(c); ok {
		if session.UserID != uuid.Nil {
			uid := session.UserID
			userID = &uid
		} else {
			sid := session.ID
			sessionID = &sid
		}
	}

	res, err := h.checkoutService.ValidateCoupon(c.Request.Context(), req.Code, userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}

const maxWebhookBodyBytes = 64 * 1024 // 64KB max webhook payload

// HandlePaymentWebhook godoc
//
//	@Summary		Ingest a payment webhook event
//	@Description	Unauthenticated provider webhook: verifies the HMAC signature, claims the transaction id idempotently, and on success transitions the order Pending->Processing with a payment audit entry.
//	@Tags			Checkout
//	@Accept			json
//	@Produce		json
//	@Param			request	body		PaymentWebhookRequest						true	"Payment event payload"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Validation error or invalid order status"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse			"Invalid webhook signature or missing secret"
//	@Failure		404		{object}	api.NotFoundErrorResponse				"Order not found"
//	@Failure		409		{object}	api.ConflictErrorResponse				"Transaction belongs to another order"
//	@Failure		413		{object}	api.PayloadTooLargeErrorResponse		"Webhook payload exceeds maximum size"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200		{object}	api.DataResponse{data=PaymentWebhookResult}	"Webhook processed (duplicate=true on replay)"
//	@Router			/checkout/webhook/payment [post]
func (h *CheckoutHandler) HandlePaymentWebhook(c *gin.Context) {
	// Bounded read prevents memory exhaustion on unauthenticated webhook endpoint
	bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxWebhookBodyBytes+1))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Failed to read request body").WithCode(apierr.CodeInvalidInput))
		return
	}
	if len(bodyBytes) > maxWebhookBodyBytes {
		_ = c.Error(apierr.FileTooLargeError("Webhook payload exceeds maximum size (64KB)"))
		return
	}
	if len(bodyBytes) == 0 {
		_ = c.Error(apierr.BadRequestError("Request body cannot be empty").WithCode(apierr.CodeInvalidInput))
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req PaymentWebhookRequest
	_ = json.Unmarshal(bodyBytes, &req)

	eventType := req.Event
	if eventType == "" {
		eventType = req.Type
	}

	var orderID uuid.UUID
	if req.OrderID != "" {
		if parsed, parseErr := uuid.Parse(req.OrderID); parseErr == nil {
			orderID = parsed
		}
	}

	sig := req.Signature
	if sig == "" {
		sig = c.GetHeader("Stripe-Signature")
	}
	if sig == "" {
		_ = c.Error(apierr.BadRequestError("Signature is required (body signature or Stripe-Signature header)").WithCode(apierr.CodeInvalidInput))
		return
	}

	res, err := h.checkoutService.HandlePaymentWebhook(c.Request.Context(), &PaymentWebhookInput{
		EventType:     eventType,
		OrderID:       orderID,
		TransactionID: req.TransactionID,
		Signature:     sig,
		RawPayload:    bodyBytes,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: res,
	})
}
