// Package product provides database repository operations for product catalog persistence,
// rich search/filtering across brands, categories, and tags, and publication status updates.
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
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// allowedProductSortFields maps user-facing sort parameters to safe SQL column names.
var allowedProductSortFields = map[string]string{
	"id":         "p.id",
	"slug":       "p.slug",
	"title":      "p.title",
	"status":     "p.status",
	"created_at": "p.created_at",
	"updated_at": "p.updated_at",
}

// PublicProductBrandReadModel represents brand summary information projected onto a product card.
type PublicProductBrandReadModel struct {
	ID   uuid.UUID
	Name string
	Link *string
}

// PublicProductCategoryReadModel represents category summary information projected onto a product card.
type PublicProductCategoryReadModel struct {
	ID   uuid.UUID
	Name string
}

// ProductCardReadModel represents a flattened, performant denormalized view of a product for storefront listings,
// including default variant pricing, aggregated inventory, and taxonomy tags.
type ProductCardReadModel struct {
	ID              uuid.UUID
	Slug            string
	Title           string
	Description     *string
	Status          model.PublicationStatus
	Brand           *PublicProductBrandReadModel
	Category        *PublicProductCategoryReadModel
	ProductType     model.ProductType
	Thumbnail       *model.Object
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	SKU             *string
	DefaultVariantID *uuid.UUID
	StockQuantity   int
	Rating          *model.RatingSummary
	TagsRaw         []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

// CreateProductMediaInput defines parameters to link a media object to a product.
type CreateProductMediaInput struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	ObjectID  uuid.UUID
	MediaType model.MediaType
	SortOrder int
}

// ProductRepository provides database access for products and related denormalized queries.
type ProductRepository struct{}

// NewProductRepository creates and returns a new ProductRepository instance.
func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

// Create inserts a new product record in draft state.
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
			slug,
			title,
			status,
			product_type,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		p.ID,
		p.BrandID,
		p.CategoryID,
		p.Slug,
		p.Title,
		p.Status,
		p.ProductType,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

// GetByID retrieves a single non-deleted product by its UUID.
func (r *ProductRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, qe, "id = $1 AND deleted_at IS NULL", id)
}

// GetByIDWithDeleted retrieves a product by UUID even if it has been soft-deleted (for administrative audit).
func (r *ProductRepository) GetByIDWithDeleted(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, qe, "id = $1", id)
}

// GetBySlug retrieves a single published, non-deleted product by its public SEO slug or UUID.
// Unpublished (draft/archived) products must not be reachable through public storefront lookups.
func (r *ProductRepository) GetBySlug(
	ctx context.Context,
	qe database.QueryExecutor,
	slug string,
) (*model.Product, error) {
	if parsedUUID, err := uuid.Parse(slug); err == nil && parsedUUID != uuid.Nil {
		return r.get(ctx, qe, "(slug = $1 OR id = $2) AND deleted_at IS NULL AND status = $3", slug, parsedUUID, model.PublicationStatusPublished)
	}
	return r.get(ctx, qe, "slug = $1 AND deleted_at IS NULL AND status = $2", slug, model.PublicationStatusPublished)
}

// SlugExists checks whether any product (including soft-deleted) has the specified slug.
func (r *ProductRepository) SlugExists(
	ctx context.Context,
	qe database.QueryExecutor,
	slug string,
) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)"
	err := qe.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return exists, nil
}

