package errcode

// Business Domain Error Codes for Coupon Subdomain.
// These constants define the canonical domain failure types returned when validating or applying coupons.
const (
	// CodeCouponNotFound indicates that the requested coupon code or ID does not exist.
	CodeCouponNotFound = "COUPON_NOT_FOUND"

	// CodeCouponAlreadyExists indicates that a coupon with this code already exists.
	CodeCouponAlreadyExists = "COUPON_ALREADY_EXISTS"

	// CodeCouponInactive indicates that the coupon status is not active.
	CodeCouponInactive = "COUPON_INACTIVE"

	// CodeCouponNotStarted indicates that the coupon start timestamp is in the future.
	CodeCouponNotStarted = "COUPON_NOT_STARTED"

	// CodeCouponExpired indicates that the coupon expiration timestamp has passed.
	CodeCouponExpired = "COUPON_EXPIRED"

	// CodeCouponUsageLimitExceeded indicates that the global redemption quota has been met.
	CodeCouponUsageLimitExceeded = "COUPON_USAGE_LIMIT_EXCEEDED"

	// CodeCouponUserLimitExceeded indicates that the user has already redeemed the maximum allowable times.
	CodeCouponUserLimitExceeded = "COUPON_USER_LIMIT_EXCEEDED"

	// CodeCouponMinCartNotMet indicates that the cart subtotal does not meet the coupon's minimum requirement.
	CodeCouponMinCartNotMet = "COUPON_MIN_CART_NOT_MET"
)
