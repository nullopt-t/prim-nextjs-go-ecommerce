package checkout

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/order"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
)

type CartService interface {
	GetOrCreateCart(ctx context.Context, userID *uuid.UUID, sessionID *string) (*model.Cart, error)
	ClearCart(ctx context.Context, userID *uuid.UUID, sessionID *string) error
}

type OrderService interface {
	CreateOrder(ctx context.Context, in *order.CreateOrderInput) (*model.Order, error)
}

type CheckoutService struct {
	cartService  CartService
	orderService OrderService
}

func NewService(cartService CartService, orderService OrderService) *CheckoutService {
	return &CheckoutService{
		cartService:  cartService,
		orderService: orderService,
	}
}

type CheckoutInput struct {
	CustomerID      uuid.UUID
	CustomerEmail   string
	ShippingAddress model.Address
	BillingAddress  model.Address
	CouponID        *uuid.UUID
}

func (s *CheckoutService) Checkout(ctx context.Context, in *CheckoutInput) (*model.Order, error) {
	cartData, err := s.cartService.GetOrCreateCart(ctx, &in.CustomerID, nil)
	if err != nil {
		return nil, err
	}

	if len(cartData.Items) == 0 {
		return nil, apierr.ErrBadRequest("Cart is empty").WithCode("EMPTY_CART")
	}

	var items []order.CreateOrderItemInput
	var totalDiscount int64 = 0 // Add coupon logic here if needed
	var currency = "USD"
	if len(cartData.Items) > 0 {
		currency = cartData.Items[0].Currency
	}

	for _, item := range cartData.Items {
		var snapshotBytes []byte
		if item.Variant != nil {
			snapshotBytes, _ = json.Marshal(item.Variant)
		} else {
			snapshotBytes = []byte("{}")
		}

		items = append(items, order.CreateOrderItemInput{
			VariantID:       item.VariantID,
			Quantity:        item.Quantity,
			PriceAtPurchase: item.PriceAtPurchase,
			ProductSnapshot: string(snapshotBytes),
		})
	}

	orderIn := &order.CreateOrderInput{
		CustomerID:      &in.CustomerID,
		CustomerEmail:   in.CustomerEmail,
		ShippingAddress: in.ShippingAddress,
		BillingAddress:  in.BillingAddress,
		CouponID:        in.CouponID,
		DiscountAmount:  totalDiscount,
		Currency:        currency,
		Items:           items,
	}

	createdOrder, err := s.orderService.CreateOrder(ctx, orderIn)
	if err != nil {
		return nil, err
	}

	// Clear the user's cart after successful order creation
	_ = s.cartService.ClearCart(ctx, &in.CustomerID, nil)

	return createdOrder, nil
}
