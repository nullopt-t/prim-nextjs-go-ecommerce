package checkout

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type CartService interface {
	GetOrCreateCart(ctx context.Context, userID *uuid.UUID, sessionID *string) (*model.Cart, error)
	ClearCart(ctx context.Context, userID *uuid.UUID, sessionID *string) error
}

type OrderService interface {
	CreateOrder(ctx context.Context, in *order.CreateOrderInput) (*model.Order, error)
}

type InventoryService interface {
	AdjustStock(ctx context.Context, in inventory.AdjustStockInput) (*model.InventoryStock, error)
}

type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type AddressService interface {
	GetAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.UserAddress, error)
}

type CheckoutService struct {
	cartService      CartService
	orderService     OrderService
	inventoryService InventoryService
	userService      UserService
	addressService   AddressService
	runner           database.Runner
}

func NewService(cartService CartService, orderService OrderService, userService UserService, addressService AddressService, inventoryService InventoryService, runner ...database.Runner) *CheckoutService {
	s := &CheckoutService{
		cartService:      cartService,
		orderService:     orderService,
		userService:      userService,
		addressService:   addressService,
		inventoryService: inventoryService,
	}
	if len(runner) > 0 {
		s.runner = runner[0]
	}
	return s
}

type CheckoutInput struct {
	CustomerID        uuid.UUID
	ShippingAddressID uuid.UUID
	CouponID          *uuid.UUID
	PaymentMethod     string
}

