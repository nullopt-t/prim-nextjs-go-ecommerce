package order

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
)

// --- Requests ---

type UpdateOrderStatusRequest struct {
	Status model.OrderStatus `json:"status" binding:"required"`
	Notes  *string           `json:"notes,omitempty"`
}

// --- Responses ---

type OrderItemResponse struct {
	ID              string `json:"id"`
	VariantID       string `json:"variant_id"`
	Quantity        int    `json:"quantity"`
	PriceAtPurchase int64  `json:"price_at_purchase"`
	ProductSnapshot string `json:"product_snapshot"`
}

type AddressResponse struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type OrderResponse struct {
	ID              string              `json:"id"`
	CustomerID      *string             `json:"customer_id,omitempty"`
	CustomerEmail   string              `json:"customer_email"`
	ShippingAddress AddressResponse     `json:"shipping_address"`
	BillingAddress  AddressResponse     `json:"billing_address"`
	Status          model.OrderStatus   `json:"status"`
	CouponID        *string             `json:"coupon_id,omitempty"`
	DiscountAmount  int64               `json:"discount_amount"`
	TotalAmount     int64               `json:"total_amount"`
	Currency        string              `json:"currency"`
	Items           []OrderItemResponse `json:"items,omitempty"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

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
		BillingAddress: AddressResponse{
			Street:     o.BillingAddress.Street,
			City:       o.BillingAddress.City,
			State:      o.BillingAddress.State,
			PostalCode: o.BillingAddress.PostalCode,
			Country:    o.BillingAddress.Country,
		},
		Status:         o.Status,
		DiscountAmount: o.DiscountAmount,
		TotalAmount:    o.TotalAmount,
		Currency:       o.Currency,
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

type OrderHandler struct {
	orderService *OrderService
}

func NewHandler(orderService *OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
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
		_ = c.Error(apierr.ErrBadRequest("Invalid order ID").WithFields(api.FieldError{
			Field:   "id",
			Message: "invalid order UUID format",
		}))
		return
	}

	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: mapOrderResponse(order),
	})
}

// UpdateOrderStatus godoc
//
//	@Summary		Update order status (Admin)
//	@Description	Updates the status of an existing order.
//	@Tags			Orders
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
		_ = c.Error(apierr.ErrBadRequest("Invalid order ID").WithFields(api.FieldError{
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
