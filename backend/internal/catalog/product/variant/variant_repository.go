// Package variant provides repository persistence for product variants, SKU management,
// variant default selection flags, and media asset associations.
package variant

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// allowedVariantSortFields maps public API sort keys to safe database column identifiers.
var allowedVariantSortFields = map[string]string{
	"id":         "pv.id",
	"title":      "pv.title",
	"price":      "pv.price",
	"created_at": "pv.created_at",
	"updated_at": "pv.updated_at",
}

// variantFilter encapsulates query criteria for single variant lookups.
type variantFilter struct {
	ID             *uuid.UUID
	SKU            *string
	ProductID      *uuid.UUID
	IncludeDeleted bool
}

// UpdateVariantFields defines optional update patches for product variants.
type UpdateVariantFields struct {
	SKU               *string
	Title             *string
	Price             *int64
	CrossedOutPrice   *int64
	Currency          *string
	Attributes        map[string]any
	ThumbnailObjectID *uuid.UUID
}

// ListVariantOptions defines filter and pagination parameters for querying variants by product.
type ListVariantOptions struct {
	ProductID      uuid.UUID
	Query          *pagination.ListQuery
	IncludeDeleted bool
}

// CreateVariantMediaInput defines parameters to insert a new media asset link for a variant.
type CreateVariantMediaInput struct {
	VariantID uuid.UUID
	ObjectID  uuid.UUID
	MediaType string
	SortOrder int
}

// VariantRepository provides database access for product_variants and variant_media tables.
type VariantRepository struct{}

// NewRepository creates and returns an initialized VariantRepository instance.
func NewRepository() *VariantRepository {
	return &VariantRepository{}
}

// --- Variant Operations ---

// Create inserts a new product variant record into the database.
func (vr *VariantRepository) Create(
	ctx context.Context,
	variant *model.ProductVariant,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	if variant.Attributes == nil {
		variant.Attributes = make(map[string]any)
	}

	query := `
		INSERT INTO product_variants (
			id,
			sku,
			product_id,
			title,
			price,
			crossed_out_price,
			currency,
			attributes,
			thumbnail_object_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
		RETURNING created_at, updated_at
	`

	err = tx.QueryRowContext(
		ctx,
		query,
		variant.ID,
		variant.SKU,
		variant.ProductID,
		variant.Title,
		variant.Price,
		variant.CrossedOutPrice,
		variant.Currency,
		variant.Attributes,
		variant.ThumbnailObjectID,
	).Scan(&variant.CreatedAt, &variant.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create variant: %w", err)
	}

	return nil
}

