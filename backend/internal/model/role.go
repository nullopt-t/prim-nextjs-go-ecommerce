package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID        int
	Code      string
	CreatedAt time.Time
}

type UserRole struct {
	UserID uuid.UUID
	RoleID int
}

func RolesToStrings(roles []*Role) []string {
	roleCodes := make([]string, 0, len(roles))

	for _, role := range roles {
		roleCodes = append(roleCodes, strings.ToLower(role.Code))
	}

	return roleCodes
}
