package model

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents the current fulfillment and processing lifecycle state of an Order.
type OrderStatus string

const (
	// OrderStatusPending indicates the order has been created but payment has not yet succeeded.
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusPaid indicates payment has been successfully processed and verified.
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusProcessing indicates the order is currently being packed/prepared for shipment.
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusShipped indicates the order has been dispatched with a logistics carrier.
	OrderStatusShipped OrderStatus = "shipped"
	// OrderStatusDelivered indicates the carrier has successfully delivered the shipment to the customer.
	OrderStatusDelivered OrderStatus = "delivered"
	// OrderStatusCanceled indicates the customer or store cancelled the order prior to fulfillment.
	OrderStatusCanceled OrderStatus = "canceled"
	// OrderStatusRefunded indicates the purchase charge has been fully reversed to the customer.
	OrderStatusRefunded OrderStatus = "refunded"
)

// OrderItem represents an immutable purchased line item inside a finalized Order.
type OrderItem struct {
	// ID is the unique database UUID for this order line item record.
	ID uuid.UUID
	// OrderID is the parent Order UUID.
	OrderID uuid.UUID
	// VariantID is the purchased ProductVariant UUID.
	VariantID uuid.UUID
	// Quantity is the number of units purchased.
	Quantity int
	// PriceAtPurchase is the unit price in cents at the moment of checkout completion.
	PriceAtPurchase int64
	// ProductSnapshot contains a JSON snapshot of the product and variant title, attributes, and media.
	ProductSnapshot string
}

// Address represents a physical postal mailing address used for billing or shipping.
type Address struct {
	// Street address line (e.g. "123 Main St, Apt 4B").
	Street string
	// City or municipality name.
	City string
	// State, province, or regional district.
	State string
	// Postal code or ZIP code.
	PostalCode string
	// Country name or ISO country code.
	Country string
}

// Order represents a finalized customer purchase order.
type Order struct {
	// ID is the unique database UUID for the order.
	ID uuid.UUID
	// CustomerID is the optional UUID of the registered user (nil for guest checkouts).
	CustomerID *uuid.UUID
	// CustomerEmail is the contact email address for order notifications.
	CustomerEmail string
	// ShippingAddress is the destination address for physical parcel delivery.
	ShippingAddress Address
	// BillingAddress is the payment card billing address.
	BillingAddress Address
	// Status is the current fulfillment lifecycle state.
	Status OrderStatus
	// CouponID is the optional applied coupon promotion UUID.
	CouponID *uuid.UUID
	// DiscountAmount is the total discount reduction in cents.
	DiscountAmount int64
	// TotalAmount is the final charged sum (subtotal - discounts + tax + shipping) in cents.
	TotalAmount int64
	// Currency is the transaction currency code (e.g. "USD").
	Currency string
	// Items contains all purchased line items in this order.
	Items []OrderItem
	// CreatedAt is the timestamp when the order was placed.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when order status was last modified.
	UpdatedAt time.Time
}
