package checkout

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type CartService interface {
	GetOrCreateCart(ctx context.Context, userID *uuid.UUID, sessionID *string) (*model.Cart, error)
	ClearCart(ctx context.Context, userID *uuid.UUID, sessionID *string) error
	// GetOrCreateCartWithTx resolves (or creates) the cart inside the caller's
	// transaction with a FOR UPDATE row lock, serializing concurrent checkouts.
	GetOrCreateCartWithTx(
		ctx context.Context,
		tx database.QueryExecutor,
		userID *uuid.UUID,
		sessionID *string,
	) (*model.Cart, error)
	// ClearCartWithTx clears the cart inside the caller's transaction so checkout
	// and cart clearance commit atomically.
	ClearCartWithTx(
		ctx context.Context,
		tx database.QueryExecutor,
		userID *uuid.UUID,
		sessionID *string,
	) error
}

type OrderService interface {
	CreateOrder(ctx context.Context, in *order.CreateOrderInput) (*model.Order, error)
	// CreateOrderWithTx creates the order inside the caller's transaction so
	// stock reservation, coupon redemption, and payment stay atomic with it.
	CreateOrderWithTx(
		ctx context.Context,
		tx database.QueryExecutor,
		in *order.CreateOrderInput,
	) (*model.Order, error)
}

type InventoryService interface {
	AdjustStock(ctx context.Context, in inventory.AdjustStockInput) (*model.InventoryStock, error)
	// ReserveStockWithTx places a temporary stock hold inside the caller's
	// transaction so checkout stays atomic.
	ReserveStockWithTx(ctx context.Context, tx database.QueryExecutor, in inventory.ReserveStockInput) (*model.InventoryReservation, error)
}

type UserService interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type AddressService interface {
	GetAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.UserAddress, error)
}

// SettingsProvider exposes the effective store configuration used to price
// orders (shipping, taxes, payment availability, reservation TTL).
type SettingsProvider interface {
	GetConfig(ctx context.Context) (setting.Config, error)
}

type CheckoutService struct {
	cartService      CartService
	orderService     OrderService
	inventoryService InventoryService
	userService      UserService
	addressService   AddressService
	settings         SettingsProvider
	runner           database.Runner
}

func NewService(
	cartService CartService,
	orderService OrderService,
	userService UserService,
	addressService AddressService,
	inventoryService InventoryService,
	settings SettingsProvider,
	runner ...database.Runner,
) *CheckoutService {
	s := &CheckoutService{
		cartService:      cartService,
		orderService:     orderService,
		userService:      userService,
		addressService:   addressService,
		inventoryService: inventoryService,
		settings:         settings,
	}
	if len(runner) > 0 {
		s.runner = runner[0]
	}
	return s
}

// resolveConfig returns the effective store configuration, falling back to the
// built-in defaults if the settings provider is absent or fails. Falling back
// keeps checkout available (with historical pricing) rather than failing closed
// on a transient settings read error.
func (s *CheckoutService) resolveConfig(ctx context.Context) setting.Config {
	if s.settings == nil {
		return setting.DefaultConfig()
	}
	cfg, err := s.settings.GetConfig(ctx)
	if err != nil {
		return setting.DefaultConfig()
	}
	return cfg
}

type CheckoutInput struct {
	CustomerID        uuid.UUID
	ShippingAddressID uuid.UUID
	CouponID          *uuid.UUID
	PaymentMethod     string
}