func (r *ProductRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	whereClause string,
	args ...any,
) (*model.Product, error) {
	if whereClause == "" {
		return nil, errors.New("get product: whereClause is required")
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			brand_id,
			category_id,
			slug,
			title,
			description,
			highlights,
			status,
			product_type,
			thumbnail_object_id,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE %s
		LIMIT 1
	`, whereClause)

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

// Update modifies mutable product columns including taxonomy, title, description, and thumbnail.
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
			slug = $3,
			title = $4,
			description = $5,
			highlights = $6,
			status = $7,
			product_type = $8,
			thumbnail_object_id = $9,
			updated_at = now()
		WHERE id = $10 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(
		ctx,
		query,
		p.BrandID,
		p.CategoryID,
		p.Slug,
		p.Title,
		p.Description,
		p.Highlights,
		p.Status,
		p.ProductType,
		p.ThumbnailObjectID,
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

// List executes a complex faceted query against products, filtering by published status,
// brand IDs/names, category IDs/names, tag IDs/names, and full-text keyword search.
// Results are projected into ProductCardReadModel with pricing and inventory from the default variant.
// The stock calculation computes total salable available inventory across all active variants of the product
// (ledger on-hand quantity minus unreleased active reservations, clamped to 0).
func (r *ProductRepository) List(
	ctx context.Context,
	qe database.QueryExecutor,
	q *ListProductsQuery,
	includeDeleted bool,
) (*pagination.PagedResult[ProductCardReadModel], error) {
	if q == nil {
		q = &ListProductsQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	argIdx := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "p.deleted_at IS NULL")
		whereClauses = append(whereClauses, "p.status = 'published'")
	}

	if len(q.BrandUUIDs) > 0 || len(q.BrandNames) > 0 {
		var brandParts []string
		if len(q.BrandUUIDs) > 0 {
			brandParts = append(brandParts, fmt.Sprintf("p.brand_id = ANY($%d)", argIdx))
			args = append(args, q.BrandUUIDs)
			argIdx++
		}
		for _, name := range q.BrandNames {
			brandParts = append(brandParts, fmt.Sprintf("b.name ILIKE $%d", argIdx))
			args = append(args, name)
			argIdx++
		}
		if len(brandParts) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(brandParts, " OR ")+")")
		}
	}

	if len(q.CategoryUUIDs) > 0 || len(q.CategoryNames) > 0 {
		var catParts []string
		if len(q.CategoryUUIDs) > 0 {
			catParts = append(catParts, fmt.Sprintf("p.category_id = ANY($%d)", argIdx))
			args = append(args, q.CategoryUUIDs)
			argIdx++
		}
		for _, name := range q.CategoryNames {
			catParts = append(catParts, fmt.Sprintf("c.name ILIKE $%d", argIdx))
			args = append(args, name)
			argIdx++
		}
		if len(catParts) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(catParts, " OR ")+")")
		}
	}

	if len(q.TagUUIDs) > 0 || len(q.TagNames) > 0 {
		var tagParts []string
		if len(q.TagUUIDs) > 0 {
			tagParts = append(tagParts, fmt.Sprintf("EXISTS (SELECT 1 FROM product_tag_assignments pta WHERE pta.product_id = p.id AND pta.tag_id = ANY($%d))", argIdx))
			args = append(args, q.TagUUIDs)
			argIdx++
		}
		for _, name := range q.TagNames {
			tagParts = append(tagParts, fmt.Sprintf("EXISTS (SELECT 1 FROM product_tag_assignments pta JOIN product_tags pt ON pta.tag_id = pt.id WHERE pta.product_id = p.id AND pt.name ILIKE $%d)", argIdx))
			args = append(args, name)
			argIdx++
		}
		if len(tagParts) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(tagParts, " OR ")+")")
		}
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.title ILIKE $%d OR p.slug ILIKE $%d OR b.name ILIKE $%d OR c.name ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
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
		return pagination.NewPagedResult([]*ProductCardReadModel{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
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
			p.slug,
			p.title,
			p.description,
			p.status,
			p.product_type,
			b.id AS brand_id,
			b.name AS brand_name,
			b.link AS brand_link,
			c.id AS category_id,
			c.name AS category_name,
			so.id as so_id,
			so.bucket as so_bucket,
			so.object_key as so_object_key,
			so.content_type as so_content_type,
			so.file_size as so_file_size,
			pv.id AS default_variant_id,
			pv.sku AS sku,
			GREATEST(
				COALESCE((SELECT SUM(il.quantity) FROM product_variants pv_all JOIN inventory_ledgers il ON il.variant_id = pv_all.id WHERE pv_all.product_id = p.id AND pv_all.deleted_at IS NULL), 0) -
				COALESCE((SELECT SUM(ir.quantity) FROM product_variants pv_all JOIN inventory_reservations ir ON ir.variant_id = pv_all.id WHERE pv_all.product_id = p.id AND pv_all.deleted_at IS NULL AND ir.released_at IS NULL AND ir.expires_at > now()), 0),
				0
			) as stock_quantity,
			pv.price as price,
			pv.crossed_out_price as crossed_out_price,
			pv.currency as currency,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		LEFT JOIN product_brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		LEFT JOIN product_categories c ON p.category_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN storage_objects so ON p.thumbnail_object_id = so.id
		LEFT JOIN product_variants pv ON p.id = pv.product_id AND pv.is_default = true AND pv.deleted_at IS NULL
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

	items := make([]*ProductCardReadModel, 0, q.PageSize)
	for rows.Next() {
		var (
			item                           ProductCardReadModel
			brandID, catID                 *uuid.UUID
			brandName, brandLink, catName  *string
			soID                           *uuid.UUID
			soBucket, soKey, soContentType *string
			soFileSize                     *int64
		)
		err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Description,
			&item.Status,
			&item.ProductType,
			&brandID,
			&brandName,
			&brandLink,
			&catID,
			&catName,
			&soID,
			&soBucket,
			&soKey,
			&soContentType,
			&soFileSize,
			&item.DefaultVariantID,
			&item.SKU,
			&item.StockQuantity,
			&item.Price,
			&item.CrossedOutPrice,
			&item.Currency,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product list item: %w", err)
		}

		if brandID != nil && brandName != nil {
			item.Brand = &PublicProductBrandReadModel{
				ID:   *brandID,
				Name: *brandName,
				Link: brandLink,
			}
		}

		if catID != nil && catName != nil {
			item.Category = &PublicProductCategoryReadModel{
				ID:   *catID,
				Name: *catName,
			}
		}

		if soID != nil {
			item.Thumbnail = &model.Object{
				ID:          *soID,
				Bucket:      *soBucket,
				Key:         *soKey,
				ContentType: *soContentType,
				FileSize:    *soFileSize,
			}
		}

		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product list items: %w", err)
	}

	return pagination.NewPagedResult(items, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// AdminList executes an administrative catalog query including drafts, archived, and optionally soft-deleted products.
func (r *ProductRepository) AdminList(
	ctx context.Context,
	qe database.QueryExecutor,
	q *ListProductsQuery,
	includeDeleted bool,
) (*pagination.PagedResult[model.Product], error) {
	if q == nil {
		q = &ListProductsQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := make([]string, 0, 5)
	args := make([]any, 0, 5)
	argIdx := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "p.deleted_at IS NULL")
	}

	if len(q.BrandUUIDs) > 0 || len(q.BrandNames) > 0 {
		var brandParts []string
		if len(q.BrandUUIDs) > 0 {
			brandParts = append(brandParts, fmt.Sprintf("p.brand_id = ANY($%d)", argIdx))
			args = append(args, q.BrandUUIDs)
			argIdx++
		}
		for _, name := range q.BrandNames {
			brandParts = append(brandParts, fmt.Sprintf("b.name ILIKE $%d", argIdx))
			args = append(args, name)
			argIdx++
		}
		if len(brandParts) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(brandParts, " OR ")+")")
		}
	}

	if len(q.CategoryUUIDs) > 0 || len(q.CategoryNames) > 0 {
		var catParts []string
		if len(q.CategoryUUIDs) > 0 {
			catParts = append(catParts, fmt.Sprintf("p.category_id = ANY($%d)", argIdx))
			args = append(args, q.CategoryUUIDs)
			argIdx++
		}
		for _, name := range q.CategoryNames {
			catParts = append(catParts, fmt.Sprintf("c.name ILIKE $%d", argIdx))
			args = append(args, name)
			argIdx++
		}
		if len(catParts) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(catParts, " OR ")+")")
		}
	}

	if len(q.TagUUIDs) > 0 || len(q.TagNames) > 0 {
		var tagParts []string
		if len(q.TagUUIDs) > 0 {
			tagParts = append(tagParts, fmt.Sprintf("EXISTS (SELECT 1 FROM product_tag_assignments pta WHERE pta.product_id = p.id AND pta.tag_id = ANY($%d))", argIdx))
			args = append(args, q.TagUUIDs)
			argIdx++
		}
		for _, name := range q.TagNames {
			tagParts = append(tagParts, fmt.Sprintf("EXISTS (SELECT 1 FROM product_tag_assignments pta JOIN product_tags pt ON pta.tag_id = pt.id WHERE pta.product_id = p.id AND pt.name ILIKE $%d)", argIdx))
			args = append(args, name)
			argIdx++
		}
		if len(tagParts) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(tagParts, " OR ")+")")
		}
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.title ILIKE $%d OR p.slug ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	// Brand/category name filters require joining the taxonomy tables; the join
	// must apply to both the count and the select, otherwise the query fails at
	// runtime (b.name/c.name reference unbound tables).
	joinStmt := ""
	if len(q.BrandNames) > 0 || len(q.CategoryNames) > 0 {
		joinStmt = `LEFT JOIN product_brands b ON b.id = p.brand_id AND b.deleted_at IS NULL
			LEFT JOIN product_categories c ON c.id = p.category_id AND c.deleted_at IS NULL`
	}

	whereStmt := ""
	if len(whereClauses) > 0 {
		whereStmt = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Total Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p %s %s", joinStmt, whereStmt)
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
			p.slug,
			p.title,
			p.description,
			p.highlights,
			p.status,
			p.product_type,
			p.thumbnail_object_id,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM products p
		%s
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, joinStmt, whereStmt, orderBy, argIdx, argIdx+1)

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

// SoftDeleteByID soft-deletes a product by its UUID (setting deleted_at = now()).
func (r *ProductRepository) SoftDeleteByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) error {
	return r.softDelete(ctx, qe, "id = $1", id)
}

func (r *ProductRepository) softDelete(
	ctx context.Context,
	qe database.QueryExecutor,
	whereClause string,
	args ...any,
) error {
	if whereClause == "" {
		return errors.New("soft delete product: whereClause is required")
	}

	query := fmt.Sprintf(`
		UPDATE products
		SET deleted_at = now(), updated_at = now()
		WHERE %s
	`, whereClause)

	cmd, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("soft delete product: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// UpdateStatus changes the publication lifecycle status (draft, published, archived) of a product.
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