// GetByID fetches a single variant by ID.
// SetProductThumbnailIfMissing assigns the product's own thumbnail from a
// variant's image when the product has none yet. A product whose variants
// carry images must carry one itself; the first variant image wins.
func (vr *VariantRepository) SetProductThumbnailIfMissing(
	ctx context.Context,
	productID, objectID uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE products
		SET thumbnail_object_id = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND thumbnail_object_id IS NULL
	`, productID, objectID)
	if err != nil {
		return fmt.Errorf("set product thumbnail: %w", err)
	}
	return nil
}

func (vr *VariantRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.ProductVariant, error) {
	return vr.get(ctx, &variantFilter{ID: &id})
}

// GetBySKU fetches a single variant by its SKU.
func (vr *VariantRepository) GetBySKU(
	ctx context.Context,
	sku string,
) (*model.ProductVariant, error) {
	return vr.get(ctx, &variantFilter{SKU: &sku})
}

func (vr *VariantRepository) get(
	ctx context.Context,
	filter *variantFilter,
) (*model.ProductVariant, error) {
	if filter == nil || (filter.ID == nil && filter.SKU == nil && filter.ProductID == nil) {
		return nil, errors.New("get variant: filter requires at least one condition")
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	whereClauses := make([]string, 0, 3)
	args := make([]any, 0, 2)
	argIdx := 1

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "pv.deleted_at IS NULL")
	}

	if filter.ID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("pv.id = $%d", argIdx))
		args = append(args, *filter.ID)
		argIdx++
	}

	if filter.SKU != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("pv.sku = $%d", argIdx))
		args = append(args, *filter.SKU)
		argIdx++
	}

	if filter.ProductID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("pv.product_id = $%d", argIdx))
		args = append(args, *filter.ProductID)
	}

	query := fmt.Sprintf(`
		SELECT
			pv.id,
			pv.sku,
			pv.product_id,
			(CASE WHEN p.default_variant_id IS NOT NULL AND pv.id = p.default_variant_id THEN true ELSE false END) AS is_default,
			pv.title,
			pv.price,
			pv.crossed_out_price,
			pv.currency,
			pv.attributes,
			pv.thumbnail_object_id,
			pv.created_at,
			pv.updated_at,
			pv.deleted_at,
			so.id as so_id,
			so.bucket as so_bucket,
			so.object_key as so_object_key,
			so.content_type as so_content_type,
			so.file_size as so_file_size
		FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		LEFT JOIN storage_objects so ON pv.thumbnail_object_id = so.id
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get variant query: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, pgx.ErrNoRows
	}

	var (
		v                              model.ProductVariant
		soID                           *uuid.UUID
		soBucket, soKey, soContentType *string
		soFileSize                     *int64
	)
	err = rows.Scan(
		&v.ID,
		&v.SKU,
		&v.ProductID,
		&v.IsDefault,
		&v.Title,
		&v.Price,
		&v.CrossedOutPrice,
		&v.Currency,
		&v.Attributes,
		&v.ThumbnailObjectID,
		&v.CreatedAt,
		&v.UpdatedAt,
		&v.DeletedAt,
		&soID,
		&soBucket,
		&soKey,
		&soContentType,
		&soFileSize,
	)
	if err != nil {
		return nil, fmt.Errorf("get variant scan: %w", err)
	}

	if soID != nil && soBucket != nil && soKey != nil {
		v.Thumbnail = &model.Object{
			ID:          *soID,
			Bucket:      *soBucket,
			Key:         *soKey,
			ContentType: *soContentType,
			FileSize:    *soFileSize,
		}
	}

	return &v, nil
}

