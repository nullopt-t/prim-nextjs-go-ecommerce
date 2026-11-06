package history

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// visibleProductPredicate keeps history aligned with storefront visibility:
// deleted and unpublished products never surface, and vendor-owned products
// disappear while their seller is not active.
const visibleProductPredicate = `
	p.deleted_at IS NULL
	AND p.status = 'published'
	AND (
		p.vendor_id IS NULL
		OR EXISTS (
			SELECT 1 FROM vendors v
			WHERE v.id = p.vendor_id AND v.status = 'active'
		)
	)
`

// Repository owns SQL for per-user browsing history.
type Repository struct{}

// NewRepository constructs a history repository.
func NewRepository() *Repository {
	return &Repository{}
}

// Touch records a view for a visible product and refreshes viewed_at when the
// same product is viewed again. It reports false when the product is unknown,
// unpublished, or otherwise not visible in the storefront.
func (r *Repository) Touch(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return false, err
	}

	cmd, err := tx.Exec(ctx, `
		INSERT INTO user_product_views (user_id, product_id, viewed_at)
		SELECT $1, p.id, now()
		FROM products p
		WHERE p.id = $2
		AND `+visibleProductPredicate+`
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET viewed_at = EXCLUDED.viewed_at
	`, userID, productID)
	if err != nil {
		return false, fmt.Errorf("record product view: %w", err)
	}
	return cmd.RowsAffected() > 0, nil
}

// Prune deletes all but the newest limit views for one account.
func (r *Repository) Prune(ctx context.Context, userID uuid.UUID, limit int) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM user_product_views
		WHERE user_id = $1
		AND id NOT IN (
			SELECT id FROM user_product_views
			WHERE user_id = $1
			ORDER BY viewed_at DESC, id DESC
			LIMIT $2
		)
	`, userID, limit)
	if err != nil {
		return fmt.Errorf("prune product views: %w", err)
	}
	return nil
}

// List returns the caller's newest visible product views.
func (r *Repository) List(ctx context.Context, userID uuid.UUID, limit int) ([]ViewedProductItem, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := executor.Query(ctx, `
		SELECT
			p.id,
			p.title,
			p.slug,
			COALESCE(pv.price, 0),
			COALESCE(NULLIF(pv.currency, ''), 'EGP'),
			so.bucket,
			so.object_key,
			v.viewed_at
		FROM user_product_views v
		JOIN products p ON p.id = v.product_id
		LEFT JOIN LATERAL (
			SELECT variant.price, variant.currency
			FROM product_variants variant
			WHERE variant.product_id = p.id
			AND variant.deleted_at IS NULL
			ORDER BY (variant.id = p.default_variant_id) DESC, variant.created_at ASC
			LIMIT 1
		) pv ON true
		LEFT JOIN storage_objects so
			ON so.id = p.thumbnail_object_id AND so.deleted_at IS NULL
		WHERE v.user_id = $1
		AND `+visibleProductPredicate+`
		ORDER BY v.viewed_at DESC, p.id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list product views: %w", err)
	}
	defer rows.Close()

	items := make([]ViewedProductItem, 0, limit)
	for rows.Next() {
		var item ViewedProductItem
		if err := rows.Scan(
			&item.ProductID,
			&item.Title,
			&item.Slug,
			&item.BasePrice,
			&item.Currency,
			&item.thumbnailBucket,
			&item.thumbnailKey,
			&item.ViewedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product view: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product views: %w", err)
	}
	return items, nil
}

// Merge resolves guest product UUIDs and slugs, inserts the visible matches,
// and preserves the client's newest-first order. Existing rows keep the newer
// of the two timestamps, so a merge never moves an item backwards.
func (r *Repository) Merge(
	ctx context.Context,
	userID uuid.UUID,
	productIDs []uuid.UUID,
	slugs []string,
) (int, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return 0, err
	}

	cmd, err := tx.Exec(ctx, `
		WITH resolved AS (
			SELECT
				p.id,
				COALESCE(
					array_position($2::uuid[], p.id),
					cardinality($2::uuid[]) + array_position($3::text[], p.slug)
				) AS position
			FROM products p
			WHERE (p.id = ANY($2::uuid[]) OR p.slug = ANY($3::text[]))
			AND `+visibleProductPredicate+`
		),
		ordered AS (
			SELECT DISTINCT ON (id) id, position
			FROM resolved
			WHERE position IS NOT NULL
			ORDER BY id, position
		)
		INSERT INTO user_product_views (user_id, product_id, viewed_at)
		SELECT
			$1,
			id,
			now() - ((position - 1) * interval '1 millisecond')
		FROM ordered
		ON CONFLICT (user_id, product_id)
		DO UPDATE SET viewed_at = GREATEST(
			user_product_views.viewed_at,
			EXCLUDED.viewed_at
		)
	`, userID, productIDs, slugs)
	if err != nil {
		return 0, fmt.Errorf("merge product views: %w", err)
	}
	return int(cmd.RowsAffected()), nil
}

// Delete removes one product from the caller's history. It reports false when
// the product was not present.
func (r *Repository) Delete(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return false, err
	}

	cmd, err := tx.Exec(ctx, `
		DELETE FROM user_product_views
		WHERE user_id = $1 AND product_id = $2
	`, userID, productID)
	if err != nil {
		return false, fmt.Errorf("delete product view: %w", err)
	}
	return cmd.RowsAffected() > 0, nil
}

// Clear removes every history row for one account.
func (r *Repository) Clear(ctx context.Context, userID uuid.UUID) error {
	tx, err := database.RequireTx(ctx)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM user_product_views WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("clear product views: %w", err)
	}
	return nil
}
