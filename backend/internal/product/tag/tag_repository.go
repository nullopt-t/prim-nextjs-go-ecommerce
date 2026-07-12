package tag

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type TagRepository struct {
}

func NewRepository() *TagRepository {
	return &TagRepository{}
}

func (tr *TagRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	tag *model.ProductTag,
) error {
	query := `
	INSERT INTO tags(
		id,
		name,
		created_at,
		updated_at
	)
	VALUES (
		$1,
		$2,
		$3,
		$4
	)
	`
	_, err := qe.Exec(
		ctx,
		query,
		tag.ID,
		tag.Name,
		tag.CreatedAt,
		tag.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create tag:%w", err)
	}

	return nil
}

type Filter struct {
	ID   *uuid.UUID
	Name *string
}

func (tr *TagRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter *Filter,
) (*model.ProductTag, error) {
	query := `
	SELECT
		id,
		name,
		created_at,
		updated_at
	FROM
		tags
	WHERE deleted_at IS NULL
	`

	args := []any{}
	argID := 1
	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", argID)
		args = append(args, *filter.ID)
		argID++
	}

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name = $%d", argID)
		args = append(args, *filter.Name)
	}

	var tag model.ProductTag
	err := qe.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get a product by id: %w", err)
	}

	return &tag, nil
}

func (tr *TagRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*model.ProductTag, *api.Page, error) {
	query := `
	SELECT
		id,
		name,
		created_at,
		updated_at
	FROM
		tags
	WHERE
		deleted_at IS NULL
	LIMIT $1
	OFFSET $2
	`

	offset := (q.Page - 1) * q.PageSize
	rows, err := qe.Query(
		ctx,
		query,
		q.PageSize,
		offset,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list tags:%w", err)
	}

	tags := make([]*model.ProductTag, 0)
	for rows.Next() {
		var tag model.ProductTag
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.CreatedAt,
			&tag.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("list tags:%w", err)
		}

		tags = append(
			tags,
			&tag,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("list tags:%w", err)
	}

	var total int
	err = qe.QueryRow(
		ctx,
		`
		SELECT
			COUNT(id)
		FROM
			tags
		`).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("list tags:%w", err)
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
	return tags, p, nil
}
