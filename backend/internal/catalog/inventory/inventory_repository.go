// Package inventory provides repository and domain service operations for managing product variant stock levels,
// append-only inventory ledgers, and temporary cart-based stock reservations.
package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

// InventoryRepository handles persistence and database query execution for inventory ledgers,
// variant row locking, stock balances, and temporary checkout reservations.
type InventoryRepository struct{}

// NewRepository creates and returns an initialized InventoryRepository instance.
func NewRepository() *InventoryRepository {
	return &InventoryRepository{}
}

// LockVariantForUpdate acquires an exclusive row-level write lock (SELECT ... FOR UPDATE) on the specified variant.
// This serializes concurrent inventory operations (such as stock adjustments and reservation creations) on the variant,
// preventing concurrent race conditions and preventing overselling.
func (r *InventoryRepository) LockVariantForUpdate(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) error {
	// Query product_variants table by ID with FOR UPDATE to block competing transactions on this SKU.
	query := `SELECT id FROM product_variants WHERE id = $1 FOR UPDATE`
	var id uuid.UUID
	err := qe.QueryRow(ctx, query, variantID).Scan(&id)
	if err != nil {
		return fmt.Errorf("lock variant for update: %w", err)
	}
	return nil
}

// CreateLedger records a new immutable append-only transaction entry into the inventory_ledgers table.
// Positive quantities indicate stock additions (restocks, customer returns), whereas negative quantities
// indicate inventory reductions (sales fulfillments, scrap/damage adjustments).
func (r *InventoryRepository) CreateLedger(
	ctx context.Context,
	qe database.QueryExecutor,
	ledger *model.InventoryLedger,
) error {
	// Insert ledger entry with server-generated created_at timestamp.
	query := `
		INSERT INTO inventory_ledgers (
			id,
			variant_id,
			quantity,
			reason,
			created_at
		)
		VALUES ($1, $2, $3, $4, now())
		RETURNING created_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		ledger.ID,
		ledger.VariantID,
		ledger.Quantity,
		ledger.Reason,
	).Scan(&ledger.CreatedAt)
	if err != nil {
		return fmt.Errorf("create inventory ledger: %w", err)
	}

	return nil
}

// GetStock computes real-time aggregated inventory quantities for a given variant:
// - on_hand_quantity: sum of all immutable inventory ledger entries for the variant.
// - reserved_quantity: sum of all unreleased, non-expired inventory reservations.
// - available_quantity: salable inventory calculated as max(on_hand_quantity - reserved_quantity, 0).
func (r *InventoryRepository) GetStock(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) (*model.InventoryStock, error) {
	// Aggregate on-hand ledger transactions and active reservations concurrently for the specified variant.
	query := `
		SELECT
			pv.id AS variant_id,
			COALESCE((SELECT SUM(quantity) FROM inventory_ledgers WHERE variant_id = pv.id), 0)::int AS on_hand_quantity,
			COALESCE((SELECT SUM(quantity) FROM inventory_reservations WHERE variant_id = pv.id AND released_at IS NULL AND expires_at > now()), 0)::int AS reserved_quantity
		FROM product_variants pv
		WHERE pv.id = $1 AND pv.deleted_at IS NULL
	`

	stock := &model.InventoryStock{}
	err := qe.QueryRow(ctx, query, variantID).Scan(&stock.VariantID, &stock.OnHandQuantity, &stock.ReservedQuantity)
	if err != nil {
		return nil, fmt.Errorf("get inventory stock: %w", err)
	}

	// Calculate salable available stock; clamp to 0 to prevent negative available numbers from being displayed to storefronts.
	stock.AvailableQuantity = max(stock.OnHandQuantity-stock.ReservedQuantity, 0)
	stock.IsInStock = stock.AvailableQuantity > 0

	return stock, nil
}

// GetStockForVariants performs batch stock computation for multiple variants in a single query using CTEs,
// returning a map keyed by variant UUID. Joins with product_variants to filter out non-existent or soft-deleted variants.
func (r *InventoryRepository) GetStockForVariants(
	ctx context.Context,
	qe database.QueryExecutor,
	variantIDs []uuid.UUID,
) (map[uuid.UUID]*model.InventoryStock, error) {
	result := make(map[uuid.UUID]*model.InventoryStock, len(variantIDs))
	if len(variantIDs) == 0 {
		return result, nil
	}

	// Use CTEs to aggregate on_hand and reserved quantities for the requested variant IDs array.
	query := `
		WITH on_hand AS (
			SELECT variant_id, COALESCE(SUM(quantity), 0)::int AS on_hand_qty
			FROM inventory_ledgers
			WHERE variant_id = ANY($1)
			GROUP BY variant_id
		),
		reserved AS (
			SELECT variant_id, COALESCE(SUM(quantity), 0)::int AS reserved_qty
			FROM inventory_reservations
			WHERE variant_id = ANY($1) AND released_at IS NULL AND expires_at > now()
			GROUP BY variant_id
		)
		SELECT
			pv.id AS variant_id,
			COALESCE(oh.on_hand_qty, 0) AS on_hand_quantity,
			COALESCE(r.reserved_qty, 0) AS reserved_quantity
		FROM product_variants pv
		LEFT JOIN on_hand oh ON oh.variant_id = pv.id
		LEFT JOIN reserved r ON r.variant_id = pv.id
		WHERE pv.id = ANY($1) AND pv.deleted_at IS NULL
	`

	rows, err := qe.Query(ctx, query, variantIDs)
	if err != nil {
		return nil, fmt.Errorf("get stock for variants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		stock := &model.InventoryStock{}
		if err := rows.Scan(&stock.VariantID, &stock.OnHandQuantity, &stock.ReservedQuantity); err != nil {
			return nil, fmt.Errorf("scan stock for variants: %w", err)
		}
		// Compute net available stock clamped to 0.
		stock.AvailableQuantity = stock.OnHandQuantity - stock.ReservedQuantity
		stock.AvailableQuantity = max(0, stock.AvailableQuantity)
		stock.IsInStock = stock.AvailableQuantity > 0
		result[stock.VariantID] = stock
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stock for variants: %w", err)
	}

	return result, nil
}

// ListVariantLedgers returns a paginated audit log of inventory ledger entries for a specific variant,
// ordered with the most recent transaction first.
func (r *InventoryRepository) ListVariantLedgers(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.InventoryLedger], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	// Count total ledger entries for pagination metadata.
	countQuery := `SELECT COUNT(*) FROM inventory_ledgers WHERE variant_id = $1`
	var total int
	if err := qe.QueryRow(ctx, countQuery, variantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count inventory ledgers: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.InventoryLedger{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	// Fetch requested slice of ledger entries.
	selectQuery := `
		SELECT
			id,
			variant_id,
			quantity,
			reason,
			created_at
		FROM inventory_ledgers
		WHERE variant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := qe.Query(ctx, selectQuery, variantID, q.PageSize, q.Offset)
	if err != nil {
		return nil, fmt.Errorf("list inventory ledgers: %w", err)
	}

	items, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.InventoryLedger])
	if err != nil {
		return nil, fmt.Errorf("collect inventory ledgers: %w", err)
	}

	return pagination.NewPagedResult(items, pagination.NewPage(q.Page, q.PageSize, total)), nil
}

