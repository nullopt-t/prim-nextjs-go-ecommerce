package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	SuperRole  UserRole = "super"
	AdminRole  UserRole = "admin"
	VendorRole UserRole = "vendor"
)

type AccountStatus string

const (
	StatusActive    AccountStatus = "active"
	StatusInactive  AccountStatus = "inactive"
	StatusSuspended AccountStatus = "suspended"
	StatusDeleted   AccountStatus = "deleted"
)

type User struct {
	ID uuid.UUID

	Identifier string
	Role       *UserRole

	EmailVerifiedAt *time.Time
	PhoneVerifiedAt *time.Time

	LastLoginAt *time.Time
	LastLoginIP *string

	Status AccountStatus

	SuspendedUntil *time.Time

	LockedUntil *time.Time

	DeletedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) IsSuspended() bool {
	return u.Status == StatusSuspended
}

func (u *User) String() string {
	return fmt.Sprintf(
		"user{id=%s, identifier=%s, status=%s}",
		u.ID,
		u.Identifier,
		u.Status,
	)
}
