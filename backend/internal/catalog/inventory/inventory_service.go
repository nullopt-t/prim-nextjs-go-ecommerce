// Package inventory provides domain operations for modifying variant inventory stock,
// ledger balance computation, holding temporary checkout stock reservations, and releasing expired holds.
package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// InventoryService provides domain logic and transaction orchestration for inventory ledgers,
// stock reservations, and inventory balance calculations.
type InventoryService struct {
	logger          log.Logger
	dr              database.Runner
	repo            *InventoryRepository
	storageProvider storage.StorageProvider
}

// NewService constructs a new InventoryService instance with the given logger,
// database runner, repository, and storage provider.
func NewService(
	logger log.Logger,
	dr database.Runner,
	repo *InventoryRepository,
	storageProvider storage.StorageProvider,
) *InventoryService {
	return &InventoryService{
		logger:          logger,
		dr:              dr,
		repo:            repo,
		storageProvider: storageProvider,
	}
}

// handleError maps persistence and database errors to domain APIErrors.
func (s *InventoryService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("Inventory resource not found").
			WithCode(errcode.CodeVariantNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced entity does not exist").
			WithCode(errcode.CodeVariantNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Inventory conflict detected").
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// AdjustStockInput defines the parameter payload for manually or programmatically adjusting stock levels.
type AdjustStockInput struct {
	// VariantID is the unique identifier of the product variant being adjusted.
	VariantID uuid.UUID
	// Quantity is the signed integer delta (+ for additions, - for reductions). Cannot be zero.
	Quantity int
	// Reason describes the business rationale (restock, sale, return, adjustment, reservation_release).
	Reason string
}

// ReserveStockInput defines the parameters for reserving stock for an active cart or checkout session.
type ReserveStockInput struct {
	// VariantID is the unique identifier of the product variant to reserve.
	VariantID uuid.UUID
	// CartID optionally associates this reservation with a customer cart.
	CartID *uuid.UUID
	// OrderID optionally associates this reservation with the order that placed
	// the hold. When set, the hold is keyed to the order (not upserted against
	// an existing cart hold) so it can be committed or released deterministically.
	OrderID *uuid.UUID
	// Quantity is the positive number of units to reserve.
	Quantity int
	// Duration specifies how long the reservation remains valid before auto-expiration (defaults to 15m).
	Duration time.Duration
}

// AdjustStock records an inventory adjustment ledger entry within a database transaction.
// It acquires an exclusive row lock on the variant to prevent race conditions and validates
// that negative adjustments do not cause on-hand quantity to drop below reserved quantity or zero.
func (s *InventoryService) AdjustStock(
	ctx context.Context,
	in AdjustStockInput,
) (*model.InventoryStock, error) {
	var updatedStock *model.InventoryStock

	txErr := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		var err error
		updatedStock, err = s.applyAdjustment(ctx, tx, in)
		return err
	})

	if txErr != nil {
		var apiErr *apierr.APIError
		if errors.As(txErr, &apiErr) {
			return nil, apiErr
		}
		mappedErr := database.MapError(txErr)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation), errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.BadRequestError("Referenced variant does not exist").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(txErr)
		default:
			return nil, s.handleError(txErr, "Failed to adjust inventory")
		}
	}

	return updatedStock, nil
}

// ApplyStockAdjustment applies a stock adjustment inside the caller's transaction.
//
// This is the participating form of AdjustStock: the caller (order cancellation,
// refund restock, variant bootstrap) already owns the unit of work, so this
// method must not begin another transaction. It joins the executor it is handed,
// so the ledger append, the row lock and the order status change either commit
// together or roll back together.
func (s *InventoryService) ApplyStockAdjustment(
	ctx context.Context,
	tx *database.Tx,
	in AdjustStockInput,
) error {
	if tx == nil {
		return apierr.InternalServerError("Stock adjustment requires a transaction").
			WithCode(apierr.CodeInternalError)
	}

	if _, err := s.applyAdjustment(ctx, tx, in); err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation), errors.Is(mappedErr, database.ErrNotFound):
			return apierr.BadRequestError("Referenced variant does not exist").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		default:
			return s.handleError(err, "Failed to adjust inventory")
		}
	}

	return nil
}

