package errcode

// Business Domain Error Codes for the Seller Subdomain.
const (
	// CodeSellerNotFound indicates the seller profile does not exist (or is soft-deleted).
	CodeSellerNotFound = "SELLER_NOT_FOUND"

	// CodeSellerSlugTaken indicates registration with a slug that already exists.
	CodeSellerSlugTaken = "SELLER_SLUG_TAKEN"

	// CodeSellerAlreadyRegistered indicates the user already has a seller profile.
	CodeSellerAlreadyRegistered = "SELLER_ALREADY_REGISTERED"

	// CodeSellerNotAuthorizedForAction indicates the seller is not in a state
	// allowing the requested action (e.g. profile edits while suspended).
	CodeSellerNotAuthorizedForAction = "SELLER_NOT_AUTHORIZED_FOR_ACTION"

	// CodePayoutNotFound indicates the payout record does not exist.
	CodePayoutNotFound = "PAYOUT_NOT_FOUND"

	// CodePayoutInsufficientBalance indicates the requested payout amount
	// exceeds the seller's available (settled minus outstanding) balance.
	CodePayoutInsufficientBalance = "PAYOUT_INSUFFICIENT_BALANCE"

	// CodePayoutInvalidAmount indicates a non-positive or over-limit payout amount.
	CodePayoutInvalidAmount = "PAYOUT_INVALID_AMOUNT"

	// CodePayoutInvalidStateTransition indicates a payout status change that
	// violates the lifecycle (e.g. already paid, or double processing).
	CodePayoutInvalidStateTransition = "PAYOUT_INVALID_STATE_TRANSITION"

	// CodePayoutIdempotencyConflict indicates an Idempotency-Key was reused for
	// a different withdrawal amount.
	CodePayoutIdempotencyConflict = "PAYOUT_IDEMPOTENCY_CONFLICT"
)