// CreateReservation inserts a new inventory reservation hold record into the database.
func (r *InventoryRepository) CreateReservation(
	ctx context.Context,
	qe database.QueryExecutor,
	res *model.InventoryReservation,
) error {
	query := `
		INSERT INTO inventory_reservations (
			id,
			variant_id,
			cart_id,
			quantity,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, now())
		RETURNING created_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		res.ID,
		res.VariantID,
		res.CartID,
		res.Quantity,
		res.ExpiresAt,
	).Scan(&res.CreatedAt)
	if err != nil {
		return fmt.Errorf("create inventory reservation: %w", err)
	}

	return nil
}

// GetReservationByID retrieves a reservation record by its UUID.
// GetReservationByID retrieves a reservation record by its UUID.
// It acquires an exclusive row lock (FOR UPDATE) when executed within a transaction to prevent race conditions during commit/release.
func (r *InventoryRepository) GetReservationByID(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) (*model.InventoryReservation, error) {
	query := `
		SELECT
			id,
			variant_id,
			cart_id,
			quantity,
			expires_at,
			created_at,
			released_at
		FROM inventory_reservations
		WHERE id = $1
		FOR UPDATE
		LIMIT 1
	`

	rows, err := qe.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("query reservation: %w", err)
	}

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.InventoryReservation])
	if err != nil {
		return nil, fmt.Errorf("scan reservation: %w", err)
	}

	return res, nil
}

// GetActiveCartReservation finds an existing active (unreleased, non-expired) reservation for a given cart and variant combination.
func (r *InventoryRepository) GetActiveCartReservation(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	variantID uuid.UUID,
) (*model.InventoryReservation, error) {
	// Find the most recent active reservation for this cart and SKU.
	query := `
		SELECT
			id,
			variant_id,
			cart_id,
			quantity,
			expires_at,
			created_at,
			released_at
		FROM inventory_reservations
		WHERE cart_id = $1 AND variant_id = $2 AND released_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1
	`

	rows, err := qe.Query(ctx, query, cartID, variantID)
	if err != nil {
		return nil, fmt.Errorf("query active cart reservation: %w", err)
	}

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[model.InventoryReservation])
	if err != nil {
		return nil, err
	}

	return res, nil
}

// UpdateReservationQuantityAndExpiry updates the quantity and extends the expiry timestamp for an unreleased reservation.
func (r *InventoryRepository) UpdateReservationQuantityAndExpiry(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
	quantity int,
	expiresAt time.Time,
) error {
	query := `
		UPDATE inventory_reservations
		SET quantity = $2, expires_at = $3
		WHERE id = $1 AND released_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, id, quantity, expiresAt)
	if err != nil {
		return fmt.Errorf("update reservation quantity and expiry: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ReleaseReservation marks a single reservation as released by setting released_at to now().
// Returns pgx.ErrNoRows if the reservation does not exist or has already been released.
func (r *InventoryRepository) ReleaseReservation(
	ctx context.Context,
	qe database.QueryExecutor,
	id uuid.UUID,
) error {
	query := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE id = $1 AND released_at IS NULL
	`

	cmd, err := qe.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("release reservation: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ReleaseCartReservations marks all unreleased reservations associated with a cart as released (e.g. upon cart clear).
func (r *InventoryRepository) ReleaseCartReservations(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
) error {
	query := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE cart_id = $1 AND released_at IS NULL
	`

	_, err := qe.Exec(ctx, query, cartID)
	if err != nil {
		return fmt.Errorf("release cart reservations: %w", err)
	}

	return nil
}

// ReleaseExpiredReservations marks all overdue reservations (expires_at <= now() with released_at IS NULL) as released.
// Returns the count of rows updated.
func (r *InventoryRepository) ReleaseExpiredReservations(
	ctx context.Context,
	qe database.QueryExecutor,
) (int64, error) {
	query := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE released_at IS NULL AND expires_at <= now()
	`

	cmd, err := qe.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("release expired reservations: %w", err)
	}

	return cmd.RowsAffected(), nil
}

type InventoryStockItem struct {
	VariantID         uuid.UUID `json:"variantId"`
	ProductID         uuid.UUID `json:"productId"`
	ProductName       string    `json:"productName"`
	SKU               string    `json:"sku"`
	VariantTitle      string    `json:"variantTitle"`
	OnHandQuantity    int       `json:"onHandQuantity"`
	ReservedQuantity  int       `json:"reservedQuantity"`
	AvailableQuantity int       `json:"availableQuantity"`
	IsInStock         bool      `json:"isInStock"`
	Price             int64     `json:"price"`
	Currency          string    `json:"currency"`
	Thumbnail         string    `json:"thumbnail"`
	CreatedAt         time.Time `json:"createdAt"`
}

type InventoryLedgerItem struct {
	ID           uuid.UUID `json:"id"`
	VariantID    uuid.UUID `json:"variantId"`
	ProductName  string    `json:"productName"`
	VariantTitle string    `json:"variantTitle"`
	SKU          string    `json:"sku"`
	Quantity     int       `json:"quantity"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (r *InventoryRepository) ListInventoryOverview(ctx context.Context, qe database.QueryExecutor, limit int, cursor *pagination.Keyset) ([]InventoryStockItem, *pagination.CursorPage, error) {
	query := `
		SELECT 
			pv.id,
			pv.product_id,
			p.title,
			pv.sku,
			pv.title,
			COALESCE((SELECT SUM(quantity) FROM inventory_ledgers WHERE variant_id = pv.id), 0)::int AS on_hand,
			COALESCE((SELECT SUM(quantity) FROM inventory_reservations WHERE variant_id = pv.id AND released_at IS NULL AND expires_at > NOW()), 0)::int AS reserved,
			pv.price,
			pv.currency,
			pv.created_at
		FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		WHERE pv.deleted_at IS NULL
	`
	args := []any{}
	i := 1

	keysetClause, keysetArgs := pagination.KeysetClause("pv", cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("pv")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, limit+1)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list inventory: %w", err)
	}
	defer rows.Close()

	var list []InventoryStockItem
	for rows.Next() {
		var item InventoryStockItem
		if err := rows.Scan(
			&item.VariantID,
			&item.ProductID,
			&item.ProductName,
			&item.SKU,
			&item.VariantTitle,
			&item.OnHandQuantity,
			&item.ReservedQuantity,
			&item.Price,
			&item.Currency,
			&item.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan inventory item: %w", err)
		}
		item.AvailableQuantity = item.OnHandQuantity - item.ReservedQuantity
		if item.AvailableQuantity < 0 {
			item.AvailableQuantity = 0
		}
		item.IsInStock = item.AvailableQuantity > 0
		list = append(list, item)
	}

	list, page := pagination.NewCursorPageFrom(list, limit, func(it InventoryStockItem) pagination.Keyset {
		return pagination.Keyset{CreatedAt: it.CreatedAt, ID: it.VariantID}
	})
	return list, page, nil
}

func (r *InventoryRepository) ListRecentLedgers(ctx context.Context, qe database.QueryExecutor, limit int, cursor *pagination.Keyset, variantID *uuid.UUID, productID *uuid.UUID) ([]InventoryLedgerItem, *pagination.CursorPage, error) {
	query := `
		SELECT il.id, il.variant_id, p.title, pv.title, pv.sku, il.quantity, il.reason, il.created_at
		FROM inventory_ledgers il
		JOIN product_variants pv ON il.variant_id = pv.id
		JOIN products p ON pv.product_id = p.id
		WHERE 1=1
	`
	args := []any{}
	i := 1

	if variantID != nil {
		query += fmt.Sprintf(" AND il.variant_id = $%d", i)
		args = append(args, *variantID)
		i++
	}

	if productID != nil {
		query += fmt.Sprintf(" AND p.id = $%d", i)
		args = append(args, *productID)
		i++
	}

	keysetClause, keysetArgs := pagination.KeysetClause("il", cursor, i)
	query += keysetClause
	args = append(args, keysetArgs...)
	i += len(keysetArgs)

	query += pagination.KeysetOrderBy("il")
	query += fmt.Sprintf(" LIMIT $%d", i)
	args = append(args, limit+1)

	rows, err := qe.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list recent ledgers: %w", err)
	}
	defer rows.Close()

	var list []InventoryLedgerItem
	for rows.Next() {
		var item InventoryLedgerItem
		if err := rows.Scan(
			&item.ID, &item.VariantID, &item.ProductName, &item.VariantTitle, &item.SKU, &item.Quantity, &item.Reason, &item.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("scan ledger: %w", err)
		}
		list = append(list, item)
	}

	list, page := pagination.NewCursorPageFrom(list, limit, func(it InventoryLedgerItem) pagination.Keyset {
		return pagination.Keyset{CreatedAt: it.CreatedAt, ID: it.ID}
	})
	return list, page, nil
}

