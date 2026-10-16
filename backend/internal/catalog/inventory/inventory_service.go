// Package inventory provides domain operations for modifying variant inventory stock,
// ledger balance computation, holding temporary checkout stock reservations, and releasing expired holds.
package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/errcode"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// InventoryService provides domain logic and transaction orchestration for inventory ledgers,
// stock reservations, and inventory balance calculations.
type InventoryService struct {
	logger log.Logger
	dr     database.Runner
	repo   *InventoryRepository
}

// NewService constructs a new InventoryService instance with the given logger, database runner, and repository.
func NewService(
	logger log.Logger,
	dr database.Runner,
	repo *InventoryRepository,
) *InventoryService {
	return &InventoryService{
		logger: logger,
		dr:     dr,
		repo:   repo,
	}
}

// AdjustStockInput defines the parameter payload for manually or programmatically adjusting stock levels.
type AdjustStockInput struct {
	// VariantID is the unique identifier of the product variant being adjusted.
	VariantID uuid.UUID
	// Quantity is the signed integer delta (+ for additions, - for reductions). Cannot be zero.
	Quantity  int
	// Reason describes the business rationale (restock, sale, return, adjustment, reservation_release).
	Reason    string
}

// ReserveStockInput defines the parameters for reserving stock for an active cart or checkout session.
type ReserveStockInput struct {
	// VariantID is the unique identifier of the product variant to reserve.
	VariantID uuid.UUID
	// CartID optionally associates this reservation with a customer cart.
	CartID    *uuid.UUID
	// Quantity is the positive number of units to reserve.
	Quantity  int
	// Duration specifies how long the reservation remains valid before auto-expiration (defaults to 15m).
	Duration  time.Duration
}

// AdjustStock records an inventory adjustment ledger entry within a database transaction.
// It acquires an exclusive row lock on the variant to prevent race conditions and validates
// that negative adjustments do not cause on-hand quantity to drop below reserved quantity or zero.
func (s *InventoryService) AdjustStock(
	ctx context.Context,
	in AdjustStockInput,
) (*model.InventoryStock, error) {
	var updatedStock *model.InventoryStock

	txErr := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		var err error
		updatedStock, err = s.adjustStockInTx(ctx, tx, in)
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
			return nil, apierr.InternalServerError("Failed to adjust inventory").
				WithCode(apierr.CodeInternalError).
				Wrap(txErr).
				WithStack()
		}
	}

	return updatedStock, nil
}

// AdjustStockWithTx adjusts variant stock within the caller's database transaction,
// so stock movements can participate in wider atomic units of work (e.g. checkout
// or order cancellation) instead of opening their own separate transaction.
func (s *InventoryService) AdjustStockWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	in AdjustStockInput,
) (*model.InventoryStock, error) {
	if _, err := validateAdjustStockInput(in); err != nil {
		return nil, err
	}
	return s.adjustStockInTx(ctx, tx, in)
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

// adjustStockInTx executes a stock adjustment inside an already-begun transaction.
// It row-locks the variant, rejects negative adjustments that would drop on-hand
// stock below active reservations, and appends a ledger entry.
func (s *InventoryService) adjustStockInTx(
	ctx context.Context,
	tx database.QueryExecutor,
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
		ID:        uuid.New(),
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		stock, repoErr = s.repo.GetStock(ctx, db, variantID)
		return repoErr
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		default:
			return nil, apierr.InternalServerError("Failed to get variant inventory stock").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return stock, nil
}

// GetStockForVariants batch calculates stock levels for multiple variant UUIDs.
func (s *InventoryService) GetStockForVariants(
	ctx context.Context,
	variantIDs []uuid.UUID,
) (map[uuid.UUID]*model.InventoryStock, error) {
	if len(variantIDs) == 0 {
		return make(map[uuid.UUID]*model.InventoryStock), nil
	}

	var stocks map[uuid.UUID]*model.InventoryStock
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		stocks, repoErr = s.repo.GetStockForVariants(ctx, db, variantIDs)
		return repoErr
	})

	if err != nil {
		return nil, apierr.InternalServerError("Failed to get inventory stock for variants").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
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
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = s.repo.ListVariantLedgers(ctx, db, variantID, q)
		return repoErr
	})

	if err != nil {
		return nil, apierr.InternalServerError("Failed to list inventory ledgers").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
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

	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		var err error
		resultReservation, err = s.reserveStockInTx(ctx, tx, in, ttl)
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
			return nil, apierr.InternalServerError("Failed to reserve inventory").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return resultReservation, nil
}

