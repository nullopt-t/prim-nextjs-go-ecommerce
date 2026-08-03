package category

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type Filter struct {
	ID   *uuid.UUID
	Name *string
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
	const query = `
		INSERT INTO product_categories (
			id,
			parent_id,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := qe.Exec(
		ctx,
		query,
		category.ID,
		category.ParentID,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	return nil
}
func (cr *CategoryRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.ProductCategory, error) {
	if filter.ID == nil && filter.Name == nil {
		return nil, errors.New("get category: at least one filter parameter is required")
	}

	query := `
		SELECT
			id,
			parent_id,
			name,
			created_at,
			updated_at
		FROM
			product_categories
		WHERE
			deleted_at IS NULL
	`

	args := make([]any, 0, 2)
	argCount := 1

	if filter.ID != nil {
		query += fmt.Sprintf(" AND id = $%d", argCount)
		args = append(args, *filter.ID)
		argCount++
	}

	if filter.Name != nil {
		query += fmt.Sprintf(" AND name = $%d", argCount)
		args = append(args, *filter.Name)
		argCount++
	}

	category := new(model.ProductCategory)
	err := qe.QueryRow(ctx, query, args...).Scan(
		&category.ID,
		&category.ParentID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get product category: %w", err)
	}

	return category, nil
}

func (cr *CategoryRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.ProductCategory, error) {
	return cr.get(ctx, qe, Filter{ID: &id})
}

func (cr *CategoryRepository) GetByName(
	ctx context.Context,
	qe database.QueryExecutor,
	name string,
) (*model.ProductCategory, error) {
	return cr.get(ctx, qe, Filter{Name: &name})
}

type UpdateCategoryFields struct {
	Name     *string
	ParentID *uuid.UUID
}

func (cr *CategoryRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	categoryID uuid.UUID,
	fields UpdateCategoryFields,
) error {
	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	argIdx := 1

	if fields.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *fields.Name)
		argIdx++
	}

	if fields.ParentID != nil {
		setClauses = append(setClauses, fmt.Sprintf("parent_id = $%d", argIdx))
		args = append(args, fields.ParentID)
		argIdx++
	}

	// No fields provided to update; return early to save a DB roundtrip
	if len(setClauses) == 0 {
		return nil
	}

	// Always touch updated_at timestamp
	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE product_categories
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "), argIdx)

	args = append(args, categoryID)

	cmdTag, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update product category: %w", err)
	}

	// If no rows were updated (e.g., ID doesn't exist or is soft-deleted),
	// return pgx.ErrNoRows so the service layer can map it to ErrNotFound.
	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (cr *CategoryRepository) GetProductCategory(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) (*model.ProductCategory, error) {
	query := `
		SELECT
			c.id,
			c.parent_id,
			c.name,
			c.created_at,
			c.updated_at
		FROM products p
		JOIN product_categories c ON p.category_id = c.id
		WHERE p.id = $1
		  AND p.deleted_at IS NULL
		  AND c.deleted_at IS NULL
	`

	category := new(model.ProductCategory)
	err := qe.QueryRow(ctx, query, productID).Scan(
		&category.ID,
		&category.ParentID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get product category: %w", err)
	}

	return category, nil
}

var allowedCategorySortFields = map[string]string{
	"id":         "id",
	"name":       "name",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

type ListCategoryOptions struct {
	ListQuery      *api.ListQuery
	IncludeDeleted bool // set true for admin views
}

func (cr *CategoryRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	opts ListCategoryOptions,
) (*api.PagedResult[model.ProductCategory], error) {
	q := opts.ListQuery
	if q == nil {
		q = &api.ListQuery{}
	}
	// Guarantee sanitized Page, PageSize, Offset, and Sort
	q.Process(api.QueryOptions{})

	whereClauses := []string{"1=1"}
	args := make([]any, 0, 2)
	argID := 1

	if !opts.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argID))
		args = append(args, "%"+search+"%")
		argID++
	}

	whereStmt := strings.Join(whereClauses, " AND ")

	// 1. Total matching count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product_categories WHERE %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("list product categories count: %w", err)
	}

	if total == 0 {
		return api.NewPagedResult([]*model.ProductCategory{}, api.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Build whitelist-guarded ORDER BY clause using q.Sort
	orderBy := "ORDER BY created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedCategorySortFields[strings.ToLower(sort.Field)]
			if !ok {
				continue
			}
			dir := "ASC"
			if sort.Order == api.SortDesc {
				dir = "DESC"
			}
			sortParts = append(sortParts, fmt.Sprintf("%s %s", dbField, dir))
		}
		if len(sortParts) > 0 {
			orderBy = "ORDER BY " + strings.Join(sortParts, ", ")
		}
	}

	// 3. Fetch Paginated Records
	selectQuery := fmt.Sprintf(`
		SELECT
			id,
			parent_id,
			name,
			created_at,
			updated_at,
			deleted_at
		FROM product_categories
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argID, argID+1)

	queryArgs := append(args, q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list product categories query: %w", err)
	}
	defer rows.Close()

	categories := make([]*model.ProductCategory, 0, q.PageSize)
	for rows.Next() {
		var cat model.ProductCategory
		if err := rows.Scan(
			&cat.ID,
			&cat.ParentID,
			&cat.Name,
			&cat.CreatedAt,
			&cat.UpdatedAt,
			&cat.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product category row: %w", err)
		}
		categories = append(categories, &cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product category rows: %w", err)
	}

	page := api.NewPage(q.Page, q.PageSize, total)
	return api.NewPagedResult(categories, page), nil
}
