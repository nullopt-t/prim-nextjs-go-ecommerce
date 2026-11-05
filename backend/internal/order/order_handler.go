// Package order implements order lifecycle, line item persistence, and order management handlers.
package order

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

// --- Requests ---

// UpdateOrderStatusRequest represents the JSON body accepted for mutating an order's status.
type UpdateOrderStatusRequest struct {
	// Status specifies the target OrderStatus transition (e.g. processing, shipped, delivered, cancelled).
	Status model.OrderStatus `json:"status" binding:"required"`
	// Notes provides optional human-readable administrative justification or carrier tracking details.
	Notes *string `json:"notes,omitempty"`
}

// --- Responses ---

// OrderItemResponse represents the serialized line item data associated with an order.
type OrderItemResponse struct {
	// ID is the unique UUID of the order line item.
	ID string `json:"id"`
	// VariantID is the product variant purchased in this line item.
	VariantID string `json:"variantId"`
	// Quantity is the number of units purchased.
	Quantity int `json:"quantity"`
	// PriceAtPurchase is the unit price in minor units (e.g. cents) at checkout time.
	PriceAtPurchase int64 `json:"priceAtPurchase"`
	// ProductSnapshot contains immutable JSON metadata captured at checkout (e.g. title, SKU, media).
	ProductSnapshot string `json:"productSnapshot"`
}

// AddressResponse represents a serialized postal shipping address.
type AddressResponse struct {
	// Street represents the primary street address lines.
	Street string `json:"street"`
	// City is the postal locality or city.
	City string `json:"city"`
	// State is the region, state, or province.
	State string `json:"state"`
	// PostalCode is the postal or ZIP code.
	PostalCode string `json:"postalCode"`
	// Country is the ISO or common country name.
	Country string `json:"country"`
}

// OrderResponse represents the complete HTTP JSON response payload describing an order.
type OrderResponse struct {
	// ID is the unique order UUID.
	ID string `json:"id"`
	// CustomerID is the optional authenticated user UUID who placed the order.
	CustomerID *string `json:"customerId,omitempty"`
	// CustomerEmail is the contact email address for transactional notifications.
	CustomerEmail string `json:"customerEmail"`
	// ShippingAddress is the destination address.
	ShippingAddress AddressResponse `json:"shippingAddress"`
	// Status is the current state of the order lifecycle.
	Status model.OrderStatus `json:"status"`
	// CouponID is the optional applied promotional voucher or coupon UUID.
	CouponID *string `json:"couponId,omitempty"`
	// CouponCode is the human-readable applied voucher or coupon code.
	CouponCode *string `json:"couponCode,omitempty"`
	// DiscountAmount is the deducted amount in minor currency units.
	DiscountAmount int64 `json:"discountAmount"`
	// ShippingAmount is the shipping charge applied to the order in minor currency units.
	ShippingAmount int64 `json:"shippingAmount"`
	// TaxAmount is the tax amount applied to the order in minor currency units.
	TaxAmount int64 `json:"taxAmount"`
	// CODFee is the cash-on-delivery surcharge applied to the order in minor currency units.
	CODFee int64 `json:"codFee"`
	// TotalAmount is the final monetary charge in minor currency units after discounts.
	TotalAmount int64 `json:"totalAmount"`
	// Currency is the 3-letter ISO currency code.
	Currency string `json:"currency"`
	// PaymentMethod is the payment method used for the order.
	PaymentMethod string `json:"paymentMethod"`
	// TrackingNumber is the carrier tracking number for shipped orders.
	TrackingNumber string `json:"trackingNumber"`
	// Items lists the individual line items in this order.
	Items []OrderItemResponse `json:"items,omitempty"`
	// CreatedAt is the RFC3339 creation timestamp.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the RFC3339 last update timestamp.
	UpdatedAt string `json:"updatedAt"`
}

// mapOrderResponse converts internal domain model.Order into the external OrderResponse DTO.
func mapOrderResponse(o *model.Order) OrderResponse {
	res := OrderResponse{
		ID:            o.ID.String(),
		CustomerEmail: o.CustomerEmail,
		ShippingAddress: AddressResponse{
			Street:     o.ShippingAddress.Street,
			City:       o.ShippingAddress.City,
			State:      o.ShippingAddress.State,
			PostalCode: o.ShippingAddress.PostalCode,
			Country:    o.ShippingAddress.Country,
		},
		Status:         o.Status,
		DiscountAmount: o.DiscountAmount,
		ShippingAmount: o.ShippingAmount,
		TaxAmount:      o.TaxAmount,
		CODFee:         o.CODFee,
		TotalAmount:    o.TotalAmount,
		Currency:       o.Currency,
		PaymentMethod:  o.PaymentMethod,
		TrackingNumber: o.TrackingNumber,
		CreatedAt:      o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      o.UpdatedAt.Format(time.RFC3339),
	}

	if o.CustomerID != nil {
		cid := o.CustomerID.String()
		res.CustomerID = &cid
	}

	if o.CouponID != nil {
		cid := o.CouponID.String()
		res.CouponID = &cid
	}

	res.CouponCode = o.CouponCode

	if len(o.Items) > 0 {
		items := make([]OrderItemResponse, 0, len(o.Items))
		for _, item := range o.Items {
			items = append(items, OrderItemResponse{
				ID:              item.ID.String(),
				VariantID:       item.VariantID.String(),
				Quantity:        item.Quantity,
				PriceAtPurchase: item.PriceAtPurchase,
				ProductSnapshot: item.ProductSnapshot,
			})
		}
		res.Items = items
	}

	return res
}

