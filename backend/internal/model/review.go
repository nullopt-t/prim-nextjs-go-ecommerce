package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ReviewStatus represents the moderation lifecycle state of a customer review.
type ReviewStatus string

const (
	// ReviewStatusPending indicates the review is submitted and awaiting moderation review.
	ReviewStatusPending ReviewStatus = "pending"
	// ReviewStatusApproved indicates the review passed moderation and is visible on product pages.
	ReviewStatusApproved ReviewStatus = "approved"
	// ReviewStatusRejected indicates the review violated policies and is hidden from public display.
	ReviewStatusRejected ReviewStatus = "rejected"
)

// ErrInvalidReviewStatus is returned when an invalid review status string is provided.
var ErrInvalidReviewStatus = errors.New("invalid review status")

// ParseReviewStatus converts a raw string into a typed ReviewStatus.
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

// String returns the string representation of ReviewStatus.
func (s ReviewStatus) String() string {
	return string(s)
}

// Review represents a verified customer evaluation and rating for a purchased product.
type Review struct {
	// ID is the unique database UUID for the review.
	ID uuid.UUID
	// ProductID is the reviewed Product UUID.
	ProductID uuid.UUID
	// UserID is the author Customer UUID.
	UserID uuid.UUID
	// AuthorName is the reviewer's display name, snapshotted from the user
	// account when the review is created and stored as reviews.user_name so
	// listing reviews needs no join and the name survives account deletion.
	AuthorName string
	// ProductTitle is the reviewed product's title, resolved live at listing
	// time so profile pages can link each review to its product.
	ProductTitle string
	// ProductImageBucket and ProductImageKey locate the reviewed product's
	// thumbnail in storage, joined at listing time for the reviewer profile.
	ProductImageBucket *string
	// ProductImageKey is the storage object key of the product thumbnail.
	ProductImageKey *string
	// ProductImageURL is the resolved public URL of the product thumbnail.
	ProductImageURL string
	// OrderItemID is the purchased line item UUID verifying buyer authenticity.
	OrderItemID uuid.UUID
	// Rating is the numerical score between 1 and 5 stars.
	Rating int16
	// Title is the optional review headline.
	Title *string
	// Body is the full customer review narrative text.
	Body *string
	// Status is the moderation state (pending, approved, rejected).
	Status ReviewStatus
	// HelpfulCount is the number of users who found this review helpful.
	HelpfulCount int
	// UserVoted indicates whether the requesting user voted this review helpful.
	UserVoted bool
	// CreatedAt is the timestamp when the review was submitted.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when the review was last modified.
	UpdatedAt time.Time
}

// RatingSummary contains aggregated rating statistics for a product.
type RatingSummary struct {
	// AverageRating is the arithmetic mean score (1.0 to 5.0).
	AverageRating float64
	// ReviewCount is the total number of approved customer reviews.
	ReviewCount int
	// Distribution maps each star score (1-5) to its frequency count.
	Distribution map[int16]int
}

// ReviewFilter narrows a review listing. A nil field means "no constraint".
// It lives in the model package because both the catalog and the product
// domain list reviews and must not depend on each other.
type ReviewFilter struct {
	// ProductID limits the listing to one product's reviews.
	ProductID *uuid.UUID
	// UserID limits the listing to one author's reviews.
	UserID *uuid.UUID
	// Status limits the listing to one moderation status.
	Status *ReviewStatus
	// Rating limits the listing to one star score.
	Rating *int16
	// VoterID marks the reviews this voter has already voted helpful.
	VoterID *uuid.UUID
}
