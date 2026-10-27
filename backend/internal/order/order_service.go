// Package order implements core order domain logic, state machine validation,
// customer/admin query operations, and transactional order creation.
package order

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// CreateOrderItemInput describes the essential pricing and variant attributes for a new order line item.
type CreateOrderItemInput struct {
	// VariantID is the unique UUID identifying the specific SKU/variant being purchased.
	VariantID uuid.UUID
	// Quantity is the number of units ordered.
	Quantity int
	// PriceAtPurchase is the unit price in minor currency units (cents).
	PriceAtPurchase int64
	// ProductSnapshot contains serialized JSON capturing product name, variant title, SKU, and image URL at checkout time.
	ProductSnapshot string
}

// CreateOrderInput encapsulates all parameters required to generate a brand new customer order.
type CreateOrderInput struct {
	// CustomerID is the optional UUID of an authenticated customer (nil for guest checkout).
	CustomerID *uuid.UUID
	// CustomerEmail is the contact email address where invoices and receipts are sent.
	CustomerEmail string
	// ShippingAddress is the physical destination address for the delivery.
	ShippingAddress *model.Address
	// CouponID is the optional applied promotional discount coupon UUID.
	CouponID *uuid.UUID
	// CouponCode is the human-readable code of the applied coupon.
	CouponCode *string
	// DiscountAmount is the deducted total from the gross order price in minor units.
	DiscountAmount int64
	// ShippingAmount is the shipping charge applied to the order in minor units.
	ShippingAmount int64
	// TaxAmount is the tax amount applied to the order in minor units.
	TaxAmount int64
	// CODFee is the cash-on-delivery surcharge applied to the order in minor units.
	CODFee int64
	// PaymentMethod is the normalized payment method selected at checkout.
	PaymentMethod string
	// Currency is the 3-letter currency code (e.g., "USD").
	Currency string
	// Items is the non-empty list of line items included in the purchase.
	Items []CreateOrderItemInput
}

// InventoryRepository defines the subset of inventory persistence the order domain
// needs for restocking cancelled orders, releasing reservations, and committing
// payment captures — all on the caller's transaction.
type InventoryRepository interface {
	// CreateLedger writes an inventory ledger entry (e.g. 'return' on cancel) within the caller's tx.
	CreateLedger(ctx context.Context, qe database.QueryExecutor, ledger *model.InventoryLedger) error
	// ReleaseReservationsForOrder releases active stock holds for the given order within the caller's tx.
	ReleaseReservationsForOrder(ctx context.Context, qe database.QueryExecutor, orderID uuid.UUID) (int64, error)
	// CommitReservationsForOrder turns holds into deductions for the given order within the caller's tx.
	CommitReservationsForOrder(ctx context.Context, qe database.QueryExecutor, orderID uuid.UUID) error
	// ExpiredReservationOrderIDs lists orders whose reservation TTL has elapsed.
	ExpiredReservationOrderIDs(ctx context.Context, qe database.QueryExecutor) ([]uuid.UUID, error)
}

// OrderService orchestrates database transactions, order status state machine validation,
// and audit log tracking for order lifecycle events.
type OrderService struct {
	qexecutor     database.Runner
	orderRepo     *Repository
	logger        log.Logger
	inventoryRepo InventoryRepository
}

// NewService instantiates an OrderService with database runner, order repository, logger, and optional inventory repository dependencies.
func NewService(
	r database.Runner,
	orderRepo *Repository,
	logger log.Logger,
	inventoryRepo ...InventoryRepository,
) *OrderService {
	s := &OrderService{
		qexecutor: r,
		orderRepo: orderRepo,
		logger:    logger,
	}
	if len(inventoryRepo) > 0 {
		s.inventoryRepo = inventoryRepo[0]
	}
	return s
}

