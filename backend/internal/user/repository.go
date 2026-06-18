package user

import (
	"context"
	"database/sql"
	"errors"
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
	ID    *uuid.UUID
	Email *string
	Phone *string
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
) error {
	var createdUserID uuid.UUID
	err := qe.QueryRow(
		ctx,
		`
		INSERT INTO users (
			email,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING id
		`,
		user.Email,
		user.PasswordHash,
	).Scan(&createdUserID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.User, error) {

	query := `
		SELECT
			id,
			email,
			phone,
			email_verified_at,
			phone_verified_at,
			role,
			status,
			password_hash,
			password_changed_at,
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

	if filter.Email != nil {
		query += fmt.Sprintf(" AND email = $%d", i)
		args = append(args, *filter.Email)
		i++
	}

	if filter.Phone != nil {
		query += fmt.Sprintf(" AND phone = $%d", i)
		args = append(args, *filter.Phone)
	}

	var u model.User

	err := qe.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.Email,
		&u.Phone,
		&u.EmailVerifiedAt,
		&u.PhoneVerifiedAt,
		&u.Role,
		&u.Status,
		&u.PasswordHash,
		&u.PasswordChangedAt,
		&u.LastLoginAt,
		&u.LastLoginIP,
		&u.LockedUntil,
		&u.SuspendedUntil,
		&u.DeletedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
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
	if filter.Email != nil {
		query += ` AND email = $2`
		args = append(args, *filter.Email)
	}
	if filter.Phone != nil {
		query += ` AND phone = $3`
		args = append(args, *filter.Phone)
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
			email,
			phone,
			email_verified_at,
			phone_verified_at,
			role,
			status,
			password_hash,
			password_changed_at,
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
			&u.Email,
			&u.Phone,
			&u.EmailVerifiedAt,
			&u.PhoneVerifiedAt,
			&u.Role,
			&u.Status,
			&u.PasswordHash,
			&u.PasswordChangedAt,
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
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    total,
	}

	return users, page, nil
}

func (r *UserRepository) VerifyIdentifier(
	ctx context.Context,
	qe database.QueryExecutor,
	identifierType IdentifierType,
	identifier string,
) error {
	query := `UPDATE users SET email_verified_at = NOW() WHERE deleted_at IS NULL AND email = $1`
	if identifierType == IdentifierTypePhone {
		query = `UPDATE users SET phone_verified_at = NOW() WHERE deleted_at IS NULL AND phone = $1`
	}

	res, err := qe.Exec(
		ctx,
		query,
		identifier,
	)
	if err != nil {
		return fmt.Errorf("verify user identifier: %w", err)
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UserRepository) UpdatePasswordHash(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
	newPassHash []byte,
) error {

	query := `UPDATE users SET password_hash = $1,
		updated_at = NOW()
	WHERE deleted_at IS NULL`

	args := []any{newPassHash}
	i := 2

	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", i)
		args = append(args, *filter.ID)
		i++
	}

	if filter.Email != nil {
		query += fmt.Sprintf(" AND email = $%d", i)
		args = append(args, *filter.Email)
		i++
	}

	if filter.Phone != nil {
		query += fmt.Sprintf(" AND phone = $%d", i)
		args = append(args, *filter.Phone)
	}

	res, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
