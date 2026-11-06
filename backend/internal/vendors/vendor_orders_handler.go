package vendors

import (
	"encoding/json"
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

// VendorOrderItemResponse is one of the merchant's own lines in an order.
// No customer PII is exposed on the vendor surface.
type VendorOrderItemResponse struct {
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
	// Frozen vendor earning for this line, in minor units.
	VendorEarning int64 `json:"vendorEarning"`
	// Frozen commission rate applied to this line.
	CommissionRateBps int `json:"commissionRateBps"`
	// Frozen platform commission amount for this line, in minor units.
	CommissionAmount int64 `json:"commissionAmount"`
	// Product/variant title from the checkout snapshot, when present.
	Title string `json:"title,omitempty"`
}

// VendorOrderResponse is the merchant-facing order projection: order state
// plus the vendor's own lines and their money math only.
type VendorOrderResponse struct {
	// Order id.
	ID uuid.UUID `json:"id"`
	// Fulfillment lifecycle status.
	Status model.OrderStatus `json:"status"`
	// Currency of the order.
	Currency string `json:"currency"`
	// The vendor's lines in this order.
	Items []VendorOrderItemResponse `json:"items"`
	// Sum of the vendor's line totals, in minor units.
	Subtotal int64 `json:"subtotal"`
	// Sum of the vendor's frozen earnings, in minor units.
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

func toVendorOrderResponse(o *model.Order) VendorOrderResponse {
	res := VendorOrderResponse{
		ID:        o.ID,
		Status:    o.Status,
		Currency:  o.Currency,
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
	}
	res.Items = make([]VendorOrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		lineTotal := item.PriceAtPurchase * int64(item.Quantity)
		res.Subtotal += lineTotal
		res.Earnings += item.VendorEarning
		res.Items = append(res.Items, VendorOrderItemResponse{
			ID:                item.ID,
			VariantID:         item.VariantID,
			Quantity:          item.Quantity,
			PriceAtPurchase:   item.PriceAtPurchase,
			LineTotal:         lineTotal,
			VendorEarning:     item.VendorEarning,
			CommissionRateBps: item.CommissionRateBps,
			CommissionAmount:  item.CommissionAmount,
			Title:             snapshotTitle(item.ProductSnapshot),
		})
	}
	return res
}

// ListMyOrders godoc
// @Summary      List my orders
// @Description  Returns orders containing the authenticated vendor's lines, newest first. Only the vendor's own lines are included; no customer PII is exposed.
// @Tags         Vendor
// @Produce      json
// @Param        status    query string false "Order status filter" Enums(pending, paid, processing, shipped, delivered, cancelled, refunded)
// @Param        page      query int    false "Page number (1-based)"
// @Param        pageSize  query int    false "Page size (max 100)"
// @Success      200 {object} api.PaginatedResponse{data=[]VendorOrderResponse}
// @Failure      400 {object} api.BadReqResponse "Invalid status filter"
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found"
// @Router       /vendor/orders [get]
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

	responses := make([]VendorOrderResponse, 0, len(result.Items))
	for i := range result.Items {
		responses = append(responses, toVendorOrderResponse(&result.Items[i]))
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

// GetMyOrder godoc
// @Summary      Get one of my orders
// @Description  Returns the vendor's own lines for one order. Orders without the vendor's lines answer 404.
// @Tags         Vendor
// @Produce      json
// @Param        id path string true "Order UUID" format(uuid)
// @Success      200 {object} api.DataResponse{data=VendorOrderResponse}
// @Failure      401 {object} api.UnauthorizedErrorResponse
// @Failure      404 {object} api.NotFoundErrorResponse "Vendor not found or order not yours"
// @Router       /vendor/orders/{id} [get]
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

	c.JSON(http.StatusOK, api.DataResponse{Data: toVendorOrderResponse(order)})
}
