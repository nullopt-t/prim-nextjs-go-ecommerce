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
	"price":      "COALESCE(pv.price, 0)",
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
	ID               uuid.UUID
	Slug             string
	Title            string
	Description      *string
	Status           model.PublicationStatus
	Brand            *PublicProductBrandReadModel
	Category         *PublicProductCategoryReadModel
	ProductType      model.ProductType
	Thumbnail        *model.Object
	Price            *int64
	CrossedOutPrice  *int64
	Currency         *string
	SKU              *string
	DefaultVariantID *uuid.UUID
	VendorID         *uuid.UUID
	StockQuantity    int
	Rating           *model.RatingSummary
	TagsRaw          []byte
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
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
	p *model.Product,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO products (
			id,
			brand_id,
			category_id,
			slug,
			title,
			description,
			highlights,
			status,
			product_type,
			variation_axes,
			vendor_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
		RETURNING created_at, updated_at
	`

	err = tx.QueryRow(
		ctx,
		query,
		p.ID,
		p.BrandID,
		p.CategoryID,
		p.Slug,
		p.Title,
		p.Description,
		p.Highlights,
		p.Status,
		p.ProductType,
		p.VariationAxes,
		p.VendorID,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}

	return nil
}

// GetByID retrieves a single non-deleted product by its UUID.
func (r *ProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, "id = $1 AND deleted_at IS NULL", id)
}

// GetByIDWithDeleted retrieves a product by UUID even if it has been soft-deleted (for administrative audit).
func (r *ProductRepository) GetByIDWithDeleted(
	ctx context.Context,
	id uuid.UUID,
) (*model.Product, error) {
	return r.get(ctx, "id = $1", id)
}

// GetBySlug retrieves a single published, non-deleted product by its public SEO slug or UUID.
// Unpublished (draft/archived) products must not be reachable through public storefront lookups.
func (r *ProductRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*model.Product, error) {
	// A vendor-owned product is only storefront-visible while its seller is
	// active. Suspended/rejected vendor catalogs must not be reachable by slug
	// or id, even though the product row itself is still published.
	const visibleVendor = `(vendor_id IS NULL OR EXISTS (
		SELECT 1 FROM vendors v WHERE v.id = products.vendor_id AND v.status = 'active'
	))`
	if parsedUUID, err := uuid.Parse(slug); err == nil && parsedUUID != uuid.Nil {
		return r.get(ctx, fmt.Sprintf("(slug = $1 OR id = $2) AND deleted_at IS NULL AND status = $3 AND %s", visibleVendor), slug, parsedUUID, model.PublicationStatusPublished)
	}
	return r.get(ctx, fmt.Sprintf("slug = $1 AND deleted_at IS NULL AND status = $2 AND %s", visibleVendor), slug, model.PublicationStatusPublished)
}

// SlugExists checks whether any product (including soft-deleted) has the specified slug.
func (r *ProductRepository) SlugExists(
	ctx context.Context,
	slug string,
) (bool, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)"
	err = executor.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check slug exists: %w", err)
	}
	return exists, nil
}

func (r *ProductRepository) get(
	ctx context.Context,
	whereClause string,
	args ...any,
) (*model.Product, error) {
	if whereClause == "" {
		return nil, errors.New("get product: whereClause is required")
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
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
			variation_axes,
			thumbnail_object_id,
			default_variant_id,
			vendor_id,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE %s
		LIMIT 1
	`, whereClause)

	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}

	product, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.Product])
	if err != nil {
		return nil, fmt.Errorf("scan product: %w", err)
	}

	return product, nil
}