// validateAdjustStockInput checks shared preconditions for stock adjustments and
// returns the parsed inventory reason.
func validateAdjustStockInput(in AdjustStockInput) (model.InventoryReason, error) {
	if in.VariantID == uuid.Nil {
		return "", apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if in.Quantity == 0 {
		return "", apierr.BadRequestError("Quantity adjustment cannot be zero").
			WithCode(apierr.CodeInvalidInput)
	}

	reason, err := model.ParseInventoryReason(in.Reason)
	if err != nil {
		return "", apierr.BadRequestError("Invalid inventory reason: " + in.Reason).
			WithCode(errcode.CodeInvalidInventoryReason).
			Wrap(err)
	}

	return reason, nil
}

// applyAdjustment executes a stock adjustment inside an already-begun transaction.
// It row-locks the variant, rejects negative adjustments that would drop on-hand
// stock below active reservations, and appends a ledger entry.
func (s *InventoryService) applyAdjustment(
	ctx context.Context,
	tx *database.Tx,
	in AdjustStockInput,
) (*model.InventoryStock, error) {
	reason, err := validateAdjustStockInput(in)
	if err != nil {
		return nil, err
	}

	if err := s.repo.LockVariantForUpdate(ctx, tx, in.VariantID); err != nil {
		return nil, err
	}

	currentStock, err := s.repo.GetStock(ctx, tx, in.VariantID)
	if err != nil {
		return nil, err
	}

	if in.Quantity < 0 && (currentStock.OnHandQuantity+in.Quantity < currentStock.ReservedQuantity) {
		return nil, apierr.BadRequestError(
			fmt.Sprintf("Insufficient available inventory: negative adjustment (%d) drops on-hand stock (%d) below active reservations (%d)",
				in.Quantity, currentStock.OnHandQuantity, currentStock.ReservedQuantity),
		).WithCode(errcode.CodeInsufficientInventory)
	}

	ledger := &model.InventoryLedger{
		VariantID: in.VariantID,
		Quantity:  in.Quantity,
		Reason:    reason,
	}

	if err := s.repo.CreateLedger(ctx, tx, ledger); err != nil {
		return nil, err
	}

	return s.repo.GetStock(ctx, tx, in.VariantID)
}

// GetStockByVariantID calculates and returns current stock levels (on-hand, reserved, available) for a variant.
func (s *InventoryService) GetStockByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.InventoryStock, error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var stock *model.InventoryStock
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var repoErr error
		stock, repoErr = s.repo.GetStock(ctx, exec, variantID)
		return repoErr
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		mappedErr := database.MapError(err)
		if errors.Is(mappedErr, database.ErrNotFound) {
			return nil, apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		}
		return nil, s.handleError(err, "Failed to get variant inventory stock")
	}

	return stock, nil
}

// GetStockForVariants returns stock levels for many variants in one query.
func (s *InventoryService) GetStockForVariants(
	ctx context.Context,
	variantIDs []uuid.UUID,
) (map[uuid.UUID]*model.InventoryStock, error) {
	if len(variantIDs) == 0 {
		return make(map[uuid.UUID]*model.InventoryStock), nil
	}

	var stocks map[uuid.UUID]*model.InventoryStock
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var repoErr error
		stocks, repoErr = s.repo.GetStockForVariants(ctx, exec, variantIDs)
		return repoErr
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to get inventory stock for variants")
	}

	return stocks, nil
}

// ListVariantLedgers returns a paginated list of inventory ledger entries for a variant.
func (s *InventoryService) ListVariantLedgers(
	ctx context.Context,
	variantID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.InventoryLedger], error) {
	if variantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var result *pagination.PagedResult[model.InventoryLedger]
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var repoErr error
		result, repoErr = s.repo.ListVariantLedgers(ctx, exec, variantID, q)
		return repoErr
	})

	if err != nil {
		return nil, s.handleError(err, "Failed to list inventory ledgers")
	}

	return result, nil
}

// ListLedgers is an alias for ListVariantLedgers.
func (s *InventoryService) ListLedgers(
	ctx context.Context,
	variantID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.InventoryLedger], error) {
	return s.ListVariantLedgers(ctx, variantID, q)
}

// ReserveStock places a temporary hold on variant stock for customer checkout.
// If an active reservation already exists for the given cart and variant, the existing reservation
// is extended in duration and quantity.
func (s *InventoryService) ReserveStock(
	ctx context.Context,
	in ReserveStockInput,
) (*model.InventoryReservation, error) {
	if in.VariantID == uuid.Nil {
		return nil, apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if in.Quantity <= 0 {
		return nil, apierr.BadRequestError("Reservation quantity must be positive").
			WithCode(apierr.CodeInvalidInput)
	}

	ttl := in.Duration
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	var resultReservation *model.InventoryReservation

	err := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		var err error
		resultReservation, err = s.placeReservation(ctx, tx, in, ttl)
		return err
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation), errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.BadRequestError("Referenced variant does not exist").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		default:
			return nil, s.handleError(err, "Failed to reserve inventory")
		}
	}

	return resultReservation, nil
}

