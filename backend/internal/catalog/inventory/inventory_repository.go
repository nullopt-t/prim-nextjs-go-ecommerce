package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type InventoryRepository struct{}

func NewRepository() *InventoryRepository {
	return &InventoryRepository{}
}

func (r *InventoryRepository) LockVariantForUpdate(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) error {
	query := `SELECT id FROM product_variants WHERE id = $1 FOR UPDATE`
	var id uuid.UUID
	err := qe.QueryRow(ctx, query, variantID).Scan(&id)
	if err != nil {
		return fmt.Errorf("lock variant for update: %w", err)
	}
	return nil
}

func (r *InventoryRepository) CreateLedger(
	ctx context.Context,
	qe database.QueryExecutor,
	ledger *model.InventoryLedger,
) error {
	query := `
		INSERT INTO inventory_ledgers (
			id,
			variant_id,
			quantity,
			reason,
			reference_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, now())
		RETURNING created_at
	`

	err := qe.QueryRow(
		ctx,
		query,
		ledger.ID,
		ledger.VariantID,
		ledger.Quantity,
		ledger.Reason,
		ledger.ReferenceID,
	).Scan(&ledger.CreatedAt)
	if err != nil {
		return fmt.Errorf("create inventory ledger: %w", err)
	}

	return nil
}

func (r *InventoryRepository) GetStock(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
) (*model.InventoryStock, error) {
	query := `
		SELECT
			$1::uuid AS variant_id,
			COALESCE((SELECT SUM(quantity) FROM inventory_ledgers WHERE variant_id = $1), 0)::int AS on_hand_quantity,
			COALESCE((SELECT SUM(quantity) FROM inventory_reservations WHERE variant_id = $1 AND released_at IS NULL AND expires_at > now()), 0)::int AS reserved_quantity
	`

	stock := &model.InventoryStock{VariantID: variantID}
	err := qe.QueryRow(ctx, query, variantID).Scan(&stock.VariantID, &stock.OnHandQuantity, &stock.ReservedQuantity)
	if err != nil {
		return nil, fmt.Errorf("get inventory stock: %w", err)
	}

	stock.AvailableQuantity = max(stock.OnHandQuantity - stock.ReservedQuantity, 0)
	stock.IsInStock = stock.AvailableQuantity > 0

	return stock, nil
}

func (r *InventoryRepository) GetStockForVariants(
	ctx context.Context,
	qe database.QueryExecutor,
	variantIDs []uuid.UUID,
) (map[uuid.UUID]*model.InventoryStock, error) {
	result := make(map[uuid.UUID]*model.InventoryStock, len(variantIDs))
	if len(variantIDs) == 0 {
		return result, nil
	}

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
			v.id AS variant_id,
			COALESCE(oh.on_hand_qty, 0) AS on_hand_quantity,
			COALESCE(r.reserved_qty, 0) AS reserved_quantity
		FROM unnest($1::uuid[]) AS v(id)
		LEFT JOIN on_hand oh ON oh.variant_id = v.id
		LEFT JOIN reserved r ON r.variant_id = v.id
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
		stock.AvailableQuantity = stock.OnHandQuantity - stock.ReservedQuantity
		if stock.AvailableQuantity < 0 {
			stock.AvailableQuantity = 0
		}
		stock.IsInStock = stock.AvailableQuantity > 0
		result[stock.VariantID] = stock
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stock for variants: %w", err)
	}

	return result, nil
}

func (r *InventoryRepository) ListLedgers(
	ctx context.Context,
	qe database.QueryExecutor,
	variantID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.InventoryLedger], error) {
	if q == nil {
		q = &pagination.ListQuery{}
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 20, MaxPageSize: 100})

	countQuery := `SELECT COUNT(*) FROM inventory_ledgers WHERE variant_id = $1`
	var total int
	if err := qe.QueryRow(ctx, countQuery, variantID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count inventory ledgers: %w", err)
	}

	if total == 0 {
		return pagination.NewPagedResult([]*model.InventoryLedger{}, pagination.NewPage(q.Page, q.PageSize, 0)), nil
	}

	selectQuery := `
		SELECT
			id,
			variant_id,
			quantity,
			reason,
			reference_id,
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

func (r *InventoryRepository) GetActiveCartReservation(
	ctx context.Context,
	qe database.QueryExecutor,
	cartID uuid.UUID,
	variantID uuid.UUID,
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