// GetVendorProductStats counts a vendor's listings and low-stock variants.
// A variant is low-stock when its salable quantity (ledger on-hand minus live
// reservations) is at or below the threshold.
func (r *ProductRepository) GetVendorProductStats(
	ctx context.Context,
	vendorID uuid.UUID,
	lowStockThreshold int,
) (*model.VendorProductStats, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	stats := new(model.VendorProductStats)
	err = executor.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE deleted_at IS NULL),
			COUNT(*) FILTER (WHERE deleted_at IS NULL AND status = 'published'),
			COUNT(*) FILTER (WHERE deleted_at IS NULL AND status = 'draft'),
			COUNT(*) FILTER (WHERE deleted_at IS NULL AND status = 'archived')
		FROM products
		WHERE vendor_id = $1
	`, vendorID).Scan(
		&stats.Total,
		&stats.Published,
		&stats.Draft,
		&stats.Archived,
	)
	if err != nil {
		return nil, fmt.Errorf("get vendor product stats: %w", err)
	}

	err = executor.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM product_variants pv
		JOIN products p ON p.id = pv.product_id
		WHERE p.vendor_id = $1
		  AND p.deleted_at IS NULL
		  AND pv.deleted_at IS NULL
		  AND GREATEST(
			COALESCE((SELECT SUM(il.quantity) FROM inventory_ledgers il WHERE il.variant_id = pv.id), 0) -
			COALESCE((SELECT SUM(ir.quantity) FROM inventory_reservations ir WHERE ir.variant_id = pv.id AND ir.released_at IS NULL AND ir.expires_at > now()), 0),
			0
		  ) <= $2
	`, vendorID, lowStockThreshold).Scan(&stats.LowStock)
	if err != nil {
		return nil, fmt.Errorf("get vendor low stock count: %w", err)
	}

	return stats, nil
}

// LockForUpdate locks the product row for update within an active transaction.
func (r *ProductRepository) LockForUpdate(
	ctx context.Context,
	id uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	var lockedID uuid.UUID
	return tx.QueryRow(ctx,
		`SELECT id FROM products WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`,
		id,
	).Scan(&lockedID)
}