// HoldStock places a temporary stock hold within the caller's database
// transaction, allowing checkout to reserve, order, and commit/release stock
// atomically instead of across independent transactions.
func (s *InventoryService) HoldStock(
	ctx context.Context,
	tx *database.Tx,
	in ReserveStockInput,
) (*model.InventoryReservation, error) {
	if err := validateReserveStockInput(in); err != nil {
		return nil, err
	}

	ttl := in.Duration
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	return s.placeReservation(ctx, tx, in, ttl)
}

// validateReserveStockInput checks shared preconditions for stock reservations.
func validateReserveStockInput(in ReserveStockInput) error {
	if in.VariantID == uuid.Nil {
		return apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if in.Quantity <= 0 {
		return apierr.BadRequestError("Reservation quantity must be positive").
			WithCode(apierr.CodeInvalidInput)
	}

	return nil
}

// placeReservation executes a stock reservation inside an already-begun transaction.
// It row-locks the variant, extends an existing active cart reservation when present,
// and otherwise inserts a new reservation after checking salable stock.
func (s *InventoryService) placeReservation(
	ctx context.Context,
	tx *database.Tx,
	in ReserveStockInput,
	ttl time.Duration,
) (*model.InventoryReservation, error) {
	// Row-level lock to prevent concurrency overselling
	if err := s.repo.LockVariantForUpdate(ctx, tx, in.VariantID); err != nil {
		return nil, err
	}

	stock, err := s.repo.GetStock(ctx, tx, in.VariantID)
	if err != nil {
		return nil, err
	}

	// Check if active cart reservation already exists to extend / upsert.
	// This path is only for plain cart holds (no order yet); an order-linked
	// hold is always a fresh row so its lifecycle stays deterministic.
	if in.CartID != nil && in.OrderID == nil {
		existing, err := s.repo.GetActiveCartReservation(ctx, tx, *in.CartID, in.VariantID)
		if err == nil && existing != nil {
			// stock.AvailableQuantity already subtracts existing.Quantity; account for existing hold
			effectiveAvailable := stock.AvailableQuantity + existing.Quantity
			newQty := existing.Quantity + in.Quantity
			if effectiveAvailable < newQty {
				return nil, apierr.BadRequestError(
					fmt.Sprintf("Insufficient available stock (%d) for reservation (%d)", effectiveAvailable, newQty),
				).WithCode(errcode.CodeInsufficientInventory)
			}

			newExpiry := time.Now().Add(ttl)
			if err := s.repo.UpdateReservationQuantityAndExpiry(ctx, tx, existing.ID, newQty, newExpiry); err != nil {
				return nil, err
			}
			existing.Quantity = newQty
			existing.ExpiresAt = newExpiry
			return existing, nil
		}
	}

	if stock.AvailableQuantity < in.Quantity {
		return nil, apierr.BadRequestError(
			fmt.Sprintf("Insufficient available stock (%d) for reservation (%d)", stock.AvailableQuantity, in.Quantity),
		).WithCode(errcode.CodeInsufficientInventory)
	}

	newRes := &model.InventoryReservation{
		ID:        uuid.New(),
		VariantID: in.VariantID,
		CartID:    in.CartID,
		OrderID:   in.OrderID,
		Quantity:  in.Quantity,
		ExpiresAt: time.Now().Add(ttl),
	}

	if err := s.repo.CreateReservation(ctx, tx, newRes); err != nil {
		return nil, err
	}

	return newRes, nil
}

// ReleaseReservation cancels a specific inventory reservation hold, immediately returning units to available stock.
func (s *InventoryService) ReleaseReservation(
	ctx context.Context,
	reservationID uuid.UUID,
) error {
	if reservationID == uuid.Nil {
		return apierr.BadRequestError("Reservation ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
		return s.repo.ReleaseReservation(ctx, exec, reservationID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Inventory reservation not found").
				WithCode(errcode.CodeReservationNotFound).
				Wrap(err)
		default:
			return s.handleError(err, "Failed to release reservation")
		}
	}

	return nil
}

// ReleaseCartReservations releases all active inventory reservation holds for a specific cart UUID.
func (s *InventoryService) ReleaseCartReservations(
	ctx context.Context,
	cartID uuid.UUID,
) error {
	if cartID == uuid.Nil {
		return apierr.BadRequestError("Cart ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
		return s.repo.ReleaseCartReservations(ctx, exec, cartID)
	})

	if err != nil {
		return s.handleError(err, "Failed to release cart reservations")
	}

	return nil
}

// ReleaseExpiredReservations batch releases all reservations where expires_at has passed.
// Typically triggered by a background cron worker or scheduler.
func (s *InventoryService) ReleaseExpiredReservations(
	ctx context.Context,
) (int64, error) {
	var count int64
	err := s.dr.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
		var repoErr error
		count, repoErr = s.repo.ReleaseExpiredReservations(ctx, exec)
		return repoErr
	})

	if err != nil {
		return 0, s.handleError(err, "Failed to release expired reservations")
	}

	return count, nil
}

