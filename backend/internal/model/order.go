package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCanceled   OrderStatus = "canceled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type OrderItem struct {
	ID              uuid.UUID
	OrderID         uuid.UUID
	VariantID       uuid.UUID
	Quantity        int
	PriceAtPurchase int64
	ProductSnapshot string
}

type Address struct {
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
}

type Order struct {
	ID              uuid.UUID
	CustomerID      *uuid.UUID
	CustomerEmail   string
	ShippingAddress Address
	BillingAddress  Address
	Status          OrderStatus
	CouponID        *uuid.UUID
	DiscountAmount  int64
	TotalAmount     int64
	Currency        string
	Items           []OrderItem
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
