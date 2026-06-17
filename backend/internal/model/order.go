package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefuneded  OrderStatus = "refuneded"
)

type OrderItem struct {
	ProductID     uuid.UUID
	Quantity      int
	PurchasePrice int
}

type Order struct {
	// order data
	ID         uuid.UUID
	Status     OrderStatus
	TotalPrice int

	// items data
	Items []OrderItem

	// timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}
