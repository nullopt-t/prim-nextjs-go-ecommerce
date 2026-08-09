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
	Name         *string
	Link         *string
	LogoObjectID *uuid.UUID
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
	query := `
		INSERT INTO product_brands (
			id,
			public_id,
			name,
			link,
			logo_object_id,
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
		brand.LogoObjectID,
	).Scan(&brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create brand: %w", err)
	}

	return nil
}

// get is a private helper to fetch a brand by a custom where clause.
func (br *BrandRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	whereClause string,
	args ...any,
) (*model.ProductBrand, error) {
	query := fmt.Sprintf(`
		SELECT
			id,
			public_id,
			name,
			link,
			logo_object_id,
			created_at,
			updated_at,
			deleted_at
		FROM product_brands
		WHERE %s AND deleted_at IS NULL
	`, whereClause)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get brand query: %w", err)
	}

	brand, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.ProductBrand])
	if err != nil {
		return nil, fmt.Errorf("get brand scan: %w", err)
	}

	return brand, nil
}

// GetByID fetches an active brand by ID.
func (br *BrandRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.ProductBrand, error) {
	return br.get(ctx, qe, "id = $1", id)
}

// GetByName fetches an active brand by Name.
func (br *BrandRepository) GetByName(
	ctx context.Context,
	qe database.QueryExecutor,
	name string,
) (*model.ProductBrand, error) {
	return br.get(ctx, qe, "name = $1", name)
}

// Update dynamically updates present fields on active records.
func (br *BrandRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
	fields UpdateBrandFields,
) error {
	if brandID == uuid.Nil {
		panic("update brand: brandID is required")
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

	if fields.LogoObjectID != nil {
		setClauses = append(setClauses, fmt.Sprintf("logo_object_id = $%d", argIdx))
		args = append(args, fields.LogoObjectID)
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
	if opts.Query == nil {
		panic("list brands: opts.Query is required")
	}
	q := opts.Query
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
			logo_object_id,
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

// delete is a private helper to soft-delete by a custom where clause.
func (br *BrandRepository) delete(
	ctx context.Context,
	qe database.QueryExecutor,
	whereClause string,
	args ...any,
) error {
	// First find the ID so we can check if it's used
	var brandID uuid.UUID
	selectQuery := fmt.Sprintf("SELECT id FROM product_brands WHERE %s AND deleted_at IS NULL", whereClause)
	if err := qe.QueryRow(ctx, selectQuery, args...).Scan(&brandID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("delete brand select: %w", err)
	}

	// Check if any products are still using this brand
	var used bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM products WHERE brand_id = $1 AND deleted_at IS NULL)"
	if err := qe.QueryRow(ctx, checkQuery, brandID).Scan(&used); err != nil {
		return fmt.Errorf("delete brand check usage: %w", err)
	}
	if used {
		return database.ErrConflict
	}

	query := fmt.Sprintf(`
		UPDATE product_brands
		SET deleted_at = now(), updated_at = now()
		WHERE %s AND deleted_at IS NULL
	`, whereClause)

	cmd, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete brand: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByID performs a soft-delete on an active brand by ID.
func (br *BrandRepository) DeleteByID(
	ctx context.Context,
	qe database.QueryExecutor,
	brandID uuid.UUID,
) error {
	if brandID == uuid.Nil {
		panic("delete brand: brandID is required")
	}
	return br.delete(ctx, qe, "id = $1", brandID)
}

// DeleteByName performs a soft-delete on an active brand by name.
func (br *BrandRepository) DeleteByName(
	ctx context.Context,
	qe database.QueryExecutor,
	name string,
) error {
	if name == "" {
		panic("delete brand: name is required")
	}
	return br.delete(ctx, qe, "name = $1", name)
}

// Delete performs a soft-delete on an active brand given the model instance.
func (br *BrandRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	brand *model.ProductBrand,
) error {
	if brand == nil || brand.ID == uuid.Nil {
		panic("delete brand: valid brand instance is required")
	}
	return br.delete(ctx, qe, "id = $1", brand.ID)
}
