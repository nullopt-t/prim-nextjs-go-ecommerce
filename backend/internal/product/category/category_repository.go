package category

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type Filter struct {
	ID *uuid.UUID
}

type CategoryRepository struct {
}

func NewRepository() *CategoryRepository {
	return &CategoryRepository{}
}

func (cr *CategoryRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	category *model.ProductCategory,
) error {
	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO
		categories (
			id,
			name,
			parent_id,
			created_by,
			updated_by,
			created_at,
			updated_at
		)
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		`,
		category.ID,
		category.Name,
		category.ParentID,
		category.CreatedBy,
		category.UpdatedBy,
		category.CreatedAt,
		category.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	return err
}

func (cr *CategoryRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.ProductCategory, error) {
	var category model.ProductCategory

	query :=
		`
			SELECT
				id,
				name,
				parent_id,
				created_by,
				updated_by,
				created_at,
				updated_at
			FROM
				categories
			WHERE
			deleted_at IS NULL
		`

	args := []any{}
	argID := 1
	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", argID)
		args = append(args, *filter.ID)
		argID++
	}

	err := qe.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&category.ID,
		&category.Name,
		&category.ParentID,
		&category.CreatedBy,
		&category.UpdatedBy,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}

	return &category, nil
}

type UpdateCategoryFields struct {
	Name      *string
	ParentID  *uuid.UUID
	UpdatedBy uuid.UUID
}

func (cr *CategoryRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	categoryID uuid.UUID,
	fields UpdateCategoryFields,
) error {
	query := `
	UPDATE
		categories
	SET
		name = COALESCE($1, name),
		parent_id = COALESCE($2, parent_id),
		updated_by = $3
	WHERE
		id = $4
	`
	_, err := qe.Exec(
		ctx,
		query,
		fields.Name,
		fields.ParentID,
		fields.UpdatedBy,
		categoryID,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}

	return nil
}

func (cr *CategoryRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*model.ProductCategory, *api.Page, error) {
	offset := (q.Page - 1) * q.PageSize

	query := `
	SELECT
		id,
		name,
		parent_id,
		created_by,
		updated_by,
		created_at,
		updated_at
	FROM
		categories
	WHERE
		deleted_at IS NULL
	LIMIT $1
	OFFSET $2
	`

	var total int
	err := qe.QueryRow(
		ctx,
		`
		SELECT
			COUNT(id)
		FROM
			categories
		`,
	).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("list categories: %w", err)
	}

	rows, err := qe.Query(
		ctx,
		query,
		q.PageSize,
		offset,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list categories: %w", err)
	}

	var result = make(
		[]*model.ProductCategory,
		0,
	)

	for rows.Next() {
		var c model.ProductCategory
		err = rows.Scan(
			&c.ID,
			&c.Name,
			&c.ParentID,
			&c.CreatedBy,
			&c.UpdatedBy,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("list categories: %w", err)
		}
		result = append(
			result,
			&c,
		)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("list categories: %w", err)
	}

	totalPages := (total + q.PageSize - 1) / q.PageSize
	p := &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		TotalItems:  total,
		TotalPages:  totalPages,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page < totalPages,
	}

	return result, p, nil
}
