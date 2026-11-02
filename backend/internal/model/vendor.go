package model

import (
	"time"

	"github.com/google/uuid"
)

// VendorStatus is the lifecycle state of a vendor profile plus how they can use
// the merchant portal and be attributed in checkout.
type VendorStatus string

const (
	// VendorStatusPending marks a submitted application awaiting admin approval.
	VendorStatusPending VendorStatus = "pending"
	// VendorStatusActive marks an approved, purchasable vendor.
	VendorStatusActive VendorStatus = "active"
	// VendorStatusSuspended marks a previously approved vendor temporarily barred from selling.
	VendorStatusSuspended VendorStatus = "suspended"
	// VendorStatusRejected marks an application denied by an administrator.
	VendorStatusRejected VendorStatus = "rejected"
)

// PayoutStatus is the lifecycle state of a payout request. Transitions are
// conditional: requested -> approved -> processing -> paid (terminal);
// rejection is only legal from requested/approved (or an explicitly-recorded
// failed transfer in processing).
type PayoutStatus string

const (
	// PayoutStatusRequested marks a newly submitted payout request.
	PayoutStatusRequested PayoutStatus = "requested"
	// PayoutStatusApproved marks a payout cleared by an administrator but not yet transferred.
	PayoutStatusApproved PayoutStatus = "approved"
	// PayoutStatusProcessing marks a payout whose external transfer is in flight.
	PayoutStatusProcessing PayoutStatus = "processing"
	// PayoutStatusPaid marks a completed payout (terminal).
	PayoutStatusPaid PayoutStatus = "paid"
	// PayoutStatusRejected marks a denied or failed payout; funds return to the balance.
	PayoutStatusRejected PayoutStatus = "rejected"
)

// Vendor represents a merchant profile: a 1:1 extension of a user account
// carrying store identity, commission terms, and marketplace standing.
type Vendor struct {
	// ID is the unique vendor profile UUID.
	ID uuid.UUID `db:"id" json:"id"`
	// UserID is the owning account (unique); carries the login identity.
	UserID uuid.UUID `db:"user_id" json:"userId"`
	// StoreName is the customer-facing merchant display name.
	StoreName string `db:"store_name" json:"storeName"`
	// Slug is the URL-safe unique store identifier.
	Slug string `db:"slug" json:"slug"`
	// Description is the optional storefront blurb.
	Description string `db:"description" json:"description"`
	// LogoObjectID / BannerObjectID reference storage_objects blobs.
	LogoObjectID *uuid.UUID `db:"logo_object_id" json:"logoObjectId,omitempty"`
	BannerObjectID *uuid.UUID `db:"banner_object_id" json:"bannerObjectId,omitempty"`
	// CommissionRateBps is the platform commission in basis points (0..10000).
	CommissionRateBps int `db:"commission_rate_bps" json:"commissionRateBps"`
	// Status is the vendor lifecycle state.
	Status VendorStatus `db:"status" json:"status"`
	// BankAccountInfo is an opaque payout-details blob; schema-free until a real payout provider is integrated.
	BankAccountInfo map[string]any `db:"bank_account_info" json:"bankAccountInfo,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
	DeletedAt *time.Time `db:"deleted_at" json:"-"`

	// LogoURL / BannerURL are resolved public URLs (hydrated, not stored).
	LogoURL string `db:"-" json:"logoUrl,omitempty"`
	BannerURL string `db:"-" json:"bannerUrl,omitempty"`
}

// VendorPayout is one entry in a vendor's payout ledger. Rows are never
// deleted: rejected requests persist for audit, and snapshot_available_balance
// records the balance the request validated against.
type VendorPayout struct {
	// ID is the unique payout record UUID.
	ID uuid.UUID `db:"id" json:"id"`
	// VendorID is the requesting vendor.
	VendorID uuid.UUID `db:"vendor_id" json:"vendorId"`
	// Amount in minor currency units (integer, > 0).
	Amount int64 `db:"amount" json:"amount"`
	// Currency is the ISO code, defaulting to EGP.
	Currency string `db:"currency" json:"currency"`
	// Status is the payout lifecycle state.
	Status PayoutStatus `db:"status" json:"status"`
	// ReferenceNumber is the external transfer reference filled in on processing/paid.
	ReferenceNumber *string `db:"reference_number" json:"referenceNumber,omitempty"`
	// Notes carries admin remarks (rejection reasons, transfer notes).
	Notes string `db:"notes" json:"notes,omitempty"`
	// SnapshotAvailableBalance is a copy of the validated available balance at request time.
	SnapshotAvailableBalance int64 `db:"snapshot_available_balance" json:"snapshotAvailableBalance"`

	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
	// ProcessedAt is when the payout reached a terminal state.
	ProcessedAt *time.Time `db:"processed_at" json:"processedAt,omitempty"`
}
