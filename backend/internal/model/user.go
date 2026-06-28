package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RoleCode string

const (
	OwnerRole  RoleCode = "OWNER"
	AdminRole  RoleCode = "ADMIN"
	VendorRole RoleCode = "VENDOR"
)

func (ur *RoleCode) String() string {
	return string(*ur)
}

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

func NewUser(
	identifier string,
) *User {
	return &User{
		Identifier: identifier,
	}
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