// CreateOrderWithTx creates an order and its associated line items inside the
// caller's database transaction, so checkout can atomically bundle stock
// reservation, coupon redemption, and payment recording with order creation.
func (s *OrderService) CreateOrderWithTx(
	ctx context.Context,
	tx *database.Tx,
	in *CreateOrderInput,
) (*model.Order, error) {
	// Validate that at least one line item is provided.
	if len(in.Items) == 0 {
		return nil, apierr.BadRequestError("Order must contain at least one item").WithCode(errcode.CodeEmptyOrderItems)
	}

	orderID := uuid.New()

	var totalAmount int64
	orderItems := make([]model.OrderItem, 0, len(in.Items))

	for _, itemInput := range in.Items {
		if itemInput.Quantity <= 0 {
			return nil, apierr.BadRequestError("Item quantity must be greater than zero")
		}
		if itemInput.PriceAtPurchase < 0 {
			return nil, apierr.BadRequestError("Item price cannot be negative")
		}

		// Validate against int64 overflow during itemTotal multiplication and running total
		qty := int64(itemInput.Quantity)
		if itemInput.PriceAtPurchase > 0 && qty > math.MaxInt64/itemInput.PriceAtPurchase {
			return nil, apierr.BadRequestError("Order total exceeds maximum allowable limit")
		}
		itemTotal := itemInput.PriceAtPurchase * qty
		if itemTotal > math.MaxInt64-totalAmount {
			return nil, apierr.BadRequestError("Order total exceeds maximum allowable limit")
		}
		totalAmount += itemTotal

		item := model.OrderItem{
			ID:              uuid.New(),
			OrderID:         orderID,
			VariantID:       itemInput.VariantID,
			Quantity:        itemInput.Quantity,
			PriceAtPurchase: itemInput.PriceAtPurchase,
			ProductSnapshot: itemInput.ProductSnapshot,
		}
		orderItems = append(orderItems, item)
	}

	finalTotal := totalAmount - in.DiscountAmount + in.ShippingAmount + in.TaxAmount + in.CODFee
	if finalTotal < 0 {
		finalTotal = 0
	}

	newOrder := &model.Order{
		ID:              orderID,
		CustomerID:      in.CustomerID,
		CustomerEmail:   in.CustomerEmail,
		ShippingAddress: in.ShippingAddress,
		Status:          model.OrderStatusPending,
		CouponID:        in.CouponID,
		CouponCode:      in.CouponCode,
		DiscountAmount:  in.DiscountAmount,
		ShippingAmount:  in.ShippingAmount,
		TaxAmount:       in.TaxAmount,
		CODFee:          in.CODFee,
		PaymentMethod:   in.PaymentMethod,
		TotalAmount:     finalTotal,
		Currency:        in.Currency,
		Items:           orderItems,
	}

	// Persist the root order record.
	if err := s.orderRepo.CreateOrder(ctx, tx, newOrder); err != nil {
		return nil, err
	}

	// Persist each line item associated with the order.
	for i := range orderItems {
		if err := s.orderRepo.AddOrderItem(ctx, tx, &orderItems[i]); err != nil {
			return nil, err
		}
	}

	// Insert initial audit history entry.
	note := "Order created with status pending"
	if err := s.orderRepo.AddStatusHistory(ctx, tx, orderID, model.OrderStatusPending, &note); err != nil {
		return nil, err
	}

	return newOrder, nil
}

// CreateOrder creates an order and its associated line items within an atomic database transaction.
// It computes the gross total, applies discount deductions, persists the initial 'pending' status,
// and creates the initial status history audit trail.
func (s *OrderService) CreateOrder(ctx context.Context, in *CreateOrderInput) (*model.Order, error) {
	var createdOrder *model.Order

	// Execute within an atomic transaction.
	err := s.qexecutor.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		dbTx, ok := tx.(*database.Tx)
		if !ok {
			return apierr.InternalServerError("Order creation requires a database transaction").
				WithCode(apierr.CodeInternalError)
		}
		var err error
		createdOrder, err = s.CreateOrderWithTx(ctx, dbTx, in)
		return err
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrConflict):
			return nil, apierr.ConflictError("Order conflict").
				WithCode(errcode.CodeOrderAlreadyExists).
				Wrap(err)
		default:
			return nil, apierr.InternalServerError("Failed to create order").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}

	return createdOrder, nil
}

// restockOrderItems releases the purchased quantity of every line item back to
// variant stock within the caller's transaction.
func (s *OrderService) restockOrderItems(
	ctx context.Context,
	tx database.QueryExecutor,
	orderID uuid.UUID,
) error {
	items, err := s.orderRepo.GetOrderItems(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("fetch order items for restock: %w", err)
	}
	return s.restockItems(ctx, tx, items)
}

