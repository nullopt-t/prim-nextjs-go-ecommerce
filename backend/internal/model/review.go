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
	ID          uuid.UUID
	ProductID   uuid.UUID
	UserID      uuid.UUID
	OrderItemID uuid.UUID
	Rating      int16
	Title       *string
	Body        *string
	Status      ReviewStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RatingSummary struct {
	AverageRating float64
	ReviewCount   int
	Distribution  map[int16]int
}
