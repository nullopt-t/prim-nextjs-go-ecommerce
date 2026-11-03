package notifier

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// OrderNumber returns the short, customer-facing reference for an order id.
func OrderNumber(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return "PRIM-" + strings.ToUpper(id.String()[:8])
}

// OrderItemSummary is the immutable, customer-facing description of one order line.
type OrderItemSummary struct {
	Title     string `json:"title"`
	SKU       string `json:"sku,omitempty"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unitPrice"`
	LineTotal int64  `json:"lineTotal"`
}

// OrderAddressSummary is the shipping destination shown in order notifications.
type OrderAddressSummary struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	Country    string `json:"country"`
}

// OrderPlacedPayload is the data contract shared by the checkout producer and
// the notification worker for an order-confirmation email and in-app alert.
type OrderPlacedPayload struct {
	OrderID         uuid.UUID           `json:"orderId"`
	OrderNumber     string              `json:"orderNumber"`
	CustomerID      *uuid.UUID          `json:"customerId,omitempty"`
	CustomerEmail   string              `json:"customerEmail"`
	CustomerName    string              `json:"customerName"`
	TotalAmount     int64               `json:"totalAmount"`
	Currency        string              `json:"currency"`
	PaymentMethod   string              `json:"paymentMethod"`
	Locale          string              `json:"locale,omitempty"`
	Items           []OrderItemSummary  `json:"items"`
	ShippingAddress OrderAddressSummary `json:"shippingAddress"`
}

// AdminOrderEvent represents the real-time payload sent across SSE to administrative dashboards.
type AdminOrderEvent struct {
	ID            uuid.UUID `json:"id"`
	OrderNumber   string    `json:"orderNumber"`
	CustomerEmail string    `json:"customerEmail"`
	CustomerName  string    `json:"customerName"`
	TotalAmount   int64     `json:"totalAmount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

// AdminStockEvent represents a real-time inventory ledger update sent across SSE to administrative dashboards.
type AdminStockEvent struct {
	VariantID    uuid.UUID `json:"variantId"`
	ProductID    uuid.UUID `json:"productId"`
	ProductName  string    `json:"productName"`
	VariantTitle string    `json:"variantTitle,omitempty"`
	SKU          string    `json:"sku"`
	Quantity     int       `json:"quantity"`     // Delta (+/-)
	CurrentStock int       `json:"currentStock"` // On-hand balance after change
	Reason       string    `json:"reason"`
	IsLowStock   bool      `json:"isLowStock"`
	Timestamp    time.Time `json:"timestamp"`
}

// AdminRealtimeMessage wraps multiple administrative event types over the single admin SSE connection.
type AdminRealtimeMessage struct {
	Type  string           `json:"type"` // "order.created" | "stock.updated"
	Order *AdminOrderEvent `json:"order,omitempty"`
	Stock *AdminStockEvent `json:"stock,omitempty"`
}

// OrderShippedPayload is the data contract for dispatch notifications.
type OrderShippedPayload struct {
	OrderID        uuid.UUID  `json:"orderId"`
	OrderNumber    string     `json:"orderNumber"`
	CustomerID     *uuid.UUID `json:"customerId,omitempty"`
	CustomerEmail  string     `json:"customerEmail"`
	CustomerName   string     `json:"customerName"`
	Carrier        string     `json:"carrier,omitempty"`
	TrackingNumber string     `json:"trackingNumber,omitempty"`
	TrackingURL    string     `json:"trackingUrl,omitempty"`
	Locale         string     `json:"locale,omitempty"`
}

// OrderCancelledPayload is the data contract for cancellation notifications.
type OrderCancelledPayload struct {
	OrderID       uuid.UUID  `json:"orderId"`
	OrderNumber   string     `json:"orderNumber"`
	CustomerID    *uuid.UUID `json:"customerId,omitempty"`
	CustomerEmail string     `json:"customerEmail"`
	CustomerName  string     `json:"customerName"`
	Reason        string     `json:"reason,omitempty"`
	Locale        string     `json:"locale,omitempty"`
}