// restockItems writes a compensating 'return' ledger entry for each line item,
// restoring previously deducted units to salable stock. On-hand stock is
// derived from the ledger sum, so the ledger entry is the restock. It runs
// entirely on the caller's transaction; the inventory repository owns the SQL.
func (s *OrderService) restockItems(
	ctx context.Context,
	tx database.QueryExecutor,
	items []model.OrderItem,
) error {
	if s.inventoryRepo == nil {
		return apierr.InternalServerError("Inventory repository is not configured").
			WithCode(apierr.CodeInternalError)
	}

	for _, item := range items {
		ledger := &model.InventoryLedger{
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			Reason:    model.InventoryReasonReturn,
		}
		if err := s.inventoryRepo.CreateLedger(ctx, tx, ledger); err != nil {
			return fmt.Errorf("restore inventory for order item %s: %w", item.ID, err)
		}
	}

	return nil
}

// releaseOrRestockOnCancel fetches the order's line items and then resolves
// inventory for a cancelled order (see releaseOrRestockCancelItems).
func (s *OrderService) releaseOrRestockOnCancel(
	ctx context.Context,
	tx database.QueryExecutor,
	orderID uuid.UUID,
) error {
	items, err := s.orderRepo.GetOrderItems(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("fetch order items for cancellation: %w", err)
	}
	return s.releaseOrRestockCancelItems(ctx, tx, orderID, items)
}

// releaseOrRestockCancelItems resolves inventory when an order is cancelled.
//
// An order's stock can be in one of two states, and only one compensation is
// correct for each:
//   - still HELD by a live reservation (order never paid): releasing the hold
//     returns the units to salable stock. No ledger entry is written because the
//     units were never deducted from on-hand.
//   - already COMMITTED as a sale (order was paid): there is no live hold left,
//     so the deducted units must be returned via a compensating 'return' ledger.
//
// Distinguishing the two prevents the previous double-count, where cancelling an
// unpaid order added stock that had never been removed.
func (s *OrderService) releaseOrRestockCancelItems(
	ctx context.Context,
	tx database.QueryExecutor,
	orderID uuid.UUID,
	items []model.OrderItem,
) error {
	if s.inventoryRepo != nil {
		released, err := s.inventoryRepo.ReleaseReservationsForOrder(ctx, tx, orderID)
		if err != nil {
			return fmt.Errorf("release reservations for cancelled order %s: %w", orderID, err)
		}
		if released > 0 {
			return nil
		}
	}
	return s.restockItems(ctx, tx, items)
}

// handleError translates repository and driver errors into typed APIErrors so
// that only apierr.APIError values leave the order service. Already-mapped
// APIErrors are returned unchanged to avoid double wrapping.
func (s *OrderService) handleError(err error, fallbackMsg string) error {
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
		return apierr.NotFoundError(fallbackMsg + ": not found").
			WithCode(apierr.CodeNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError(fallbackMsg + ": conflict").
			WithCode(apierr.CodeResourceConflict).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.BadRequestError(fallbackMsg + ": referenced entity does not exist").
			WithCode(apierr.CodeInvalidReference).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrCheckViolation),
		errors.Is(mappedErr, database.ErrNotNullViolation),
		errors.Is(mappedErr, database.ErrInvalidInput):
		return apierr.BadRequestError(fallbackMsg + ": invalid input").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrTimeout), errors.Is(mappedErr, database.ErrCanceled):
		return apierr.New(http.StatusGatewayTimeout, "database operation timed out or canceled").
			WithCode(apierr.CodeGatewayTimeout).
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// GetOrderByID loads an order by its UUID within a read-only database query execution block.
func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order *model.Order
	err := s.qexecutor.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		res, err := s.orderRepo.GetOrderByID(ctx, exec, id)
		if err != nil {
			return err
		}
		order = res
		return nil
	})

	if err != nil {
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return nil, apierr.NotFoundError("Order not found").
				WithCode(errcode.CodeOrderNotFound)
		default:
			return nil, apierr.InternalServerError("Failed to fetch order").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}
	return order, nil
}

// GetCustomerOrders fetches all orders belonging to the specified customer UUID.
func (s *OrderService) GetCustomerOrders(ctx context.Context, customerID uuid.UUID) ([]model.Order, error) {
	var orders []model.Order
	err := s.qexecutor.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		res, err := s.orderRepo.GetOrdersByCustomerID(ctx, exec, customerID)
		if err != nil {
			return err
		}
		orders = res
		return nil
	})

	if err != nil {
		return nil, apierr.InternalServerError("Failed to fetch customer orders").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
	return orders, nil
}

