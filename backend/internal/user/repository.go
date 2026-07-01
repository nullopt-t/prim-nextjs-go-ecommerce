package user

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"

	"github.com/google/uuid"
)

type IdentifierType string

const (
	IdentifierTypeEmail IdentifierType = "email"
	IdentifierTypePhone IdentifierType = "phone"
)

type Filter struct {
	ID         *uuid.UUID
	Identifier *string
}

type UserRepository struct {
}

func NewPostgresRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	user *model.User,
) (uuid.UUID, error) {
	var createdUserID uuid.UUID
	err := qe.QueryRow(
		ctx,
		`
		INSERT INTO users (
			identifier
		)
		VALUES ($1)
		RETURNING id
		`,
		user.Identifier,
	).Scan(&createdUserID)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}
	return createdUserID, nil
}

func (r *UserRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.User, error) {

	query := `
		SELECT
			id,
			identifier,
			status,
			last_login_at,
			last_login_ip,
			locked_until,
			suspended_until,
			deleted_at,
			created_at,
			updated_at
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
		query += fmt.Sprintf(" AND identifier = $%d", i)
		args = append(args, *filter.Identifier)
		i++
	}

	var u model.User

	err := qe.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.Identifier,
		&u.Status,
		&u.LastLoginAt,
		&u.LastLoginIP,
		&u.LockedUntil,
		&u.SuspendedUntil,
		&u.DeletedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &u, nil
}
func (r *UserRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) error {
	query := `UPDATE users SET deleted_at = NOW(), status = 'deleted' WHERE deleted_at IS NULL`
	args := []interface{}{}

	if filter.ID != nil {
		query += ` AND id = $1`
		args = append(args, *filter.ID)
	}
	if filter.Identifier != nil {
		query += ` AND identifier = $2`
		args = append(args, *filter.Identifier)
	}

	res, err := qe.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("delete user by id: %w", err)
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UserRepository) GetAll(
	ctx context.Context,
	qe database.QueryExecutor,
	q api.PageQuery,
) ([]model.User, api.Page, error) {

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}

	offset := (q.Page - 1) * q.PageSize

	var users []model.User

	rows, err := qe.Query(
		ctx,
		`SELECT
			id,
			identifier,
			status,
			last_login_at,
			last_login_ip,
			suspended_until,
			locked_until,
			deleted_at,
			created_at,
			updated_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		q.PageSize,
		offset,
	)
	if err != nil {
		return nil, api.Page{}, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u model.User

		if err := rows.Scan(
			&u.ID,
			&u.Identifier,
			&u.Status,
			&u.LastLoginAt,
			&u.LastLoginIP,
			&u.SuspendedUntil,
			&u.LockedUntil,
			&u.DeletedAt,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, api.Page{}, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, api.Page{}, fmt.Errorf("iterate users: %w", err)
	}

	var total int
	err = qe.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	).Scan(&total)
	if err != nil {
		return nil, api.Page{}, fmt.Errorf("count users: %w", err)
	}

	page := api.Page{
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalItems: total,
	}

	return users, page, nil
}
