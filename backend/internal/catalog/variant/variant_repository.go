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
	IsDefault         *bool
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
	ID        uuid.UUID
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
	qe database.QueryExecutor,
	variant *model.ProductVariant,
) error {
	if variant.Attributes == nil {
		variant.Attributes = make(map[string]any)
	}

	query := `
		INSERT INTO product_variants (
			id,
			sku,
			product_id,
			is_default,
			title,
			price,
			crossed_out_price,
			currency,
			attributes,
			thumbnail_object_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		variant.ID,
		variant.SKU,
		variant.ProductID,
		variant.IsDefault,
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
func (vr *VariantRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.ProductVariant, error) {
	return vr.get(ctx, qe, &variantFilter{ID: &id})
}

// GetBySKU fetches a single variant by its SKU.
func (vr *VariantRepository) GetBySKU(
	ctx context.Context,
	qe database.QueryExecutor,
	sku string,
) (*model.ProductVariant, error) {
	return vr.get(ctx, qe, &variantFilter{SKU: &sku})
}

func (vr *VariantRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter *variantFilter,
) (*model.ProductVariant, error) {
	if filter == nil || (filter.ID == nil && filter.SKU == nil && filter.ProductID == nil) {
		return nil, errors.New("get variant: filter requires at least one condition")
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
			pv.is_default,
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
		LEFT JOIN storage_objects so ON pv.thumbnail_object_id = so.id
		WHERE %s
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	rows, err := qe.Query(ctx, query, args...)
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
	qe database.QueryExecutor,
	variantID uuid.UUID,
	fields UpdateVariantFields,
) error {
	if variantID == uuid.Nil {
		return errors.New("update variant: variantID is required")
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
		setClauses = append(setClauses, fmt.Sprintf("crossed_out_price = $%d", argIdx))
		args = append(args, fields.CrossedOutPrice)
		argIdx++
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

	if fields.IsDefault != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_default = $%d", argIdx))
		args = append(args, *fields.IsDefault)
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

	cmd, err := qe.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update variant: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ClearDefaultFlags sets is_default = false on all active variants of the specified product.
func (vr *VariantRepository) ClearDefaultFlags(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) error {
	if productID == uuid.Nil {
		return errors.New("clear default flags: productID is required")
	}

	query := `
		UPDATE product_variants
		SET is_default = false, updated_at = now()
		WHERE product_id = $1 AND is_default = true AND deleted_at IS NULL
	`

	_, err := qe.Exec(ctx, query, productID)
	if err != nil {
		return fmt.Errorf("clear default flags: %w", err)
	}

	return nil
}

// ListByProductID executes a paginated query returning all variants belonging to a product.
// Supports full-text title search and dynamic sorting.
func (vr *VariantRepository) ListByProductID(
	ctx context.Context,
	qe database.QueryExecutor,
	opts ListVariantOptions,
) (*pagination.PagedResult[model.ProductVariant], error) {
	if opts.ProductID == uuid.Nil {
		return nil, errors.New("list variants: productID is required")
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
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("list variants count: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.ProductVariant{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	orderBy := "ORDER BY pv.is_default DESC, pv.created_at ASC"
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
			pv.is_default,
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
		LEFT JOIN storage_objects so ON pv.thumbnail_object_id = so.id
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
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

// CountByProductID returns the count of active (non-deleted) variants for a given product.
func (vr *VariantRepository) CountByProductID(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) (int, error) {
	if productID == uuid.Nil {
		return 0, errors.New("count variants: productID is required")
	}

	query := `SELECT COUNT(*) FROM product_variants WHERE product_id = $1 AND deleted_at IS NULL`
	var count int
	err := qe.QueryRow(ctx, query, productID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count variants: %w", err)
	}

	return count, nil
}

// Delete marks a variant as soft-deleted by setting deleted_at = now().
func (vr *VariantRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return errors.New("delete variant: variantID is required")
	}

	query := `
		UPDATE product_variants
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, variantID)
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
	if _, err := qe.Exec(ctx, releaseReservationsQuery, variantID); err != nil {
		return fmt.Errorf("release variant reservations: %w", err)
	}

	return nil
}

// SetFirstAvailableAsDefault sets the oldest active variant of a product as the new default.
func (vr *VariantRepository) SetFirstAvailableAsDefault(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) error {
	if productID == uuid.Nil {
		return errors.New("set first available as default: productID is required")
	}

	query := `
		UPDATE product_variants
		SET is_default = true, updated_at = now()
		WHERE id = (
			SELECT id
			FROM product_variants
			WHERE product_id = $1 AND deleted_at IS NULL
			ORDER BY created_at ASC
			LIMIT 1
		)
	`

	_, err := qe.Exec(ctx, query, productID)
	if err != nil {
		return fmt.Errorf("set first available as default: %w", err)
	}

	return nil
}

