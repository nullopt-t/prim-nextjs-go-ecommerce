package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

type InventoryService struct {
	logger log.Logger
	dr     database.Runner
	repo   *InventoryRepository
}

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

type AdjustStockInput struct {
	VariantID   uuid.UUID
	Quantity    int
	Reason      string
	ReferenceID *string
}

type ReserveStockInput struct {
	VariantID uuid.UUID
	CartID    *uuid.UUID
	Quantity  int
	Duration  time.Duration
}

func (s *InventoryService) AdjustStock(
	ctx context.Context,
	in AdjustStockInput,
) (*model.InventoryStock, error) {
	if in.VariantID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if in.Quantity == 0 {
		return nil, apierr.ErrBadRequest("Quantity adjustment cannot be zero").
			WithCode(apierr.CodeInvalidInput)
	}

	reason, err := model.ParseInventoryReason(in.Reason)
	if err != nil {
		return nil, apierr.ErrBadRequest("Invalid inventory reason: "+in.Reason).
			WithCode(errcode.CodeInvalidInventoryReason).
			Wrap(err)
	}

	var updatedStock *model.InventoryStock

	txErr := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		_ = s.repo.LockVariantForUpdate(ctx, tx, in.VariantID)

		currentStock, err := s.repo.GetStock(ctx, tx, in.VariantID)
		if err != nil {
			return err
		}

		if in.Quantity < 0 && (currentStock.OnHandQuantity+in.Quantity < 0) {
			return apierr.ErrBadRequest(
				fmt.Sprintf("Insufficient on-hand inventory (%d) for adjustment (%d)", currentStock.OnHandQuantity, in.Quantity),
			).WithCode(errcode.CodeInsufficientInventory)
		}

		ledger := &model.InventoryLedger{
			ID:          uuid.New(),
			VariantID:   in.VariantID,
			Quantity:    in.Quantity,
			Reason:      reason,
			ReferenceID: in.ReferenceID,
		}

		if err := s.repo.CreateLedger(ctx, tx, ledger); err != nil {
			return err
		}

		updatedStock, err = s.repo.GetStock(ctx, tx, in.VariantID)
		return err
	})

	if txErr != nil {
		mappedErr := database.MapError(txErr)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.ErrBadRequest("Referenced variant does not exist").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(txErr)
		default:
			var apiErr *apierr.APIError
			if errors.As(txErr, &apiErr) {
				return nil, apiErr
			}
			return nil, apierr.ErrInternalError("Failed to adjust inventory").
				WithCode(apierr.CodeInternalError).
				Wrap(txErr).
				WithStack()
		}
	}

	return updatedStock, nil
}

func (s *InventoryService) GetStockByVariantID(
	ctx context.Context,
	variantID uuid.UUID,
) (*model.InventoryStock, error) {
	if variantID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var stock *model.InventoryStock
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		stock, repoErr = s.repo.GetStock(ctx, db, variantID)
		return repoErr
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to get variant inventory stock").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return stock, nil
}

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
		return nil, apierr.ErrInternalError("Failed to get inventory stock for variants").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return stocks, nil
}

func (s *InventoryService) ListLedgers(
	ctx context.Context,
	variantID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.InventoryLedger], error) {
	if variantID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var result *pagination.PagedResult[model.InventoryLedger]
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var repoErr error
		result, repoErr = s.repo.ListLedgers(ctx, db, variantID, q)
		return repoErr
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to list inventory ledgers").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return result, nil
}

