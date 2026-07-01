package brand

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

type BrandRepository struct {
}

func NewRepository() *BrandRepository {
	return &BrandRepository{}
}

func (br *BrandRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	brand *model.Brand,
) error {
	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO brands(
			id,
			name,
			logo_url,
			logo_label,
			created_at,
			updated_at
		) 
		VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		)
		`,
		brand.ID,
		brand.Name,
		brand.LogoURL,
		brand.LogoLabel,
		brand.CreatedAt,
		brand.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create brand :%w", err)
	}
	return nil
}

func (br *BrandRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter *Filter,
) (*model.Brand, error) {
	query := `
	SELECT
		id,
		name,
		logo_url,
		logo_label,
		created_at,
		updated_at
	FROM
		brands
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

	var brand model.Brand
	err := qe.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&brand.ID,
		&brand.Name,
		&brand.LogoURL,
		&brand.LogoLabel,
		&brand.CreatedAt,
		&brand.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get brand :%w", err)
	}
	return &brand, nil
}

func (br *BrandRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.PageQuery,
) ([]*model.Brand, *api.Page, error) {
	query := `
	SELECT 
		id,
		name,
		logo_url,
		logo_label,
		created_at,
		updated_at
	FROM
		brands
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
			brands
		`,
	).Scan(&total)
	if err != nil {
		return nil, nil, fmt.Errorf("list brands :%w", err)
	}

	offset := (q.Page - 1) * q.PageSize
	rows, err := qe.Query(
		ctx,
		query,
		q.PageSize,
		offset,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("list brands:%w", err)
	}

	var brands = make([]*model.Brand, 0)
	for rows.Next() {
		var brand model.Brand
		err = rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.LogoURL,
			&brand.LogoLabel,
			&brand.CreatedAt,
			&brand.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("list brands: %w", err)
		}

		brands = append(
			brands,
			&brand,
		)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("list brands: %w", err)
	}

	totalPages := (total + q.PageSize - 1) / q.PageSize
	page := &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page < totalPages,
		TotalItems:  total,
		TotalPages:  totalPages,
	}

	return brands, page, nil
}
