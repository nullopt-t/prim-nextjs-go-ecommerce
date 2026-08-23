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
	SuperRole UserRole = "super"
	// AdminRole designates staff members with store and catalog management permissions.
	AdminRole UserRole = "admin"
	// VendorRole designates merchant/seller partners with localized product access.
	VendorRole UserRole = "vendor"
)

// AccountStatus represents the state of a customer or staff user account.
type AccountStatus string

const (
	// StatusActive indicates the account is in good standing and can authenticate.
	StatusActive AccountStatus = "active"
	// StatusInactive indicates the account has not been activated or verified.
	StatusInactive AccountStatus = "inactive"
	// StatusSuspended indicates the account is temporarily blocked from access.
	StatusSuspended AccountStatus = "suspended"
	// StatusDeleted indicates the account has been soft-deleted or closed.
	StatusDeleted AccountStatus = "deleted"
)

// User represents an account identity (customer, staff, or vendor) in the system.
type User struct {
	// ID is the unique database UUID for the user account.
	ID uuid.UUID

	// Identifier is the primary authentication handle (e.g. email address or phone).
	Identifier string
	// Role is the administrative role tier (nil for regular store customers).
	Role *UserRole

	// EmailVerifiedAt is the timestamp when the user confirmed email ownership.
	EmailVerifiedAt *time.Time
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

	// LockedUntil is the expiration timestamp for temporary rate-limit / brute-force security lockouts.
	LockedUntil *time.Time

	// DeletedAt is the optional timestamp when the account was soft-deleted.
	DeletedAt *time.Time

	// CreatedAt is the timestamp when the account registered.
	CreatedAt time.Time
	// UpdatedAt is the timestamp when user profile/security metadata was last modified.
	UpdatedAt time.Time
}

// IsActive returns true if the user's account status is active.
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsSuspended returns true if the user's account status is currently suspended.
func (u *User) IsSuspended() bool {
	return u.Status == StatusSuspended
}

// String returns a safe debug summary of the User record (excluding sensitive timestamps).
func (u *User) String() string {
	return fmt.Sprintf(
		"user{id=%s, identifier=%s, status=%s}",
		u.ID,
		u.Identifier,
		u.Status,
	)
}