// Restore un-deletes a variant by resetting deleted_at to NULL.
func (vr *VariantRepository) Restore(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) error {
	if variantID == uuid.Nil {
		return errors.New("restore variant: variantID is required")
	}

	query := `
		UPDATE product_variants
		SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
	`

	cmd, err := qe.Exec(ctx, query, variantID)
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
	qe database.QueryExecutor,
	in CreateVariantMediaInput,
) (*model.VariantMedia, error) {
	var sortOrder int
	if in.SortOrder >= 0 {
		// Shift existing media at or above in.SortOrder to make room
		shiftQuery := `
			UPDATE variant_media
			SET sort_order = sort_order + 1
			WHERE variant_id = $1 AND sort_order >= $2
		`
		if _, err := qe.Exec(ctx, shiftQuery, in.VariantID, in.SortOrder); err != nil {
			return nil, fmt.Errorf("shift variant media sort order: %w", err)
		}

		insertQuery := `
			INSERT INTO variant_media (
				id,
				variant_id,
				object_id,
				media_type,
				sort_order
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING sort_order
		`
		err := qe.QueryRow(
			ctx,
			insertQuery,
			in.ID,
			in.VariantID,
			in.ObjectID,
			in.MediaType,
			in.SortOrder,
		).Scan(&sortOrder)
		if err != nil {
			return nil, fmt.Errorf("add variant media: %w", err)
		}
	} else {
		insertQuery := `
			INSERT INTO variant_media (
				id,
				variant_id,
				object_id,
				media_type,
				sort_order
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				COALESCE((SELECT MAX(sort_order) + 1 FROM variant_media WHERE variant_id = $2), 0)
			)
			RETURNING sort_order
		`
		err := qe.QueryRow(
			ctx,
			insertQuery,
			in.ID,
			in.VariantID,
			in.ObjectID,
			in.MediaType,
		).Scan(&sortOrder)
		if err != nil {
			return nil, fmt.Errorf("add variant media: %w", err)
		}
	}

	return &model.VariantMedia{
		ID:        in.ID,
		VariantID: in.VariantID,
		ObjectID:  in.ObjectID,
		MediaType: in.MediaType,
		SortOrder: sortOrder,
	}, nil
}

// ListMediaByVariantID retrieves all media attachments for a single variant ordered by sort_order.
func (vr *VariantRepository) ListMediaByVariantID(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) ([]*model.VariantMedia, error) {
	if variantID == uuid.Nil {
		return nil, errors.New("list variant media: variantID is required")
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

	rows, err := qe.Query(ctx, query, variantID)
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
	qe database.QueryExecutor,
	variantIDs []uuid.UUID,
) (map[uuid.UUID][]*model.VariantMedia, error) {
	result := make(map[uuid.UUID][]*model.VariantMedia, len(variantIDs))
	if len(variantIDs) == 0 {
		return result, nil
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

	rows, err := qe.Query(ctx, query, variantIDs)
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
	qe database.QueryExecutor,
	variantID uuid.UUID,
	mediaID uuid.UUID,
) error {
	if variantID == uuid.Nil || mediaID == uuid.Nil {
		return errors.New("remove variant media: variantID and mediaID are required")
	}

	query := `
		DELETE FROM variant_media
		WHERE id = $1 AND variant_id = $2
	`

	cmd, err := qe.Exec(ctx, query, mediaID, variantID)
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
	_, _ = qe.Exec(ctx, resequenceQuery, variantID)

	return nil
}

// ReorderMedia applies a batch update of sort_order values for a variant's media attachments.
func (vr *VariantRepository) ReorderMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if variantID == uuid.Nil || len(orderedMediaIDs) == 0 {
		return nil
	}

	orders := make([]int, len(orderedMediaIDs))
	for i := range orderedMediaIDs {
		orders[i] = i
	}

	query := `
		UPDATE variant_media AS vm
		SET sort_order = batch.sort_order
		FROM (
			SELECT unnest($1::uuid[]) AS id, unnest($2::int[]) AS sort_order
		) AS batch
		WHERE vm.id = batch.id AND vm.variant_id = $3
	`

	_, err := qe.Exec(ctx, query, orderedMediaIDs, orders, variantID)
	if err != nil {
		return fmt.Errorf("reorder variant media: %w", err)
	}

	return nil
}
