package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// userFilter defines internal query criteria for retrieving a single user by primary UUID or phone identifier.
type userFilter struct {
	ID    *uuid.UUID
	Phone *string
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
	user *model.User,
) (uuid.UUID, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	role := "customer"
	if user.Role != "" {
		role = string(user.Role)
	}

	fullName := strings.TrimSpace(user.FullName)
	if fullName == "" {
		fullName = user.Phone
	}
	user.FullName = fullName

	query := `
		INSERT INTO users (id, full_name, phone, phone_verified, country_code, role, preference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
		RETURNING id, created_at, updated_at
	`
	var createdUserID uuid.UUID
	err = tx.QueryRowContext(ctx, query, user.ID, fullName, user.Phone, user.PhoneVerified, user.CountryCode, role, "{}").Scan(&createdUserID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create user: %w", err)
	}
	return createdUserID, nil
}

// GetByID retrieves an active (non-deleted) user record by primary UUID key.
func (r *UserRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.User, error) {
	return r.get(ctx, userFilter{ID: &id})
}

// GetByPhone retrieves an active (non-deleted) user record by their phone number using case-insensitive comparison.
func (r *UserRepository) GetByPhone(
	ctx context.Context,
	phone string,
) (*model.User, error) {
	return r.get(ctx, userFilter{Phone: &phone})
}

// ExistsByID reports whether a non-deleted user with the given UUID exists.
// It runs a lightweight existence probe (no row materialization) so session
// validation on the request path does not pay the cost of a full user fetch.
func (r *UserRepository) ExistsByID(
	ctx context.Context,
	id uuid.UUID,
) (bool, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return false, err
	}

	query := `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`

	var exists bool
	if err := executor.QueryRowContext(ctx, query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user existence: %w", err)
	}
	return exists, nil
}

// get executes a parameterized SELECT query against users with deleted_at IS NULL,
// scanning account status, suspension windows, and timestamps into the model.User struct.
func (r *UserRepository) get(
	ctx context.Context,
	filter userFilter,
) (*model.User, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, full_name, phone, phone_verified, country_code, role, status, suspended_until, created_at, updated_at, deleted_at
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

	if filter.Phone != nil {
		query += fmt.Sprintf(" AND lower(phone) = lower($%d)", i)
		args = append(args, *filter.Phone)
	}

	var u model.User
	var roleStr string
	var statusStr string

	err = executor.QueryRowContext(ctx, query, args...).Scan(
		&u.ID,
		&u.FullName,
		&u.Phone,
		&u.PhoneVerified,
		&u.CountryCode,
		&roleStr,
		&statusStr,
		&u.SuspendedUntil,
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
	// Cursor, when set, switches the query to keyset pagination on (created_at, id) DESC.
	Cursor *pagination.Keyset
}

// ListUsers dynamically builds and executes a parameterized SQL query to retrieve
// a paginated list of non-deleted users matching role, status, or name/phone substring searches.
func (r *UserRepository) ListUsers(
	ctx context.Context,
	filter UserListFilter,
) ([]model.User, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, full_name, phone, phone_verified, country_code, role, status, suspended_until, created_at, updated_at, deleted_at
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
		query += fmt.Sprintf(" AND (lower(phone) LIKE lower($%d) OR lower(full_name) LIKE lower($%d))", i, i)
		args = append(args, pattern)
		i++
	}

	// Keyset pagination: restrict to rows strictly older than the cursor, then
	// over-fetch a single extra row so the caller can detect whether more remain.
	keysetClause, keysetArgs := pagination.KeysetClause("", filter.Cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, filter.Limit+1)

	rows, err := executor.QueryContext(ctx, query, args...)
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
			&u.FullName,
			&u.Phone,
			&u.PhoneVerified,
			&u.CountryCode,
			&roleStr,
			&statusStr,
			&u.SuspendedUntil,
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
	filter UserListFilter,
) (int, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return 0, err
	}

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
		query += fmt.Sprintf(" AND (lower(phone) LIKE lower($%d) OR lower(full_name) LIKE lower($%d))", i, i)
		args = append(args, pattern)
	}

	var count int
	if err := executor.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count users query: %w", err)
	}

	return count, nil
}

// ListByRoles returns every non-deleted user whose role is in the given set.
// It is intentionally unpaginated: session oversight only needs the (small)
// population of staff accounts, never the full customer table.
func (r *UserRepository) ListByRoles(
	ctx context.Context,
	roles []model.UserRole,
) ([]model.User, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(roles))
	for i, role := range roles {
		names[i] = string(role)
	}

	query := `
		SELECT id, full_name, phone, phone_verified, country_code, role, status, suspended_until, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL AND role::text = ANY($1)
		ORDER BY created_at
	`

	rows, err := executor.QueryContext(ctx, query, names)
	if err != nil {
		return nil, fmt.Errorf("list users by roles query: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		var roleStr string
		var statusStr string

		if err := rows.Scan(
			&u.ID,
			&u.FullName,
			&u.Phone,
			&u.PhoneVerified,
			&u.CountryCode,
			&roleStr,
			&statusStr,
			&u.SuspendedUntil,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user by role row: %w", err)
		}

		u.Role = model.UserRole(roleStr)
		u.Status = model.AccountStatus(statusStr)
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("users by roles iteration error: %w", err)
	}

	return users, nil
}

// UpdateRole changes the user's authorization tier (e.g. customer, admin, superadmin).
// Returns database.ErrNotFound if the user does not exist or has been soft-deleted.
func (r *UserRepository) UpdateRole(
	ctx context.Context,
	userID uuid.UUID,
	role model.UserRole,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET role = $1::user_role, updated_at = now()
		WHERE id = $2 AND deleted_at IS NULL
	`
	tag, err := tx.ExecContext(ctx, query, string(role), userID)
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
	userID uuid.UUID,
	status model.AccountStatus,
	suspendedUntil *time.Time,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET status = $1::user_status, suspended_until = $2, updated_at = now()
		WHERE id = $3 AND deleted_at IS NULL
	`
	tag, err := tx.ExecContext(ctx, query, string(status), suspendedUntil, userID)
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
	userID uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET status = 'deleted', deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	tag, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// UpdateProfile updates the display name and/or phone number of a user.
func (r *UserRepository) UpdateProfile(
	ctx context.Context,
	id uuid.UUID,
	fullName *string,
	phone *string,
) (*model.User, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	setClauses := []string{"updated_at = now()"}
	args := []any{id}
	argIdx := 2

	if fullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argIdx))
		args = append(args, strings.TrimSpace(*fullName))
		argIdx++
	}

	if phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argIdx))
		trimmed := strings.TrimSpace(*phone)
		if trimmed == "" {
			args = append(args, nil)
		} else {
			args = append(args, trimmed)
		}
	}

	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, full_name, phone, phone_verified, country_code, role, status, suspended_until, created_at, updated_at, deleted_at
	`, strings.Join(setClauses, ", "))

	var u model.User
	var roleStr string
	var statusStr string

	err = executor.QueryRowContext(ctx, query, args...).Scan(
		&u.ID,
		&u.FullName,
		&u.Phone,
		&u.PhoneVerified,
		&u.CountryCode,
		&roleStr,
		&statusStr,
		&u.SuspendedUntil,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	u.Role = model.UserRole(roleStr)
	u.Status = model.AccountStatus(statusStr)
	return &u, nil
}
