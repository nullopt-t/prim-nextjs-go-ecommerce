// Package order implements core order domain logic, state machine validation,
// customer/admin query operations, and transactional order creation.
package order

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/internal/setting"
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
	// VendorID is the selling vendor for this line (nil = platform-owned; resolved from the product at checkout).
	VendorID *uuid.UUID
	// VendorEarning is the vendor share (item total minus commission, minor units), frozen at checkout.
	VendorEarning int64
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
	// Locale is the customer preferred language (e.g. "en", "ar") for communications.
	Locale string
	// Items is the non-empty list of line items included in the purchase.
	Items []CreateOrderItemInput
}

// CreatePaymentInput defines the parameters for recording a payment transaction on an order.
type CreatePaymentInput struct {
	OrderID               uuid.UUID
	Amount                int64
	Currency              string
	Status                model.PaymentStatus
	Type                  string
	Provider              string
	ProviderTransactionID *string
}

// InventoryReader is the read-only subset of inventory persistence the order
// domain needs. Reads carry no invariants, so the consumer-owned repository keeps
// order independent of the inventory service for lookups.
type InventoryReader interface {
	// ExpiredReservationOrderIDs lists orders whose reservation TTL has elapsed.
	ExpiredReservationOrderIDs(ctx context.Context) ([]uuid.UUID, error)
}

// InventoryWriter owns the stock invariants the order lifecycle depends on:
// appending compensating ledger entries, and releasing or committing
// reservations. It participates in the caller's transaction — every method
// receives the executor the order use case already holds, so stock and order
// state commit atomically.
type InventoryWriter interface {
	// ApplyStockAdjustment appends a signed ledger entry within the active tx.
	ApplyStockAdjustment(ctx context.Context, in inventory.AdjustStockInput) error
	// ReleaseReservationsForOrder releases live holds and reports how many.
	ReleaseReservationsForOrder(ctx context.Context, orderID uuid.UUID) (int64, error)
	// CommitReservationsForOrder turns holds into deductions on payment capture.
	CommitReservationsForOrder(ctx context.Context, orderID uuid.UUID) error
}

// OrderNotifier is the consumer-owned contract for post-commit lifecycle
// notifications. A failure is logged and never reverts the committed status.
type OrderNotifier interface {
	NotifyOrderShipped(ctx context.Context, in notifier.OrderShippedPayload) error
	NotifyOrderCancelled(ctx context.Context, in notifier.OrderCancelledPayload) error
}

// SettingsProvider exposes the effective store configuration used to gate
// optional lifecycle notifications.
type SettingsProvider interface {
	GetConfig(ctx context.Context) (setting.Config, error)
}

// UserReader is the read-only projection of user profiles the order domain needs
// for customer communications.
type UserReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// OrderService orchestrates database transactions, order status state machine validation,
// and audit log tracking for order lifecycle events.
type OrderService struct {
	qexecutor       database.Runner
	orderRepo       *Repository
	logger          log.Logger
	inventoryReader InventoryReader
	inventory       InventoryWriter
	userReader      UserReader
	settings        SettingsProvider
	orderNotifier   OrderNotifier
}

// NewService instantiates an OrderService with its database runner, order
// repository, logger, the inventory writer that owns stock mutation
// invariants, inventory read repository, user reader, settings, and order
// notifier.
func NewService(
	r database.Runner,
	orderRepo *Repository,
	logger log.Logger,
	inventoryReader InventoryReader,
	inventoryWriter InventoryWriter,
	userReader UserReader,
	settings SettingsProvider,
	orderNotifier OrderNotifier,
) *OrderService {
	if r == nil {
		panic("order: database runner is required")
	}
	if orderRepo == nil {
		panic("order: order repository is required")
	}
	if logger == nil {
		panic("order: logger is required")
	}
	if inventoryReader == nil {
		panic("order: inventory reader is required")
	}
	if inventoryWriter == nil {
		panic("order: inventory writer is required")
	}
	if userReader == nil {
		panic("order: user reader is required")
	}
	if settings == nil {
		panic("order: settings provider is required")
	}
	if orderNotifier == nil {
		panic("order: order notifier is required")
	}
	return &OrderService{
		qexecutor:       r,
		orderRepo:       orderRepo,
		logger:          logger,
		inventoryReader: inventoryReader,
		inventory:       inventoryWriter,
		userReader:      userReader,
		settings:        settings,
		orderNotifier:   orderNotifier,
	}
}

