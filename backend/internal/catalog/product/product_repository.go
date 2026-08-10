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
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

var allowedProductSortFields = map[string]string{
	"id":         "p.id",
	"public_id":  "p.public_id",
	"title":      "p.title",
	"status":     "p.status",
	"created_at": "p.created_at",
	"updated_at": "p.updated_at",
}

type productFilter struct {
	ID             *uuid.UUID
	PublicID       *string
	IncludeDeleted bool
}

type PublicProductBrandReadModel struct {
	ID       uuid.UUID
	PublicID string
	Name     string
	Link     *string
}

type PublicProductCategoryReadModel struct {
	ID       uuid.UUID
	PublicID string
	Name     string
}

type PublicProductListReadModel struct {
	ID          uuid.UUID
	PublicID    string
	Title       string
	Description string
	Status      model.PublicationStatus
	Brand          *PublicProductBrandReadModel
	Category       *PublicProductCategoryReadModel
	ProductType    model.ProductType
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type CreateProductMediaInput struct {
	ID              uuid.UUID
	PublicID        string
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
	query := `
		INSERT INTO products (
			id,
			brand_id,
			category_id,
			public_id,
			title,
			description,
			status,
			product_type,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		p.ID,
		p.BrandID,
		p.CategoryID,
		p.PublicID,
		p.Title,
		p.Description,
		p.Status,
		p.ProductType,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

func (r *ProductRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, qe, productFilter{ID: &id})
}

func (r *ProductRepository) GetByIDWithDeleted(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, qe, productFilter{ID: &id, IncludeDeleted: true})
}

func (r *ProductRepository) GetByPublicID(
	ctx context.Context,
	qe database.QueryExecutor,
	publicID string,
) (*model.Product, error) {
	return r.get(ctx, qe, productFilter{PublicID: &publicID})
}

func (r *ProductRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter productFilter,
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
			product_type,
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

	query := `
		UPDATE products
		SET
			brand_id = $1,
			category_id = $2,
			title = $3,
			description = $4,
			highlights = $5,
			status = $6,
			product_type = $7,
			updated_at = now()
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
		p.ProductType,
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

// List performs paginated searching and listing of products for the public storefront.
func (r *ProductRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[PublicProductListReadModel], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	argIdx := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "p.deleted_at IS NULL")
		whereClauses = append(whereClauses, "p.status = 'published'")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.title ILIKE $%d OR p.public_id ILIKE $%d OR b.name ILIKE $%d OR c.name ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := ""
	if len(whereClauses) > 0 {
		whereStmt = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Total Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p LEFT JOIN product_brands b ON p.brand_id = b.id LEFT JOIN product_categories c ON p.category_id = c.id %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count products: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*PublicProductListReadModel{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Sorting
	orderBy := "ORDER BY p.created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedProductSortFields[strings.ToLower(sort.Field)]
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

	// 3. Paginated Select
	selectQuery := fmt.Sprintf(`
		SELECT
			p.id,
			p.public_id,
			p.title,
			p.description,
			p.status,
			p.product_type,
			b.id as brand_id,
			b.public_id as brand_public_id,
			b.name as brand_name,
			b.link as brand_link,
			c.id as category_id,
			c.public_id as category_public_id,
			c.name as category_name,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		LEFT JOIN product_brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		LEFT JOIN product_categories c ON p.category_id = c.id AND c.deleted_at IS NULL
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list products select: %w", err)
	}
	defer rows.Close()

	items := make([]*PublicProductListReadModel, 0, q.PageSize)
	for rows.Next() {
		var (
			item                                PublicProductListReadModel
			brandID                             *uuid.UUID
			brandPublicID, brandName, brandLink *string
			catID                               *uuid.UUID
			catPublicID, catName                *string
		)
		err := rows.Scan(
			&item.ID,
			&item.PublicID,
			&item.Title,
			&item.Description,
			&item.Status,
			&item.ProductType,
			&brandID,
			&brandPublicID,
			&brandName,
			&brandLink,
			&catID,
			&catPublicID,
			&catName,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product list item: %w", err)
		}

		if brandID != nil && brandPublicID != nil && brandName != nil {
			item.Brand = &PublicProductBrandReadModel{
				ID:       *brandID,
				PublicID: *brandPublicID,
				Name:     *brandName,
				Link:     brandLink,
			}
		}

		if catID != nil && catPublicID != nil && catName != nil {
			item.Category = &PublicProductCategoryReadModel{
				ID:       *catID,
				PublicID: *catPublicID,
				Name:     *catName,
			}
		}

		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product list items: %w", err)
	}

	return pagination.NewPagedResult(items, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// AdminList performs paginated searching and listing of products for administrative management without joining brand/category details.
func (r *ProductRepository) AdminList(
	ctx context.Context,
	qe database.QueryExecutor,
	q *pagination.ListQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.Product], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	argIdx := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "p.deleted_at IS NULL")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.title ILIKE $%d OR p.public_id ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := ""
	if len(whereClauses) > 0 {
		whereStmt = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Total Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admin products: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.Product{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// 2. Sorting
	orderBy := "ORDER BY p.created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedProductSortFields[strings.ToLower(sort.Field)]
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

	// 3. Paginated Select
	selectQuery := fmt.Sprintf(`
		SELECT
			p.id,
			p.brand_id,
			p.category_id,
			p.public_id,
			p.title,
			p.description,
			p.highlights,
			p.status,
			p.product_type,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list admin products select: %w", err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.Product])
	if err != nil {
		return nil, fmt.Errorf("collect admin products: %w", err)
	}

	return pagination.NewPagedResult(items, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// SoftDeleteByID soft-deletes a product by ID.
func (r *ProductRepository) SoftDeleteByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) error {
	return r.softDelete(ctx, qe, productFilter{ID: &id})
}

func (r *ProductRepository) softDelete(
	ctx context.Context,
	qe database.QueryExecutor,
	filter productFilter,
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
	}

	query := fmt.Sprintf(`
		UPDATE products
		SET deleted_at = now(), updated_at = now()
		WHERE %s
	`, strings.Join(whereClauses, " AND "))

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

	query := `
		UPDATE products
		SET status = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, productID, status)
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
	query := `
		INSERT INTO product_media (
			id,
			public_id,
			product_id,
			storage_object_id,
			media_type,
			sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := qe.Exec(
		ctx,
		query,
		in.ID,
		in.PublicID,
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
		PublicID:        in.PublicID,
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
			m.public_id,
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
			&m.PublicID,
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

// GetMaxMediaSortOrder returns the current highest sort_order for a product's media.
// Returns -1 if no media exists (so the first item gets sort_order 0).
func (r *ProductRepository) GetMaxMediaSortOrder(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) (int, error) {
	var max int
	err := qe.QueryRow(ctx,
		`SELECT COALESCE(MAX(sort_order), -1) FROM product_media WHERE product_id = $1`,
		productID,
	).Scan(&max)
	if err != nil {
		return 0, fmt.Errorf("get max media sort_order: %w", err)
	}
	return max, nil
}