func (s *CheckoutService) Checkout(ctx context.Context, in *CheckoutInput) (*model.Order, error) {
	cartData, err := s.cartService.GetOrCreateCart(ctx, &in.CustomerID, nil)
	if err != nil {
		return nil, err
	}

	if len(cartData.Items) == 0 {
		return nil, apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
	}

	var items []order.CreateOrderItemInput
	var grossTotal int64 = 0
	var totalDiscount int64 = 0
	var currency = "EGP"
	if len(cartData.Items) > 0 && cartData.Items[0].Currency != "" {
		currency = cartData.Items[0].Currency
	}

	user, err := s.userService.GetUserByID(ctx, in.CustomerID)
	if err != nil {
		return nil, err
	}

	shippingAddr, err := s.addressService.GetAddress(ctx, in.ShippingAddressID, in.CustomerID)
	if err != nil {
		return nil, err
	}

	for _, item := range cartData.Items {
		var snapshotBytes []byte

		// Build a complete, immutable snapshot of the product & variant at the moment of checkout
		snapshotMap := make(map[string]any)
		if item.Variant != nil {
			snapshotMap["variantId"] = item.Variant.ID.String()
			snapshotMap["sku"] = item.Variant.SKU
			snapshotMap["title"] = item.Variant.Title
			snapshotMap["price"] = item.Variant.Price
			if item.Variant.Thumbnail != nil && item.Variant.Thumbnail.PublicURL != "" {
				snapshotMap["thumbnailUrl"] = item.Variant.Thumbnail.PublicURL
				snapshotMap["image"] = item.Variant.Thumbnail.PublicURL
			}
			if len(item.Variant.Media) > 0 && item.Variant.Media[0].Object != nil && item.Variant.Media[0].Object.PublicURL != "" {
				snapshotMap["image"] = item.Variant.Media[0].Object.PublicURL
				if _, ok := snapshotMap["thumbnailUrl"]; !ok {
					snapshotMap["thumbnailUrl"] = item.Variant.Media[0].Object.PublicURL
				}
			}
		}

		if item.ThumbnailURL != "" {
			snapshotMap["thumbnailUrl"] = item.ThumbnailURL
			snapshotMap["image"] = item.ThumbnailURL
		}

		if item.Product != nil {
			snapshotMap["productId"] = item.Product.ID.String()
			snapshotMap["productTitle"] = item.Product.Title
			snapshotMap["slug"] = item.Product.Slug
			if item.Product.Thumbnail != nil && item.Product.Thumbnail.PublicURL != "" {
				if _, ok := snapshotMap["thumbnailUrl"]; !ok {
					snapshotMap["thumbnailUrl"] = item.Product.Thumbnail.PublicURL
					snapshotMap["image"] = item.Product.Thumbnail.PublicURL
				}
			}
			if title, ok := snapshotMap["title"].(string); !ok || title == "" {
				snapshotMap["title"] = item.Product.Title
			}
		}

		snapshotBytes, _ = json.Marshal(snapshotMap)

		qty := int64(item.Quantity)
		grossTotal += item.PriceAtPurchase * qty

		items = append(items, order.CreateOrderItemInput{
			VariantID:       item.VariantID,
			Quantity:        item.Quantity,
			PriceAtPurchase: item.PriceAtPurchase,
			ProductSnapshot: string(snapshotBytes),
		})
	}

	var appliedCouponCode *string

	// Validate coupon if provided
	if in.CouponID != nil && s.runner != nil {
		err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
			var (
				cID           uuid.UUID
				code          string
				discountType  string
				discountValue int64
				minCartAmount sql.NullInt64
				usageLimit    sql.NullInt32
				perUserLimit  sql.NullInt32
				startsAt      *time.Time
				expiresAt     *time.Time
				status        string
			)

			query := `
				SELECT id, code, discount_type, discount_value, min_cart_amount,
				       usage_limit, per_user_limit, starts_at, expires_at, status
				FROM coupons
				WHERE id = $1
			`
			scanErr := db.QueryRow(ctx, query, *in.CouponID).Scan(
				&cID, &code, &discountType, &discountValue, &minCartAmount,
				&usageLimit, &perUserLimit, &startsAt, &expiresAt, &status,
			)
			if scanErr != nil {
				if errors.Is(scanErr, sql.ErrNoRows) {
					return apierr.BadRequestError("Coupon not found").WithCode("COUPON_NOT_FOUND")
				}
				return apierr.InternalServerError("Failed to validate coupon").Wrap(scanErr)
			}

			appliedCouponCode = &code

			if status != "active" {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' is not active", code)).WithCode("COUPON_INACTIVE")
			}

			now := time.Now()
			if startsAt != nil && now.Before(*startsAt) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has not started yet", code)).WithCode("COUPON_NOT_STARTED")
			}
			if expiresAt != nil && now.After(*expiresAt) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has expired", code)).WithCode("COUPON_EXPIRED")
			}

			if minCartAmount.Valid && grossTotal < minCartAmount.Int64 {
				return apierr.BadRequestError(fmt.Sprintf("Order total does not meet the minimum requirement of %.2f for coupon '%s'", float64(minCartAmount.Int64)/100, code)).WithCode("COUPON_MIN_AMOUNT_NOT_MET")
			}

			// Check global usage limit
			if usageLimit.Valid {
				var usedCount int
				_ = db.QueryRow(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1`, cID).Scan(&usedCount)
				if usedCount >= int(usageLimit.Int32) {
					return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' usage limit has been reached", code)).WithCode("COUPON_USAGE_LIMIT_EXCEEDED")
				}
			}

			// Check per-user usage limit
			var userLimit = 1
			if perUserLimit.Valid && perUserLimit.Int32 > 0 {
				userLimit = int(perUserLimit.Int32)
			}
			var userUsedCount int
			_ = db.QueryRow(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`, cID, in.CustomerID).Scan(&userUsedCount)
			if userUsedCount >= userLimit {
				return apierr.BadRequestError(fmt.Sprintf("You have already used coupon '%s' the maximum allowable times", code)).WithCode("COUPON_USER_LIMIT_EXCEEDED")
			}

			// Calculate discount
			if discountType == "percentage" {
				totalDiscount = (grossTotal * discountValue) / 100
			} else if discountType == "fixed_amount" {
				totalDiscount = discountValue
			}
			if totalDiscount > grossTotal {
				totalDiscount = grossTotal
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	addr := &model.Address{
		Street:  shippingAddr.Line1,
		City:    shippingAddr.City,
		Country: shippingAddr.Country,
	}

	if shippingAddr.State != nil {
		addr.State = *shippingAddr.State
	}
	if shippingAddr.PostalCode != nil {
		addr.PostalCode = *shippingAddr.PostalCode
	}

	// Charge shipping consistently with the cart-quote rule:
	// free shipping at or above 15000 minor units, 999 otherwise.
	var shippingCost int64
	if grossTotal > 0 && grossTotal < 15000 {
		shippingCost = 999
	}

	orderIn := &order.CreateOrderInput{
		CustomerID:      &in.CustomerID,
		CustomerEmail:   user.Identifier,
		ShippingAddress: addr,
		CouponID:        in.CouponID,
		CouponCode:      appliedCouponCode,
		DiscountAmount:  totalDiscount,
		ShippingAmount:  shippingCost,
		Currency:        currency,
		Items:           items,
	}

	// Reserve inventory / deduct stock atomically during checkout
	if s.inventoryService != nil {
		deducted := make([]order.CreateOrderItemInput, 0, len(items))
		for _, item := range items {
			_, err := s.inventoryService.AdjustStock(ctx, inventory.AdjustStockInput{
				VariantID: item.VariantID,
				Quantity:  -item.Quantity,
				Reason:    string(model.InventoryReasonSale),
			})
			if err != nil {
				// Compensate already deducted items if a subsequent item fails
				for _, prev := range deducted {
					_, _ = s.inventoryService.AdjustStock(ctx, inventory.AdjustStockInput{
						VariantID: prev.VariantID,
						Quantity:  prev.Quantity,
						Reason:    string(model.InventoryReasonAdjustment),
					})
				}
				return nil, err
			}
			deducted = append(deducted, item)
		}
	}

	createdOrder, err := s.orderService.CreateOrder(ctx, orderIn)
	if err != nil {
		if s.inventoryService != nil {
			// Roll back stock deductions if order creation failed
			for _, item := range items {
				_, _ = s.inventoryService.AdjustStock(ctx, inventory.AdjustStockInput{
					VariantID: item.VariantID,
					Quantity:  item.Quantity,
					Reason:    string(model.InventoryReasonAdjustment),
				})
			}
		}
		return nil, err
	}

	// Record coupon redemption and payment details if runner is available
	if s.runner != nil {
		_ = s.runner.WithTx(ctx, func(tx database.QueryExecutor) error {
			// Record coupon redemption
			if in.CouponID != nil && totalDiscount > 0 {
				redemptionQuery := `
					INSERT INTO coupon_redemptions (id, coupon_id, user_id, order_id, discount_amount, created_at)
					VALUES ($1, $2, $3, $4, $5, NOW())
				`
				_, _ = tx.Exec(ctx, redemptionQuery, uuid.New(), *in.CouponID, in.CustomerID, createdOrder.ID, totalDiscount)
			}

			// Record payment transaction
			provider := "cash_on_delivery"
			providerTxID := fmt.Sprintf("COD-%s", createdOrder.ID.String()[:8])
			if in.PaymentMethod != "" && in.PaymentMethod != "cod" && in.PaymentMethod != "cash_on_delivery" {
				provider = in.PaymentMethod
				providerTxID = fmt.Sprintf("PAY-%s", createdOrder.ID.String()[:8])
			}

			paymentQuery := `
				INSERT INTO payments (id, order_id, amount, currency, status, provider, provider_transaction_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, 'pending'::payment_status, $5, $6, NOW(), NOW())
			`
			_, _ = tx.Exec(ctx, paymentQuery, uuid.New(), createdOrder.ID, createdOrder.TotalAmount, createdOrder.Currency, provider, providerTxID)

			return nil
		})
	}

	// Clear the user's cart after successful order creation
	_ = s.cartService.ClearCart(ctx, &in.CustomerID, nil)

	return createdOrder, nil
}

type CouponValidationResult struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	DiscountType    string    `json:"discountType"`
	DiscountValue   int64     `json:"discountValue"`
	DiscountPercent *int      `json:"discountPercent,omitempty"`
	FixedDiscount   *float64  `json:"fixedDiscount,omitempty"`
	MinCartAmount   *int64    `json:"minCartAmount,omitempty"`
	DiscountAmount  int64     `json:"discountAmount"`
	Subtotal        int64     `json:"subtotal"`
	Shipping        int64     `json:"shipping"`
	Total           int64     `json:"total"`
	Message         string    `json:"message"`
}

