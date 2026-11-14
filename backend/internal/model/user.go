package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserRole defines the authorization permission tier for administrative or staff accounts.
type UserRole string

const (
	// SuperRole designates the system superadministrator with unrestricted platform access.
	SuperRole UserRole = "superadmin"
	// AdminRole designates staff members with store and catalog management permissions.
	AdminRole UserRole = "admin"
	// SupportRole designates staff members with customer service and support permissions.
	SupportRole UserRole = "support"
	// SellerRole designates merchant/seller partners with localized product access.
	SellerRole UserRole = "seller"
	// CustomerRole designates standard shoppers.
	CustomerRole UserRole = "customer"
)

// AccountStatus represents the state of a customer or staff user account.
type AccountStatus string

const (
	// StatusActive indicates the account is in good standing and can authenticate.
	StatusActive AccountStatus = "active"
	// StatusSuspended indicates the account is temporarily blocked from access.
	StatusSuspended AccountStatus = "suspended"
	// StatusDeleted indicates the account has been soft-deleted or closed.
	StatusDeleted AccountStatus = "deleted"
)

// User represents an account identity (customer, staff, or seller) in the system.
type User struct {
	// ID is the unique database UUID for the user account.
	ID uuid.UUID

	// Phone is the primary authentication handle (phone number).
	Phone string
	// FullName is the user's display name.
	FullName string
	// Role is the administrative role tier, defaulting to CustomerRole.
	Role UserRole
	// PhoneVerified indicates whether the user confirmed their phone number.
	PhoneVerified bool
	// CountryCode is the user's country calling code (e.g. +20).
	CountryCode string

	// PhoneVerifiedAt is the timestamp when the user confirmed phone ownership.
	PhoneVerifiedAt *time.Time

	// LastLoginAt is the timestamp of the most recent successful authentication.
	LastLoginAt *time.Time
	// LastLoginIP is the IPv4/IPv6 address used during the last login session.
	LastLoginIP *string

	// Status is the account operational state (active, suspended, deleted).
	Status AccountStatus

	// SuspendedUntil is the expiration timestamp for temporary suspensions (nil if indefinite).
	SuspendedUntil *time.Time

	// DeletedAt is the optional timestamp when the account was soft-deleted.
	DeletedAt *time.Time

	// Preference stores arbitrary user preferences as key-value pairs.
	Preference map[string]any

	// CreatedAt is the timestamp when the account registered.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when user profile/security metadata was last modified.
	UpdatedAt time.Time
}

// IsActive returns true if the user's account status is active and not suspended.
func (u *User) IsActive() bool {
	if u.Status != StatusActive {
		return false
	}
	if u.SuspendedUntil != nil && time.Now().Before(*u.SuspendedUntil) {
		return false
	}
	return true
}

// IsSuspended returns true if the user's account status is currently suspended.
func (u *User) IsSuspended() bool {
	return u.Status == StatusSuspended || (u.SuspendedUntil != nil && time.Now().Before(*u.SuspendedUntil))
}

// String returns a safe debug summary of the User record (excluding sensitive timestamps).
func (u *User) String() string {
	return fmt.Sprintf(
		"user{id=%s, phone=%s, status=%s}",
		u.ID,
		u.Phone,
		u.Status,
	)
}