// CommitReservation transitions a reservation hold into a finalized sale deduction.
// It acquires row locks on the reservation and variant to serialize concurrent commits and releases,
// and validates on-hand inventory before writing a negative sale ledger entry.
func (s *InventoryService) CommitReservation(
	ctx context.Context,
	reservationID uuid.UUID,
) error {
	if reservationID == uuid.Nil {
		return apierr.BadRequestError("Reservation ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		res, err := s.repo.GetReservationByID(ctx, tx, reservationID)
		if err != nil {
			return err
		}

		if res.ReleasedAt != nil {
			return apierr.BadRequestError("Reservation has already been released").
				WithCode(errcode.CodeReservationExpired)
		}

		if time.Now().After(res.ExpiresAt) {
			return apierr.BadRequestError("Reservation has expired").
				WithCode(errcode.CodeReservationExpired)
		}

		// Acquire row-level lock on the variant to serialize stock deductions
		if err := s.repo.LockVariantForUpdate(ctx, tx, res.VariantID); err != nil {
			return err
		}

		// Verify sufficient physical on-hand stock is available before writing a negative sale ledger entry
		stock, err := s.repo.GetStock(ctx, tx, res.VariantID)
		if err != nil {
			return err
		}
		if stock.OnHandQuantity < res.Quantity {
			return apierr.BadRequestError(
				fmt.Sprintf("Insufficient physical on-hand stock (%d) to commit reservation (%d)", stock.OnHandQuantity, res.Quantity),
			).WithCode(errcode.CodeInsufficientInventory)
		}

		if err := s.repo.ReleaseReservation(ctx, tx, reservationID); err != nil {
			return err
		}

		ledger := &model.InventoryLedger{
			VariantID: res.VariantID,
			Quantity:  -res.Quantity,
			Reason:    model.InventoryReasonSale,
		}

		return s.repo.CreateLedger(ctx, tx, ledger)
	})

	if err != nil {
		return s.handleError(err, "Failed to commit reservation")
	}

	return nil
}

