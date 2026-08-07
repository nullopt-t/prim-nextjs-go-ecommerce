package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type userFilter struct {
	ID         *uuid.UUID
	Identifier *string
}

type UserRepository struct{}

func NewPostgresRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	user *model.User,
) (uuid.UUID, error) {
	role := "customer"
	if user.Role != nil {
		role = string(*user.Role)
	}

	query := `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, now(), now())
		RETURNING id, created_at, updated_at
	`
	var createdUserID uuid.UUID
	err := qe.QueryRow(ctx, query, user.ID, user.Identifier, role).Scan(&createdUserID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}
	return createdUserID, nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.User, error) {
	return r.get(ctx, qe, userFilter{ID: &id})
}

func (r *UserRepository) GetByIdentifier(
	ctx context.Context,
	qe database.QueryExecutor,
	identifier string,
) (*model.User, error) {
	return r.get(ctx, qe, userFilter{Identifier: &identifier})
}

func (r *UserRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter userFilter,
) (*model.User, error) {
	query := `
		SELECT id, email, role, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
	`
	args := []any{}
	i := 1

	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", i)
		args = append(args, *filter.ID)
		i++
	}

	if filter.Identifier != nil {
		query += fmt.Sprintf(" AND lower(email) = lower($%d)", i)
		args = append(args, *filter.Identifier)
	}

	var u model.User
	var roleStr string

	err := qe.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.Identifier,
		&roleStr,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	rVal := model.UserRole(roleStr)
	u.Role = &rVal
	u.Status = model.StatusActive
	return &u, nil
}
