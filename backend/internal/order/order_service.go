// Package order implements core order domain logic, state machine validation,
// customer/admin query operations, and transactional order creation.
package order

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order/errcode"
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
	// Currency is the 3-letter currency code (e.g., "USD").
	Currency string
	// Items is the non-empty list of line items included in the purchase.
	Items []CreateOrderItemInput
}

// InventoryService defines the subset of inventory functionality needed by OrderService for restoring stock upon cancellation.
type InventoryService interface {
	AdjustStock(ctx context.Context, in inventory.AdjustStockInput) (*model.InventoryStock, error)
	// AdjustStockWithTx restores/adjusts stock within the caller's transaction so
	// cancellation and restock remain atomic with the order status transition.
	AdjustStockWithTx(ctx context.Context, tx database.QueryExecutor, in inventory.AdjustStockInput) (*model.InventoryStock, error)
}

// OrderService orchestrates database transactions, order status state machine validation,
// and audit log tracking for order lifecycle events.
type OrderService struct {
	qexecuter        database.Runner
	orderRepo        *Repository
	logger           log.Logger
	inventoryService InventoryService
}

// NewService instantiates an OrderService with database runner, order repository, logger, and optional inventory service dependencies.
func NewService(
	r database.Runner,
	orderRepo *Repository,
	logger log.Logger,
	inventoryService ...InventoryService,
) *OrderService {
	s := &OrderService{
		qexecuter: r,
		orderRepo: orderRepo,
		logger:    logger,
	}
	if len(inventoryService) > 0 {
		s.inventoryService = inventoryService[0]
	}
	return s
}

// CreateOrderWithTx creates an order and its associated line items inside the
// caller's database transaction, so checkout can atomically bundle stock
// reservation, coupon redemption, and payment recording with order creation.
func (s *OrderService) CreateOrderWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
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

	finalTotal := totalAmount - in.DiscountAmount + in.ShippingAmount + in.TaxAmount
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
	err := s.qexecuter.WithTx(ctx, func(tx database.QueryExecutor) error {
		var err error
		createdOrder, err = s.CreateOrderWithTx(ctx, tx, in)
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

	for _, item := range items {
		if s.inventoryService != nil {
			if _, err := s.inventoryService.AdjustStockWithTx(
				ctx,
				tx,
				inventory.AdjustStockInput{
					VariantID: item.VariantID,
					Quantity:  item.Quantity,
					Reason:    string(model.InventoryReasonReturn),
				},
			); err != nil {
				return fmt.Errorf("restore inventory for order item %s: %w", item.ID, err)
			}
		} else {
			// Directly record inventory restoration ledger within the transaction
			ledgerQuery := `
				INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at)
				VALUES ($1, $2, $3, 'return'::inventory_reason, now())
			`
			if _, err := tx.Exec(ctx, ledgerQuery, uuid.New(), item.VariantID, item.Quantity); err != nil {
				return fmt.Errorf("restore inventory ledger for order item %s: %w", item.ID, err)
			}
		}
	}

	return nil
}

// GetOrderByID loads an order by its UUID within a read-only database query execution block.
func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order *model.Order
	err := s.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		res, err := s.orderRepo.GetOrderByID(ctx, db, id)
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
	err := s.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		res, err := s.orderRepo.GetOrdersByCustomerID(ctx, db, customerID)
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

// ListOrders retrieves a paginated set of orders filtered by optional customer ID or status.
func (s *OrderService) ListOrders(
	ctx context.Context,
	customerID *uuid.UUID,
	status *string,
	page int,
	pageSize int,
) ([]model.Order, *pagination.Page, error) {
	var orders []model.Order
	var pageMeta *pagination.Page

	err := s.qexecuter.WithDB(ctx, func(db database.QueryExecutor) error {
		if pageSize <= 0 {
			pageSize = 10
		}
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * pageSize

		filter := OrderListFilter{
			CustomerID: customerID,
			Status:     status,
			Limit:      pageSize,
			Offset:     offset,
		}

		totalCount, err := s.orderRepo.CountOrders(ctx, db, filter)
		if err != nil {
			return err
		}

		res, err := s.orderRepo.ListOrders(ctx, db, filter)
		if err != nil {
			return err
		}

		orders = res
		pageMeta = pagination.NewPage(page, pageSize, totalCount)
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
	err := s.qexecuter.WithTx(ctx, func(tx database.QueryExecutor) error {
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

		restocking := ord.Status != model.OrderStatusCancelled && status == model.OrderStatusCancelled

		if err := s.orderRepo.UpdateOrderStatus(ctx, tx, id, status); err != nil {
			return err
		}
		if err := s.orderRepo.AddStatusHistory(ctx, tx, id, status, notes); err != nil {
			return err
		}

		// Restore stock when an order is cancelled (by admin or customer) so the
		// reserved/purchased quantities return to salable stock in the same transaction.
		if restocking {
			if err := s.restockOrderItems(ctx, tx, id); err != nil {
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
	return s.qexecuter.WithTx(ctx, func(tx database.QueryExecutor) error {
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, tx, id)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return apierr.NotFoundError("Order not found").WithCode(errcode.CodeOrderNotFound)
			}
			return apierr.InternalServerError("Failed to fetch order").Wrap(err)
		}

		if ord.CustomerID == nil || *ord.CustomerID != customerID {
			return apierr.ForbiddenError("You are not authorized to cancel this order")
		}

		if !CanTransitionOrderStatus(ord.Status, model.OrderStatusCancelled) {
			return apierr.BadRequestError("Order cannot be cancelled in its current state").WithCode(errcode.CodeInvalidOrderStatus)
		}

		note := "Cancelled by customer"
		if err := s.orderRepo.UpdateOrderStatus(ctx, tx, id, model.OrderStatusCancelled); err != nil {
			return err
		}

		if err := s.orderRepo.AddStatusHistory(ctx, tx, id, model.OrderStatusCancelled, &note); err != nil {
			return err
		}

		// Fetch order items to restore stock
		items, err := s.orderRepo.GetOrderItems(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("fetch order items for cancellation: %w", err)
		}

		// Restore stock for purchased items within the same transaction
		for _, item := range items {
			if s.inventoryService != nil {
				_, err := s.inventoryService.AdjustStockWithTx(ctx, tx, inventory.AdjustStockInput{
					VariantID: item.VariantID,
					Quantity:  item.Quantity,
					Reason:    string(model.InventoryReasonReturn),
				})
				if err != nil {
					return fmt.Errorf("restore inventory for cancelled item %s: %w", item.ID, err)
				}
			} else {
				// Directly record inventory restoration ledger within the transaction
				ledgerQuery := `
					INSERT INTO inventory_ledgers (id, variant_id, quantity, reason, created_at)
					VALUES ($1, $2, $3, 'return'::inventory_reason, now())
				`
				if _, err := tx.Exec(ctx, ledgerQuery, uuid.New(), item.VariantID, item.Quantity); err != nil {
					return fmt.Errorf("restore inventory ledger for cancelled item %s: %w", item.ID, err)
				}
			}
		}

		// Update any associated captured or pending payment records to refunded / failed
		paymentRefundQuery := `
			UPDATE payments
			SET status = CASE
				WHEN status = 'captured'::payment_status THEN 'refunded'::payment_status
				ELSE 'failed'::payment_status
			END,
			updated_at = now()
			WHERE order_id = $1 AND status NOT IN ('refunded'::payment_status, 'failed'::payment_status)
		`
		if _, err := tx.Exec(ctx, paymentRefundQuery, id); err != nil {
			return fmt.Errorf("refund payment for cancelled order %s: %w", id, err)
		}

		return nil
	})
}
