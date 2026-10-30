package category

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type CategoryRepository struct {
}

func NewRepository() *CategoryRepository {
	return &CategoryRepository{}
}

func (cr *CategoryRepository) Create(
	ctx context.Context,
	category *model.ProductCategory,
) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}
	const query = `
		INSERT INTO product_categories (
			id,
			parent_id,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, now(), now())
		RETURNING created_at, updated_at
	`

	err = executor.QueryRow(
		ctx,
		query,
		category.ID,
		category.ParentID,
		category.Name,
	).Scan(&category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	return nil
}

func (cr *CategoryRepository) get(
	ctx context.Context,
	whereClause string,
	args ...any,
) (*model.ProductCategory, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf(`
		SELECT
			id,
			parent_id,
			name,
			(SELECT COUNT(*) FROM products p WHERE p.category_id = product_categories.id AND p.deleted_at IS NULL),
			created_at,
			updated_at
		FROM product_categories
		WHERE %s
	`, whereClause)

	category := new(model.ProductCategory)
	err = executor.QueryRow(ctx, query, args...).Scan(
		&category.ID,
		&category.ParentID,
		&category.Name,
		&category.ProductCount,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get product category: %w", err)
	}

	// Resolve the parent's display name for callers that expose the parent
	// (notably the admin category detail endpoint).
	if err := cr.attachParentNames(ctx, []*model.ProductCategory{category}); err != nil {
		return nil, err
	}
	return category, nil
}

func (cr *CategoryRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.ProductCategory, error) {
	return cr.get(ctx, "deleted_at IS NULL AND id = $1", id)
}

func (cr *CategoryRepository) GetByName(
	ctx context.Context,
	name string,
) (*model.ProductCategory, error) {
	return cr.get(ctx, "deleted_at IS NULL AND name = $1", name)
}

type UpdateCategoryFields struct {
	Name     *string
	ParentID *uuid.UUID
}

func (cr *CategoryRepository) Update(
	ctx context.Context,
	categoryID uuid.UUID,
	fields UpdateCategoryFields,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	argIdx := 1

	if fields.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *fields.Name)
		argIdx++
	}

	if fields.ParentID != nil {
		if *fields.ParentID == uuid.Nil {
			// NULL explicitly clears the parent (top-level category); binding
			// uuid.Nil would violate the foreign key constraint.
			setClauses = append(setClauses, "parent_id = NULL")
		} else {
			setClauses = append(setClauses, fmt.Sprintf("parent_id = $%d", argIdx))
			args = append(args, *fields.ParentID)
			argIdx++
		}
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

	cmdTag, err := tx.Exec(ctx, query, args...)
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
	productID uuid.UUID,
) (*model.ProductCategory, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

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
	err = executor.QueryRow(ctx, query, productID).Scan(
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
	ListQuery        *pagination.ListQuery
	IncludeDeleted   bool // set true for admin views
	OnlyWithProducts bool // set true for public customer views to omit empty categories
	// WithParentName resolves each category's parent display name. Used by admin
	// views, which show a Parent column but only store the parent UUID.
	WithParentName bool
	// WithProductCount resolves the number of active products per category. Used
	// by admin views (table "Products" column and the category tree map).
	WithProductCount bool
	// Cursor, when set, switches the query to keyset pagination on (created_at, id) DESC.
	Cursor *pagination.Keyset
}

// attachParentNames resolves the display names of the given categories' parents
// in a single query and stores them on each category. It is a no-op when no
// category has a parent, so public listings that do not need the names pay
// nothing beyond this check.
func (cr *CategoryRepository) attachParentNames(
	ctx context.Context,
	categories []*model.ProductCategory,
) error {
	executor, err := database.Executor(ctx)
	if err != nil {
		return err
	}

	parentIDs := make([]uuid.UUID, 0, len(categories))
	seen := make(map[uuid.UUID]struct{}, len(categories))
	for _, category := range categories {
		if category.ParentID == nil || *category.ParentID == uuid.Nil {
			continue
		}
		if _, ok := seen[*category.ParentID]; ok {
			continue
		}
		seen[*category.ParentID] = struct{}{}
		parentIDs = append(parentIDs, *category.ParentID)
	}

	if len(parentIDs) == 0 {
		return nil
	}

	rows, err := executor.Query(ctx, `SELECT id, name FROM product_categories WHERE id = ANY($1)`, parentIDs)
	if err != nil {
		return fmt.Errorf("resolve category parent names: %w", err)
	}
	defer rows.Close()

	names := make(map[uuid.UUID]string, len(parentIDs))
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return fmt.Errorf("scan category parent name: %w", err)
		}
		names[id] = name
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate category parent names: %w", err)
	}

	for _, category := range categories {
		if category.ParentID != nil {
			category.ParentName = names[*category.ParentID]
		}
	}
	return nil
}