// Update applies partial updates to variant fields and refreshes the updated_at timestamp.
func (vr *VariantRepository) Update(
	ctx context.Context,
	variantID uuid.UUID,
	fields UpdateVariantFields,
) error {
	if variantID == uuid.Nil {
		return errors.New("update variant: variantID is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	setClauses := make([]string, 0, 8)
	args := make([]any, 0, 8)
	argIdx := 1

	if fields.SKU != nil {
		setClauses = append(setClauses, fmt.Sprintf("sku = $%d", argIdx))
		args = append(args, *fields.SKU)
		argIdx++
	}

	if fields.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *fields.Title)
		argIdx++
	}

	if fields.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argIdx))
		args = append(args, fields.Price)
		argIdx++
	}

	if fields.CrossedOutPrice != nil {
		if *fields.CrossedOutPrice <= 0 {
			setClauses = append(setClauses, "crossed_out_price = NULL")
		} else {
			setClauses = append(setClauses, fmt.Sprintf("crossed_out_price = $%d", argIdx))
			args = append(args, fields.CrossedOutPrice)
			argIdx++
		}
	}

	if fields.Currency != nil {
		setClauses = append(setClauses, fmt.Sprintf("currency = $%d", argIdx))
		args = append(args, fields.Currency)
		argIdx++
	}

	if fields.Attributes != nil {
		setClauses = append(setClauses, fmt.Sprintf("attributes = $%d", argIdx))
		args = append(args, fields.Attributes)
		argIdx++
	}

	if fields.ThumbnailObjectID != nil {
		setClauses = append(setClauses, fmt.Sprintf("thumbnail_object_id = $%d", argIdx))
		args = append(args, fields.ThumbnailObjectID)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = now()")

	query := fmt.Sprintf(`
		UPDATE product_variants
		SET %s
		WHERE id = $%d AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "), argIdx)

	args = append(args, variantID)

	cmd, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update variant: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ClearDefaultFlags unsets default_variant_id on the specified product.
func (vr *VariantRepository) ClearDefaultFlags(
	ctx context.Context,
	productID uuid.UUID,
) error {
	if productID == uuid.Nil {
		return errors.New("clear default flags: productID is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE products
		SET default_variant_id = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, query, productID)
	if err != nil {
		return fmt.Errorf("clear default flags: %w", err)
	}

	return nil
}

// SetDefaultVariant sets default_variant_id on the specified product to the given variantID.
func (vr *VariantRepository) SetDefaultVariant(
	ctx context.Context,
	productID uuid.UUID,
	variantID uuid.UUID,
) error {
	if productID == uuid.Nil || variantID == uuid.Nil {
		return errors.New("set default variant: productID and variantID are required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE products
		SET default_variant_id = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, query, productID, variantID)
	if err != nil {
		return fmt.Errorf("set default variant: %w", err)
	}

	return nil
}

// ListByProductID executes a paginated query returning all variants belonging to a product.
// Supports full-text title search and dynamic sorting.
func (vr *VariantRepository) ListByProductID(
	ctx context.Context,
	opts ListVariantOptions,
) (*pagination.PagedResult[model.ProductVariant], error) {
	if opts.ProductID == uuid.Nil {
		return nil, errors.New("list variants: productID is required")
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	q := opts.Query
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{})

	whereClauses := []string{"pv.product_id = $1"}
	args := []any{opts.ProductID}
	argIdx := 2

	if !opts.IncludeDeleted {
		whereClauses = append(whereClauses, "pv.deleted_at IS NULL")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("pv.title ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product_variants pv WHERE %s", whereStmt)
	var total int
	if err := executor.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("list variants count: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.ProductVariant{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	orderBy := "ORDER BY (CASE WHEN p.default_variant_id IS NOT NULL AND pv.id = p.default_variant_id THEN 0 ELSE 1 END) ASC, pv.created_at ASC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			dbField, ok := allowedVariantSortFields[strings.ToLower(sort.Field)]
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

	selectQuery := fmt.Sprintf(`
		SELECT
			pv.id,
			pv.sku,
			pv.product_id,
			(CASE WHEN p.default_variant_id IS NOT NULL AND pv.id = p.default_variant_id THEN true ELSE false END) AS is_default,
			pv.title,
			pv.price,
			pv.crossed_out_price,
			pv.currency,
			pv.attributes,
			pv.thumbnail_object_id,
			pv.created_at,
			pv.updated_at,
			pv.deleted_at,
			so.id as so_id,
			so.bucket as so_bucket,
			so.object_key as so_object_key,
			so.content_type as so_content_type,
			so.file_size as so_file_size
		FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		LEFT JOIN storage_objects so ON pv.thumbnail_object_id = so.id
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := executor.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list variants select: %w", err)
	}
	defer rows.Close()

	variants := make([]*model.ProductVariant, 0, q.PageSize)
	for rows.Next() {
		var (
			v                              model.ProductVariant
			soID                           *uuid.UUID
			soBucket, soKey, soContentType *string
			soFileSize                     *int64
		)
		err = rows.Scan(
			&v.ID,
			&v.SKU,
			&v.ProductID,
			&v.IsDefault,
			&v.Title,
			&v.Price,
			&v.CrossedOutPrice,
			&v.Currency,
			&v.Attributes,
			&v.ThumbnailObjectID,
			&v.CreatedAt,
			&v.UpdatedAt,
			&v.DeletedAt,
			&soID,
			&soBucket,
			&soKey,
			&soContentType,
			&soFileSize,
		)
		if err != nil {
			return nil, fmt.Errorf("list variants scan: %w", err)
		}
		if soID != nil && soBucket != nil && soKey != nil {
			v.Thumbnail = &model.Object{
				ID:          *soID,
				Bucket:      *soBucket,
				Key:         *soKey,
				ContentType: *soContentType,
				FileSize:    *soFileSize,
			}
		}
		variants = append(variants, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list variants iterate: %w", err)
	}

	return pagination.NewPagedResult(variants, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// ProductVariationContext is the parent-product state needed to validate a
// variant against its chosen variation theme.
type ProductVariationContext struct {
	ProductID     uuid.UUID
	Title         string
	CategoryID    uuid.UUID
	SellerID      *uuid.UUID
	ProductType   model.ProductType
	VariationAxes []string
}

// GetVariationContext loads the parent product's variation axes and category.
// Pass forUpdate=true when the caller holds a transaction and needs to serialize
// concurrent variant writes.
func (vr *VariantRepository) GetVariationContext(
	ctx context.Context,
	productID uuid.UUID,
	forUpdate bool,
) (*ProductVariationContext, error) {
	if productID == uuid.Nil {
		return nil, errors.New("get variation context: productID is required")
	}

	var qe database.QueryExecutor
	var err error
	if forUpdate {
		qe, err = database.RequireTx(ctx)
	} else {
		qe, err = database.Executor(ctx)
	}
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, title, category_id, seller_id, product_type, variation_axes
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	var out ProductVariationContext
	if err := qe.QueryRowContext(ctx, query, productID).Scan(
		&out.ProductID,
		&out.Title,
		&out.CategoryID,
		&out.SellerID,
		&out.ProductType,
		&out.VariationAxes,
	); err != nil {
		return nil, fmt.Errorf("get variation context: %w", err)
	}
	return &out, nil
}

// CombinationExists reports whether another active variant already uses the
// exact same attribute combination. attributes is compared as canonical jsonb.
func (vr *VariantRepository) CombinationExists(
	ctx context.Context,
	productID uuid.UUID,
	attributes map[string]any,
	excludeVariantID *uuid.UUID,
) (bool, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return false, err
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM product_variants
			WHERE product_id = $1
			  AND attributes = $2::jsonb
			  AND deleted_at IS NULL
			  AND ($3::uuid IS NULL OR id <> $3)
		)
	`
	var exists bool
	if err := executor.QueryRowContext(ctx, query, productID, attributes, excludeVariantID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check variant combination: %w", err)
	}
	return exists, nil
}

// CountByProductID returns the count of active (non-deleted) variants for a given product.
func (vr *VariantRepository) CountByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (int, error) {
	if productID == uuid.Nil {
		return 0, errors.New("count variants: productID is required")
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return 0, err
	}

	query := `SELECT COUNT(*) FROM product_variants WHERE product_id = $1 AND deleted_at IS NULL`
	var count int
	err = executor.QueryRowContext(ctx, query, productID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count variants: %w", err)
	}

	return count, nil
}

// GetProductType returns the classification (simple or variable) of a product.
func (vr *VariantRepository) GetProductType(
	ctx context.Context,
	productID uuid.UUID,
) (model.ProductType, error) {
	if productID == uuid.Nil {
		return "", errors.New("get product type: productID is required")
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return "", err
	}

	var productType model.ProductType
	err = executor.QueryRowContext(ctx,
		`SELECT product_type FROM products WHERE id = $1 AND deleted_at IS NULL`,
		productID,
	).Scan(&productType)
	if err != nil {
		return "", fmt.Errorf("get product type: %w", err)
	}

	return productType, nil
}

// Delete marks a variant as soft-deleted by setting deleted_at = now().
func (vr *VariantRepository) Delete(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return errors.New("delete variant: variantID is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE product_variants
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := tx.ExecContext(ctx, query, variantID)
	if err != nil {
		return fmt.Errorf("delete variant: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// Release any active, unreleased inventory reservations holding stock for this deleted variant
	releaseReservationsQuery := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE variant_id = $1 AND released_at IS NULL
	`
	if _, err := tx.ExecContext(ctx, releaseReservationsQuery, variantID); err != nil {
		return fmt.Errorf("release variant reservations: %w", err)
	}

	return nil
}

// SetFirstAvailableAsDefault sets the oldest active variant of a product as the new default.
func (vr *VariantRepository) SetFirstAvailableAsDefault(
	ctx context.Context,
	productID uuid.UUID,
) error {
	if productID == uuid.Nil {
		return errors.New("set first available as default: productID is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE products
		SET default_variant_id = (
			SELECT id
			FROM product_variants
			WHERE product_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC
			LIMIT 1
		), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, query, productID)
	if err != nil {
		return fmt.Errorf("set first available as default: %w", err)
	}

	return nil
}

// GetDeletedVariantInfo retrieves product_id and default_variant_id for a deleted variant.
func (vr *VariantRepository) GetDeletedVariantInfo(
	ctx context.Context,
	variantID uuid.UUID,
) (uuid.UUID, *uuid.UUID, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return uuid.Nil, nil, err
	}

	var (
		productID        uuid.UUID
		defaultVariantID *uuid.UUID
	)
	err = tx.QueryRowContext(ctx,
		`SELECT pv.product_id, p.default_variant_id
		 FROM product_variants pv
		 JOIN products p ON pv.product_id = p.id
		 WHERE pv.id = $1 AND pv.deleted_at IS NOT NULL`,
		variantID,
	).Scan(&productID, &defaultVariantID)
	return productID, defaultVariantID, err
}

// LockProduct locks the product row for update.
func (vr *VariantRepository) LockProduct(
	ctx context.Context,
	productID uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	var pID uuid.UUID
	return tx.QueryRowContext(ctx, `SELECT id FROM products WHERE id = $1 FOR UPDATE`, productID).Scan(&pID)
}

// Restore un-deletes a variant by resetting deleted_at to NULL.
func (vr *VariantRepository) Restore(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return errors.New("restore variant: variantID is required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE product_variants
		SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
	`

	cmd, err := tx.ExecContext(ctx, query, variantID)
	if err != nil {
		return fmt.Errorf("restore variant: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// --- Variant Media Operations ---

// AddMedia links an uploaded media storage object to a variant.
// If in.SortOrder is non-negative and specified, existing media at or above that sort order are shifted up.
// If in.SortOrder is negative, it appends to the end with MAX(sort_order) + 1.
func (vr *VariantRepository) AddMedia(
	ctx context.Context,
	in CreateVariantMediaInput,
) (*model.VariantMedia, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return nil, err
	}

	var (
		mediaID   int64
		sortOrder int
	)
	if in.SortOrder >= 0 {
		// Shift existing media at or above in.SortOrder to make room
		shiftQuery := `
			UPDATE variant_media
			SET sort_order = sort_order + 1
			WHERE variant_id = $1 AND sort_order >= $2
		`
		if _, err := tx.ExecContext(ctx, shiftQuery, in.VariantID, in.SortOrder); err != nil {
			return nil, fmt.Errorf("shift variant media sort order: %w", err)
		}

		insertQuery := `
			INSERT INTO variant_media (
				variant_id,
				object_id,
				media_type,
				sort_order
			)
			VALUES ($1, $2, $3, $4)
			RETURNING id, sort_order
		`
		err := tx.QueryRowContext(
			ctx,
			insertQuery,
			in.VariantID,
			in.ObjectID,
			in.MediaType,
			in.SortOrder,
		).Scan(&mediaID, &sortOrder)
		if err != nil {
			return nil, fmt.Errorf("add variant media: %w", err)
		}
	} else {
		insertQuery := `
			INSERT INTO variant_media (
				variant_id,
				object_id,
				media_type,
				sort_order
			)
			VALUES (
				$1,
				$2,
				$3,
				COALESCE((SELECT MAX(sort_order) + 1 FROM variant_media WHERE variant_id = $1), 0)
			)
			RETURNING id, sort_order
		`
		err := tx.QueryRowContext(
			ctx,
			insertQuery,
			in.VariantID,
			in.ObjectID,
			in.MediaType,
		).Scan(&mediaID, &sortOrder)
		if err != nil {
			return nil, fmt.Errorf("add variant media: %w", err)
		}
	}

	return &model.VariantMedia{
		ID:        mediaID,
		VariantID: in.VariantID,
		ObjectID:  in.ObjectID,
		MediaType: in.MediaType,
		SortOrder: sortOrder,
	}, nil
}

// ListMediaByVariantID retrieves all media attachments for a single variant ordered by sort_order.
func (vr *VariantRepository) ListMediaByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {
	if variantID == uuid.Nil {
		return nil, errors.New("list variant media: variantID is required")
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			m.id,
			m.variant_id,
			m.object_id,
			m.media_type,
			m.sort_order,
			so.id,
			so.bucket,
			so.object_key as key,
			so.content_type,
			so.file_size,
			so.status
		FROM variant_media m
		JOIN storage_objects so ON m.object_id = so.id
		WHERE m.variant_id = $1
		ORDER BY m.sort_order ASC
	`

	rows, err := executor.QueryContext(ctx, query, variantID)
	if err != nil {
		return nil, fmt.Errorf("list variant media query: %w", err)
	}
	defer rows.Close()

	var result []*model.VariantMedia
	for rows.Next() {
		m := &model.VariantMedia{
			Object: &model.Object{},
		}
		if err := rows.Scan(
			&m.ID,
			&m.VariantID,
			&m.ObjectID,
			&m.MediaType,
			&m.SortOrder,
			&m.Object.ID,
			&m.Object.Bucket,
			&m.Object.Key,
			&m.Object.ContentType,
			&m.Object.FileSize,
			&m.Object.Status,
		); err != nil {
			return nil, fmt.Errorf("list variant media scan: %w", err)
		}
		result = append(result, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list variant media iterate: %w", err)
	}

	return result, nil
}

// ListMediaByVariantIDs executes a batch query retrieving media items for multiple variants.
func (vr *VariantRepository) ListMediaByVariantIDs(
	ctx context.Context,
	variantIDs []uuid.UUID,
) (map[uuid.UUID][]*model.VariantMedia, error) {
	result := make(map[uuid.UUID][]*model.VariantMedia, len(variantIDs))
	if len(variantIDs) == 0 {
		return result, nil
	}

	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			m.id,
			m.variant_id,
			m.object_id,
			m.media_type,
			m.sort_order,
			so.id,
			so.bucket,
			so.object_key as key,
			so.content_type,
			so.file_size,
			so.status
		FROM variant_media m
		JOIN storage_objects so ON m.object_id = so.id
		WHERE m.variant_id = ANY($1)
		ORDER BY m.sort_order ASC
	`

	rows, err := executor.QueryContext(ctx, query, variantIDs)
	if err != nil {
		return nil, fmt.Errorf("list media by variant IDs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		m := &model.VariantMedia{
			Object: &model.Object{},
		}
		if err := rows.Scan(
			&m.ID,
			&m.VariantID,
			&m.ObjectID,
			&m.MediaType,
			&m.SortOrder,
			&m.Object.ID,
			&m.Object.Bucket,
			&m.Object.Key,
			&m.Object.ContentType,
			&m.Object.FileSize,
			&m.Object.Status,
		); err != nil {
			return nil, fmt.Errorf("scan variant media: %w", err)
		}
		result[m.VariantID] = append(result[m.VariantID], m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list media by variant IDs iterate: %w", err)
	}

	return result, nil
}

// RemoveMedia deletes a media attachment and resequences remaining media to keep sort_order contiguous 0..N-1.
func (vr *VariantRepository) RemoveMedia(
	ctx context.Context,
	variantID uuid.UUID,
	mediaID int64,
) error {
	if variantID == uuid.Nil || mediaID == 0 {
		return errors.New("remove variant media: variantID and mediaID are required")
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		DELETE FROM variant_media
		WHERE id = $1 AND variant_id = $2
	`

	cmd, err := tx.ExecContext(ctx, query, mediaID, variantID)
	if err != nil {
		return fmt.Errorf("remove variant media: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// Resequence remaining media items to maintain contiguous 0..N-1 sort orders
	resequenceQuery := `
		UPDATE variant_media AS vm
		SET sort_order = ranked.new_order
		FROM (
			SELECT id, ROW_NUMBER() OVER (ORDER BY sort_order ASC) - 1 AS new_order
			FROM variant_media
			WHERE variant_id = $1
		) AS ranked
		WHERE vm.id = ranked.id AND vm.variant_id = $1
	`
	_, _ = tx.ExecContext(ctx, resequenceQuery, variantID)

	return nil
}

// ReorderMedia applies a batch update of sort_order values for a variant's media attachments.
func (vr *VariantRepository) ReorderMedia(
	ctx context.Context,
	variantID uuid.UUID,
	orderedMediaIDs []int64,
) error {
	if variantID == uuid.Nil || len(orderedMediaIDs) == 0 {
		return nil
	}

	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	orders := make([]int, len(orderedMediaIDs))
	for i := range orderedMediaIDs {
		orders[i] = i
	}

	query := `
		UPDATE variant_media AS vm
		SET sort_order = batch.sort_order
		FROM (
			SELECT unnest($1::bigint[]) AS id, unnest($2::int[]) AS sort_order
		) AS batch
		WHERE vm.id = batch.id AND vm.variant_id = $3
	`

	_, err = tx.ExecContext(ctx, query, orderedMediaIDs, orders, variantID)
	if err != nil {
		return fmt.Errorf("reorder variant media: %w", err)
	}

	return nil
}

// ClearThumbnail unlinks the thumbnail storage object from a product variant.
func (r *VariantRepository) ClearThumbnail(
	ctx context.Context,
	variantID uuid.UUID,
) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	query := `
		UPDATE product_variants
		SET thumbnail_object_id = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err = tx.ExecContext(ctx, query, variantID)
	return err
}
