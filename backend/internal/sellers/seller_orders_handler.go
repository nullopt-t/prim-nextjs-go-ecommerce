package sellers

import (
	"encoding/json"
	"net/http"
	"strings"
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

// SellerOrderItemResponse is one of the merchant's own lines in an order.
// No customer PII is exposed on the seller surface.
type SellerOrderItemResponse struct {
	// Line item id.
	ID uuid.UUID `json:"id"`
	// Purchased variant id.
	VariantID uuid.UUID `json:"variantId"`
	// Units purchased.
	Quantity int `json:"quantity"`
	// Unit price in minor units, frozen at checkout.
	PriceAtPurchase int64 `json:"priceAtPurchase"`
	// Line total in minor units.
	LineTotal int64 `json:"lineTotal"`
	// Frozen seller earning for this line, in minor units.
	SellerEarning int64 `json:"sellerEarning"`
	// Frozen commission rate applied to this line.
	CommissionRateBps int `json:"commissionRateBps"`
	// Frozen platform commission amount for this line, in minor units.
	CommissionAmount int64 `json:"commissionAmount"`
	// Product/variant title from the checkout snapshot, when present.
	Title string `json:"title,omitempty"`
	// Fulfillment status for this item: pending, shipped, delivered.
	FulfillmentStatus model.FulfillmentStatus `json:"fulfillmentStatus"`
	// Carrier tracking number.
	TrackingNumber *string `json:"trackingNumber,omitempty"`
	// Logistics carrier name (e.g. "Aramex").
	Carrier *string `json:"carrier,omitempty"`
	// Timestamp when marked shipped.
	ShippedAt *string `json:"shippedAt,omitempty"`
}

// SellerOrderResponse is the merchant-facing order projection: order state
// plus the seller's own lines and their money math only.
type SellerOrderResponse struct {
	// Order id.
	ID uuid.UUID `json:"id"`
	// Fulfillment lifecycle status.
	Status model.OrderStatus `json:"status"`
	// Currency of the order.
	Currency string `json:"currency"`
	// The seller's lines in this order.
	Items []SellerOrderItemResponse `json:"items"`
	// Applied coupon code, when the buyer used one.
	CouponCode *string `json:"couponCode,omitempty"`
	// Total discount on the order, in minor units, when a coupon applied.
	DiscountAmount int64 `json:"discountAmount,omitempty"`
	// Sum of the seller's line totals, in minor units.
	Subtotal int64 `json:"subtotal"`
	// Sum of the seller's frozen earnings, in minor units.
	Earnings int64 `json:"earnings"`
	// Order created timestamp (RFC3339).
	CreatedAt string `json:"createdAt"`
	// Order last updated timestamp (RFC3339).
	UpdatedAt string `json:"updatedAt"`
}

// snapshotTitle best-effort extracts a display title from a product snapshot.
func snapshotTitle(raw string) string {
	var snap struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		return ""
	}
	return snap.Title
}

func toSellerOrderResponse(o *model.Order) SellerOrderResponse {
	res := SellerOrderResponse{
		ID:             o.ID,
		Status:         o.Status,
		Currency:       o.Currency,
		CouponCode:     o.CouponCode,
		DiscountAmount: o.DiscountAmount,
		CreatedAt:      o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      o.UpdatedAt.Format(time.RFC3339),
	}
	res.Items = make([]SellerOrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		lineTotal := item.PriceAtPurchase * int64(item.Quantity)
		res.Subtotal += lineTotal
		res.Earnings += item.SellerEarning
		var shippedAt *string
		if item.ShippedAt != nil {
			s := item.ShippedAt.UTC().Format(time.RFC3339)
			shippedAt = &s
		}
		fulfillmentStatus := item.FulfillmentStatus
		if fulfillmentStatus == "" {
			fulfillmentStatus = model.FulfillmentStatusPending
		}
		res.Items = append(res.Items, SellerOrderItemResponse{
			ID:                item.ID,
			VariantID:         item.VariantID,
			Quantity:          item.Quantity,
			PriceAtPurchase:   item.PriceAtPurchase,
			LineTotal:         lineTotal,
			SellerEarning:     item.SellerEarning,
			CommissionRateBps: item.CommissionRateBps,
			CommissionAmount:  item.CommissionAmount,
			Title:             snapshotTitle(item.ProductSnapshot),
			FulfillmentStatus: fulfillmentStatus,
			TrackingNumber:    item.TrackingNumber,
			Carrier:           item.Carrier,
			ShippedAt:         shippedAt,
		})
	}
	return res
}