// --- Handlers ---

// OrderHandler encapsulates HTTP controllers processing customer and administrative order actions.
type OrderHandler struct {
	orderService *OrderService
	broadcaster  Broadcaster
}

// NewHandler constructs a new OrderHandler with the required OrderService and Broadcaster dependencies.
func NewHandler(orderService *OrderService, broadcaster Broadcaster) *OrderHandler {
	if orderService == nil {
		panic("order: service is required")
	}
	if broadcaster == nil {
		panic("order: broadcaster is required")
	}
	return &OrderHandler{
		orderService: orderService,
		broadcaster:  broadcaster,
	}
}

// StreamEvents godoc
//
//	@Summary		Stream administrative real-time events (SSE)
//	@Description	Establishes a long-lived Server-Sent Events (SSE) channel for real-time notifications of new orders.
//	@Tags			Admin Orders
//	@Produce		text/event-stream
//	@Success		200	{string}	string	"text/event-stream"
//	@Router			/admin/events [get]
func (h *OrderHandler) StreamEvents(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	ch := h.broadcaster.Subscribe()
	defer h.broadcaster.Unsubscribe(ch)

	// Send initial connected handshake event
	_, _ = fmt.Fprintf(c.Writer, "event: connected\ndata: {}\n\n")
	c.Writer.Flush()

	notifyDone := c.Request.Context().Done()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notifyDone:
			return
		case <-ticker.C:
			// Heartbeat comment to keep intermediate proxies/NAT from timing out
			_, err := fmt.Fprintf(c.Writer, ": keepalive\n\n")
			if err != nil {
				return
			}
			c.Writer.Flush()
		case msg, ok := <-ch:
			if !ok {
				return
			}
			eventType := msg.Type
			if eventType == "" {
				eventType = "order.created"
			}
			var payload any
			switch {
			case msg.Type == "stock.updated" && msg.Stock != nil:
				payload = msg.Stock
			case msg.Order != nil:
				payload = msg.Order
			default:
				payload = msg
			}

			data, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			_, err = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", eventType, data)
			if err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}

// GetOrders godoc
//
//	@Summary		List customer orders
//	@Description	Returns a list of orders belonging to the authenticated customer.
//	@Tags			Orders
//	@Produce		json
//	@Failure		401	{object}	api.UnauthorizedErrorResponse			"Authentication required"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=[]OrderResponse}	"List of orders"
//	@Router			/orders [get]
func (h *OrderHandler) GetOrders(c *gin.Context) {
	// Extract authenticated user ID from context established by middleware.Authenticate.
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	// Fetch all orders associated with the customer ID.
	orders, err := h.orderService.GetCustomerOrders(c.Request.Context(), customerID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ordersRes := make([]OrderResponse, 0, len(orders))
	for i := range orders {
		ordersRes = append(ordersRes, mapOrderResponse(&orders[i]))
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: ordersRes,
	})
}

// GetOrderByID godoc
//
//	@Summary		Get order by ID
//	@Description	Retrieves full order details by its UUID.
//	@Tags			Orders
//	@Produce		json
//	@Param			id	path		string									true	"Order UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse				"Invalid UUID format"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse			"Authentication required"
//	@Failure		404	{object}	api.NotFoundErrorResponse				"Order not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=OrderResponse}	"Order details"
//	@Router			/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	idParam := c.Param("id")
	orderID, err := uuid.Parse(idParam)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid order ID").WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid order UUID format",
		}))
		return
	}

	// Fetch order details by ID from the service layer.
	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Enforce customer ownership check for non-admin requests to prevent IDOR vulnerabilities.
	role, _ := middleware.GetUserRole(c)
	if role != model.AdminRole && role != model.SuperRole {
		customerID, ok := middleware.GetUserID(c)
		if !ok {
			_ = c.Error(apierr.UnauthorizedError("Authentication required"))
			return
		}
		if order.CustomerID == nil || *order.CustomerID != customerID {
			_ = c.Error(apierr.ForbiddenError("You are not authorized to view this order"))
			return
		}
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapOrderResponse(order),
	})
}

