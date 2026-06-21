package role

import (
	"context"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type RoleRepository struct{}

func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

func (r *RoleRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	roleID int,
) (*model.Role, error) {
	query := `
	SELECT
		id,
		code,
		created_at
	FROM
		roles
	WHERE id = $1
	`

	var role model.Role
	err := qe.QueryRow(
		ctx,
		query,
		roleID,
	).Scan(
		&role.ID,
		&role.Code,
		&role.CreatedAt,
	)

	return &role, err
}

func (r *RoleRepository) GetByCode(
	ctx context.Context,
	qe database.QueryExecutor,
	code model.RoleCode,
) (*model.Role, error) {
	query := `
	SELECT
		id,
		code,
		created_at
	FROM
		roles
	WHERE code = $1
	`

	var role model.Role
	err := qe.QueryRow(
		ctx,
		query,
		code,
	).Scan(
		&role.ID,
		&role.Code,
		&role.CreatedAt,
	)

	return &role, err
}

func (r *RoleRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q api.PageQuery,
) ([]*model.Role, *api.Page, error) {

	var total int
	err := qe.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM roles`,
	).Scan(&total)
	if err != nil {
		return nil, nil, err
	}

	offset := (q.Page - 1) * q.PageSize
	len := min(q.PageSize, total)
	var roles = make([]*model.Role, 0, len)
	query := `
	SELECT
		id,
		code,
		created_at
	FROM
		roles
	LIMIT $1
	OFFSET $2
	`

	rows, err := qe.Query(ctx, query, len, offset)
	if err != nil {
		return nil, nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var role model.Role
		err := rows.Scan(
			&role.ID,
			&role.Code,
			&role.CreatedAt,
		)
		if err != nil {
			return nil, nil, err
		}

		roles = append(
			roles,
			&role,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	page := &api.Page{
		Page:     q.Page,
		PageSize: len,
		Total:    total,
	}

	return roles, page, nil
}

func (r *RoleRepository) Assign(
	ctx context.Context,
	qe database.QueryExecutor,
	ur *model.UserRole,
) error {
	query := `
	INSERT INTO
		user_roles(
		user_id,
	 	role_id
		)
	VALUES
	($1, $2)
	`
	_, err := qe.Exec(
		ctx,
		query,
		ur.UserID,
		ur.RoleID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) Revoke(
	ctx context.Context,
	qe database.QueryExecutor,
	ur *model.UserRole,
) error {
	query := `
		DELETE FROM
		user_roles
		WHERE
		user_id = $1
		AND
		role_id = $2
	`
	_, err := qe.Exec(
		ctx,
		query,
		ur.UserID,
		ur.RoleID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RoleRepository) HasRole(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	roleCode model.RoleCode,
) (bool, error) {
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
		  AND r.code = $2
	)
	`

	var exists bool

	err := qe.QueryRow(ctx, query, userID, roleCode).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *RoleRepository) UserRoles(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) ([]*model.Role, error) {
	query := `
	SELECT
		r.id,
		r.code,
		r.created_at
	FROM roles r
	JOIN user_roles ur
		ON ur.role_id = r.id
	WHERE ur.user_id = $1
	`

	rows, err := qe.Query(
		ctx,
		query,
		userID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var roles = make([]*model.Role, 0)
	for rows.Next() {
		var role model.Role
		err := rows.Scan(
			&role.ID,
			&role.Code,
			&role.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(
			roles,
			&role,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}