// ListMyOrders godoc
// @Summary      List my orders
// @Description  Returns orders containing the authenticated seller's lines, newest first. Only the seller's own lines are included; no customer PII is exposed.
// @Tags         Seller
// @Produce      json
// @Param        status    query string false "Order status filter" Enums(pending, paid, processing, shipped, delivered, cancelled, refunded)
// @Param        page      query int    false "Page number (1-based)"
// @Param        pageSize  query int    false "Page size (max 100)"
// @Success      200 {object} api.PaginatedResponse{data=[]SellerOrderResponse}
// @Failure      400 {object} api.BadReqResponse "Invalid status filter"
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found"
// @Router       /seller/orders [get]
// @Security     SessionAuth
func (h *Handler) ListMyOrders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	var status *model.OrderStatus
	if raw := c.Query("status"); raw != "" {
		s := model.OrderStatus(raw)
		switch s {
		case model.OrderStatusPending, model.OrderStatusPaid, model.OrderStatusProcessing,
			model.OrderStatusShipped, model.OrderStatusDelivered,
			model.OrderStatusCancelled, model.OrderStatusRefunded:
		default:
			_ = c.Error(apierr.BadRequestError("Invalid order status filter"))
			return
		}
		status = &s
	}

	var q pagination.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		validation.ValidationError(c, err)
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	result, err := h.service.ListOwnOrders(c.Request.Context(), userID, status, &q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	responses := make([]SellerOrderResponse, 0, len(result.Items))
	for i := range result.Items {
		responses = append(responses, toSellerOrderResponse(&result.Items[i]))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

// FulfillMyOrderRequest is the JSON body for fulfilling specific line items or advancing one of the seller's orders.
type FulfillMyOrderRequest struct {
	// Specific item IDs to fulfill. When empty, all items owned by the seller on this order are fulfilled.
	ItemIDs []string `json:"item_ids,omitempty"`
	// CamelCase alias for item_ids.
	ItemIDsAlt []string `json:"itemIds,omitempty"`
	// Carrier tracking number.
	TrackingNumber string `json:"tracking_number,omitempty"`
	// CamelCase alias for tracking_number.
	TrackingNumAlt string `json:"trackingNumber,omitempty"`
	// Logistics carrier name, e.g. "Aramex".
	Carrier string `json:"carrier,omitempty"`
	// Target lifecycle status (optional, defaults to shipped).
	Status *model.OrderStatus `json:"status,omitempty"`
	// Direct carrier tracking URL.
	TrackingURL *string `json:"trackingUrl,omitempty" binding:"omitempty,url"`
	// Optional note recorded in the status history.
	Notes *string `json:"notes,omitempty"`
}

// FulfillMyOrder godoc
// @Summary      Fulfill one of my orders
// @Description  Fulfills specific line items or advances one of the authenticated seller's orders to processing or shipped and records the tracking details.
// @Tags         Seller
// @Accept       json
// @Produce      json
// @Param        id      path  string                   true  "Order UUID" format(uuid)
// @Param        body    body  FulfillMyOrderRequest    true  "Fulfillment details"
// @Success      200     {object} api.DataResponse{data=SellerOrderResponse}
// @Failure      400     {object} api.BadReqResponse "Invalid transition or missing tracking number"
// @Failure      401     {object} api.UnauthorizedErrorResponse
// @Failure      404     {object} api.NotFoundErrorResponse "Seller not found or order not yours"
// @Router       /seller/orders/{id}/fulfill [post]
// @Security     SessionAuth
func (h *Handler) FulfillMyOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.InvalidUUIDError().WithFields(api.FieldError{
			Field: "id", Constraint: "uuid", Message: "invalid order id",
		}))
		return
	}

	var body FulfillMyOrderRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	rawItemIDs := body.ItemIDs
	if len(rawItemIDs) == 0 && len(body.ItemIDsAlt) > 0 {
		rawItemIDs = body.ItemIDsAlt
	}

	trackingNumber := strings.TrimSpace(body.TrackingNumber)
	if trackingNumber == "" && body.TrackingNumAlt != "" {
		trackingNumber = strings.TrimSpace(body.TrackingNumAlt)
	}

	carrier := strings.TrimSpace(body.Carrier)

	var itemIDs []uuid.UUID
	for _, raw := range rawItemIDs {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		parsed, err := uuid.Parse(trimmed)
		if err != nil {
			_ = c.Error(apierr.BadRequestError("Invalid item ID: must be a valid UUID").WithCode(apierr.CodeInvalidInput))
			return
		}
		itemIDs = append(itemIDs, parsed)
	}

	var order *model.Order
	if len(itemIDs) > 0 || carrier != "" || (body.Status == nil && trackingNumber != "") {
		order, err = h.service.FulfillOwnOrderItems(c.Request.Context(), userID, orderID, itemIDs, trackingNumber, carrier)
	} else {
		status := model.OrderStatusShipped
		if body.Status != nil {
			status = *body.Status
		}
		var carrierPtr *string
		if carrier != "" {
			carrierPtr = &carrier
		}
		order, err = h.service.FulfillOwnOrder(c.Request.Context(), userID, orderID, FulfillOwnOrderInput{
			Status:         status,
			TrackingNumber: trackingNumber,
			Carrier:        carrierPtr,
			TrackingURL:    body.TrackingURL,
			Notes:          body.Notes,
		})
	}
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerOrderResponse(order)})
}

// GetMyOrder godoc
// @Summary      Get one of my orders
// @Description  Returns the seller's own lines for one order. Orders without the seller's lines answer 404.
// @Tags         Seller
// @Produce      json
// @Param        id path string true "Order UUID" format(uuid)
// @Success      200 {object} api.DataResponse{data=SellerOrderResponse}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Seller not found or order not yours"
// @Router       /seller/orders/{id} [get]
// @Security     SessionAuth
func (h *Handler) GetMyOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		_ = c.Error(apierr.UnauthorizedError("Authentication required"))
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.InvalidUUIDError().WithFields(api.FieldError{
			Field: "id", Constraint: "uuid", Message: "invalid order id",
		}))
		return
	}

	order, err := h.service.GetOwnOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: toSellerOrderResponse(order)})
}
