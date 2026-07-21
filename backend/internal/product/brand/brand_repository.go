package brand

import (
	"context"
	"fmt"
	"slices"
	"strings"
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
	if brand.ID == uuid.Nil {
		brand.ID = uuid.New()
	}

	now := time.Now().UTC()
	if brand.CreatedAt.IsZero() {
		brand.CreatedAt = now
	}

	if brand.UpdatedAt.IsZero() {
		brand.UpdatedAt = now
	}

	_, err := qe.Exec(
		ctx,
		`
		INSERT INTO brands(
			id,
			name,
			slug,
			status,
			logo_url,
			logo_alt,
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
			$7,
			$8
		)
		`,
		brand.ID,
		brand.Name,
		brand.Slug,
		brand.Status,
		brand.LogoURL,
		brand.LogoAlt,
		brand.CreatedAt,
		brand.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"create brand :%w",
			err,
		)
	}
	return nil
}

func (br *BrandRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.ProductBrand, error) {
	if filter.ID == nil && filter.Slug == nil {
		return nil, fmt.Errorf(
			"filter must have id or slug",
		)
	}

	query := `
	SELECT
		id,
		name,
		slug,
		status,
		logo_url,
		logo_alt,
		created_at,
		updated_at,
		archived_at,
		deleted_at
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

	brand := &model.ProductBrand{}
	err := qe.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.Status,
		&brand.LogoURL,
		&brand.LogoAlt,
		&brand.CreatedAt,
		&brand.UpdatedAt,
		&brand.ArchivedAt,
		&brand.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get brand :%w",
			err,
		)
	}
	return brand, nil
}

type UpdateBrandFields struct {
	Name    *string
	LogoURL *string
	LogoAlt *string
}

func (br *BrandRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
	fields UpdateBrandFields,
) error {
	query := `
	UPDATE
		brands
	SET
		name = COALESCE($1, name),
		logo_url = COALESCE($2, logo_url),
		logo_alt = COALESCE($3, logo_alt),
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

func (br *BrandRepository) Archive(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
) error {
	query := `
	UPDATE
		brands
	SET
		archived_at = $1,
		status = 'archived',
		updated_at = $2
	WHERE
		id = $3
	`
	args := []any{
		time.Now().UTC(),
		time.Now().UTC(),
		brandID,
	}

	cmd, err := qe.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("archive brand :%w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("archive brand :no rows affected")
	}

	return nil
}

func (br *BrandRepository) Restore(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
) error {
	query := `
	UPDATE
		brands
	SET
		status = 'active',
		archived_at = NULL,
		updated_at = $1
	WHERE
		id = $2
	`
	args := []any{
		time.Now().UTC(),
		brandID,
	}

	cmd, err := qe.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("restore brand :%w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("restore brand :no rows affected")
	}

	return nil
}

type BrandList struct {
	Brands []*model.ProductBrand
	Page   *api.Page
}

func (br *BrandRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
) (*BrandList, error) {
	var query strings.Builder
	var countQuery strings.Builder

	query.WriteString(`
		SELECT
			id,
			name,
			slug,
			status,
			logo_url,
			logo_alt,
			created_at,
			updated_at
		FROM brands
		WHERE 1=1
	`)

	countQuery.WriteString(`
		SELECT
			COUNT(id)
		FROM brands
		WHERE 1=1
	`)

	args := make([]any, 0)
	argID := 1

	// Search
	if q.Search != "" {
		condition := fmt.Sprintf(`
			AND (
				name ILIKE $%d
				OR slug ILIKE $%d
			)
		`, argID, argID)

		query.WriteString(condition)
		countQuery.WriteString(condition)

		args = append(
			args,
			"%"+q.Search+"%",
		)

		argID++
	}

	// Sorting
	if len(q.Sort) > 0 {
		query.WriteString(" ORDER BY ")

		for i, sort := range q.Sort {
			fmt.Fprintf(
				&query,
				"%s %s",
				sort.Field,
				sort.Order,
			)

			if i < len(q.Sort)-1 {
				query.WriteString(", ")
			}
		}
	} else {
		query.WriteString(`
			ORDER BY created_at DESC
		`)
	}

	// Pagination
	fmt.Fprintf(
		&query,
		`
		LIMIT $%d
		OFFSET $%d
		`,
		argID,
		argID+1,
	)

	queryArgs := append(
		slices.Clone(args),
		q.PageSize,
		(q.Page-1)*q.PageSize,
	)

	// Count
	var total int

	err := qe.QueryRow(
		ctx,
		countQuery.String(),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf(
			"list brands count: %w",
			err,
		)
	}

	rows, err := qe.Query(
		ctx,
		query.String(),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list brands: %w",
			err,
		)
	}
	defer rows.Close()

	brands := make([]*model.ProductBrand, 0)

	for rows.Next() {
		var brand model.ProductBrand

		err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Slug,
			&brand.Status,
			&brand.LogoURL,
			&brand.LogoAlt,
			&brand.CreatedAt,
			&brand.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"list brands scan: %w",
				err,
			)
		}

		brands = append(
			brands,
			&brand,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"list brands rows: %w",
			err,
		)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + q.PageSize - 1) / q.PageSize
	}

	page := &api.Page{
		Page:        q.Page,
		PageSize:    q.PageSize,
		HasPrevious: q.Page > 1,
		HasNext:     q.Page < totalPages,
		TotalItems:  total,
		TotalPages:  totalPages,
	}

	return &BrandList{
		Brands: brands,
		Page:   page,
	}, nil
}

func (br *BrandRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) error {
	if filter.ID == nil && filter.Slug == nil {
		return nil
	}

	query := `
	UPDATE
		brands
	SET
		deleted_at = $1,
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
