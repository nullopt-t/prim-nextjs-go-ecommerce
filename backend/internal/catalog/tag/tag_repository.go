package tag

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

var allowedTagSortFields = map[string]string{
	"id":         "id",
	"name":       "name",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

type tagFilter struct {
	ID             *uuid.UUID
	Name           *string
	IncludeDeleted bool
}

type UpdateTagFields struct {
	Name *string
}

type ListTagOptions struct {
	Query          *pagination.ListQuery
	IncludeDeleted bool
}

type TagRepository struct{}

func NewRepository() *TagRepository {
	return &TagRepository{}
}

// Create inserts a new product tag.
func (tr *TagRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	tag *model.ProductTag,
) error {
	query := `
		INSERT INTO product_tags (
			id,
			name,
			created_at,
			updated_at
		)
		VALUES ($1, $2, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		tag.ID,
		tag.Name,
	).Scan(&tag.CreatedAt, &tag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create tag: %w", err)
	}

	return nil
}

// GetByID fetches a single tag by ID.
func (tr *TagRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.ProductTag, error) {
	return tr.get(ctx, qe, &tagFilter{ID: &id})
}

// GetByName fetches a single tag by Name.
func (tr *TagRepository) GetByName(
	ctx context.Context,
	qe database.QueryExecutor,
	name string,
) (*model.ProductTag, error) {
	return tr.get(ctx, qe, &tagFilter{Name: &name})
}

func (tr *TagRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter *tagFilter,
) (*model.ProductTag, error) {
	if filter == nil || (filter.ID == nil && filter.Name == nil) {
		return nil, errors.New("get tag: filter ID or Name is required")
	}

	whereClauses := make([]string, 0, 3)
	args := make([]any, 0, 2)
	argIdx := 1

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if filter.ID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("id = $%d", argIdx))
		args = append(args, *filter.ID)
		argIdx++
	}

	if filter.Name != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *filter.Name)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			created_at,
			updated_at,
			deleted_at
		FROM product_tags
		WHERE %s
	`, strings.Join(whereClauses, " AND "))

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get tag query: %w", err)
	}

	tag, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.ProductTag])
	if err != nil {
		return nil, fmt.Errorf("get tag scan: %w", err)
	}

	return tag, nil
}

// Update dynamically modifies an active tag.
func (tr *TagRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	tagID uuid.UUID,
	fields UpdateTagFields,
) error {
	if tagID == uuid.Nil {
		return errors.New("update tag: tagID is required")
	}

	if fields.Name == nil {
		return nil // Nothing to update
	}

	query := `
		UPDATE product_tags
		SET name = $1, updated_at = now()
		WHERE id = $2 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, *fields.Name, tagID)
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// List handles unified paginated tag retrieval for public and admin contexts.
func (tr *TagRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	opts ListTagOptions,
) (*pagination.PagedResult[model.ProductTag], error) {
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

	// 1. Total Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product_tags WHERE %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("list tags count: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.ProductTag{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Whitelisted ORDER BY
	orderBy := "ORDER BY created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedTagSortFields[strings.ToLower(sort.Field)]
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
			name,
			created_at,
			updated_at,
			deleted_at
		FROM product_tags
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list tags select: %w", err)
	}

	tags, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.ProductTag])
	if err != nil {
		return nil, fmt.Errorf("list tags collect rows: %w", err)
	}

	return pagination.NewPagedResult(tags, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// Delete performs a soft-delete on an active tag by ID.
func (tr *TagRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	tagID uuid.UUID,
) error {
	if tagID == uuid.Nil {
		return errors.New("delete tag: tagID is required")
	}

	query := `
		UPDATE product_tags
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, tagID)
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ReplaceTagsForProduct replaces all tag assignments for a product with the given tag IDs.
// Pass an empty slice to remove all tags.
func (tr *TagRepository) ReplaceTagsForProduct(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
	tagIDs []uuid.UUID,
) error {
	if productID == uuid.Nil {
		return errors.New("replace product tags: productID is required")
	}

	// Delete all existing assignments
	_, err := qe.Exec(ctx,
		`DELETE FROM product_tag_assignments WHERE product_id = $1`,
		productID,
	)
	if err != nil {
		return fmt.Errorf("replace product tags: delete existing: %w", err)
	}

	if len(tagIDs) == 0 {
		return nil
	}

	// Bulk insert new assignments
	args := make([]any, 0, len(tagIDs)*2)
	valueParts := make([]string, 0, len(tagIDs))
	for i, tagID := range tagIDs {
		base := i * 2
		valueParts = append(valueParts, fmt.Sprintf("($%d, $%d, now())", base+1, base+2))
		args = append(args, productID, tagID)
	}

	insertQuery := fmt.Sprintf(
		`INSERT INTO product_tag_assignments (product_id, tag_id, created_at) VALUES %s ON CONFLICT DO NOTHING`,
		strings.Join(valueParts, ", "),
	)

	if _, err := qe.Exec(ctx, insertQuery, args...); err != nil {
		return fmt.Errorf("replace product tags: insert: %w", err)
	}

	return nil
}

// GetTagsByProductID fetches all active tags assigned to a product.
func (tr *TagRepository) GetTagsByProductID(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) ([]*model.ProductTag, error) {
	if productID == uuid.Nil {
		return nil, errors.New("get product tags: productID is required")
	}

	query := `
		SELECT t.id, t.name, t.created_at, t.updated_at, t.deleted_at
		FROM product_tags t
		INNER JOIN product_tag_assignments pta ON t.id = pta.tag_id
		WHERE pta.product_id = $1 AND t.deleted_at IS NULL
		ORDER BY t.name ASC
	`

	rows, err := qe.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("get product tags query: %w", err)
	}

	tags, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.ProductTag])
	if err != nil {
		return nil, fmt.Errorf("get product tags collect: %w", err)
	}

	return tags, nil
}
