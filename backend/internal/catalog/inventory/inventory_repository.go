// Package inventory provides repository and domain service operations for managing product variant stock levels,
// append-only inventory ledgers, and temporary cart-based stock reservations.
package inventory

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

// ErrInsufficientStock is returned when a reservation cannot be committed as a
// sale because the physical on-hand stock (ledger sum) is lower than the held
// quantity. Consumers map it to their own API error type; the repository only
// deals in sentinel/driver errors.
var ErrInsufficientStock = errors.New("insufficient physical on-hand stock to commit reservation")

// CommitReservationsForOrder converts every unreleased reservation for an order
// into a finalized sale deduction on the caller's query executor. Each held
// unit becomes a negative 'sale' ledger entry and the hold is marked released.
// Holds whose TTL just lapsed but which the reaper has not yet released are
// still honoured (physical on-hand is re-checked beneath the variant row lock,
// so truly unavailable units fail closed with ErrInsufficientStock, which
// rolls the whole operation back). The operation is idempotent — a second call
// finds no unreleased holds and does nothing.
func (r *InventoryRepository) CommitReservationsForOrder(
	ctx context.Context,
	tx *database.Tx,
	orderID uuid.UUID,
) error {
	reservations, err := r.GetUnreleasedReservationsForOrder(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("fetch reservations for order %s: %w", orderID, err)
	}

	// Lock variants in a deterministic order so two concurrent commits that
	// share variants cannot deadlock (consistent lock ordering).
	sort.Slice(reservations, func(i, j int) bool {
		if reservations[i].VariantID != reservations[j].VariantID {
			return reservations[i].VariantID.String() < reservations[j].VariantID.String()
		}
		return reservations[i].ID.String() < reservations[j].ID.String()
	})

	for _, res := range reservations {
		// Serialize stock deductions per variant across concurrent commits.
		if err := r.LockVariantForUpdate(ctx, tx, res.VariantID); err != nil {
			return fmt.Errorf("lock variant %s for commit: %w", res.VariantID, err)
		}

		stock, err := r.GetStock(ctx, tx, res.VariantID)
		if err != nil {
			return fmt.Errorf("get stock for variant %s: %w", res.VariantID, err)
		}
		if stock.OnHandQuantity < res.Quantity {
			return fmt.Errorf("commit reservation %s (variant %s): %w", res.ID, res.VariantID, ErrInsufficientStock)
		}

		if err := r.ReleaseReservation(ctx, tx, res.ID); err != nil {
			return fmt.Errorf("release reservation %s: %w", res.ID, err)
		}

		ledger := &model.InventoryLedger{
			VariantID: res.VariantID,
			Quantity:  -res.Quantity,
			Reason:    model.InventoryReasonSale,
		}
		if err := r.CreateLedger(ctx, tx, ledger); err != nil {
			return fmt.Errorf("create sale ledger for variant %s: %w", res.VariantID, err)
		}
	}

	// Defensive: release anything still unreleased for this order (e.g. a hold
	// inserted concurrently) so it cannot linger and be committed twice.
	if _, err := r.ReleaseReservationsForOrder(ctx, tx, orderID); err != nil {
		return fmt.Errorf("release lingering reservations for order %s: %w", orderID, err)
	}

	return nil
}