func (s *CheckoutService) ValidateCoupon(ctx context.Context, code string, userID *uuid.UUID, sessionID *string) (*CouponValidationResult, error) {
	if s.runner == nil {
		return nil, apierr.InternalServerError("Database runner unavailable")
	}

	var grossTotal int64 = 0
	if userID != nil || sessionID != nil {
		if cartData, err := s.cartService.GetOrCreateCart(ctx, userID, sessionID); err == nil && cartData != nil {
			for _, item := range cartData.Items {
				unitPrice := item.PriceAtPurchase
				if item.Variant != nil && item.Variant.Price != nil {
					unitPrice = *item.Variant.Price
				}
				grossTotal += int64(item.Quantity) * unitPrice
			}
		}
	}

	// grossTotal is computed exclusively from the server-side cart and the
	// authoritative live variant prices. Client-supplied subtotals are never
	// trusted for pricing decisions.
	var result *CouponValidationResult

	err := s.runner.WithTx(ctx, func(db database.QueryExecutor) error {
		var (
			cID           uuid.UUID
			cCode         string
			discountType  string
			discountValue int64
			minCartAmount sql.NullInt64
			usageLimit    sql.NullInt32
			perUserLimit  sql.NullInt32
			startsAt      *time.Time
			expiresAt     *time.Time
			status        string
		)

		query := `
			SELECT id, code, discount_type, discount_value, min_cart_amount,
			       usage_limit, per_user_limit, starts_at, expires_at, status
			FROM coupons
			WHERE LOWER(code) = LOWER($1)
		`
		scanErr := db.QueryRow(ctx, query, code).Scan(
			&cID, &cCode, &discountType, &discountValue, &minCartAmount,
			&usageLimit, &perUserLimit, &startsAt, &expiresAt, &status,
		)
		if scanErr != nil {
			if errors.Is(scanErr, sql.ErrNoRows) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' not found", code)).WithCode("COUPON_NOT_FOUND")
			}
			return apierr.InternalServerError("Failed to validate coupon").Wrap(scanErr)
		}

		if status != "active" {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' is not active", cCode)).WithCode("COUPON_INACTIVE")
		}

		now := time.Now()
		if startsAt != nil && now.Before(*startsAt) {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has not started yet", cCode)).WithCode("COUPON_NOT_STARTED")
		}
		if expiresAt != nil && now.After(*expiresAt) {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has expired", cCode)).WithCode("COUPON_EXPIRED")
		}

		// Check global usage limit
		if usageLimit.Valid {
			var usedCount int
			_ = db.QueryRow(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1`, cID).Scan(&usedCount)
			if usedCount >= int(usageLimit.Int32) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' usage limit has been reached", cCode)).WithCode("COUPON_USAGE_LIMIT_EXCEEDED")
			}
		}

		// Check per-user limit if userID is provided
		if userID != nil {
			var userLimit = 1
			if perUserLimit.Valid && perUserLimit.Int32 > 0 {
				userLimit = int(perUserLimit.Int32)
			}
			var userUsedCount int
			_ = db.QueryRow(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`, cID, *userID).Scan(&userUsedCount)
			if userUsedCount >= userLimit {
				return apierr.BadRequestError(fmt.Sprintf("You have already used coupon '%s' the maximum allowable times", cCode)).WithCode("COUPON_USER_LIMIT_EXCEEDED")
			}
		}

		res := &CouponValidationResult{
			ID:            cID,
			Code:          cCode,
			DiscountType:  discountType,
			DiscountValue: discountValue,
			Message:       fmt.Sprintf("Coupon %s applied successfully!", cCode),
		}

		if minCartAmount.Valid {
			res.MinCartAmount = &minCartAmount.Int64
		}

		var calculatedDiscount int64 = 0
		if discountType == "percentage" {
			pct := int(discountValue)
			res.DiscountPercent = &pct
			if grossTotal > 0 {
				calculatedDiscount = (grossTotal * discountValue) / 100
			}
		} else if discountType == "fixed_amount" {
			fixed := float64(discountValue) / 100.0
			res.FixedDiscount = &fixed
			calculatedDiscount = discountValue
		}

		if calculatedDiscount > grossTotal && grossTotal > 0 {
			calculatedDiscount = grossTotal
		}

		// Standard shipping threshold: 15000 cents (150 EGP), cost: 999 cents (9.99 EGP)
		var shippingCost int64 = 0
		if grossTotal > 0 && grossTotal < 15000 {
			shippingCost = 999
		}

		finalTotal := grossTotal - calculatedDiscount + shippingCost
		if finalTotal < 0 {
			finalTotal = 0
		}

		res.Subtotal = grossTotal
		res.DiscountAmount = calculatedDiscount
		res.Shipping = shippingCost
		res.Total = finalTotal

		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
