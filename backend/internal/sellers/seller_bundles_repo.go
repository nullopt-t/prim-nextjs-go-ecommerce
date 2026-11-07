package sellers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// BundleItem is one component of a product bundle.
type BundleItem struct {
	ID        uuid.UUID  `json:"id"`
	BundleID  uuid.UUID  `json:"-"`
	ProductID uuid.UUID  `json:"productId"`
	VariantID *uuid.UUID `json:"variantId,omitempty"`
	Quantity  int        `json:"quantity"`
	Title     string     `json:"title,omitempty"`
	CreatedAt time.Time  `json:"-"`
}

// SellerBundle is the merchant-facing bundle projection.
type SellerBundle struct {
	ID           uuid.UUID    `json:"id"`
	ProductID    uuid.UUID    `json:"productId"`
	ProductSlug  string       `json:"productSlug"`
	ProductTitle string       `json:"productTitle"`
	Status       string       `json:"status"`
	BundlePrice  int64        `db:"bundle_price" json:"bundlePrice"`
	Currency     string       `db:"currency" json:"currency"`
	Items        []BundleItem `db:"-" json:"items"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// BundleStore is the seller portal's window into bundle persistence.
type BundleStore interface {
	CreateBundle(ctx context.Context, sellerID uuid.UUID, bundlePrice int64, currency string, productID uuid.UUID, items []BundleItem) (*SellerBundle, error)
	ListBundles(ctx context.Context, sellerID uuid.UUID) ([]SellerBundle, error)
	DeleteBundle(ctx context.Context, sellerID, bundleID uuid.UUID) (productID uuid.UUID, err error)
}

type bundleStore struct {
	runner database.Runner
}

// NewBundleStore builds the bundle persistence store on the shared runner.
func NewBundleStore(runner database.Runner) *bundleStore {
	return &bundleStore{runner: runner}
}

func (b *bundleStore) CreateBundle(ctx context.Context, sellerID uuid.UUID, bundlePrice int64, currency string, productID uuid.UUID, items []BundleItem) (*SellerBundle, error) {
	var bundle *SellerBundle
	err := b.runner.WithTx(ctx, func(ctx context.Context) error {
		tx, err := database.RequireTx(ctx)
		if err != nil {
			return err
		}
		if currency == "" {
			currency = "EGP"
		}
		bundle = &SellerBundle{
			ID:          uuid.New(),
			ProductID:   productID,
			BundlePrice: bundlePrice,
			Currency:    currency,
			Items:       make([]BundleItem, 0, len(items)),
		}
		if err := tx.QueryRow(ctx, `
			INSERT INTO product_bundles (id, seller_id, product_id, bundle_price, currency, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, now(), now())
			RETURNING created_at, updated_at
		`, bundle.ID, sellerID, productID, bundlePrice, currency).Scan(&bundle.CreatedAt, &bundle.UpdatedAt); err != nil {
			return fmt.Errorf("create bundle: %w", err)
		}
		for _, item := range items {
			item.ID = uuid.New()
			item.BundleID = bundle.ID
			if err := tx.QueryRow(ctx, `
				INSERT INTO bundle_items (id, bundle_id, product_id, variant_id, quantity, created_at)
				VALUES ($1, $2, $3, $4, $5, now())
				RETURNING created_at
			`, item.ID, bundle.ID, item.ProductID, item.VariantID, item.Quantity).Scan(&item.CreatedAt); err != nil {
				return fmt.Errorf("create bundle item: %w", err)
			}
			bundle.Items = append(bundle.Items, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func (b *bundleStore) ListBundles(ctx context.Context, sellerID uuid.UUID) ([]SellerBundle, error) {
	executor, err := database.Executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := executor.Query(ctx, `
		SELECT pb.id, pb.product_id, p.slug AS product_slug, p.title AS product_title, p.status::text, pb.bundle_price, pb.currency, pb.created_at, pb.updated_at
		FROM product_bundles pb
		JOIN products p ON p.id = pb.product_id AND p.deleted_at IS NULL
		WHERE pb.seller_id = $1
		ORDER BY pb.created_at DESC
	`, sellerID)
	if err != nil {
		return nil, fmt.Errorf("list bundles: %w", err)
	}
	defer rows.Close()

	bundles, err := pgx.CollectRows(rows, pgx.RowToStructByName[SellerBundle])
	if err != nil {
		return nil, fmt.Errorf("scan bundles: %w", err)
	}
	if len(bundles) == 0 {
		return bundles, nil
	}
	// Batch-load every bundle's composition in one query instead of one
	// round trip per bundle.
	bundleIDs := make([]uuid.UUID, len(bundles))
	indexByID := make(map[uuid.UUID]int, len(bundles))
	for i := range bundles {
		bundleIDs[i] = bundles[i].ID
		indexByID[bundles[i].ID] = i
	}
	items, err := b.listItemsForBundles(ctx, executor, bundleIDs)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		bundles[indexByID[item.BundleID]].Items = append(bundles[indexByID[item.BundleID]].Items, item)
	}
	return bundles, nil
}

func (b *bundleStore) listItemsForBundles(ctx context.Context, executor database.QueryExecutor, bundleIDs []uuid.UUID) ([]BundleItem, error) {
	rows, err := executor.Query(ctx, `
		SELECT bi.id, bi.bundle_id, bi.product_id, bi.variant_id, bi.quantity, bi.created_at,
		       COALESCE(pv.title, p.title) AS title
		FROM bundle_items bi
		JOIN products p ON p.id = bi.product_id
		LEFT JOIN product_variants pv ON pv.id = bi.variant_id
		WHERE bi.bundle_id = ANY($1)
		ORDER BY bi.created_at ASC
	`, bundleIDs)
	if err != nil {
		return nil, fmt.Errorf("list bundle items: %w", err)
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[BundleItem])
}

func (b *bundleStore) DeleteBundle(ctx context.Context, sellerID, bundleID uuid.UUID) (uuid.UUID, error) {
	var productID uuid.UUID
	err := b.runner.WithTx(ctx, func(ctx context.Context) error {
		tx, err := database.RequireTx(ctx)
		if err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `
			SELECT product_id FROM product_bundles WHERE seller_id = $1 AND id = $2
		`, sellerID, bundleID).Scan(&productID); err != nil {
			return fmt.Errorf("get bundle product: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM product_bundles WHERE seller_id = $1 AND id = $2`, sellerID, bundleID); err != nil {
			return fmt.Errorf("delete bundle: %w", err)
		}
		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	return productID, nil
}
