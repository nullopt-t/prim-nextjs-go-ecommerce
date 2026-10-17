package errcode

// Business Domain Error Codes for Review Subdomain.
// These constants define the canonical domain failure types returned by review moderation operations.
const (
	// CodeReviewAlreadyExists indicates that the user has already submitted a product review for this item.
	CodeReviewAlreadyExists = "REVIEW_ALREADY_EXISTS"

	// CodeReviewNotFound indicates that the requested review UUID does not exist.
	CodeReviewNotFound = "REVIEW_NOT_FOUND"

	// CodeReviewNotEligible indicates that the user has not purchased the item and cannot post a verified review.
	CodeReviewNotEligible = "REVIEW_NOT_ELIGIBLE"
)
