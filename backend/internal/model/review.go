package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

var ErrInvalidReviewStatus = errors.New("invalid review status")

func ParseReviewStatus(s string) (ReviewStatus, error) {
	switch s {
	case "pending":
		return ReviewStatusPending, nil
	case "approved":
		return ReviewStatusApproved, nil
	case "rejected":
		return ReviewStatusRejected, nil
	default:
		return "", ErrInvalidReviewStatus
	}
}

func (s ReviewStatus) String() string {
	return string(s)
}

type Review struct {
	ID          uuid.UUID    `db:"id"`
	ProductID   uuid.UUID    `db:"product_id"`
	UserID      uuid.UUID    `db:"user_id"`
	OrderItemID uuid.UUID    `db:"order_item_id"`
	Rating      int16        `db:"rating"`
	Title       *string      `db:"title"`
	Body        *string      `db:"body"`
	Status      ReviewStatus `db:"status"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}