// PlaceOrder creates an order and its associated line items within the
// active database transaction (creating one if none exists), so checkout
// can atomically bundle stock reservation, coupon redemption, and payment
// recording with order creation.
func (s *OrderService) PlaceOrder(
	ctx context.Context,
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

	locale := in.Locale
	if locale != "ar" && locale != "en" {
		locale = "en"
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
		Locale:          locale,
		Items:           orderItems,
	}

	err := s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		// Persist the root order record.
		if err := s.orderRepo.CreateOrder(ctx, newOrder); err != nil {
			return err
		}

		// Persist each line item associated with the order.
		for i := range orderItems {
			if err := s.orderRepo.AddOrderItem(ctx, &orderItems[i]); err != nil {
				return err
			}
		}

		// Insert initial audit history entry.
		note := "Order created with status pending"
		if err := s.orderRepo.AddStatusHistory(ctx, orderID, model.OrderStatusPending, &note); err != nil {
			return err
		}

		return nil
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

	return newOrder, nil
}

// RecordPayment records a payment transaction for an order within the active transaction.
func (s *OrderService) RecordPayment(ctx context.Context, in *CreatePaymentInput) (*model.Payment, error) {
	if in.OrderID == uuid.Nil {
		return nil, apierr.BadRequestError("Order ID is required")
	}
	if in.Amount < 0 {
		return nil, apierr.BadRequestError("Payment amount cannot be negative")
	}

	p := &model.Payment{
		ID:                    uuid.New(),
		OrderID:               in.OrderID,
		Amount:                in.Amount,
		Currency:              in.Currency,
		Status:                in.Status,
		Type:                  in.Type,
		Provider:              in.Provider,
		ProviderTransactionID: in.ProviderTransactionID,
	}

	err := s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		if err := s.orderRepo.CreatePayment(ctx, p); err != nil {
			return s.handleError(err, "Failed to record payment transaction")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// CreateOrder creates an order and its associated line items within an atomic database transaction.
// It computes the gross total, applies discount deductions, persists the initial 'pending' status,
// and creates the initial status history audit trail.
func (s *OrderService) CreateOrder(ctx context.Context, in *CreateOrderInput) (*model.Order, error) {
	return s.PlaceOrder(ctx, in)
}

// restockItems asks the inventory domain to append a compensating 'return'
// ledger entry for each line item, restoring previously deducted units to
// salable stock. On-hand stock is derived from the ledger sum, so the entry is
// the restock. The inventory writer runs it on the active transaction.
func (s *OrderService) restockItems(
	ctx context.Context,
	items []model.OrderItem,
) error {
	for _, item := range items {
		in := inventory.AdjustStockInput{
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			Reason:    string(model.InventoryReasonReturn),
		}
		if err := s.inventory.ApplyStockAdjustment(ctx, in); err != nil {
			return fmt.Errorf("restore inventory for order item %s: %w", item.ID, err)
		}
	}

	return nil
}

// releaseOrRestockOnCancel fetches the order's line items and then resolves
// inventory for a cancelled order (see releaseOrRestockCancelItems).
func (s *OrderService) releaseOrRestockOnCancel(
	ctx context.Context,
	orderID uuid.UUID,
) error {
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("fetch order items for cancellation: %w", err)
	}
	return s.releaseOrRestockCancelItems(ctx, orderID, items)
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
	orderID uuid.UUID,
	items []model.OrderItem,
) error {
	released, err := s.inventory.ReleaseReservationsForOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("release reservations for cancelled order %s: %w", orderID, err)
	}
	if released > 0 {
		return nil
	}
	return s.restockItems(ctx, items)
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

// resolveConfig returns the effective store configuration, falling back to the
// built-in defaults if settings cannot be read. Notifications are best-effort
// and must not fail an otherwise successful order action.
func (s *OrderService) resolveConfig(ctx context.Context) setting.Config {
	cfg, err := s.settings.GetConfig(ctx)
	if err != nil {
		return setting.DefaultConfig()
	}
	return cfg
}

// resolveCustomerName safely looks up the customer display name inside the active
// transaction context, falling back to empty string when not found or guest checkout.
func (s *OrderService) resolveCustomerName(ctx context.Context, ord *model.Order) string {
	if ord == nil || ord.CustomerID == nil || *ord.CustomerID == uuid.Nil {
		return ""
	}
	u, err := s.userReader.GetByID(ctx, *ord.CustomerID)
	if err != nil || u == nil {
		return ""
	}
	return strings.TrimSpace(u.FullName)
}

func (s *OrderService) notifyOrderShipped(ctx context.Context, o *model.Order, customerName string) {
	if o == nil || !s.resolveConfig(ctx).NotifyOrderShipped {
		return
	}

	notifyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	carrier := ""
	if o.Carrier != nil {
		carrier = *o.Carrier
	}
	trackingURL := ""
	if o.TrackingURL != nil {
		trackingURL = *o.TrackingURL
	}

	if err := s.orderNotifier.NotifyOrderShipped(notifyCtx, notifier.OrderShippedPayload{
		OrderID:        o.ID,
		OrderNumber:    notifier.OrderNumber(o.ID),
		CustomerID:     o.CustomerID,
		CustomerEmail:  o.CustomerEmail,
		CustomerName:   customerName,
		Carrier:        carrier,
		TrackingNumber: o.TrackingNumber,
		TrackingURL:    trackingURL,
		Locale:         o.Locale,
	}); err != nil {
		s.logger.Warn("failed to publish order shipped notification", log.Meta{
			"orderId": o.ID.String(),
			"error":   err.Error(),
		})
	}
}

func (s *OrderService) notifyOrderCancelled(ctx context.Context, o *model.Order, reason string, customerName string) {
	if o == nil {
		return
	}

	notifyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := s.orderNotifier.NotifyOrderCancelled(notifyCtx, notifier.OrderCancelledPayload{
		OrderID:       o.ID,
		OrderNumber:   notifier.OrderNumber(o.ID),
		CustomerID:    o.CustomerID,
		CustomerEmail: o.CustomerEmail,
		CustomerName:  customerName,
		Reason:        reason,
		Locale:        o.Locale,
	}); err != nil {
		s.logger.Warn("failed to publish order cancelled notification", log.Meta{
			"orderId": o.ID.String(),
			"error":   err.Error(),
		})
	}
}

// GetOrderByID loads an order by its UUID within a read-only database query execution block.
func (s *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order *model.Order
	err := s.qexecutor.WithExecutor(ctx, func(ctx context.Context) error {
		res, err := s.orderRepo.GetOrderByID(ctx, id)
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
	err := s.qexecutor.WithExecutor(ctx, func(ctx context.Context) error {
		res, err := s.orderRepo.GetOrdersByCustomerID(ctx, customerID)
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

	err = s.qexecutor.WithExecutor(ctx, func(ctx context.Context) error {
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

		res, err := s.orderRepo.ListOrders(ctx, filter)
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
	var changedOrder *model.Order
	var statusChanged bool
	var notificationReason string
	var customerName string

	err := s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		// Acquire an exclusive FOR UPDATE row-level lock on the order to prevent concurrent race conditions
		// during status transitions.
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}

		if !CanTransitionOrderStatus(ord.Status, status) {
			return apierr.BadRequestError(
				fmt.Sprintf("Cannot transition order status from '%s' to '%s'", ord.Status, status),
			).WithCode(errcode.CodeInvalidOrderStatus)
		}

		statusChanged = ord.Status != status
		cancelling := ord.Status != model.OrderStatusCancelled && status == model.OrderStatusCancelled

		if err := s.orderRepo.UpdateOrderStatus(ctx, id, status); err != nil {
			return err
		}
		if err := s.orderRepo.AddStatusHistory(ctx, id, status, notes); err != nil {
			return err
		}

		// Resolve inventory on cancellation: release a still-held reservation
		// (unpaid order) or restock a committed sale (paid order), in the same
		// transaction as the status change.
		if cancelling {
			if err := s.releaseOrRestockOnCancel(ctx, id); err != nil {
				return err
			}
		}

		ord.Status = status
		changedOrder = ord
		if notes != nil {
			notificationReason = *notes
		}
		customerName = s.resolveCustomerName(ctx, ord)
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

	if changedOrder != nil && statusChanged {
		switch changedOrder.Status {
		case model.OrderStatusShipped:
			s.notifyOrderShipped(ctx, changedOrder, customerName)
		case model.OrderStatusCancelled:
			reason := notificationReason
			if reason == "" {
				reason = "Order cancelled"
			}
			s.notifyOrderCancelled(ctx, changedOrder, reason, customerName)
		}
	}

	return nil
}

// CancelOrder handles customer-initiated cancellation for pending or confirmed orders.
// Verifies customer authorization and valid transition before setting status to cancelled.
func (s *OrderService) CancelOrder(ctx context.Context, id uuid.UUID, customerID uuid.UUID) error {
	var cancelledOrder *model.Order
	var customerName string

	err := s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, id)
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
		wasAlreadyCancelled := ord.Status == model.OrderStatusCancelled

		note := "Cancelled by customer"
		if err := s.orderRepo.UpdateOrderStatus(ctx, id, model.OrderStatusCancelled); err != nil {
			return s.handleError(err, "Failed to cancel order")
		}

		if err := s.orderRepo.AddStatusHistory(ctx, id, model.OrderStatusCancelled, &note); err != nil {
			return s.handleError(err, "Failed to cancel order")
		}

		// Fetch order items to restore stock
		items, err := s.orderRepo.GetOrderItems(ctx, id)
		if err != nil {
			return s.handleError(err, "Failed to fetch order items")
		}

		// Resolve inventory: release the still-held reservation (unpaid order) or
		// restock the committed sale (paid order) within the same transaction.
		if err := s.releaseOrRestockCancelItems(ctx, id, items); err != nil {
			return s.handleError(err, "Failed to restore inventory")
		}

		// Update any associated captured or pending payment records to refunded / failed.
		if err := s.orderRepo.RefundOrFailPaymentsForOrder(ctx, id); err != nil {
			return s.handleError(err, "Failed to refund payment")
		}

		if !wasAlreadyCancelled {
			ord.Status = model.OrderStatusCancelled
			cancelledOrder = ord
			customerName = s.resolveCustomerName(ctx, ord)
		}

		return nil
	})
	if err != nil {
		return s.handleError(err, "Failed to cancel order")
	}

	if cancelledOrder != nil {
		s.notifyOrderCancelled(ctx, cancelledOrder, "Cancelled by customer", customerName)
	}
	return nil
}

// ReapExpiredReservations cancels pending orders whose checkout hold has expired
// and releases those holds, returning the number of orders cancelled. It is
// safe to run repeatedly and from multiple instances: each order is locked FOR
// UPDATE and only transitions while still pending, so concurrent runs are
// idempotent (the loser observes a non-pending status and does nothing).
func (s *OrderService) ReapExpiredReservations(ctx context.Context) (int, error) {
	var orderIDs []uuid.UUID
	if err := s.qexecutor.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		orderIDs, err = s.inventoryReader.ExpiredReservationOrderIDs(ctx)
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
	var cancelledOrder *model.Order
	var customerName string

	err := s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, orderID)
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

		if err := s.orderRepo.UpdateOrderStatus(ctx, orderID, model.OrderStatusCancelled); err != nil {
			return err
		}
		note := "Auto-cancelled: checkout reservation expired"
		if err := s.orderRepo.AddStatusHistory(ctx, orderID, model.OrderStatusCancelled, &note); err != nil {
			return err
		}

		// The hold lapsed without ever being committed, so there is no stock to
		// restock — releasing the reservation is the complete reversal.
		if _, err := s.inventory.ReleaseReservationsForOrder(ctx, orderID); err != nil {
			return err
		}

		// Fail the never-captured payment so it no longer looks receivable.
		if err := s.orderRepo.FailPendingPaymentsForOrder(ctx, orderID); err != nil {
			return s.handleError(err, "Failed to fail stale payment")
		}

		ord.Status = model.OrderStatusCancelled
		cancelledOrder = ord
		customerName = s.resolveCustomerName(ctx, ord)
		didCancel = true
		return nil
	})

	if didCancel && cancelledOrder != nil {
		s.notifyOrderCancelled(ctx, cancelledOrder, "Order reservation expired", customerName)
	}
	return didCancel, err
}

