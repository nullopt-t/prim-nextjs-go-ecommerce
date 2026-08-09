package variant

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

var allowedVariantSortFields = map[string]string{
	"id":         "id",
	"title":      "title",
	"price":      "price",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

type variantFilter struct {
	ID             *uuid.UUID
	PublicID       *string
	ProductID      *uuid.UUID
	IncludeDeleted bool
}

type UpdateVariantFields struct {
	Title           *string
	Price           *int64
	CrossedOutPrice *int64
	Currency        *string
	Attributes      map[string]any
	IsDefault       *bool
}

type ListVariantOptions struct {
	ProductID      uuid.UUID
	Query          *pagination.ListQuery
	IncludeDeleted bool
}

type CreateVariantMediaInput struct {
	ID              uuid.UUID
	VariantID       uuid.UUID
	StorageObjectID uuid.UUID
	MediaType       string
	SortOrder       int
}

type VariantRepository struct{}

func NewRepository() *VariantRepository {
	return &VariantRepository{}
}

// --- Variant Operations ---

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
			public_id,
			product_id,
			is_default,
			title,
			price,
			crossed_out_price,
			currency,
			attributes,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
		RETURNING created_at, updated_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		variant.ID,
		variant.PublicID,
		variant.ProductID,
		variant.IsDefault,
		variant.Title,
		variant.Price,
		variant.CrossedOutPrice,
		variant.Currency,
		variant.Attributes,
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

// GetByPublicID fetches a single variant by its public ID.
func (vr *VariantRepository) GetByPublicID(
	ctx context.Context,
	qe database.QueryExecutor,
	publicID string,
) (*model.ProductVariant, error) {
	return vr.get(ctx, qe, &variantFilter{PublicID: &publicID})
}

func (vr *VariantRepository) get(
	ctx context.Context,
	qe database.QueryExecutor,
	filter *variantFilter,
) (*model.ProductVariant, error) {
	if filter == nil || (filter.ID == nil && filter.PublicID == nil && filter.ProductID == nil) {
		return nil, errors.New("get variant: filter requires at least one condition")
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
	
	if filter.PublicID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("public_id = $%d", argIdx))
		args = append(args, *filter.PublicID)
		argIdx++
	}

	if filter.ProductID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("product_id = $%d", argIdx))
		args = append(args, *filter.ProductID)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			public_id,
			product_id,
			is_default,
			title,
			price,
			crossed_out_price,
			currency,
			attributes,
			created_at,
			updated_at,
			deleted_at
		FROM product_variants
		WHERE %s
	`, strings.Join(whereClauses, " AND "))

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get variant query: %w", err)
	}

	variant, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.ProductVariant])
	if err != nil {
		return nil, fmt.Errorf("get variant scan: %w", err)
	}

	return variant, nil
}

func (vr *VariantRepository) Update(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
	fields UpdateVariantFields,
) error {
	if variantID == uuid.Nil {
		return errors.New("update variant: variantID is required")
	}

	setClauses := make([]string, 0, 7)
	args := make([]any, 0, 7)
	argIdx := 1

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

	whereClauses := []string{"product_id = $1"}
	args := []any{opts.ProductID}
	argIdx := 2

	if !opts.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	search := strings.TrimSpace(q.Search)
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("title ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereStmt := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product_variants WHERE %s", whereStmt)
	var total int
	if err := qe.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("list variants count: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.ProductVariant{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	orderBy := "ORDER BY is_default DESC, created_at ASC"
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
			id,
			public_id,
			product_id,
			is_default,
			title,
			price,
			crossed_out_price,
			currency,
			attributes,
			created_at,
			updated_at,
			deleted_at
		FROM product_variants
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereStmt, orderBy, argIdx, argIdx+1)

	queryArgs := append(slices.Clone(args), q.PageSize, q.Offset)

	rows, err := qe.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list variants select: %w", err)
	}

	variants, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.ProductVariant])
	if err != nil {
		return nil, fmt.Errorf("list variants collect rows: %w", err)
	}

	return pagination.NewPagedResult(variants, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

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

	return nil
}

// --- Variant Media Operations ---

func (vr *VariantRepository) AddMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	in CreateVariantMediaInput,
) (*model.VariantMedia, error) {
	query := `
		INSERT INTO variant_media (
			id,
			variant_id,
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
		in.VariantID,
		in.StorageObjectID,
		in.MediaType,
		in.SortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("add variant media: %w", err)
	}

	return &model.VariantMedia{
		ID:              in.ID,
		VariantID:       in.VariantID,
		StorageObjectID: in.StorageObjectID,
		MediaType:       in.MediaType,
		SortOrder:       in.SortOrder,
	}, nil
}

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
			m.storage_object_id,
			m.media_type,
			m.sort_order,
			so.id,
			so.bucket,
			so.object_key as key,
			so.content_type,
			so.file_size,
			so.status
		FROM variant_media m
		JOIN storage_objects so ON m.storage_object_id = so.id
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
			return nil, fmt.Errorf("list variant media scan: %w", err)
		}
		result = append(result, m)
	}

	return result, nil
}

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

	return nil
}

func (vr *VariantRepository) ReorderMedia(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
	orderedMediaIDs []uuid.UUID,
) error {
	if variantID == uuid.Nil || len(orderedMediaIDs) == 0 {
		return nil
	}

	query := `
		UPDATE variant_media
		SET sort_order = $1
		WHERE id = $2 AND variant_id = $3
	`

	for idx, mediaID := range orderedMediaIDs {
		if _, err := qe.Exec(ctx, query, idx, mediaID, variantID); err != nil {
			return fmt.Errorf("reorder variant media item %s: %w", mediaID, err)
		}
	}

	return nil
}