// ListOrders retrieves a keyset-paginated set of orders filtered by optional customer ID, status, or date range.
func (s *OrderService) ListOrders(
	ctx context.Context,
	customerID *uuid.UUID,
	status *string,
	dateFrom, dateTo *time.Time,
	q pagination.ListQuery,
) ([]model.Order, *pagination.CursorPage, error) {
	var orders []model.Order
	var pageMeta *pagination.CursorPage

	cursor, err := q.Keyset()
	if err != nil {
		return nil, nil, apierr.BadRequestError("Invalid pagination cursor").Wrap(err)
	}

	err = s.qexecutor.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		pageSize := q.PageSize
		if pageSize <= 0 {
			pageSize = 10
		}

		filter := OrderListFilter{
			CustomerID: customerID,
			Status:     status,
			DateFrom:   dateFrom,
			DateTo:     dateTo,
			Limit:      pageSize,
			Cursor:     cursor,
		}

		res, err := s.orderRepo.ListOrders(ctx, exec, filter)
		if err != nil {
			return err
		}

		orders, pageMeta = pagination.NewCursorPageFrom(res, pageSize, func(o model.Order) pagination.Keyset {
			return pagination.Keyset{CreatedAt: o.CreatedAt, ID: o.ID}
		})
		return nil
	})

	if err != nil {
		return nil, nil, apierr.InternalServerError("Failed to list orders").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return orders, pageMeta, nil
}

// CanTransitionOrderStatus implements the finite state machine transitions for the order lifecycle:
//   - pending -> paid, processing, cancelled
//   - paid -> processing, shipped, refunded, cancelled
//   - processing -> shipped, refunded, cancelled
//   - shipped -> delivered, refunded
//   - delivered -> refunded
//   - cancelled -> (terminal state)
//   - refunded -> (terminal state)
func CanTransitionOrderStatus(current, next model.OrderStatus) bool {
	if current == next {
		return true
	}
	switch current {
	case model.OrderStatusPending:
		// Pending can transition to paid, processing, or cancelled
		return next == model.OrderStatusPaid || next == model.OrderStatusProcessing || next == model.OrderStatusCancelled
	case model.OrderStatusPaid:
		// Paid can transition to processing, shipped, refunded, or cancelled
		return next == model.OrderStatusProcessing || next == model.OrderStatusShipped || next == model.OrderStatusRefunded || next == model.OrderStatusCancelled
	case model.OrderStatusProcessing:
		// Processing can transition to shipped, refunded, or cancelled
		return next == model.OrderStatusShipped || next == model.OrderStatusRefunded || next == model.OrderStatusCancelled
	case model.OrderStatusShipped:
		// Shipped can only transition to delivered or refunded (shipped orders in transit cannot be cancelled)
		return next == model.OrderStatusDelivered || next == model.OrderStatusRefunded
	case model.OrderStatusDelivered:
		// Delivered is a terminal delivery state; can only be refunded
		return next == model.OrderStatusRefunded
	case model.OrderStatusCancelled:
		// Once cancelled, an order is terminal and cannot transition to delivered, shipped, or processing
		return false
	case model.OrderStatusRefunded:
		// Terminal state
		return false
	default:
		return false
	}
}

// UpdateOrderStatus modifies an existing order's state after validating the state machine transition.
// It executes within a transaction and inserts an audit log entry in order_status_history.
func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus, notes *string) error {
	err := s.qexecutor.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		// Acquire an exclusive FOR UPDATE row-level lock on the order to prevent concurrent race conditions
		// during status transitions.
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}

		if !CanTransitionOrderStatus(ord.Status, status) {
			return apierr.BadRequestError(
				fmt.Sprintf("Cannot transition order status from '%s' to '%s'", ord.Status, status),
			).WithCode(errcode.CodeInvalidOrderStatus)
		}

		cancelling := ord.Status != model.OrderStatusCancelled && status == model.OrderStatusCancelled

		if err := s.orderRepo.UpdateOrderStatus(ctx, tx, id, status); err != nil {
			return err
		}
		if err := s.orderRepo.AddStatusHistory(ctx, tx, id, status, notes); err != nil {
			return err
		}

		// Resolve inventory on cancellation: release a still-held reservation
		// (unpaid order) or restock a committed sale (paid order), in the same
		// transaction as the status change.
		if cancelling {
			if err := s.releaseOrRestockOnCancel(ctx, tx, id); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		var appErr *apierr.APIError
		if errors.As(err, &appErr) {
			return err
		}
		mappedErr := database.MapError(err)
		switch {
		case errors.Is(mappedErr, database.ErrNotFound):
			return apierr.NotFoundError("Order not found").
				WithCode(errcode.CodeOrderNotFound)
		default:
			return apierr.InternalServerError("Failed to update order status").
				WithCode(apierr.CodeInternalError).
				Wrap(err).
				WithStack()
		}
	}
	return nil
}

