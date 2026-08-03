package product

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

var allowedProductSortFields = map[string]string{
	"id":         "id",
	"public_id":  "public_id",
	"title":      "title",
	"status":     "status",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

type Filter struct {
	ID             *uuid.UUID
	PublicID       *string
	IncludeDeleted bool
}

type ProductListItem struct {
	ID         uuid.UUID               `json:"id"`
	PublicID   string                  `json:"public_id"`
	Title      string                  `json:"title"`
	Status     model.PublicationStatus `json:"status"`
	CategoryID uuid.UUID               `json:"category_id"`
	BrandID    *uuid.UUID              `json:"brand_id,omitempty"`
	CreatedAt  time.Time               `json:"created_at"`
}

type CreateProductMediaInput struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	StorageObjectID uuid.UUID
	MediaType       model.MediaType
	SortOrder       int
}

type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

// --- Product CRUD ---

// Create inserts a new product record into the products table.
func (r *ProductRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	p *model.Product,
) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}

	query := `
		INSERT INTO products (
			id,
			brand_id,
			category_id,
			public_id,
			title,
			description,
			highlights,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := qe.Exec(
		ctx,
		query,
		p.ID,
		p.BrandID,
		p.CategoryID,
		p.PublicID,
		p.Title,
		p.Description,
		p.Highlights,
		p.Status,
		p.CreatedAt,
		p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

// Get fetches a single product matching filter criteria.
func (r *ProductRepository) Get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) (*model.Product, error) {
	if filter.ID == nil && filter.PublicID == nil {
		return nil, errors.New("get product: filter ID or PublicID is required")
	}

	whereClauses := make([]string, 0, 3)
	args := make([]any, 0, 2)
	argID := 1

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if filter.ID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("id = $%d", argID))
		args = append(args, *filter.ID)
		argID++
	}

	if filter.PublicID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("public_id = $%d", argID))
		args = append(args, *filter.PublicID)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			brand_id,
			category_id,
			public_id,
			title,
			description,
			highlights,
			status,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}

	product, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.Product])
	if err != nil {
		return nil, fmt.Errorf("scan product: %w", err)
	}

	return product, nil
}

// Update modifies core product attributes.
func (r *ProductRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	p *model.Product,
) error {
	if p.ID == uuid.Nil {
		return errors.New("update product: product ID is required")
	}

	now := time.Now().UTC().Truncate(time.Millisecond)

	query := `
		UPDATE products
		SET
			brand_id = $1,
			category_id = $2,
			title = $3,
			description = $4,
			highlights = $5,
			status = $6,
			updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(
		ctx,
		query,
		p.BrandID,
		p.CategoryID,
		p.Title,
		p.Description,
		p.Highlights,
		p.Status,
		now,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// List performs paginated searching and listing of products.
func (r *ProductRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *api.ListQuery,
	includeDeleted bool,
) (*api.PagedResult[ProductListItem], error) {
	if q == nil {
		q = &api.ListQuery{}
	}
	q.Process(api.QueryOptions{})

	whereClauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	argIdx := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(title ILIKE $%d OR public_id ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := ""
	if len(whereClauses) > 0 {
		whereStmt = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Total Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count products: %w", err)
	}

	if total == 0 {
		return api.NewPagedResult([]*ProductListItem{}, api.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Sorting
	orderBy := "ORDER BY created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedProductSortFields[strings.ToLower(sort.Field)]
			if !ok {
				continue
			}
			direction := "ASC"
			if sort.Order == api.SortDesc {
				direction = "DESC"
			}
			sortParts = append(sortParts, fmt.Sprintf("%s %s", dbField, direction))
		}
		if len(sortParts) > 0 {
			orderBy = "ORDER BY " + strings.Join(sortParts, ", ")
		}
	}

	// 3. Paginated Select
	selectQuery := fmt.Sprintf(`
		SELECT
			id,
			public_id,
			title,
			status,
			category_id,
			brand_id,
			created_at
		FROM products
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list products select: %w", err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[ProductListItem])
	if err != nil {
		return nil, fmt.Errorf("collect products: %w", err)
	}

	return api.NewPagedResult(items, api.NewPage(q.Page, q.PageSize, total)), nil
}

// SoftDelete soft-deletes a product by ID or public_id.
func (r *ProductRepository) SoftDelete(
	ctx context.Context,
	qe database.QueryExecutor,
	filter Filter,
) error {
	if filter.ID == nil && filter.PublicID == nil {
		return errors.New("soft delete product: filter ID or PublicID is required")
	}

	whereClauses := []string{"deleted_at IS NULL"}
	args := make([]any, 0, 2)
	argID := 1

	if filter.ID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("id = $%d", argID))
		args = append(args, *filter.ID)
		argID++
	}

	if filter.PublicID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("public_id = $%d", argID))
		args = append(args, *filter.PublicID)
		argID++
	}

	now := time.Now().UTC().Truncate(time.Millisecond)
	query := fmt.Sprintf(`
		UPDATE products
		SET deleted_at = $%d, updated_at = $%d
		WHERE %s
	`, argID, argID, strings.Join(whereClauses, " AND "))

	args = append(args, now)

	cmd, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("soft delete product: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateStatus changes the publication status of a product.
func (r *ProductRepository) UpdateStatus(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
	status model.PublicationStatus,
) error {
	if productID == uuid.Nil {
		return errors.New("update product status: productID is required")
	}

	now := time.Now().UTC().Truncate(time.Millisecond)

	query := `
		UPDATE products
		SET status = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, productID, status, now)
	if err != nil {
		return fmt.Errorf("update product status: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// --- Product Media Operations ---

// AddMedia links a storage object to a product.
func (r *ProductRepository) AddMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	in CreateProductMediaInput,
) (*model.ProductMedia, error) {
	if in.ID == uuid.Nil {
		in.ID = uuid.New()
	}

	query := `
		INSERT INTO product_media (
			id,
			product_id,
			storage_object_id,
			media_type,
			sort_order
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := qe.Exec(
		ctx,
		query,
		in.ID,
		in.ProductID,
		in.StorageObjectID,
		in.MediaType,
		in.SortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("add product media: %w", err)
	}

	return &model.ProductMedia{
		ID:              in.ID,
		ProductID:       in.ProductID,
		StorageObjectID: in.StorageObjectID,
		MediaType:       in.MediaType,
		SortOrder:       in.SortOrder,
	}, nil
}

// ListMediaByProductID fetches all media assigned to a product joined with storage_objects.
func (r *ProductRepository) ListMediaByProductID(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) ([]*model.ProductMedia, error) {
	if productID == uuid.Nil {
		return nil, errors.New("list product media: productID is required")
	}

	query := `
		SELECT
			m.id,
			m.product_id,
			m.storage_object_id,
			m.media_type,
			m.sort_order,
			so.id,
			so.bucket,
			so.object_key,
			so.content_type,
			so.file_size,
			so.status
		FROM product_media m
		JOIN storage_objects so ON m.storage_object_id = so.id
		WHERE m.product_id = $1
		ORDER BY m.sort_order ASC
	`

	rows, err := qe.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("list product media query: %w", err)
	}
	defer rows.Close()

	var result []*model.ProductMedia
	for rows.Next() {
		m := &model.ProductMedia{
			Object: &model.Object{},
		}
		if err := rows.Scan(
			&m.ID,
			&m.ProductID,
			&m.StorageObjectID,
			&m.MediaType,
			&m.SortOrder,
			&m.Object.ID,
			&m.Object.Bucket,
			&m.Object.Key,
			&m.Object.ContentType,
			&m.Object.FileSize,
			&m.Object.Status,
		); err != nil {
			return nil, fmt.Errorf("list product media scan: %w", err)
		}
		result = append(result, m)
	}

	return result, nil
}

// RemoveMedia unlinks a specific media item from a product.
func (r *ProductRepository) RemoveMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
	mediaID uuid.UUID,
) error {
	if productID == uuid.Nil || mediaID == uuid.Nil {
		return errors.New("remove product media: productID and mediaID are required")
	}

	query := `
		DELETE FROM product_media
		WHERE id = $1 AND product_id = $2
	`

	cmd, err := qe.Exec(ctx, query, mediaID, productID)
	if err != nil {
		return fmt.Errorf("remove product media: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ReorderMedia updates media sort orders for a product in batch.
func (r *ProductRepository) ReorderMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if productID == uuid.Nil || len(orderedMediaIDs) == 0 {
		return nil
	}

	query := `
		UPDATE product_media
		SET sort_order = $1
		WHERE id = $2 AND product_id = $3
	`

	for idx, mediaID := range orderedMediaIDs {
		if _, err := qe.Exec(ctx, query, idx, mediaID, productID); err != nil {
			return fmt.Errorf("reorder product media item %s: %w", mediaID, err)
		}
	}

	return nil
}