func (cr *CategoryRepository) List(
	ctx context.Context,
	opts ListCategoryOptions,
) (*pagination.PagedResult[model.ProductCategory], error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	if opts.ListQuery == nil {
		panic("list categories: ListQuery is required and cannot be nil")
	}
	q := opts.ListQuery
	// Guarantee sanitized Page, PageSize, Offset, and Sort
	q.Process(pagination.QueryOptions{})

	whereClauses := []string{"1=1"}
	args := make([]any, 0, 2)
	argID := 1

	if !opts.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if opts.OnlyWithProducts {
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM products p WHERE p.category_id = product_categories.id AND p.deleted_at IS NULL AND p.status = 'published')")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argID))
		args = append(args, "%"+search+"%")
		argID++
	}

	cursorMode := opts.Cursor != nil || q.KeysetMode
	if opts.Cursor != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(created_at, id) < ($%d, $%d)", argID, argID+1))
		args = append(args, opts.Cursor.CreatedAt, opts.Cursor.ID)
		argID += 2
	}

	whereStmt := strings.Join(whereClauses, " AND ")

	// Correlated count for admin views. Zero when not requested so scans stay
	// uniform across public and admin queries.
	productCountColumn := "0"
	if opts.WithProductCount {
		productCountColumn = "(SELECT COUNT(*) FROM products p WHERE p.category_id = product_categories.id AND p.deleted_at IS NULL)"
	}

	if cursorMode {
		selectQuery := fmt.Sprintf(`
		SELECT
			id,
			parent_id,
			name,
			%s,
			created_at,
			updated_at,
			deleted_at
		FROM product_categories
		WHERE %s
		%s
		LIMIT $%d
	`, productCountColumn, whereStmt, pagination.KeysetOrderBy(""), argID)

		args = append(args, q.PageSize+1)

		rows, err := executor.Query(ctx, selectQuery, args...)
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
				&cat.ProductCount,
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

		if opts.WithParentName {
			if err := cr.attachParentNames(ctx, categories); err != nil {
				return nil, err
			}
		}

		items, page := pagination.NewCursorPageFrom(categories, q.PageSize, func(c *model.ProductCategory) pagination.Keyset {
			return pagination.Keyset{CreatedAt: c.CreatedAt, ID: c.ID}
		})
		return pagination.NewCursorPagedResult(items, page), nil
	}

	// 1. Total matching count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product_categories WHERE %s", whereStmt)
	var total int
	if err := executor.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("list product categories count: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.ProductCategory{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
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
			if sort.Order == pagination.SortDesc {
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
			%s,
			created_at,
			updated_at,
			deleted_at
		FROM product_categories
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, productCountColumn, whereStmt, orderBy, argID, argID+1)

	args = append(args, q.PageSize, q.Offset)

	rows, err := executor.Query(ctx, selectQuery, args...)
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
			&cat.ProductCount,
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

	if opts.WithParentName {
		if err := cr.attachParentNames(ctx, categories); err != nil {
			return nil, err
		}
	}

	page := pagination.NewPage(q.Page, q.PageSize, total)
	return pagination.NewPagedResult(categories, page), nil
}

func (cr *CategoryRepository) SoftDeleteCategory(ctx context.Context, id uuid.UUID) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `UPDATE product_categories SET deleted_at = now(), updated_at = now() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