// UpdateOrderStatus godoc
//
//	@Summary		Update order status (Admin)
//	@Description	Updates the status of an existing order.
//	@Tags			Admin Orders
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Order UUID"	format(uuid)
//	@Param			status	body		UpdateOrderStatusRequest		true	"Status update details"
//	@Failure		400		{object}	api.BadRequestErrorResponse		"Validation error or invalid UUID format"
//	@Failure		404		{object}	api.NotFoundErrorResponse		"Order not found"
//	@Failure		500		{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200		{object}	api.MessageResponse				"Status update confirmation"
//	@Router			/admin/orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid order ID").WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid order UUID format",
		}))
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, req.Status, req.Notes); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "Order status updated successfully",
	})
}

// CancelOrder godoc
//
//	@Summary		Cancel order
//	@Description	Allows a customer to cancel their order if it is in pending or confirmed status.
//	@Tags			Orders
//	@Produce		json
//	@Param			id	path		string							true	"Order UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse		"Invalid UUID or order cannot be cancelled"
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Authentication required"
//	@Failure		403	{object}	api.ForbiddenErrorResponse		"Not authorized to cancel this order"
//	@Failure		404	{object}	api.NotFoundErrorResponse		"Order not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.MessageResponse				"Cancellation confirmation"
//	@Router			/orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	customerID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid order ID").WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid order UUID format",
		}))
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), orderID, customerID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "Order cancelled successfully",
	})
}

// AdminListOrders godoc
//
//	@Summary		List all orders (Admin)
//	@Description	Returns a paginated list of orders across the platform with optional filters.
//	@Tags			Admin Orders
//	@Produce		json
//	@Param			page		query		int								false	"Page number (default 1)"
//	@Param			pageSize	query		int								false	"Page size (default 10)"
//	@Param			customerId	query		string							false	"Filter by customer UUID"
//	@Param			status		query		string							false	"Filter by order status"
//	@Success		200			{object}	api.SuccessResponse				"Paginated order list"
//	@Router			/admin/orders [get]
func (h *OrderHandler) AdminListOrders(c *gin.Context) {
	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}

	q.Process(pagination.QueryOptions{DefaultPageSize: 10, MaxPageSize: 100})

	var customerIDPtr *uuid.UUID
	if cidStr := c.Query("customerId"); cidStr != "" {
		if parsed, err := uuid.Parse(cidStr); err == nil {
			customerIDPtr = &parsed
		}
	}

	var statusPtr *string
	if s := c.Query("status"); s != "" {
		statusPtr = &s
	}

	dateFrom, dateTo, err := parseDateRangeQuery(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	orders, page, err := h.orderService.ListOrders(
		c.Request.Context(),
		customerIDPtr,
		statusPtr,
		dateFrom,
		dateTo,
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	ordersRes := make([]OrderResponse, 0, len(orders))
	for i := range orders {
		ordersRes = append(ordersRes, mapOrderResponse(&orders[i]))
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Data: ordersRes,
		Meta: page,
	})
}

// AdminGetOrderByID godoc
//
//	@Summary		Get order by ID (Admin)
//	@Description	Retrieves complete order details for administrative inspection.
//	@Tags			Admin Orders
//	@Produce		json
//	@Param			id	path		string									true	"Order UUID"	format(uuid)
//	@Failure		400	{object}	api.BadRequestErrorResponse				"Invalid UUID format"
//	@Failure		404	{object}	api.NotFoundErrorResponse				"Order not found"
//	@Failure		500	{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=OrderResponse}	"Order details"
//	@Router			/admin/orders/{id} [get]
func (h *OrderHandler) AdminGetOrderByID(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid order ID").WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid order UUID format",
		}))
		return
	}

	ord, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapOrderResponse(ord),
	})
}

// GetDashboard godoc
//
//	@Summary		Get administrative dashboard metrics
//	@Description	Aggregates platform KPIs including revenue, order volume, and customer statistics.
//	@Tags			Admin Dashboard
//	@Produce		json
//	@Success		200	{object}	api.DataResponse
//	@Router			/admin/dashboard [get]
func (h *OrderHandler) GetDashboard(c *gin.Context) {
	stats, err := h.orderService.GetDashboardStats(c.Request.Context())
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch dashboard stats").Wrap(err))
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: stats})
}

