package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// userFilter defines internal query criteria for retrieving a single user by primary UUID or email identifier.
type userFilter struct {
	ID         *uuid.UUID
	Identifier *string
}

// UserRepository provides PostgreSQL database persistence operations for the users table:
// user creation, single-record lookups with case-insensitive email matching,
// dynamic paginated listing with role/status filters, and role/status mutations.
type UserRepository struct{}

// NewPostgresRepository constructs a new UserRepository instance.
func NewPostgresRepository() *UserRepository {
	return &UserRepository{}
}

// Create inserts a new user record into the users table with default or specified role tier,
// and returns the newly assigned user UUID and database-generated timestamps.
func (r *UserRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	user *model.User,
) (uuid.UUID, error) {
	role := "customer"
	if user.Role != "" {
		role = string(user.Role)
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

// GetByID retrieves an active (non-deleted) user record by primary UUID key.
func (r *UserRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.User, error) {
	return r.get(ctx, qe, userFilter{ID: &id})
}

// GetByIdentifier retrieves an active (non-deleted) user record by their email address using case-insensitive comparison.
func (r *UserRepository) GetByIdentifier(
	ctx context.Context,
	qe database.QueryExecutor,
	identifier string,
) (*model.User, error) {
	return r.get(ctx, qe, userFilter{Identifier: &identifier})
}

// get executes a parameterized SELECT query against users with deleted_at IS NULL,
// scanning account status, suspension/lockout windows, and timestamps into the model.User struct.
func (r *UserRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter userFilter,
) (*model.User, error) {
	query := `
		SELECT id, email, full_name, role, is_email_verified, status, suspended_until, locked_until, created_at, updated_at, deleted_at
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
	var statusStr string

	err := qe.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.Identifier,
		&u.FullName,
		&roleStr,
		&u.IsEmailVerified,
		&statusStr,
		&u.SuspendedUntil,
		&u.LockedUntil,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	rVal := model.UserRole(roleStr)
	u.Role = rVal
	u.Status = model.AccountStatus(statusStr)
	return &u, nil
}

// UserListFilter encapsulates filtering, search, and pagination criteria for administrative user queries.
type UserListFilter struct {
	Role   *string
	Status *string
	Search *string
	Limit  int
	Offset int
}

// ListUsers dynamically builds and executes a parameterized SQL query to retrieve
// a paginated list of non-deleted users matching role, status, or name/email substring searches.
func (r *UserRepository) ListUsers(
	ctx context.Context,
	qe database.QueryExecutor,
	filter UserListFilter,
) ([]model.User, error) {
	query := `
		SELECT id, email, full_name, role, is_email_verified, status, suspended_until, locked_until, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
	`
	args := []any{}
	i := 1

	if filter.Role != nil && *filter.Role != "" {
		query += fmt.Sprintf(" AND role = $%d::user_role", i)
		args = append(args, *filter.Role)
		i++
	}

	if filter.Status != nil && *filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, *filter.Status)
		i++
	}

	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		query += fmt.Sprintf(" AND (lower(email) LIKE lower($%d) OR lower(full_name) LIKE lower($%d))", i, i)
		args = append(args, pattern)
		i++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list users query: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		var roleStr string
		var statusStr string

		if err := rows.Scan(
			&u.ID,
			&u.Identifier,
			&u.FullName,
			&roleStr,
			&u.IsEmailVerified,
			&statusStr,
			&u.SuspendedUntil,
			&u.LockedUntil,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}

		rVal := model.UserRole(roleStr)
		u.Role = rVal
		u.Status = model.AccountStatus(statusStr)
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

// CountUsers returns the total count of non-deleted users matching the given filters,
// utilized to construct pagination metadata (total pages, total records).
func (r *UserRepository) CountUsers(
	ctx context.Context,
	qe database.QueryExecutor,
	filter UserListFilter,
) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	args := []any{}
	i := 1

	if filter.Role != nil && *filter.Role != "" {
		query += fmt.Sprintf(" AND role = $%d::user_role", i)
		args = append(args, *filter.Role)
		i++
	}

	if filter.Status != nil && *filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, *filter.Status)
		i++
	}

	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		query += fmt.Sprintf(" AND (lower(email) LIKE lower($%d) OR lower(full_name) LIKE lower($%d))", i, i)
		args = append(args, pattern)
		i++
	}

	var count int
	if err := qe.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count users query: %w", err)
	}

	return count, nil
}

// UpdateRole changes the user's authorization tier (e.g. customer, admin, superadmin).
// Returns database.ErrNotFound if the user does not exist or has been soft-deleted.
func (r *UserRepository) UpdateRole(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	role model.UserRole,
) error {
	query := `
		UPDATE users
		SET role = $1::user_role, updated_at = now()
		WHERE id = $2 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, string(role), userID)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// UpdateStatus changes the account standing and optional temporary suspension timestamp.
// Returns database.ErrNotFound if the user does not exist or has been soft-deleted.
func (r *UserRepository) UpdateStatus(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	status model.AccountStatus,
	suspendedUntil *time.Time,
) error {
	query := `
		UPDATE users
		SET status = $1, suspended_until = $2, updated_at = now()
		WHERE id = $3 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, string(status), suspendedUntil, userID)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// SoftDelete marks a user account as deleted and sets deleted_at timestamp.
// Returns database.ErrNotFound if the user does not exist or has already been soft-deleted.
func (r *UserRepository) SoftDelete(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) error {
	query := `
		UPDATE users
		SET status = 'deleted', deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	tag, err := qe.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

