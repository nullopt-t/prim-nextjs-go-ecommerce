package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

type CreateOrderItemInput struct {
	VariantID       uuid.UUID
	Quantity        int
	PriceAtPurchase int64
	ProductSnapshot string
}

type CreateOrderInput struct {
	CustomerID      *uuid.UUID
	CustomerEmail   string
	ShippingAddress model.Address
	CouponID        *uuid.UUID
	DiscountAmount  int64
	Currency        string
	Items           []CreateOrderItemInput
}

type OrderService struct {
	qexecuter database.Runner
	orderRepo *Repository
	logger    log.Logger
}

func NewService(
	r database.Runner,
	orderRepo *Repository,
	logger log.Logger,
) *OrderService {
	return &OrderService{
		qexecuter: r,
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, in *CreateOrderInput) (*model.Order, error) {
	if len(in.Items) == 0 {
		return nil, apierr.BadRequestError("Order must contain at least one item").WithCode(errcode.CodeEmptyOrderItems)
	}

	var createdOrder *model.Order

	err := s.qexecuter.WithTx(ctx, func(tx database.QueryExecutor) error {
		orderID := uuid.New()

		var totalAmount int64
		orderItems := make([]model.OrderItem, 0, len(in.Items))

		for _, itemInput := range in.Items {
			itemTotal := itemInput.PriceAtPurchase * int64(itemInput.Quantity)
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

		finalTotal := totalAmount - in.DiscountAmount
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
			DiscountAmount:  in.DiscountAmount,
			TotalAmount:     finalTotal,
			Currency:        in.Currency,
			Items:           orderItems,
		}

		if err := s.orderRepo.CreateOrder(ctx, tx, newOrder); err != nil {
			return err
		}

		for i := range orderItems {
			if err := s.orderRepo.AddOrderItem(ctx, tx, &orderItems[i]); err != nil {
				return err
			}
		}

		note := "Order created with status pending"
		if err := s.orderRepo.AddStatusHistory(ctx, tx, orderID, model.OrderStatusPending, &note); err != nil {
			return err
		}

		createdOrder = newOrder
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

	return createdOrder, nil
}

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

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus, notes *string) error {
	err := s.qexecuter.WithTx(ctx, func(tx database.QueryExecutor) error {
		if err := s.orderRepo.UpdateOrderStatus(ctx, tx, id, status); err != nil {
			return err
		}
		return s.orderRepo.AddStatusHistory(ctx, tx, id, status, notes)
	})

	if err != nil {
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

func (s *OrderService) CancelOrder(ctx context.Context, id uuid.UUID, customerID uuid.UUID) error {
	return s.qexecuter.WithTx(ctx, func(tx database.QueryExecutor) error {
		ord, err := s.orderRepo.GetOrderByID(ctx, tx, id)
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

		if ord.Status != model.OrderStatusPending {
			return apierr.BadRequestError("Order cannot be cancelled in its current state").WithCode("INVALID_ORDER_STATUS")
		}

		note := "Cancelled by customer"
		if err := s.orderRepo.UpdateOrderStatus(ctx, tx, id, model.OrderStatusCanceled); err != nil {
			return err
		}
		return s.orderRepo.AddStatusHistory(ctx, tx, id, model.OrderStatusCanceled, &note)
	})
}