// GetDashboardStats aggregates live metrics for administrative reporting.
func (s *OrderService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	var stats *DashboardStats
	err := s.qexecutor.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		stats, err = s.orderRepo.GetDashboardStats(ctx)
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
	err := s.qexecutor.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		payments, page, err = s.orderRepo.ListPayments(ctx, filter)
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
	return s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		// Lock the payment row and resolve its order up front.
		orderID, err := s.orderRepo.GetPaymentOrderIDForUpdate(ctx, id)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Payment not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to update payment status")
		}

		if err := s.orderRepo.UpdatePaymentStatus(ctx, id, status); err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Payment not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to update payment status")
		}

		// Capture = sale. Commit the order's holds (deduct on-hand stock). The
		// operation is idempotent, so re-capturing cannot double-deduct. The
		// inventory service owns the insufficient-stock error mapping.
		if status == string(model.StatusCaptured) {
			if err := s.inventory.CommitReservationsForOrder(ctx, orderID); err != nil {
				return s.handleError(err, "Failed to update payment status")
			}
		}
		return nil
	})
}

// RefundPayment sets the payment status to refunded.
func (s *OrderService) RefundPayment(ctx context.Context, id uuid.UUID) error {
	return s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		status, err := s.orderRepo.GetPaymentStatusForUpdate(ctx, id)
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
			return s.orderRepo.UpdatePaymentStatus(ctx, id, "refunded")
		case "refunded":
			return apierr.ConflictError("Payment has already been refunded")
		default:
			return apierr.ConflictError("Only captured payments can be refunded")
		}
	})
}