// LockVariantForUpdate acquires an exclusive row-level write lock (SELECT ... FOR UPDATE) on the specified variant.
// This serializes concurrent inventory operations (such as stock adjustments and reservation creations) on the variant,
// preventing concurrent race conditions and preventing overselling.
func (r *InventoryRepository) LockVariantForUpdate(
	ctx context.Context,
	tx *database.Tx,
	variantID uuid.UUID,
) error {
	// Query product_variants table by ID with FOR UPDATE to block competing transactions on this SKU.
	query := `SELECT id FROM product_variants WHERE id = $1 FOR UPDATE`
	var id uuid.UUID
	err := tx.QueryRow(ctx, query, variantID).Scan(&id)
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
	tx *database.Tx,
	ledger *model.InventoryLedger,
) error {
	// Insert ledger entry with server-owned id and created_at timestamp.
	query := `
		INSERT INTO inventory_ledgers (
			variant_id,
			quantity,
			reason,
			created_at
		)
		VALUES ($1, $2, $3, now())
		RETURNING id, created_at
	`

	err := tx.QueryRow(
		ctx,
		query,
		ledger.VariantID,
		ledger.Quantity,
		ledger.Reason,
	).Scan(&ledger.ID, &ledger.CreatedAt)
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
	tx *database.Tx,
	res *model.InventoryReservation,
) error {
	query := `
		INSERT INTO inventory_reservations (
			id,
			variant_id,
			cart_id,
			order_id,
			quantity,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING created_at
	`

	err := tx.QueryRow(
		ctx,
		query,
		res.ID,
		res.VariantID,
		res.CartID,
		res.OrderID,
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
			order_id,
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
			order_id,
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
	tx *database.Tx,
	id uuid.UUID,
	quantity int,
	expiresAt time.Time,
) error {
	query := `
		UPDATE inventory_reservations
		SET quantity = $2, expires_at = $3
		WHERE id = $1 AND released_at IS NULL
	`

	cmd, err := tx.Exec(ctx, query, id, quantity, expiresAt)
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
	tx *database.Tx,
	id uuid.UUID,
) error {
	query := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE id = $1 AND released_at IS NULL
	`

	cmd, err := tx.Exec(ctx, query, id)
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
	tx *database.Tx,
	cartID uuid.UUID,
) error {
	query := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE cart_id = $1 AND released_at IS NULL
	`

	_, err := tx.Exec(ctx, query, cartID)
	if err != nil {
		return fmt.Errorf("release cart reservations: %w", err)
	}

	return nil
}

// ReleaseExpiredReservations marks all overdue reservations (expires_at <= now() with released_at IS NULL) as released.
// Returns the count of rows updated.
func (r *InventoryRepository) ReleaseExpiredReservations(
	ctx context.Context,
	tx *database.Tx,
) (int64, error) {
	query := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE released_at IS NULL AND expires_at <= now()
	`

	cmd, err := tx.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("release expired reservations: %w", err)
	}

	return cmd.RowsAffected(), nil
}

// GetUnreleasedReservationsForOrder returns every unreleased reservation
// belonging to an order, row-locking them FOR UPDATE so concurrent
// commit/release operations for the same order serialize.
//
// It deliberately does not filter on expires_at: a payment can be captured in
// the window between a hold lapsing and the reaper releasing it. The caller
// re-checks physical on-hand stock before committing, so an expired hold can
// still be honoured safely (or fail closed if the units are truly gone).
func (r *InventoryRepository) GetUnreleasedReservationsForOrder(
	ctx context.Context,
	qe database.QueryExecutor,
	orderID uuid.UUID,
) ([]*model.InventoryReservation, error) {
	query := `
		SELECT
			id,
			variant_id,
			cart_id,
			order_id,
			quantity,
			expires_at,
			created_at,
			released_at
		FROM inventory_reservations
		WHERE order_id = $1 AND released_at IS NULL
		ORDER BY created_at ASC
		FOR UPDATE
	`

	rows, err := qe.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("query unreleased order reservations: %w", err)
	}

	reservations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[model.InventoryReservation])
	if err != nil {
		return nil, fmt.Errorf("collect unreleased order reservations: %w", err)
	}

	return reservations, nil
}

// ReleaseReservationsForOrder marks every unreleased reservation for an order as
// released and returns the number of holds released. A zero count means the
// order had no live hold left (it was already committed or already released),
// which callers use to decide whether stock needs restoring.
func (r *InventoryRepository) ReleaseReservationsForOrder(
	ctx context.Context,
	tx *database.Tx,
	orderID uuid.UUID,
) (int64, error) {
	query := `
		UPDATE inventory_reservations
		SET released_at = now()
		WHERE order_id = $1 AND released_at IS NULL
	`

	cmd, err := tx.Exec(ctx, query, orderID)
	if err != nil {
		return 0, fmt.Errorf("release order reservations: %w", err)
	}

	return cmd.RowsAffected(), nil
}

// ExpiredReservationOrderIDs returns the distinct order IDs that still hold an
// expired, unreleased reservation. The maintenance reaper uses this to cancel
// stale pending orders whose checkout hold lapsed.
func (r *InventoryRepository) ExpiredReservationOrderIDs(
	ctx context.Context,
	qe database.QueryExecutor,
) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT order_id
		FROM inventory_reservations
		WHERE order_id IS NOT NULL AND released_at IS NULL AND expires_at <= now()
	`

	rows, err := qe.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query expired reservation orders: %w", err)
	}
	defer rows.Close()

	ids := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan expired reservation order: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired reservation orders: %w", err)
	}

	return ids, nil
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
	// Thumbnail is the resolved public URL of the variant's cover image. The
	// service fills it in from thumbBucket/thumbKey after the query runs.
	Thumbnail string    `json:"thumbnail"`
	CreatedAt time.Time `json:"createdAt"`

	// thumbBucket/thumbKey are the storage coordinates of the image chosen for
	// this row (variant thumbnail, else product thumbnail, else first variant
	// media). They are unexported so they never leak into the JSON response.
	thumbBucket string
	thumbKey    string
}

type InventoryLedgerItem struct {
	ID           int64     `json:"id"`
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
			pv.created_at,
			COALESCE(vso.bucket, pso.bucket, vmso.bucket) AS thumb_bucket,
			COALESCE(vso.object_key, pso.object_key, vmso.object_key) AS thumb_key
		FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		LEFT JOIN storage_objects vso ON pv.thumbnail_object_id = vso.id
		LEFT JOIN storage_objects pso ON p.thumbnail_object_id = pso.id
		LEFT JOIN LATERAL (
			SELECT so.bucket, so.object_key
			FROM variant_media vm
			JOIN storage_objects so ON vm.object_id = so.id
			WHERE vm.variant_id = pv.id
			ORDER BY vm.sort_order ASC
			LIMIT 1
		) vmso ON TRUE
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
		var thumbBucket, thumbKey *string
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
			&thumbBucket,
			&thumbKey,
		); err != nil {
			return nil, nil, fmt.Errorf("scan inventory item: %w", err)
		}
		if thumbBucket != nil && thumbKey != nil {
			item.thumbBucket = *thumbBucket
			item.thumbKey = *thumbKey
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
		return pagination.Keyset{CreatedAt: it.CreatedAt, IDInt: it.ID}
	})
	return list, page, nil
}

// GetPrimaryVariantIDByProductID returns the default variant UUID for a product,
// or the oldest created variant UUID if no default is designated.
func (r *InventoryRepository) GetPrimaryVariantIDByProductID(
	ctx context.Context,
	qe database.QueryExecutor,
	productID uuid.UUID,
) (uuid.UUID, error) {
	var variantID uuid.UUID
	query := `
		SELECT pv.id FROM product_variants pv
		JOIN products p ON pv.product_id = p.id
		WHERE pv.product_id = $1 AND pv.deleted_at IS NULL
		ORDER BY (CASE WHEN p.default_variant_id IS NOT NULL AND pv.id = p.default_variant_id THEN 0 ELSE 1 END) ASC, pv.created_at ASC
		LIMIT 1
	`
	err := qe.QueryRow(ctx, query, productID).Scan(&variantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get primary variant for product: %w", err)
	}
	return variantID, nil
}