// CancelOrder handles customer-initiated cancellation for pending or confirmed orders.
// Verifies customer authorization and valid transition before setting status to cancelled.
func (s *OrderService) CancelOrder(ctx context.Context, id uuid.UUID, customerID uuid.UUID) error {
	err := s.qexecutor.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, tx, id)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Order not found").WithCode(errcode.CodeOrderNotFound)
			}
			return s.handleError(err, "Failed to fetch order")
		}

		if ord.CustomerID == nil || *ord.CustomerID != customerID {
			return apierr.ForbiddenError("You are not authorized to cancel this order")
		}

		if !CanTransitionOrderStatus(ord.Status, model.OrderStatusCancelled) {
			return apierr.BadRequestError("Order cannot be cancelled in its current state").WithCode(errcode.CodeInvalidOrderStatus)
		}

		note := "Cancelled by customer"
		if err := s.orderRepo.UpdateOrderStatus(ctx, tx, id, model.OrderStatusCancelled); err != nil {
			return s.handleError(err, "Failed to cancel order")
		}

		if err := s.orderRepo.AddStatusHistory(ctx, tx, id, model.OrderStatusCancelled, &note); err != nil {
			return s.handleError(err, "Failed to cancel order")
		}

		// Fetch order items to restore stock
		items, err := s.orderRepo.GetOrderItems(ctx, tx, id)
		if err != nil {
			return s.handleError(err, "Failed to fetch order items")
		}

		// Resolve inventory: release the still-held reservation (unpaid order) or
		// restock the committed sale (paid order) within the same transaction.
		if err := s.releaseOrRestockCancelItems(ctx, tx, id, items); err != nil {
			return s.handleError(err, "Failed to restore inventory")
		}

		// Update any associated captured or pending payment records to refunded / failed.
		if err := s.orderRepo.RefundOrFailPaymentsForOrder(ctx, tx, id); err != nil {
			return s.handleError(err, "Failed to refund payment")
		}

		return nil
	})
	if err != nil {
		return s.handleError(err, "Failed to cancel order")
	}
	return nil
}

// ReapExpiredReservations cancels pending orders whose checkout hold has expired
// and releases those holds, returning the number of orders cancelled. It is
// safe to run repeatedly and from multiple instances: each order is locked FOR
// UPDATE and only transitions while still pending, so concurrent runs are
// idempotent (the loser observes a non-pending status and does nothing).
func (s *OrderService) ReapExpiredReservations(ctx context.Context) (int, error) {
	if s.inventoryRepo == nil {
		return 0, nil
	}

	var orderIDs []uuid.UUID
	if err := s.qexecutor.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		orderIDs, err = s.inventoryRepo.ExpiredReservationOrderIDs(ctx, exec)
		return err
	}); err != nil {
		return 0, s.handleError(err, "Failed to list expired reservation orders")
	}

	cancelled := 0
	for _, orderID := range orderIDs {
		didCancel, err := s.expirePendingOrder(ctx, orderID)
		if err != nil {
			s.logger.Error("failed to expire stale pending order", log.Meta{
				"orderId": orderID.String(),
				"error":   err.Error(),
			})
			continue
		}
		if didCancel {
			cancelled++
		}
	}

	return cancelled, nil
}