func (s *CheckoutService) Checkout(ctx context.Context, in *CheckoutInput) (*model.Order, error) {
	if s.runner == nil {
		return nil, apierr.InternalServerError("Database runner unavailable")
	}

	var createdOrder *model.Order

	cfg := s.resolveConfig(ctx)

	// One transaction covers cart lock, authoritative pricing, coupon validation
	// and redemption, stock reservation, order creation, payment recording, and
	// cart clearance. Any failure rolls everything back (including reservations),
	// so there is no deducted-but-no-order or order-without-payment outcome.
	err := s.runner.WithTx(ctx, func(tx database.QueryExecutor) error {
		// 1. Lock the customer's cart FOR UPDATE so concurrent checkouts serialize.
		cart, err := s.cartService.GetOrCreateCartWithTx(ctx, tx, &in.CustomerID, nil)
		if err != nil {
			return err
		}

		if len(cart.Items) == 0 {
			return apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
		}

		// 2. Resolve the user and shipping address (read-only lookups).
		user, err := s.userService.GetUserByID(ctx, in.CustomerID)
		if err != nil {
			return err
		}

		shippingAddr, err := s.addressService.GetAddress(ctx, in.ShippingAddressID, in.CustomerID)
		if err != nil {
			return err
		}

		// 3. Authoritative pricing: re-read live variant prices, reject items with
		//    a missing price, and compute the gross total server-side.
		var orderItems []order.CreateOrderItemInput
		var grossTotal int64
		currency := "EGP"
		for _, item := range cart.Items {
			if item.Variant == nil || item.Variant.Price == nil {
				return apierr.BadRequestError(
					fmt.Sprintf("Variant %s has no defined price", item.VariantID),
				).WithCode(apierr.CodeInvalidInput)
			}
			unitPrice := *item.Variant.Price
			if item.Variant.Currency != nil && *item.Variant.Currency != "" {
				currency = *item.Variant.Currency
			}

			qty := int64(item.Quantity)
			if unitPrice > 0 && qty > math.MaxInt64/unitPrice {
				return apierr.BadRequestError("Order total exceeds maximum allowable limit")
			}
			itemTotal := unitPrice * qty
			if itemTotal > math.MaxInt64-grossTotal {
				return apierr.BadRequestError("Order total exceeds maximum allowable limit")
			}
			grossTotal += itemTotal

			// Build a complete, immutable snapshot of the product & variant at checkout.
			snapshotMap := make(map[string]any)
			if item.Variant != nil {
				snapshotMap["variantId"] = item.Variant.ID.String()
				snapshotMap["sku"] = item.Variant.SKU
				snapshotMap["title"] = item.Variant.Title
				snapshotMap["price"] = unitPrice
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

			snapshotBytes, _ := json.Marshal(snapshotMap)

			orderItems = append(orderItems, order.CreateOrderItemInput{
				VariantID:       item.VariantID,
				Quantity:        item.Quantity,
				PriceAtPurchase: unitPrice,
				ProductSnapshot: string(snapshotBytes),
			})
		}

		// 4. Validate and apply the coupon inside the same transaction, locking the coupon row
		//    so concurrent redemptions are serialized; the unique index on
		//    coupon_redemptions(coupon_id, user_id) backstops the per-user limit.
		var appliedCouponCode *string
		var totalDiscount int64
		if in.CouponID != nil {
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
				couponStatus  string
			)

			scanErr := tx.QueryRow(ctx, `
				SELECT id, code, discount_type, discount_value, min_cart_amount,
				       usage_limit, per_user_limit, starts_at, expires_at, status
				FROM coupons
				WHERE id = $1
				FOR UPDATE
			`, *in.CouponID).Scan(
				&cID, &code, &discountType, &discountValue, &minCartAmount,
				&usageLimit, &perUserLimit, &startsAt, &expiresAt, &couponStatus,
			)
			if scanErr != nil {
				if errors.Is(scanErr, sql.ErrNoRows) {
					return apierr.BadRequestError("Coupon not found").WithCode("COUPON_NOT_FOUND")
				}
				return apierr.InternalServerError("Failed to validate coupon").Wrap(scanErr)
			}

			appliedCouponCode = &code

			if couponStatus != "active" {
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
				if scanErr := tx.QueryRow(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1`, cID).Scan(&usedCount); scanErr != nil {
					return apierr.InternalServerError("Failed to check coupon usage").Wrap(scanErr)
				}
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
			if scanErr := tx.QueryRow(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`, cID, in.CustomerID).Scan(&userUsedCount); scanErr != nil {
				return apierr.InternalServerError("Failed to check per-user coupon limit").Wrap(scanErr)
			}
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
		}

		// 5. Apply the admin-configured shipping rule: free at/above the
		//    configured threshold, otherwise the configured flat rate.
		var shippingCost int64
		if threshold := cfg.FreeShippingThresholdMinorUnits(); grossTotal > 0 && grossTotal < threshold {
			shippingCost = cfg.ShippingMinorUnits()
		}

		// 5b. Resolve and validate the requested payment method against the
		//     admin-configured payment availability, then apply any COD fee.
		paymentMethod := strings.ToLower(strings.TrimSpace(in.PaymentMethod))
		isCOD := paymentMethod == "" || paymentMethod == "cod" || paymentMethod == "cash_on_delivery"
		if isCOD && !cfg.AllowCOD {
			return apierr.BadRequestError("Cash on delivery is currently unavailable").
				WithCode("PAYMENT_METHOD_UNAVAILABLE")
		}
		if !isCOD && !cfg.AllowCards {
			return apierr.BadRequestError("Online card payments are currently unavailable").
				WithCode("PAYMENT_METHOD_UNAVAILABLE")
		}
		normalizedPaymentMethod := "cash_on_delivery"
		if !isCOD {
			normalizedPaymentMethod = paymentMethod
		}
		var codFee int64
		if isCOD {
			codFee = cfg.CODFeeMinorUnits()
		}

		// 5c. Apply tax unless prices are already tax-inclusive. Tax is computed
		//     on the discounted subtotal so it is never charged on discounts.
		var taxAmount int64
		if !cfg.PricesIncludeTax && cfg.TaxRatePercent > 0 {
			taxable := grossTotal - totalDiscount
			if taxable < 0 {
				taxable = 0
			}
			taxAmount = int64(math.Round(float64(taxable) * cfg.TaxRatePercent / 100.0))
		}

		// 6. Reserve stock for every item (reservation at checkout; committed on
		//    payment, released on expiry or cancellation). All reservations live in
		//    this transaction and roll back with it on any failure.
		if s.inventoryService != nil {
			for _, item := range orderItems {
				if _, err := s.inventoryService.ReserveStockWithTx(ctx, tx, inventory.ReserveStockInput{
					VariantID: item.VariantID,
					CartID:    &cart.ID,
					Quantity:  item.Quantity,
					Duration:  time.Duration(cfg.ReservationTTLMinutes()) * time.Minute,
				}); err != nil {
					return err
				}
			}
		}

		// 7. Create the order inside this transaction.
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

		orderIn := &order.CreateOrderInput{
			CustomerID:      &in.CustomerID,
			CustomerEmail:   user.Identifier,
			ShippingAddress: addr,
			CouponID:        in.CouponID,
			CouponCode:      appliedCouponCode,
			DiscountAmount:  totalDiscount,
			ShippingAmount:  shippingCost,
			TaxAmount:       taxAmount,
			CODFee:          codFee,
			PaymentMethod:   normalizedPaymentMethod,
			Currency:        currency,
			Items:           orderItems,
		}

		var createErr error
		createdOrder, createErr = s.orderService.CreateOrderWithTx(ctx, tx, orderIn)
		if createErr != nil {
			return createErr
		}

		// 8. Record the coupon redemption; surface errors so the transaction rolls back.
		if in.CouponID != nil && totalDiscount > 0 {
			redemptionQuery := `
				INSERT INTO coupon_redemptions (id, coupon_id, user_id, order_id, discount_amount, created_at)
				VALUES ($1, $2, $3, $4, $5, NOW())
			`
			if _, execErr := tx.Exec(ctx, redemptionQuery, uuid.New(), *in.CouponID, in.CustomerID, createdOrder.ID, totalDiscount); execErr != nil {
				return apierr.InternalServerError("Failed to record coupon redemption").Wrap(execErr)
			}
		}

		// 9. Record the payment transaction; surface errors so the transaction rolls back.
		provider := normalizedPaymentMethod
		providerTxID := fmt.Sprintf("COD-%s", createdOrder.ID.String()[:8])
		if provider != "cash_on_delivery" {
			providerTxID = fmt.Sprintf("PAY-%s", createdOrder.ID.String()[:8])
		}

		paymentQuery := `
			INSERT INTO payments (id, order_id, amount, currency, status, provider, provider_transaction_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'pending'::payment_status, $5, $6, NOW(), NOW())
		`
		if _, execErr := tx.Exec(ctx, paymentQuery, uuid.New(), createdOrder.ID, createdOrder.TotalAmount, createdOrder.Currency, provider, providerTxID); execErr != nil {
			return apierr.InternalServerError("Failed to record payment").Wrap(execErr)
		}

		// 10. Clear the cart inside the same transaction so checkout and cart
		//     clearance commit atomically (no duplicate orders from a failed clear).
		if err := s.cartService.ClearCartWithTx(ctx, tx, &in.CustomerID, nil); err != nil {
			return apierr.InternalServerError("Failed to clear cart").Wrap(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

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

	cfg := s.resolveConfig(ctx)

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

		// Shipping follows the admin-configured rule: free at/above the configured
		// threshold, otherwise the configured flat rate.
		var shippingCost int64 = 0
		if threshold := cfg.FreeShippingThresholdMinorUnits(); grossTotal > 0 && grossTotal < threshold {
			shippingCost = cfg.ShippingMinorUnits()
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