// ReserveStockWithTx places a temporary stock hold within the caller's database
// transaction, allowing checkout to reserve, order, and commit/release stock
// atomically instead of across independent transactions.
func (s *InventoryService) ReserveStockWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	in ReserveStockInput,
) (*model.InventoryReservation, error) {
	if err := validateReserveStockInput(in); err != nil {
		return nil, err
	}

	ttl := in.Duration
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	return s.reserveStockInTx(ctx, tx, in, ttl)
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

// reserveStockInTx executes a stock reservation inside an already-begun transaction.
// It row-locks the variant, extends an existing active cart reservation when present,
// and otherwise inserts a new reservation after checking salable stock.
func (s *InventoryService) reserveStockInTx(
	ctx context.Context,
	tx database.QueryExecutor,
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

	// Check if active cart reservation already exists to extend / upsert
	if in.CartID != nil {
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

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.ReleaseReservation(ctx, db, reservationID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Inventory reservation not found").
				WithCode(errcode.CodeReservationNotFound).
				Wrap(err)
		default:
			return apierr.InternalServerError("Failed to release reservation").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
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

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.ReleaseCartReservations(ctx, db, cartID)
	})

	if err != nil {
		return apierr.InternalServerError("Failed to release cart reservations").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}

// ReleaseExpiredReservations batch releases all reservations where expires_at has passed.
// Typically triggered by a background cron worker or scheduler.
func (s *InventoryService) ReleaseExpiredReservations(
	ctx context.Context,
) (int64, error) {
	var count int64
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		count, err = s.repo.ReleaseExpiredReservations(ctx, db)
		return err
	})

	if err != nil {
		return 0, apierr.InternalServerError("Failed to release expired reservations").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
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

	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
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
			ID:        uuid.New(),
			VariantID: res.VariantID,
			Quantity:  -res.Quantity,
			Reason:    model.InventoryReasonSale,
		}

		return s.repo.CreateLedger(ctx, tx, ledger)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.InternalServerError("Failed to commit reservation").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
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

func (s *InventoryService) ListInventoryOverview(ctx context.Context, limit, offset int) ([]InventoryStockItem, int, error) {
	var items []InventoryStockItem
	var total int
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		items, total, err = s.repo.ListInventoryOverview(ctx, db, limit, offset)
		return err
	})
	return items, total, err
}

func (s *InventoryService) ListRecentLedgers(ctx context.Context, limit int, variantID *uuid.UUID, productID *uuid.UUID) ([]InventoryLedgerItem, error) {
	var list []InventoryLedgerItem
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		list, err = s.repo.ListRecentLedgers(ctx, db, limit, variantID, productID)
		return err
	})
	return list, err
}

func (s *InventoryService) AdjustInventory(ctx context.Context, variantID uuid.UUID, delta int, reason string) error {
	return s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		ledger := &model.InventoryLedger{
			ID:        uuid.New(),
			VariantID: variantID,
			Quantity:  delta,
			Reason:    model.InventoryReason(reason),
			CreatedAt: time.Now(),
		}
		query := `
			INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at)
			VALUES ($1, $2, $3, $4::inventory_reason, NOW())
		`
		_, err := tx.Exec(ctx, query, ledger.ID, ledger.VariantID, ledger.Quantity, ledger.Reason)
		return err
	})
}

func (s *InventoryService) AdjustProductInventory(ctx context.Context, productID uuid.UUID, delta int, reason string) error {
	return s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		var variantID uuid.UUID
		err := tx.QueryRow(ctx, `
			SELECT id FROM product_variants
			WHERE product_id = $1 AND deleted_at IS NULL
			ORDER BY is_default DESC, created_at ASC
			LIMIT 1
		`, productID).Scan(&variantID)
		if err != nil {
			return err
		}

		ledger := &model.InventoryLedger{
			ID:        uuid.New(),
			VariantID: variantID,
			Quantity:  delta,
			Reason:    model.InventoryReason(reason),
			CreatedAt: time.Now(),
		}
		query := `
			INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at)
			VALUES ($1, $2, $3, $4::inventory_reason, NOW())
		`
		_, err = tx.Exec(ctx, query, ledger.ID, ledger.VariantID, ledger.Quantity, ledger.Reason)
		return err
	})
}