// expirePendingOrder cancels one pending order and releases its lapsed hold in a
// single transaction, reporting whether a transition occurred.
func (s *OrderService) expirePendingOrder(ctx context.Context, orderID uuid.UUID) (bool, error) {
	didCancel := false
	err := s.qexecutor.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, tx, orderID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return nil // order already removed; nothing to do
			}
			return err
		}

		// Only a still-pending order is eligible. A paid/fulfilled order has
		// already committed its hold and would not appear in the expired set.
		if ord.Status != model.OrderStatusPending {
			return nil
		}
		if !CanTransitionOrderStatus(ord.Status, model.OrderStatusCancelled) {
			return nil
		}

		if err := s.orderRepo.UpdateOrderStatus(ctx, tx, orderID, model.OrderStatusCancelled); err != nil {
			return err
		}
		note := "Auto-cancelled: checkout reservation expired"
		if err := s.orderRepo.AddStatusHistory(ctx, tx, orderID, model.OrderStatusCancelled, &note); err != nil {
			return err
		}

		// The hold lapsed without ever being committed, so there is no stock to
		// restock — releasing the reservation is the complete reversal.
		if _, err := s.inventoryRepo.ReleaseReservationsForOrder(ctx, tx, orderID); err != nil {
			return err
		}

		// Fail the never-captured payment so it no longer looks receivable.
		if err := s.orderRepo.FailPendingPaymentsForOrder(ctx, tx, orderID); err != nil {
			return s.handleError(err, "Failed to fail stale payment")
		}

		didCancel = true
		return nil
	})
	return didCancel, err
}

// GetDashboardStats aggregates live metrics for administrative reporting.
func (s *OrderService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats *DashboardStats
	err := s.qexecutor.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		stats, err = s.orderRepo.GetDashboardStats(ctx, exec)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to load dashboard stats")
	}
	return stats, nil
}

// ListPayments returns a keyset-paginated list of payment transactions.
func (s *OrderService) ListPayments(ctx context.Context, filter PaymentListFilter) ([]PaymentItem, *pagination.CursorPage, error) {
	var payments []PaymentItem
	var page *pagination.CursorPage
	err := s.qexecutor.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		payments, page, err = s.orderRepo.ListPayments(ctx, exec, filter)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list payments")
	}
	return payments, page, nil
}

// UpdatePaymentStatus changes the status of a payment record. Transitioning a
// payment to 'captured' is the point of sale: the owning order's live stock
// reservations are committed into sale deductions in the same transaction.
func (s *OrderService) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status string) error {
	if !model.PaymentStatus(status).IsValid() {
		return apierr.BadRequestError(fmt.Sprintf("Invalid payment status '%s'", status)).WithCode(apierr.CodeInvalidInput)
	}
	return s.qexecutor.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		// Lock the payment row and resolve its order up front.
		orderID, err := s.orderRepo.GetPaymentOrderIDForUpdate(ctx, tx, id)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Payment not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to update payment status")
		}

		if err := s.orderRepo.UpdatePaymentStatus(ctx, tx, id, status); err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Payment not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to update payment status")
		}

		// Capture = sale. Commit the order's holds (deduct on-hand stock). The
		// operation is idempotent, so re-capturing cannot double-deduct.
		if status == string(model.StatusCaptured) && s.inventoryRepo != nil {
			if err := s.inventoryRepo.CommitReservationsForOrder(ctx, tx, orderID); err != nil {
				if errors.Is(err, inventory.ErrInsufficientStock) {
					return apierr.BadRequestError(
						fmt.Sprintf("Insufficient physical on-hand stock to capture payment for order %s", orderID),
					).WithCode(errcode.CodeInsufficientInventory)
				}
				return s.handleError(err, "Failed to update payment status")
			}
		}
		return nil
	})
}

// RefundPayment sets the payment status to refunded.
func (s *OrderService) RefundPayment(ctx context.Context, id uuid.UUID) error {
	return s.qexecutor.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		status, err := s.orderRepo.GetPaymentStatusForUpdate(ctx, tx, id)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Payment not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to refund payment")
		}

		// Only a captured payment holds funds that can be returned. Refunding a
		// pending/authorized/failed payment would fabricate a refund, and a second
		// refund of an already-refunded payment would double-count it.
		switch status {
		case "captured":
			return s.orderRepo.UpdatePaymentStatus(ctx, tx, id, "refunded")
		case "refunded":
			return apierr.ConflictError("Payment has already been refunded")
		default:
			return apierr.ConflictError("Only captured payments can be refunded")
		}
	})
}

// UpdateOrderTracking updates the carrier tracking number for an order.
func (s *OrderService) UpdateOrderTracking(ctx context.Context, id uuid.UUID, trackingNumber string) error {
	err := s.qexecutor.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return s.orderRepo.UpdateOrderTracking(ctx, tx, id, trackingNumber)
	})
	if err != nil {
		return s.handleError(err, "Failed to update order tracking")
	}
	return nil
}
