package errcode

// Business Domain Error Codes for the Vendor Subdomain.
const (
	// CodeVendorNotFound indicates the vendor profile does not exist (or is soft-deleted).
	CodeVendorNotFound = "VENDOR_NOT_FOUND"

	// CodeVendorSlugTaken indicates registration with a slug that already exists.
	CodeVendorSlugTaken = "VENDOR_SLUG_TAKEN"

	// CodeVendorAlreadyRegistered indicates the user already has a vendor profile.
	CodeVendorAlreadyRegistered = "VENDOR_ALREADY_REGISTERED"

	// CodeVendorNotAuthorizedForAction indicates the vendor is not in a state
	// allowing the requested action (e.g. profile edits while suspended).
	CodeVendorNotAuthorizedForAction = "VENDOR_NOT_AUTHORIZED_FOR_ACTION"

	// CodePayoutNotFound indicates the payout record does not exist.
	CodePayoutNotFound = "PAYOUT_NOT_FOUND"

	// CodePayoutInsufficientBalance indicates the requested payout amount
	// exceeds the vendor's available (settled minus outstanding) balance.
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