// ListPayments godoc
//
//	@Summary		List payments (Admin)
//	@Description	Returns a paginated list of all payment transactions across the platform.
//	@Tags			Admin Payments
//	@Produce		json
//	@Param			page		query	int	false	"Page number"
//	@Param			pageSize	query	int	false	"Items per page"
//	@Success		200			{object}	api.PaginatedResponse{data=[]PaymentItem,meta=pagination.CursorPage}
//	@Router			/admin/payments [get]
func (h *OrderHandler) ListPayments(c *gin.Context) {
	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 50, MaxPageSize: 100})

	cursor, err := q.Keyset()
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid pagination cursor").Wrap(err))
		return
	}

	var statusPtr *string
	if s := c.Query("status"); s != "" {
		statusPtr = &s
	}

	dateFrom, dateTo, err := parseDateRangeQuery(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	payments, page, err := h.orderService.ListPayments(c.Request.Context(), PaymentListFilter{
		Status:   statusPtr,
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Limit:    q.PageSize,
		Cursor:   cursor,
	})
	if err != nil {
		_ = c.Error(apierr.InternalServerError("Failed to fetch payments").Wrap(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": payments,
		"meta": page,
	})
}

// UpdatePaymentStatus godoc
//
//	@Summary		Update payment status (Superadmin)
//	@Description	Updates status of a payment transaction. Superadmin only.
//	@Tags			Admin Payments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Payment UUID"
//	@Param			status	body		map[string]string	true	"Status"
//	@Failure		400		{object}	api.BadRequestErrorResponse	"Invalid UUID, body, or payment status"
//	@Failure		403		{object}	api.ForbiddenErrorResponse	"Superadmin privileges required"
//	@Failure		404		{object}	api.NotFoundErrorResponse	"Payment not found"
//	@Success		200		{object}	api.MessageResponse
//	@Router			/admin/payments/{id}/status [patch]
func (h *OrderHandler) UpdatePaymentStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid payment ID"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid status body"))
		return
	}

	if err := h.orderService.UpdatePaymentStatus(c.Request.Context(), id, req.Status); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Payment status updated successfully"})
}

// RefundPayment godoc
//
//	@Summary		Refund payment (Superadmin)
//	@Description	Marks a payment transaction as refunded. Superadmin only.
//	@Tags			Admin Payments
//	@Produce		json
//	@Param			id	path		string	true	"Payment UUID"
//	@Failure		400		{object}	api.BadRequestErrorResponse	"Invalid payment UUID"
//	@Failure		403		{object}	api.ForbiddenErrorResponse	"Superadmin privileges required"
//	@Failure		404		{object}	api.NotFoundErrorResponse	"Payment not found"
//	@Failure		409		{object}	api.ConflictErrorResponse	"Payment is not in a refundable state"
//	@Success		200	{object}	api.MessageResponse
//	@Router			/admin/payments/{id}/refund [post]
func (h *OrderHandler) RefundPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid payment ID"))
		return
	}

	if err := h.orderService.RefundPayment(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Payment marked as refunded"})
}

// UpdateOrderTrackingRequest represents the payload for updating logistics tracking.
type UpdateOrderTrackingRequest struct {
	TrackingNumber string  `json:"trackingNumber" binding:"required"`
	Carrier        *string `json:"carrier"`
	TrackingURL    *string `json:"trackingUrl"`
}

// UpdateOrderTracking godoc
//
//	@Summary		Update order tracking details (Admin)
//	@Description	Updates the carrier, tracking number, and tracking URL for an order.
//	@Tags			Admin Orders
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Order UUID"	format(uuid)
//	@Param			tracking	body		UpdateOrderTrackingRequest	true	"Tracking details"
//	@Success		200		{object}	api.MessageResponse
//	@Router			/admin/orders/{id}/tracking [patch]
func (h *OrderHandler) UpdateOrderTracking(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid order ID").WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid order UUID format",
		}))
		return
	}

	var req UpdateOrderTrackingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apierr.BadRequestError("Invalid request body").Wrap(err))
		return
	}

	if err := h.orderService.UpdateOrderTracking(c.Request.Context(), orderID, UpdateOrderTrackingInput(req)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, api.MessageResponse{Message: "Tracking details updated successfully"})
}

// parseDateRangeQuery reads optional dateFrom/dateTo query params into a time
// range. Each value accepts RFC3339 or a plain YYYY-MM-DD date; a bare date for
// dateTo is expanded to the end of that day.
func parseDateRangeQuery(c *gin.Context) (*time.Time, *time.Time, error) {
	var from, to *time.Time

	if v := c.Query("dateFrom"); v != "" {
		t, err := parseOrderTime(v, false)
		if err != nil {
			return nil, nil, apierr.BadRequestError("Invalid dateFrom (RFC3339 or YYYY-MM-DD expected)").
				WithCode(apierr.CodeInvalidInput)
		}
		from = &t
	}
	if v := c.Query("dateTo"); v != "" {
		t, err := parseOrderTime(v, true)
		if err != nil {
			return nil, nil, apierr.BadRequestError("Invalid dateTo (RFC3339 or YYYY-MM-DD expected)").
				WithCode(apierr.CodeInvalidInput)
		}
		to = &t
	}

	return from, to, nil
}

func parseOrderTime(v string, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return t, nil
}
