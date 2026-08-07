package brand

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

var allowedBrandSortFields = map[string]string{
	"id":         "id",
	"name":       "name",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

type UpdateBrandFields struct {
	Name                *string
	Link                *string
	LogoStorageObjectID *uuid.UUID
}

type ListBrandOptions struct {
	Query          *pagination.ListQuery
	IncludeDeleted bool
}

type BrandRepository struct{}

func NewRepository() *BrandRepository {
	return &BrandRepository{}
}

// Create inserts a new product brand matching the full table schema.
func (br *BrandRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	brand *model.ProductBrand,
) error {
	if brand.ID == uuid.Nil {
		brand.ID = uuid.New()
	}

	if brand.PublicID == "" {
		brand.PublicID = uuid.NewString()
	}

	query := `
		INSERT INTO product_brands (
			id,
			public_id,
			name,
			link,
			logo_storage_object_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		brand.ID,
		brand.PublicID,
		brand.Name,
		brand.Link,
		brand.LogoStorageObjectID,
	).Scan(&brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create brand: %w", err)
	}

	return nil
}

// Get fetches an active brand by ID.
func (br *BrandRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.ProductBrand, error) {
	query := `
		SELECT
			id,
			public_id,
			name,
			link,
			logo_storage_object_id,
			created_at,
			updated_at,
			deleted_at
		FROM product_brands
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := qe.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get brand query: %w", err)
	}

	brand, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.ProductBrand])
	if err != nil {
		return nil, fmt.Errorf("get brand scan: %w", err)
	}

	return brand, nil
}

// Update dynamically updates present fields on active records.
func (br *BrandRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
	fields UpdateBrandFields,
) error {
	if brandID == uuid.Nil {
		return errors.New("update brand: brandID is required")
	}

	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	argIdx := 1

	if fields.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *fields.Name)
		argIdx++
	}

	if fields.Link != nil {
		setClauses = append(setClauses, fmt.Sprintf("link = $%d", argIdx))
		args = append(args, fields.Link)
		argIdx++
	}

	if fields.LogoStorageObjectID != nil {
		setClauses = append(setClauses, fmt.Sprintf("logo_storage_object_id = $%d", argIdx))
		args = append(args, fields.LogoStorageObjectID)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // No updates supplied
	}

	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE product_brands
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "), argIdx)

	args = append(args, brandID)

	cmd, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update brand: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// List unified query method for public and admin listings.
func (br *BrandRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	opts ListBrandOptions,
) (*pagination.PagedResult[model.ProductBrand], error) {
	q := opts.Query
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := []string{"1=1"}
	args := make([]any, 0, 2)
	argIdx := 1

	if !opts.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := strings.Join(whereClauses, " AND ")

	// 1. Total Count Query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product_brands WHERE %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("list brands count: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.ProductBrand{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Whitelist-guarded ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedBrandSortFields[strings.ToLower(sort.Field)]
			if !ok {
				continue
			}
			direction := "ASC"
			if sort.Order == pagination.SortDesc {
				direction = "DESC"
			}
			sortParts = append(sortParts, fmt.Sprintf("%s %s", dbField, direction))
		}
		if len(sortParts) > 0 {
			orderBy = "ORDER BY " + strings.Join(sortParts, ", ")
		}
	}

	// 3. Paginated Select Query
	selectQuery := fmt.Sprintf(`
		SELECT
			id,
			public_id,
			name,
			link,
			logo_storage_object_id,
			created_at,
			updated_at,
			deleted_at
		FROM product_brands
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list brands select: %w", err)
	}

	brands, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.ProductBrand])
	if err != nil {
		return nil, fmt.Errorf("list brands collect rows: %w", err)
	}

	return pagination.NewPagedResult(brands, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// Delete performs a soft-delete on an active brand.
func (br *BrandRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
) error {
	if brandID == uuid.Nil {
		return errors.New("delete brand: brandID is required")
	}

	query := `
		UPDATE product_brands
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, brandID)
	if err != nil {
		return fmt.Errorf("delete brand: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
