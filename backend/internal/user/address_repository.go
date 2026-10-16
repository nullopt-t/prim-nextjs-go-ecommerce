package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

var (
	// ErrAddressNotFound indicates that no active address matching the specified ID and UserID was found.
	ErrAddressNotFound = errors.New("address not found")
)

// AddressRepository handles persistence of user addresses in PostgreSQL.
// All queries enforce user_id tenancy filters and soft-deletion constraints (deleted_at IS NULL)
// to guarantee multi-tenant data isolation and prevent unauthorized data access across users.
type AddressRepository struct{}

// NewAddressRepository creates a new AddressRepository.
func NewAddressRepository() *AddressRepository {
	return &AddressRepository{}
}

// CountByUser returns the count of non-deleted addresses for a user.
// Used during creation to determine if the new address should automatically become the default address.
func (r *AddressRepository) CountByUser(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (int, error) {
	query := `SELECT COUNT(*) FROM addresses WHERE user_id = $1 AND deleted_at IS NULL`
	var count int
	err := qe.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count addresses by user: %w", err)
	}
	return count, nil
}

// ListByUser retrieves all non-deleted addresses for a user, sorted with default first, then newest updated.
func (r *AddressRepository) ListByUser(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) ([]model.UserAddress, error) {
	query := `
		SELECT 
			id, user_id, label, full_name, phone, line1, line2,
			city, state, postal_code, country, is_default,
			created_at, updated_at
		FROM addresses
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY is_default DESC, updated_at DESC, created_at DESC
	`
	rows, err := qe.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses by user: %w", err)
	}
	defer rows.Close()

	var list []model.UserAddress
	for rows.Next() {
		var a model.UserAddress
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.Label,
			&a.FullName,
			&a.Phone,
			&a.Line1,
			&a.Line2,
			&a.City,
			&a.State,
			&a.PostalCode,
			&a.Country,
			&a.IsDefault,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan address: %w", err)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error listing addresses: %w", err)
	}
	return list, nil
}

// GetByIDAndUser retrieves a single non-deleted address belonging to the user.
// Scoping by user_id ensures that users cannot view addresses belonging to other customers.
func (r *AddressRepository) GetByIDAndUser(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
	userID uuid.UUID,
) (*model.UserAddress, error) {
	query := `
		SELECT 
			id, user_id, label, full_name, phone, line1, line2,
			city, state, postal_code, country, is_default,
			created_at, updated_at
		FROM addresses
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	var a model.UserAddress
	err := qe.QueryRow(ctx, query, id, userID).Scan(
		&a.ID,
		&a.UserID,
		&a.Label,
		&a.FullName,
		&a.Phone,
		&a.Line1,
		&a.Line2,
		&a.City,
		&a.State,
		&a.PostalCode,
		&a.Country,
		&a.IsDefault,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAddressNotFound
		}
		return nil, fmt.Errorf("get address by id and user: %w", err)
	}
	return &a, nil
}

// ClearDefault unsets is_default for all non-deleted addresses of the specified user.
// Executed inside database transactions prior to setting a new default address.
func (r *AddressRepository) ClearDefault(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) error {
	query := `
		UPDATE addresses
		SET is_default = false, updated_at = now()
		WHERE user_id = $1 AND is_default = true AND deleted_at IS NULL
	`
	_, err := qe.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("clear default address: %w", err)
	}
	return nil
}

// SetDefault marks a specific address as default for the user.
func (r *AddressRepository) SetDefault(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	query := `
		UPDATE addresses
		SET is_default = true, updated_at = now()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("set default address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAddressNotFound
	}
	return nil
}

// Create inserts a new address.
// Default address uniqueness per user is enforced at the database level by the partial unique index:
// idx_addresses_user_default ON addresses (user_id) WHERE is_default = true AND deleted_at IS NULL.
func (r *AddressRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	a *model.UserAddress,
) error {
	query := `
		INSERT INTO addresses (
			id, user_id, label, full_name, phone, line1, line2,
			city, state, postal_code, country, is_default,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			now(), now()
		)
		RETURNING created_at, updated_at
	`
	return qe.QueryRow(
		ctx,
		query,
		a.ID,
		a.UserID,
		a.Label,
		a.FullName,
		a.Phone,
		a.Line1,
		a.Line2,
		a.City,
		a.State,
		a.PostalCode,
		a.Country,
		a.IsDefault,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

// Update updates fields of an existing address.
func (r *AddressRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	a *model.UserAddress,
) error {
	query := `
		UPDATE addresses
		SET 
			label = $1,
			full_name = $2,
			phone = $3,
			line1 = $4,
			line2 = $5,
			city = $6,
			state = $7,
			postal_code = $8,
			country = $9,
			is_default = $10,
			updated_at = now()
		WHERE id = $11 AND user_id = $12 AND deleted_at IS NULL
		RETURNING updated_at
	`
	err := qe.QueryRow(
		ctx,
		query,
		a.Label,
		a.FullName,
		a.Phone,
		a.Line1,
		a.Line2,
		a.City,
		a.State,
		a.PostalCode,
		a.Country,
		a.IsDefault,
		a.ID,
		a.UserID,
	).Scan(&a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAddressNotFound
		}
		return fmt.Errorf("update address: %w", err)
	}
	return nil
}

// SoftDelete marks an address as deleted and unsets is_default to prevent deleted addresses from acting as defaults.
func (r *AddressRepository) SoftDelete(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	query := `
		UPDATE addresses
		SET deleted_at = now(), is_default = false, updated_at = now()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("soft delete address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAddressNotFound
	}
	return nil
}

// FindMostRecentActive retrieves the most recently updated or created active address for promotion.
// Invoked when the user's current default address is deleted, promoting the next active address to default.
func (r *AddressRepository) FindMostRecentActive(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (*model.UserAddress, error) {
	query := `
		SELECT 
			id, user_id, label, full_name, phone, line1, line2,
			city, state, postal_code, country, is_default,
			created_at, updated_at
		FROM addresses
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY updated_at DESC, created_at DESC
		LIMIT 1
	`
	var a model.UserAddress
	err := qe.QueryRow(ctx, query, userID).Scan(
		&a.ID,
		&a.UserID,
		&a.Label,
		&a.FullName,
		&a.Phone,
		&a.Line1,
		&a.Line2,
		&a.City,
		&a.State,
		&a.PostalCode,
		&a.Country,
		&a.IsDefault,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAddressNotFound
		}
		return nil, fmt.Errorf("find most recent active address: %w", err)
	}
	return &a, nil
}