// CommitReservationsForOrder converts every unreleased reservation for an
// order into a finalized sale deduction within the caller's transaction. It is
// called when payment is captured: each held unit becomes a negative 'sale'
// ledger entry and the hold is marked released. Holds whose TTL just lapsed but
// which the reaper has not yet released are still honoured (physical on-hand is
// re-checked, so truly unavailable units fail closed). The operation is
// idempotent — a second call finds no unreleased holds and does nothing.
func (s *InventoryService) CommitReservationsForOrder(
	ctx context.Context,
	tx *database.Tx,
	orderID uuid.UUID,
) error {
	if orderID == uuid.Nil {
		return apierr.BadRequestError("Order ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if err := s.repo.CommitReservationsForOrder(ctx, tx, orderID); err != nil {
		if errors.Is(err, ErrInsufficientStock) {
			return apierr.BadRequestError(
				fmt.Sprintf("Insufficient physical on-hand stock to commit reservation (%s)", orderID),
			).WithCode(errcode.CodeInsufficientInventory)
		}
		return s.handleError(err, "Failed to commit reservation")
	}

	return nil
}

// ReleaseReservationsForOrder releases every unreleased hold for an order
// and reports how many were released. A zero count means the order had no live
// hold left — i.e. its stock was already committed as a sale — which callers
// use to decide whether stock needs restoring on cancellation.
func (s *InventoryService) ReleaseReservationsForOrder(
	ctx context.Context,
	tx *database.Tx,
	orderID uuid.UUID,
) (int64, error) {
	if orderID == uuid.Nil {
		return 0, apierr.BadRequestError("Order ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	count, err := s.repo.ReleaseReservationsForOrder(ctx, tx, orderID)
	if err != nil {
		return 0, s.handleError(err, "Failed to release reservations for order")
	}
	return count, nil
}

// ExpiredReservationOrderIDs lists the distinct orders whose checkout hold has
// expired without being committed or released. The reservation reaper consumes
// this to cancel stale pending orders and return their held stock to salable.
func (s *InventoryService) ExpiredReservationOrderIDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var e error
		ids, e = s.repo.ExpiredReservationOrderIDs(ctx, exec)
		return e
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to list expired reservation orders")
	}
	return ids, nil
}

// CheckAvailability checks whether a variant has at least quantity units of salable stock available.
func (s *InventoryService) CheckAvailability(
	ctx context.Context,
	variantID uuid.UUID,
	quantity int,
) (bool, *model.InventoryStock, error) {
	stock, err := s.GetStockByVariantID(ctx, variantID)
	if err != nil {
		return false, nil, err
	}

	return stock.AvailableQuantity >= quantity, stock, nil
}

// ListInventoryOverview retrieves an aggregated overview of product variants and stock levels with cursor pagination.
func (s *InventoryService) ListInventoryOverview(ctx context.Context, limit int, cursor *pagination.Keyset) ([]InventoryStockItem, *pagination.CursorPage, error) {
	var items []InventoryStockItem
	var page *pagination.CursorPage
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		items, page, err = s.repo.ListInventoryOverview(ctx, exec, limit, cursor)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list inventory overview")
	}

	// Build public image URLs outside the DB callback: URL construction is pure
	// and must not hold a pooled connection.
	if s.storageProvider != nil {
		for i := range items {
			if items[i].thumbBucket != "" && items[i].thumbKey != "" {
				items[i].Thumbnail = s.storageProvider.GetPublicURL(items[i].thumbBucket, items[i].thumbKey)
			}
		}
	}

	return items, page, nil
}

// ListRecentLedgers retrieves recent inventory ledger audit logs with cursor pagination and optional filtering.
func (s *InventoryService) ListRecentLedgers(ctx context.Context, limit int, cursor *pagination.Keyset, variantID *uuid.UUID, productID *uuid.UUID) ([]InventoryLedgerItem, *pagination.CursorPage, error) {
	var list []InventoryLedgerItem
	var page *pagination.CursorPage
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		list, page, err = s.repo.ListRecentLedgers(ctx, exec, limit, cursor, variantID, productID)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list recent ledgers")
	}
	return list, page, nil
}

// AdjustInventory writes an inventory ledger entry for a specific variant.
func (s *InventoryService) AdjustInventory(ctx context.Context, variantID uuid.UUID, delta int, reason string) error {
	if variantID == uuid.Nil {
		return apierr.BadRequestError("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}
	if delta == 0 {
		return apierr.BadRequestError("Quantity adjustment cannot be zero").
			WithCode(apierr.CodeInvalidInput)
	}
	parsedReason, err := model.ParseInventoryReason(reason)
	if err != nil {
		return apierr.BadRequestError("Invalid inventory reason: " + reason).
			WithCode(errcode.CodeInvalidInventoryReason).
			Wrap(err)
	}

	err = s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		ledger := &model.InventoryLedger{
			VariantID: variantID,
			Quantity:  delta,
			Reason:    parsedReason,
		}
		return s.repo.CreateLedger(ctx, tx, ledger)
	})
	if err != nil {
		return s.handleError(err, "Failed to adjust inventory")
	}
	return nil
}

// AdjustProductInventory writes an inventory ledger entry for a product's primary variant.
func (s *InventoryService) AdjustProductInventory(ctx context.Context, productID uuid.UUID, delta int, reason string) error {
	if productID == uuid.Nil {
		return apierr.BadRequestError("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}
	if delta == 0 {
		return apierr.BadRequestError("Quantity adjustment cannot be zero").
			WithCode(apierr.CodeInvalidInput)
	}
	parsedReason, err := model.ParseInventoryReason(reason)
	if err != nil {
		return apierr.BadRequestError("Invalid inventory reason: " + reason).
			WithCode(errcode.CodeInvalidInventoryReason).
			Wrap(err)
	}

	err = s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		variantID, err := s.repo.GetPrimaryVariantIDByProductID(ctx, tx, productID)
		if err != nil {
			return err
		}

		ledger := &model.InventoryLedger{
			VariantID: variantID,
			Quantity:  delta,
			Reason:    parsedReason,
		}
		return s.repo.CreateLedger(ctx, tx, ledger)
	})
	if err != nil {
		return s.handleError(err, "Failed to adjust product inventory")
	}
	return nil
}
