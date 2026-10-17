// Package errcode defines standard machine-readable business error codes used
// across the application's domains (catalog, order, checkout, ...).
// These business error codes are returned in API error responses under the "code"
// field, allowing client applications to programmatically identify specific error
// conditions without relying on brittle HTTP status codes or human-readable messages.
//
// Codes are grouped by subdomain in sibling files: product.go, variant.go,
// brand.go, category.go, tag.go, media.go, inventory.go, review.go, wishlist.go,
// coupon.go, and order.go.
package errcode
