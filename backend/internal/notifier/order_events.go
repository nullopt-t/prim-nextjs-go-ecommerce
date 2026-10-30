package notifier

import (
	"strings"

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
