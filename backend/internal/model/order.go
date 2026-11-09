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
	// OrderStatusCancelled indicates the customer or store cancelled the order prior to fulfillment.
	OrderStatusCancelled OrderStatus = "cancelled"
	// OrderStatusRefunded indicates the purchase charge has been fully reversed to the customer.
	OrderStatusRefunded OrderStatus = "refunded"
)

// OrderMilestone is an Amazon-granular fulfillment sub-status derived from the
// order lifecycle status (and, once shipped, whether a carrier tracking number
// exists). See docs/order-milestones.md for the contract and percent table.
type OrderMilestone string

const (
	// OrderMilestonePreTransit: order placed, not yet picked (0%).
	OrderMilestonePreTransit OrderMilestone = "pre_transit"
	// OrderMilestonePickedAtWarehouse: picked/packed at the warehouse (5%).
	OrderMilestonePickedAtWarehouse OrderMilestone = "picked_at_warehouse"
	// OrderMilestoneArrivedRegionalHub: in transit, arrived at regional hub (10%).
	OrderMilestoneArrivedRegionalHub OrderMilestone = "arrived_regional_hub"
	// OrderMilestoneArrivedLocalDepot: arrived at the local depot (25%).
	// Reserved for future carrier scan events; derivation does not emit it yet.
	OrderMilestoneArrivedLocalDepot OrderMilestone = "arrived_local_depot"
	// OrderMilestoneOutForDelivery: out for final delivery (50%).
	OrderMilestoneOutForDelivery OrderMilestone = "out_for_delivery"
	// OrderMilestoneDelivered: delivered to the customer (100%).
	OrderMilestoneDelivered OrderMilestone = "delivered"
)

// FulfillmentStatus defines the seller-specific shipping status of an individual order item.
type FulfillmentStatus string

const (
	FulfillmentStatusPending   FulfillmentStatus = "pending"
	FulfillmentStatusShipped   FulfillmentStatus = "shipped"
	FulfillmentStatusDelivered FulfillmentStatus = "delivered"
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
	// SellerID is the selling seller at purchase time (nil = platform-owned line).
	SellerID *uuid.UUID
	// SellerEarning is the seller share in minor units: item total minus platform commission, frozen at checkout.
	SellerEarning int64
	// CommissionRateBps is the platform commission rate applied to this line, frozen at checkout.
	CommissionRateBps int
	// CommissionAmount is the platform commission in minor units, frozen at checkout.
	CommissionAmount int64
	// FulfillmentStatus is the seller fulfillment state for this line item.
	FulfillmentStatus FulfillmentStatus
	// TrackingNumber is the carrier shipment tracking number for this line item.
	TrackingNumber *string
	// Carrier is the logistics delivery service (e.g. "Aramex", "Bosta").
	Carrier *string
	// ShippedAt is the timestamp when the seller marked this line item as shipped.
	ShippedAt *time.Time
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

// OrderStatusHistory is one row of the order_status_history audit trail:
// a status transition (or admin note) with its timestamp, in chronological
// order. Note the primary key is a bigint identity in schema, not a UUID.
type OrderStatusHistory struct {
	// ID is the bigint identity of the history row.
	ID int64 `json:"id"`
	// OrderID is the parent Order UUID.
	OrderID uuid.UUID `json:"order_id"`
	// Status is the status the order transitioned to (or was noted under).
	Status OrderStatus `json:"status"`
	// Notes is the optional admin/operational note for this entry.
	Notes *string `json:"notes,omitempty"`
	// CreatedAt is when the entry was recorded.
	CreatedAt time.Time `json:"created_at"`
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
	ShippingAddress *Address
	// Status is the current fulfillment lifecycle state.
	Status OrderStatus
	// CouponID is the optional applied coupon promotion UUID.
	CouponID *uuid.UUID
	// CouponCode is the human-readable code of the applied coupon (e.g. "SUMMER20").
	CouponCode *string
	// DiscountAmount is the total discount reduction in cents.
	DiscountAmount int64
	// ShippingAmount is the shipping charge applied to the order in cents.
	ShippingAmount int64
	// TaxAmount is the tax amount applied to the order in cents.
	TaxAmount int64
	// CODFee is the cash-on-delivery surcharge applied to the order in cents.
	CODFee int64
	// TotalAmount is the final charged sum (subtotal - discounts + tax + shipping) in cents.
	TotalAmount int64
	// Currency is the transaction currency code (e.g. "USD").
	Currency string
	// PaymentMethod is the payment method used for the order (e.g. "cash_on_delivery", "stripe").
	PaymentMethod string
	// TrackingNumber is the carrier tracking number for shipped orders.
	TrackingNumber string
	// Locale is the customer preferred language (e.g. "en", "ar") for notifications.
	Locale string
	// Carrier is the logistics carrier name (e.g. "Aramex", "DHL", "FedEx").
	Carrier *string
	// TrackingURL is the direct tracking link provided by the carrier.
	TrackingURL *string
	// Items contains all purchased line items in this order.
	Items []OrderItem
	// StatusHistory is the order_status_history audit trail in chronological
	// order, oldest first. Empty (never nil) on detail reads.
	StatusHistory []OrderStatusHistory `json:"status_history"`
	// CreatedAt is the timestamp when the order was placed.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when order status was last modified.
	UpdatedAt time.Time
}
