package wishlist

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

var allowedWishlistSortFields = map[string]string{
	"created_at": "wi.created_at",
	"title":      "p.title",
	"price":      "pv.price",
}

type WishlistRepository struct{}

func NewWishlistRepository() *WishlistRepository {
	return &WishlistRepository{}
}

func (r *WishlistRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	item *model.WishlistItem,
) error {
	query := `
		INSERT INTO wishlist_items (id, user_id, product_id, variant_id, created_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING created_at
	`
	return qe.QueryRow(ctx, query, item.ID, item.UserID, item.ProductID, item.VariantID).Scan(&item.CreatedAt)
}

func (r *WishlistRepository) GetByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
	userID uuid.UUID,
) (*model.WishlistItem, error) {
	query := `
		SELECT
			wi.id,
			wi.user_id,
			wi.product_id,
			wi.variant_id,
			wi.created_at,
			p.title,
			p.slug,
			p.description,
			p.product_type,
			p.status,
			b.name AS brand_name,
			c.name AS category_name,
			COALESCE(vso.bucket, so.bucket) AS thumb_bucket,
			COALESCE(vso.object_key, so.object_key) AS thumb_key,
			pv.id AS variant_id_val,
			pv.title AS variant_title,
			pv.sku AS variant_sku,
			pv.price AS variant_price,
			pv.crossed_out_price AS variant_crossed_out_price,
			pv.currency AS variant_currency,
			COALESCE(
				EXISTS (
					SELECT 1 FROM product_variants v
					WHERE v.product_id = p.id AND v.deleted_at IS NULL
					  AND (wi.variant_id IS NULL OR v.id = wi.variant_id)
					  AND COALESCE((SELECT SUM(il.quantity) FROM inventory_ledgers il WHERE il.variant_id = v.id), 0) > 0
				),
				false
			) AS in_stock
		FROM wishlist_items wi
		JOIN products p ON wi.product_id = p.id AND p.deleted_at IS NULL
		LEFT JOIN product_brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		LEFT JOIN product_categories c ON p.category_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN storage_objects so ON p.thumbnail_object_id = so.id
		LEFT JOIN product_variants pv ON pv.id = COALESCE(
			wi.variant_id,
			(SELECT def.id FROM product_variants def WHERE def.product_id = p.id AND def.is_default = true AND def.deleted_at IS NULL LIMIT 1),
			(SELECT first_v.id FROM product_variants first_v WHERE first_v.product_id = p.id AND first_v.deleted_at IS NULL ORDER BY first_v.created_at ASC LIMIT 1)
		)
		LEFT JOIN storage_objects vso ON pv.thumbnail_object_id = vso.id
		WHERE wi.id = $1 AND wi.user_id = $2
	`
	row := qe.QueryRow(ctx, query, id, userID)
	item, err := scanWishlistItem(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *WishlistRepository) GetByUserAndProduct(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	productID uuid.UUID,
) (*model.WishlistItem, error) {
	query := `
		SELECT
			wi.id,
			wi.user_id,
			wi.product_id,
			wi.variant_id,
			wi.created_at,
			p.title,
			p.slug,
			p.description,
			p.product_type,
			p.status,
			b.name AS brand_name,
			c.name AS category_name,
			COALESCE(vso.bucket, so.bucket) AS thumb_bucket,
			COALESCE(vso.object_key, so.object_key) AS thumb_key,
			pv.id AS variant_id_val,
			pv.title AS variant_title,
			pv.sku AS variant_sku,
			pv.price AS variant_price,
			pv.crossed_out_price AS variant_crossed_out_price,
			pv.currency AS variant_currency,
			COALESCE(
				EXISTS (
					SELECT 1 FROM product_variants v
					WHERE v.product_id = p.id AND v.deleted_at IS NULL
					  AND (wi.variant_id IS NULL OR v.id = wi.variant_id)
					  AND COALESCE((SELECT SUM(il.quantity) FROM inventory_ledgers il WHERE il.variant_id = v.id), 0) > 0
				),
				false
			) AS in_stock
		FROM wishlist_items wi
		JOIN products p ON wi.product_id = p.id AND p.deleted_at IS NULL
		LEFT JOIN product_brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		LEFT JOIN product_categories c ON p.category_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN storage_objects so ON p.thumbnail_object_id = so.id
		LEFT JOIN product_variants pv ON pv.id = COALESCE(
			wi.variant_id,
			(SELECT def.id FROM product_variants def WHERE def.product_id = p.id AND def.is_default = true AND def.deleted_at IS NULL LIMIT 1),
			(SELECT first_v.id FROM product_variants first_v WHERE first_v.product_id = p.id AND first_v.deleted_at IS NULL ORDER BY first_v.created_at ASC LIMIT 1)
		)
		LEFT JOIN storage_objects vso ON pv.thumbnail_object_id = vso.id
		WHERE wi.user_id = $1 AND wi.product_id = $2
	`
	row := qe.QueryRow(ctx, query, userID, productID)
	item, err := scanWishlistItem(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *WishlistRepository) ListByUser(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	q *pagination.ListQuery,
) ([]model.WishlistItem, int, error) {
	countQuery := `
		SELECT COUNT(1)
		FROM wishlist_items wi
		JOIN products p ON wi.product_id = p.id AND p.deleted_at IS NULL
		WHERE wi.user_id = $1
	`
	var total int
	if err := qe.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.WishlistItem{}, 0, nil
	}

	orderBy := "ORDER BY wi.created_at DESC"
	if len(q.Sort) > 0 {
		sortParts := make([]string, 0, len(q.Sort))
		for _, sort := range q.Sort {
			col, ok := allowedWishlistSortFields[strings.ToLower(sort.Field)]
			if !ok {
				continue
			}
			dir := "ASC"
			if sort.Order == pagination.SortDesc {
				dir = "DESC"
			}
			sortParts = append(sortParts, fmt.Sprintf("%s %s", col, dir))
		}
		if len(sortParts) > 0 {
			orderBy = "ORDER BY " + strings.Join(sortParts, ", ")
		}
	}

	selectQuery := fmt.Sprintf(`
		SELECT
			wi.id,
			wi.user_id,
			wi.product_id,
			wi.variant_id,
			wi.created_at,
			p.title,
			p.slug,
			p.description,
			p.product_type,
			p.status,
			b.name AS brand_name,
			c.name AS category_name,
			COALESCE(vso.bucket, so.bucket) AS thumb_bucket,
			COALESCE(vso.object_key, so.object_key) AS thumb_key,
			pv.id AS variant_id_val,
			pv.title AS variant_title,
			pv.sku AS variant_sku,
			pv.price AS variant_price,
			pv.crossed_out_price AS variant_crossed_out_price,
			pv.currency AS variant_currency,
			COALESCE(
				EXISTS (
					SELECT 1 FROM product_variants v
					WHERE v.product_id = p.id AND v.deleted_at IS NULL
					  AND (wi.variant_id IS NULL OR v.id = wi.variant_id)
					  AND COALESCE((SELECT SUM(il.quantity) FROM inventory_ledgers il WHERE il.variant_id = v.id), 0) > 0
				),
				false
			) AS in_stock
		FROM wishlist_items wi
		JOIN products p ON wi.product_id = p.id AND p.deleted_at IS NULL
		LEFT JOIN product_brands b ON p.brand_id = b.id AND b.deleted_at IS NULL
		LEFT JOIN product_categories c ON p.category_id = c.id AND c.deleted_at IS NULL
		LEFT JOIN storage_objects so ON p.thumbnail_object_id = so.id
		LEFT JOIN product_variants pv ON pv.id = COALESCE(
			wi.variant_id,
			(SELECT def.id FROM product_variants def WHERE def.product_id = p.id AND def.is_default = true AND def.deleted_at IS NULL LIMIT 1),
			(SELECT first_v.id FROM product_variants first_v WHERE first_v.product_id = p.id AND first_v.deleted_at IS NULL ORDER BY first_v.created_at ASC LIMIT 1)
		)
		LEFT JOIN storage_objects vso ON pv.thumbnail_object_id = vso.id
		WHERE wi.user_id = $1
		%s
		LIMIT $2 OFFSET $3
	`, orderBy)

	rows, err := qe.Query(ctx, selectQuery, userID, q.PageSize, q.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.WishlistItem, 0, q.PageSize)
	for rows.Next() {
		item, err := scanWishlistItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *WishlistRepository) CountByUser(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (int, error) {
	query := `
		SELECT COUNT(1)
		FROM wishlist_items wi
		JOIN products p ON wi.product_id = p.id AND p.deleted_at IS NULL
		WHERE wi.user_id = $1
	`
	var count int
	err := qe.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *WishlistRepository) DeleteByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	query := `DELETE FROM wishlist_items WHERE id = $1 AND user_id = $2`
	tag, err := qe.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *WishlistRepository) DeleteByProduct(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) error {
	if variantID != nil {
		query := `DELETE FROM wishlist_items WHERE user_id = $1 AND product_id = $2 AND variant_id = $3`
		tag, err := qe.Exec(ctx, query, userID, productID, *variantID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return nil
	}

	query := `DELETE FROM wishlist_items WHERE user_id = $1 AND product_id = $2`
	tag, err := qe.Exec(ctx, query, userID, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *WishlistRepository) ClearByUser(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) error {
	query := `DELETE FROM wishlist_items WHERE user_id = $1`
	_, err := qe.Exec(ctx, query, userID)
	return err
}

func (r *WishlistRepository) Exists(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) (bool, *uuid.UUID, error) {
	var query string
	var id uuid.UUID
	var err error

	if variantID != nil {
		query = `SELECT id FROM wishlist_items WHERE user_id = $1 AND product_id = $2 AND variant_id = $3`
		err = qe.QueryRow(ctx, query, userID, productID, *variantID).Scan(&id)
	} else {
		query = `SELECT id FROM wishlist_items WHERE user_id = $1 AND product_id = $2 LIMIT 1`
		err = qe.QueryRow(ctx, query, userID, productID).Scan(&id)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, &id, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanWishlistItem(scanner rowScanner) (*model.WishlistItem, error) {
	var (
		item          model.WishlistItem
		pTypeStr      string
		pStatusStr    string
		brandName     *string
		categoryName  *string
		thumbBucket   *string
		thumbKey      *string
		variantIDVal  *uuid.UUID
		variantTitle  *string
		variantSKU    *string
		price         *int64
		crossedPrice  *int64
		currency      *string
		inStock       bool
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.VariantID,
		&item.CreatedAt,
		&item.ProductTitle,
		&item.ProductSlug,
		&item.ProductDescription,
		&pTypeStr,
		&pStatusStr,
		&brandName,
		&categoryName,
		&thumbBucket,
		&thumbKey,
		&variantIDVal,
		&variantTitle,
		&variantSKU,
		&price,
		&crossedPrice,
		&currency,
		&inStock,
	)
	if err != nil {
		return nil, err
	}

	item.ProductType = model.ProductType(pTypeStr)
	item.ProductStatus = model.PublicationStatus(pStatusStr)
	item.BrandName = brandName
	item.CategoryName = categoryName
	item.ThumbnailBucket = thumbBucket
	item.ThumbnailKey = thumbKey
	if item.VariantID == nil && variantIDVal != nil {
		item.VariantID = variantIDVal
	}
	item.VariantTitle = variantTitle
	item.VariantSKU = variantSKU
	item.Price = price
	item.CrossedOutPrice = crossedPrice
	item.Currency = currency
	item.InStock = inStock

	return &item, nil
}
