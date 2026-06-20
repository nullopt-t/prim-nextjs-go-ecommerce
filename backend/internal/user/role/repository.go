package role

import (
	"context"

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