// UpdateOrderTrackingInput carries carrier tracking details for an order.
type UpdateOrderTrackingInput struct {
	TrackingNumber string
	Carrier        *string
	TrackingURL    *string
}

// UpdateOrderTracking updates the carrier tracking details for an order.
// Lifecycle notifications are not resent here to avoid duplicate shipping emails.
func (s *OrderService) UpdateOrderTracking(ctx context.Context, id uuid.UUID, in UpdateOrderTrackingInput) error {
	trackingNumber := strings.TrimSpace(in.TrackingNumber)
	if trackingNumber == "" {
		return apierr.BadRequestError("Tracking number is required").WithCode(apierr.CodeInvalidInput)
	}

	var carrier *string
	if in.Carrier != nil {
		c := strings.TrimSpace(*in.Carrier)
		if c != "" {
			carrier = &c
		}
	}

	var trackingURL *string
	if in.TrackingURL != nil {
		u := strings.TrimSpace(*in.TrackingURL)
		if u != "" {
			if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
				return apierr.BadRequestError("Tracking URL must start with http:// or https://").WithCode(apierr.CodeInvalidInput)
			}
			trackingURL = &u
		}
	}

	err := s.qexecutor.WithTx(ctx, func(ctx context.Context) error {
		ord, err := s.orderRepo.GetOrderByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}

		if err := s.orderRepo.UpdateOrderTracking(ctx, id, trackingNumber, carrier, trackingURL); err != nil {
			return err
		}

		note := fmt.Sprintf("Tracking updated: %s", trackingNumber)
		if carrier != nil {
			note = fmt.Sprintf("Tracking updated: %s via %s", trackingNumber, *carrier)
		}
		_ = s.orderRepo.AddStatusHistory(ctx, ord.ID, ord.Status, &note)
		return nil
	})
	if err != nil {
		return s.handleError(err, "Failed to update order tracking")
	}

	return nil
}