func (s *InventoryService) ReserveStock(
	ctx context.Context,
	in ReserveStockInput,
) (*model.InventoryReservation, error) {
	if in.VariantID == uuid.Nil {
		return nil, apierr.ErrBadRequest("Variant ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if in.Quantity <= 0 {
		return nil, apierr.ErrBadRequest("Reservation quantity must be positive").
			WithCode(apierr.CodeInvalidInput)
	}

	ttl := in.Duration
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	var resultReservation *model.InventoryReservation

	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		// Row-level lock to prevent concurrency overselling
		if err := s.repo.LockVariantForUpdate(ctx, tx, in.VariantID); err != nil {
			return err
		}

		stock, err := s.repo.GetStock(ctx, tx, in.VariantID)
		if err != nil {
			return err
		}

		if stock.AvailableQuantity < in.Quantity {
			return apierr.ErrBadRequest(
				fmt.Sprintf("Insufficient available stock (%d) for reservation (%d)", stock.AvailableQuantity, in.Quantity),
			).WithCode(errcode.CodeInsufficientInventory)
		}

		// Check if active cart reservation already exists to extend / upsert
		if in.CartID != nil {
			existing, err := s.repo.GetActiveCartReservation(ctx, tx, *in.CartID, in.VariantID)
			if err == nil && existing != nil {
				newQty := existing.Quantity + in.Quantity
				newExpiry := time.Now().Add(ttl)
				if err := s.repo.UpdateReservationQuantityAndExpiry(ctx, tx, existing.ID, newQty, newExpiry); err != nil {
					return err
				}
				existing.Quantity = newQty
				existing.ExpiresAt = newExpiry
				resultReservation = existing
				return nil
			}
		}

		newRes := &model.InventoryReservation{
			ID:        uuid.New(),
			VariantID: in.VariantID,
			CartID:    in.CartID,
			Quantity:  in.Quantity,
			ExpiresAt: time.Now().Add(ttl),
		}

		if err := s.repo.CreateReservation(ctx, tx, newRes); err != nil {
			return err
		}

		resultReservation = newRes
		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrForeignKeyViolation):
			return nil, apierr.ErrBadRequest("Referenced variant does not exist").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		default:
			var apiErr *apierr.APIError
			if errors.As(err, &apiErr) {
				return nil, apiErr
			}
			return nil, apierr.ErrInternalError("Failed to reserve inventory").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return resultReservation, nil
}

func (s *InventoryService) ReleaseReservation(
	ctx context.Context,
	reservationID uuid.UUID,
) error {
	if reservationID == uuid.Nil {
		return apierr.ErrBadRequest("Reservation ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.ReleaseReservation(ctx, db, reservationID)
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.ErrNotFound("Inventory reservation not found").
				WithCode(errcode.CodeReservationNotFound).
				Wrap(err)
		default:
			return apierr.ErrInternalError("Failed to release reservation").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return nil
}

func (s *InventoryService) ReleaseCartReservations(
	ctx context.Context,
	cartID uuid.UUID,
) error {
	if cartID == uuid.Nil {
		return apierr.ErrBadRequest("Cart ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.ReleaseCartReservations(ctx, db, cartID)
	})

	if err != nil {
		return apierr.ErrInternalError("Failed to release cart reservations").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}

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
		return 0, apierr.ErrInternalError("Failed to release expired reservations").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return count, nil
}

func (s *InventoryService) CommitReservation(
	ctx context.Context,
	reservationID uuid.UUID,
	referenceID *string,
) error {
	if reservationID == uuid.Nil {
		return apierr.ErrBadRequest("Reservation ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		res, err := s.repo.GetReservationByID(ctx, tx, reservationID)
		if err != nil {
			return err
		}

		if res.ReleasedAt != nil {
			return apierr.ErrBadRequest("Reservation has already been released").
				WithCode(errcode.CodeReservationExpired)
		}

		if time.Now().After(res.ExpiresAt) {
			return apierr.ErrBadRequest("Reservation has expired").
				WithCode(errcode.CodeReservationExpired)
		}

		if err := s.repo.ReleaseReservation(ctx, tx, reservationID); err != nil {
			return err
		}

		ledger := &model.InventoryLedger{
			ID:          uuid.New(),
			VariantID:   res.VariantID,
			Quantity:    -res.Quantity,
			Reason:      model.InventoryReasonSale,
			ReferenceID: referenceID,
		}

		return s.repo.CreateLedger(ctx, tx, ledger)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.ErrInternalError("Failed to commit reservation").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}

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