// Update modifies mutable product columns including taxonomy, title, description, and thumbnail.
func (r *ProductRepository) Update(
	ctx context.Context,
	p *model.Product,
) error {
	if p.ID == uuid.Nil {
		return errors.New("update product: product ID is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
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
			variation_axes = $9,
			thumbnail_object_id = $10,
			default_variant_id = $11,
			updated_at = now()
		WHERE id = $12 AND deleted_at IS NULL
	`

	cmd, err := tx.Exec(
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
		p.VariationAxes,
		p.ThumbnailObjectID,
		p.DefaultVariantID,
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
	q *ListProductsQuery,
	includeDeleted bool,
) (*pagination.PagedResult[ProductCardReadModel], error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
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

	if len(q.VendorUUIDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("p.vendor_id = ANY($%d)", argIdx))
		args = append(args, q.VendorUUIDs)
		argIdx++
	}

	if q.IncludeUnpublished {
		// Admin listings may filter by lifecycle status.
		if q.Status != nil && strings.TrimSpace(*q.Status) != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("p.status = $%d", argIdx))
			args = append(args, strings.TrimSpace(*q.Status))
			argIdx++
		}
	} else {
		// Storefront listings are always published-only, regardless of any
		// status query parameter supplied by the caller. They also exclude
		// products whose vendor is suspended/rejected: those items are not
		// purchasable and must not surface in search or browse.
		whereClauses = append(whereClauses, "p.status = 'published'")
		whereClauses = append(whereClauses, `(p.vendor_id IS NULL OR EXISTS (
			SELECT 1 FROM vendors v WHERE v.id = p.vendor_id AND v.status = 'active'
		))`)
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

	if q.MinPrice != nil && *q.MinPrice >= 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(pv.price, 0) >= $%d", argIdx))
		args = append(args, *q.MinPrice)
		argIdx++
	}

	if q.MaxPrice != nil && *q.MaxPrice >= 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(pv.price, 0) <= $%d", argIdx))
		args = append(args, *q.MaxPrice)
		argIdx++
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.title ILIKE $%d OR p.slug ILIKE $%d OR b.name ILIKE $%d OR c.name ILIKE $%d)", argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	cursor, err := q.Keyset()
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	cursorMode := cursor != nil || q.KeysetMode
	if cursor != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.created_at, p.id) < ($%d, $%d)", argIdx, argIdx+1))
		args = append(args, cursor.CreatedAt, cursor.ID)
		argIdx += 2
	}

	whereStmt := ""
	if len(whereClauses) > 0 {
		whereStmt = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Total Count (skipped for keyset pagination, which needs no total)
	var total int
	if !cursorMode {
		// Count and select must use identical joins. In particular the brand and
		// category joins carry the same `deleted_at IS NULL` predicate as the select:
		// a soft-deleted brand/category must not satisfy a `b.name`/`c.name` search
		// in the count while the select (which nulls those joins out) excludes the
		// row, otherwise the reported total disagrees with the returned rows.
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p LEFT JOIN product_brands b ON p.brand_id = b.id AND b.deleted_at IS NULL LEFT JOIN product_categories c ON p.category_id = c.id AND c.deleted_at IS NULL LEFT JOIN product_variants pv ON p.default_variant_id = pv.id AND pv.deleted_at IS NULL %s", whereStmt)
		if err := executor.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return nil, fmt.Errorf("count products: %w", err)
		}

		if total == 0 {
			return pagination.NewPagedResult([]*ProductCardReadModel{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
		}
	}

	// 2. Sorting
	orderBy := "ORDER BY p.created_at DESC"
	if cursorMode {
		orderBy = pagination.KeysetOrderBy("p")
	} else if len(q.Sort) > 0 {
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
	tail := fmt.Sprintf("LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)
	if cursorMode {
		tail = fmt.Sprintf("LIMIT $%d", argIdx)
		queryArgs = append(slices.Clone(args), q.PageSize+1)
	}

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
			p.vendor_id,
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
		LEFT JOIN product_variants pv ON p.default_variant_id = pv.id AND pv.deleted_at IS NULL
		%s
		%s
		%s
	`, whereStmt, orderBy, tail)

	rows, err := executor.Query(ctx, selectQuery, queryArgs...)
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
			&item.VendorID,
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

	if cursorMode {
		items, page := pagination.NewCursorPageFrom(items, q.PageSize, func(p *ProductCardReadModel) pagination.Keyset {
			return pagination.Keyset{CreatedAt: p.CreatedAt, ID: p.ID}
		})
		return pagination.NewCursorPagedResult(items, page), nil
	}

	return pagination.NewPagedResult(items, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// SoftDeleteByID soft-deletes a product by its UUID (setting deleted_at = now()).
func (r *ProductRepository) SoftDeleteByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	return r.softDelete(ctx, "id = $1", id)
}

// CascadeSoftDeleteVariants soft-deletes all variants of a product.
func (r *ProductRepository) CascadeSoftDeleteVariants(
	ctx context.Context,
	productID uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}
	query := `
		UPDATE product_variants
		SET deleted_at = now(), updated_at = now()
		WHERE product_id = $1 AND deleted_at IS NULL
	`
	_, err = tx.Exec(ctx, query, productID)
	return err
}

func (r *ProductRepository) softDelete(
	ctx context.Context,
	whereClause string,
	args ...any,
) error {
	if whereClause == "" {
		return errors.New("soft delete product: whereClause is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE products
		SET deleted_at = now(), updated_at = now()
		WHERE %s
	`, whereClause)

	cmd, err := tx.Exec(ctx, query, args...)
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
	productID uuid.UUID,
	status model.PublicationStatus,
) error {
	if productID == uuid.Nil {
		return errors.New("update product status: productID is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE products
		SET status = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := tx.Exec(ctx, query, productID, status)
	if err != nil {
		return fmt.Errorf("update product status: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// SetDefaultVariant updates the default_variant_id for a product.
func (r *ProductRepository) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID *uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE products
		SET default_variant_id = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmd, err := tx.Exec(ctx, query, productID, variantID)
	if err != nil {
		return fmt.Errorf("set default variant: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
