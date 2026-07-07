package brand

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type Filter struct {
	ID   *uuid.UUID
	Slug *string
}

type BrandRepository struct {
}

func NewRepository() *BrandRepository {
	return &BrandRepository{}
}

func (br *BrandRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	brand *model.ProductBrand,
) error {
	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO brands(
			id,
			name,
			slug,
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
		brand.Slug,
		brand.LogoURL,
		brand.LogoAlt,
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
) (*model.ProductBrand, error) {
	query := `
	SELECT
		id,
		name,
		slug,
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

	var brand model.ProductBrand
	err := qe.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.LogoURL,
		&brand.LogoAlt,
		&brand.CreatedAt,
		&brand.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get brand :%w", err)
	}
	return &brand, nil
}

type UpdateBrandFields struct {
	Name      *string
	LogoURL   *string
	LogoAlt   *string
	UpdatedBy uuid.UUID
}

func (br *BrandRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
	fields *UpdateBrandFields,
) error {
	query := `
	UPDATE
		brands
	SET
		name = COALESCE($1, name),
		logo_url = COALESCE($2, logo_url),
		logo_label = COALESCE($3, logo_label),
		updated_at = $4
	WHERE
		id = $5
	`
	args := []any{
		fields.Name,
		fields.LogoURL,
		fields.LogoAlt,
		time.Now().UTC(),
		brandID,
	}

	cmd, err := qe.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("update brand :%w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("update brand :no rows affected")
	}

	return nil
}

func (br *BrandRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) ([]*model.ProductBrand, *api.Page, error) {
	query := `
	SELECT
		id,
		name,
		slug,
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

	var brands = make([]*model.ProductBrand, 0)
	for rows.Next() {
		var brand model.ProductBrand
		err = rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Slug,
			&brand.LogoURL,
			&brand.LogoAlt,
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

func (br *BrandRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	filter *Filter,
) error {
	query := `
	DELETE FROM
		brands
	WHERE
		deleted_at IS NULL
	`
	args := []any{}
	argID := 1

	if filter.ID != nil {
		query += fmt.Sprintf(` AND id = $%d`, argID)
		args = append(args, filter.ID)
		argID++
	}

	if filter.Slug != nil {
		query += fmt.Sprintf(` AND slug = $%d`, argID)
		args = append(args, filter.Slug)
	}

	cmd, err := qe.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("delete brand: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return nil
	}

	return nil
}
