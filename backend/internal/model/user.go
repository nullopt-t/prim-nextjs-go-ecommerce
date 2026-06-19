package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
)

type UserRole string

const (
	CustomerRole UserRole = "customer"
	AdminRole    UserRole = "admin"
	VendorRole   UserRole = "vendor"
)

func (ur *UserRole) String() string {
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
	ID    uuid.UUID
	Email string
	Phone *string

	EmailVerifiedAt *time.Time
	PhoneVerifiedAt *time.Time

	PasswordHash      []byte
	PasswordChangedAt *time.Time

	LastLoginAt *time.Time
	LastLoginIP *string

	Role   UserRole
	Status AccountStatus

	SuspendedUntil *time.Time

	LockedUntil *time.Time

	DeletedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(
	email string,
	passwordHash []byte,
) *User {
	return &User{
		Email:        email,
		PasswordHash: passwordHash,
	}
}

func (u *User) IsAdmin() bool {
	return u.Role == AdminRole
}

func (u *User) IsVendor() bool {
	return u.Role == VendorRole
}

func (u *User) IsCustomer() bool {
	return u.Role == CustomerRole
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) IsSuspended() bool {
	return u.Status == StatusSuspended
}

func (u *User) VerifyPassword(password string) error {
	return utils.ComparePassword(u.PasswordHash, password)
}

func (u *User) String() string {
	email := u.Email
	if email == "" {
		email = "nil"
	}

	return fmt.Sprintf(
		"user{id=%s, email=%s, role=%s, status=%s}",
		u.ID,
		email,
		u.Role,
		u.Status,
	)
}
