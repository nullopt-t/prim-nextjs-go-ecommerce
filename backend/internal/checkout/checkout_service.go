package checkout

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// CartRepository is the subset of cart persistence checkout needs.
type CartRepository interface {
	LockCartByUserID(ctx context.Context, qe database.QueryExecutor, userID uuid.UUID) error
	GetCartByUserID(ctx context.Context, qe database.QueryExecutor, userID uuid.UUID) (*model.Cart, error)
	GetCartBySessionID(ctx context.Context, qe database.QueryExecutor, sessionID string) (*model.Cart, error)
	ClearCart(ctx context.Context, qe database.QueryExecutor, cartID uuid.UUID) error
}

// OrderWriter owns order creation because it enforces order invariants
// (line-item validation, totals, status history). It participates in the
// caller's transaction without opening a second connection.
type OrderWriter interface {
	CreateOrderWithTx(ctx context.Context, tx *database.Tx, in *order.CreateOrderInput) (*model.Order, error)
}

// StockWriter owns stock reservation/commitment invariants and participates in
// the caller's transaction without opening a second connection.
type StockWriter interface {
	ReserveStockWithTx(ctx context.Context, tx *database.Tx, in inventory.ReserveStockInput) (*model.InventoryReservation, error)
	CommitReservationsForOrderWithTx(ctx context.Context, tx *database.Tx, orderID uuid.UUID) error
}

// UserRepository is the subset of user persistence checkout needs.
type UserRepository interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.User, error)
}

// AddressRepository is the subset of address persistence checkout needs.
type AddressRepository interface {
	GetByIDAndUser(ctx context.Context, qe database.QueryExecutor, id uuid.UUID, userID uuid.UUID) (*model.UserAddress, error)
}

// CouponRepository is the subset of coupon persistence checkout needs.
type CouponRepository interface {
	GetByIDForUpdate(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Coupon, error)
	GetByCode(ctx context.Context, qe database.QueryExecutor, code string) (*model.Coupon, error)
	GetRedemptionCount(ctx context.Context, qe database.QueryExecutor, couponID uuid.UUID) (int, error)
	GetUserRedemptionCount(ctx context.Context, qe database.QueryExecutor, couponID uuid.UUID, userID uuid.UUID) (int, error)
	RecordRedemption(ctx context.Context, qe database.QueryExecutor, red *model.CouponRedemption) error
}

// VariantRepository is the subset of variant persistence checkout needs for
// authoritative pricing and order snapshots.
type VariantRepository interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.ProductVariant, error)
}

// ProductRepository is the subset of product persistence checkout needs for
// order snapshots.
type ProductRepository interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Product, error)
}

// ObjectRepository is the subset of storage-object persistence checkout needs
// to resolve thumbnail public URLs.
type ObjectRepository interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Object, error)
}

// PaymentRepository persists payment records owned by the order domain.
type PaymentRepository interface {
	CreatePayment(ctx context.Context, qe database.QueryExecutor, p *model.Payment) error
}

// SettingsProvider exposes the effective store configuration used to price
// orders (shipping, taxes, payment availability, reservation TTL).
type SettingsProvider interface {
	GetConfig(ctx context.Context) (setting.Config, error)
}

type CheckoutService struct {
	cartRepo    CartRepository
	orderWriter OrderWriter
	stockWriter StockWriter
	userRepo    UserRepository
	addressRepo AddressRepository
	couponRepo  CouponRepository
	variantRepo VariantRepository
	productRepo ProductRepository
	objectRepo  ObjectRepository
	paymentRepo PaymentRepository
	storage     storage.StorageProvider
	settings    SettingsProvider
	runner      database.Runner
}

func NewService(
	cartRepo CartRepository,
	orderWriter OrderWriter,
	stockWriter StockWriter,
	userRepo UserRepository,
	addressRepo AddressRepository,
	couponRepo CouponRepository,
	variantRepo VariantRepository,
	productRepo ProductRepository,
	objectRepo ObjectRepository,
	paymentRepo PaymentRepository,
	storageProvider storage.StorageProvider,
	settings SettingsProvider,
	runner database.Runner,
) *CheckoutService {
	return &CheckoutService{
		cartRepo:    cartRepo,
		orderWriter: orderWriter,
		stockWriter: stockWriter,
		userRepo:    userRepo,
		addressRepo: addressRepo,
		couponRepo:  couponRepo,
		variantRepo: variantRepo,
		productRepo: productRepo,
		objectRepo:  objectRepo,
		paymentRepo: paymentRepo,
		storage:     storageProvider,
		settings:    settings,
		runner:      runner,
	}
}

// handleError translates repository and driver errors into typed APIErrors so
// only apierr.APIError values leave the checkout service.
func (s *CheckoutService) handleError(err error, fallbackMsg string) error {
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
	err := s.runner.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		dbTx, ok := tx.(*database.Tx)
		if !ok {
			return apierr.InternalServerError("Checkout requires a database transaction").
				WithCode(apierr.CodeInternalError)
		}

		// 1. Lock the customer's cart FOR UPDATE so concurrent checkouts serialize.
		//    Without the lock two requests can both read the same items, acquire
		//    variant locks in different orders, and deadlock (or create duplicate
		//    orders) before either clears the cart.
		if err := s.cartRepo.LockCartByUserID(ctx, tx, in.CustomerID); err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
			}
			return s.handleError(err, "Failed to lock cart")
		}

		cart, err := s.cartRepo.GetCartByUserID(ctx, tx, in.CustomerID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
			}
			return err
		}

		if cart == nil || cart.ID == uuid.Nil || len(cart.Items) == 0 {
			return apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
		}

		// 2. Resolve the user and shipping address (read-only lookups on the
		//    same connection, through consumer repository interfaces).
		buyer, err := s.userRepo.GetByID(ctx, tx, in.CustomerID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("User not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to resolve user")
		}

		shippingAddr, err := s.addressRepo.GetByIDAndUser(ctx, tx, in.ShippingAddressID, in.CustomerID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Shipping address not found").WithCode(apierr.CodeNotFound)
			}
			return s.handleError(err, "Failed to resolve shipping address")
		}

		// 3. Authoritative pricing: re-read live variant prices, reject items with
		//    a missing price, and compute the gross total server-side. Thumbnails
		//    are resolved directly from storage (bucket/key), never from SQL.
		var orderItems []order.CreateOrderItemInput
		var grossTotal int64
		currency := "EGP"
		for _, item := range cart.Items {
			v, err := s.variantRepo.GetByID(ctx, tx, item.VariantID)
			if err != nil {
				if errors.Is(database.MapError(err), database.ErrNotFound) {
					return apierr.BadRequestError(
						fmt.Sprintf("Variant %s is no longer available", item.VariantID),
					).WithCode(apierr.CodeInvalidInput)
				}
				return s.handleError(err, "Failed to load variant")
			}
			if v.Price == nil {
				return apierr.BadRequestError(
					fmt.Sprintf("Variant %s has no defined price", item.VariantID),
				).WithCode(apierr.CodeInvalidInput)
			}
			unitPrice := *v.Price
			if v.Currency != nil && *v.Currency != "" {
				currency = *v.Currency
			}

			prod, err := s.productRepo.GetByID(ctx, tx, v.ProductID)
			if err != nil {
				if errors.Is(database.MapError(err), database.ErrNotFound) {
					return apierr.BadRequestError(
						fmt.Sprintf("Product for variant %s is no longer available", item.VariantID),
					).WithCode(apierr.CodeInvalidInput)
				}
				return s.handleError(err, "Failed to load product")
			}

			var thumbnailURL string
			thumbnailID := v.ThumbnailObjectID
			if thumbnailID == nil {
				thumbnailID = prod.ThumbnailObjectID
			}
			if thumbnailID != nil && s.objectRepo != nil && s.storage != nil {
				obj, err := s.objectRepo.GetByID(ctx, tx, *thumbnailID)
				if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
					return s.handleError(err, "Failed to resolve thumbnail")
				}
				if obj != nil {
					thumbnailURL = s.storage.GetPublicURL(obj.Bucket, obj.Key)
				}
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

			snapshotMap := make(map[string]any)
			snapshotMap["variantId"] = item.VariantID.String()
			snapshotMap["sku"] = v.SKU
			snapshotMap["title"] = v.Title
			if prod.Title != "" {
				snapshotMap["title"] = prod.Title + " - " + v.Title
			}
			snapshotMap["price"] = unitPrice
			if thumbnailURL != "" {
				snapshotMap["thumbnailUrl"] = thumbnailURL
				snapshotMap["image"] = thumbnailURL
			}
			snapshotMap["productId"] = prod.ID.String()
			snapshotMap["productTitle"] = prod.Title
			snapshotMap["slug"] = prod.Slug

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
			coupon, err := s.couponRepo.GetByIDForUpdate(ctx, tx, *in.CouponID)
			if err != nil {
				if errors.Is(database.MapError(err), database.ErrNotFound) {
					return apierr.BadRequestError("Coupon not found").WithCode("COUPON_NOT_FOUND")
				}
				return s.handleError(err, "Failed to validate coupon")
			}
			code := coupon.Code
			appliedCouponCode = &code

			if coupon.Status != model.CouponStatusActive {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' is not active", code)).WithCode("COUPON_INACTIVE")
			}

			now := time.Now()
			if coupon.StartsAt != nil && now.Before(*coupon.StartsAt) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has not started yet", code)).WithCode("COUPON_NOT_STARTED")
			}
			if coupon.ExpiresAt != nil && now.After(*coupon.ExpiresAt) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has expired", code)).WithCode("COUPON_EXPIRED")
			}

			if coupon.MinCartAmount != nil && grossTotal < *coupon.MinCartAmount {
				return apierr.BadRequestError(fmt.Sprintf("Order total does not meet the minimum requirement of %.2f for coupon '%s'", float64(*coupon.MinCartAmount)/100, code)).WithCode("COUPON_MIN_AMOUNT_NOT_MET")
			}

			// Check global usage limit
			if coupon.UsageLimit != nil {
				usedCount, err := s.couponRepo.GetRedemptionCount(ctx, tx, coupon.ID)
				if err != nil {
					return s.handleError(err, "Failed to check coupon usage")
				}
				if usedCount >= *coupon.UsageLimit {
					return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' usage limit has been reached", code)).WithCode("COUPON_USAGE_LIMIT_EXCEEDED")
				}
			}

			// Check per-user usage limit
			userLimit := 1
			if coupon.PerUserLimit != nil && *coupon.PerUserLimit > 0 {
				userLimit = *coupon.PerUserLimit
			}
			userUsedCount, err := s.couponRepo.GetUserRedemptionCount(ctx, tx, coupon.ID, in.CustomerID)
			if err != nil {
				return s.handleError(err, "Failed to check per-user coupon limit")
			}
			if userUsedCount >= userLimit {
				return apierr.BadRequestError(fmt.Sprintf("You have already used coupon '%s' the maximum allowable times", code)).WithCode("COUPON_USER_LIMIT_EXCEEDED")
			}

			// Calculate discount
			switch coupon.DiscountType {
			case model.CouponDiscountTypePercentage:
				totalDiscount = (grossTotal * coupon.DiscountValue) / 100
			case model.CouponDiscountTypeFixedAmount:
				totalDiscount = coupon.DiscountValue
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

		// 6. Create the order inside this transaction. It is created before the
		//    stock holds so each reservation can be linked to the order id and
		//    later committed (on payment) or released (on cancellation/expiry)
		//    deterministically.
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
			CustomerEmail:   buyer.Identifier,
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
		createdOrder, createErr = s.orderWriter.CreateOrderWithTx(ctx, dbTx, orderIn)
		if createErr != nil {
			return createErr
		}

		// 7. Reserve stock for every item, linked to the order. The hold is
		//    committed into a sale when payment is captured and released if the
		//    order is cancelled or the hold expires. All reservations live in
		//    this transaction and roll back with it on any failure.
		//
		//    Reserve variants in a deterministic order so two concurrent
		//    checkouts that share variants cannot deadlock by acquiring the same
		//    row locks in opposite orders (mirrors CommitReservationsForOrder).
		if s.stockWriter != nil {
			sortedItems := make([]order.CreateOrderItemInput, len(orderItems))
			copy(sortedItems, orderItems)
			sort.Slice(sortedItems, func(i, j int) bool {
				return sortedItems[i].VariantID.String() < sortedItems[j].VariantID.String()
			})

			for _, item := range sortedItems {
				if _, err := s.stockWriter.ReserveStockWithTx(ctx, dbTx, inventory.ReserveStockInput{
					VariantID: item.VariantID,
					CartID:    &cart.ID,
					OrderID:   &createdOrder.ID,
					Quantity:  item.Quantity,
					Duration:  time.Duration(cfg.ReservationTTLMinutes()) * time.Minute,
				}); err != nil {
					return err
				}
			}

			// Cash-on-delivery orders are confirmed the moment they are placed:
			// payment is collected on delivery, so there is no online payment
			// window to hold stock for. Commit the deduction immediately so the
			// units leave salable stock and the expiry reaper never cancels a
			// legitimate COD order. Card orders keep their hold until the
			// payment is captured.
			if isCOD {
				if err := s.stockWriter.CommitReservationsForOrderWithTx(ctx, dbTx, createdOrder.ID); err != nil {
					return err
				}
			}
		}

		// 8. Record the coupon redemption; surface errors so the transaction rolls back.
		if in.CouponID != nil && totalDiscount > 0 {
			redemption := &model.CouponRedemption{
				CouponID:       *in.CouponID,
				UserID:         in.CustomerID,
				OrderID:        createdOrder.ID,
				DiscountAmount: totalDiscount,
			}
			if err := s.couponRepo.RecordRedemption(ctx, tx, redemption); err != nil {
				return s.handleError(err, "Failed to record coupon redemption")
			}
		}

		// 9. Record the payment transaction; surface errors so the transaction rolls back.
		provider := normalizedPaymentMethod
		providerTxID := fmt.Sprintf("COD-%s", createdOrder.ID.String()[:8])
		if provider != "cash_on_delivery" {
			providerTxID = fmt.Sprintf("PAY-%s", createdOrder.ID.String()[:8])
		}

		payment := &model.Payment{
			ID:                    uuid.New(),
			OrderID:               createdOrder.ID,
			Amount:                createdOrder.TotalAmount,
			Currency:              createdOrder.Currency,
			Status:                model.StatusPending,
			Type:                  normalizedPaymentMethod,
			Provider:              provider,
			ProviderTransactionID: &providerTxID,
		}
		if err := s.paymentRepo.CreatePayment(ctx, tx, payment); err != nil {
			return s.handleError(err, "Failed to record payment")
		}

		// 10. Clear the cart inside the same transaction so checkout and cart
		//     clearance commit atomically (no duplicate orders from a failed clear).
		if err := s.cartRepo.ClearCart(ctx, tx, cart.ID); err != nil {
			return s.handleError(err, "Failed to clear cart")
		}

		return nil
	})

	if err != nil {
		return nil, s.handleError(err, "Checkout failed")
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

	var grossTotal int64
	var result *CouponValidationResult

	// One read connection for the whole use case: cart pricing, coupon lookup and
	// redemption counts all run on the same QueryExecutor (no connection hopping).
	err := s.runner.WithDB(ctx, func(db database.QueryExecutor) error {
		if userID != nil || sessionID != nil {
			var (
				cartData *model.Cart
				err      error
			)
			if userID != nil {
				cartData, err = s.cartRepo.GetCartByUserID(ctx, db, *userID)
			} else {
				cartData, err = s.cartRepo.GetCartBySessionID(ctx, db, *sessionID)
			}
			if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
				return s.handleError(err, "Failed to load cart")
			}
			if err == nil && cartData != nil && cartData.ID != uuid.Nil {
				for _, item := range cartData.Items {
					// grossTotal is computed from the authoritative live variant
					// price; the snapshot price is only a fallback when the variant
					// has no live price left.
					v, vErr := s.variantRepo.GetByID(ctx, db, item.VariantID)
					if vErr == nil && v.Price != nil {
						grossTotal += int64(item.Quantity) * *v.Price
						continue
					}
					if vErr != nil && !errors.Is(database.MapError(vErr), database.ErrNotFound) {
						return s.handleError(vErr, "Failed to load variant price")
					}
					grossTotal += int64(item.Quantity) * item.PriceAtPurchase
				}
			}
		}

		// Client-supplied subtotals are never trusted for pricing decisions.
		coupon, err := s.couponRepo.GetByCode(ctx, db, code)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' not found", code)).WithCode("COUPON_NOT_FOUND")
			}
			return s.handleError(err, "Failed to validate coupon")
		}

		if coupon.Status != model.CouponStatusActive {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' is not active", coupon.Code)).WithCode("COUPON_INACTIVE")
		}

		now := time.Now()
		if coupon.StartsAt != nil && now.Before(*coupon.StartsAt) {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has not started yet", coupon.Code)).WithCode("COUPON_NOT_STARTED")
		}
		if coupon.ExpiresAt != nil && now.After(*coupon.ExpiresAt) {
			return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' has expired", coupon.Code)).WithCode("COUPON_EXPIRED")
		}

		// Check global usage limit
		if coupon.UsageLimit != nil {
			usedCount, err := s.couponRepo.GetRedemptionCount(ctx, db, coupon.ID)
			if err != nil {
				return s.handleError(err, "Failed to check coupon usage")
			}
			if usedCount >= *coupon.UsageLimit {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' usage limit has been reached", coupon.Code)).WithCode("COUPON_USAGE_LIMIT_EXCEEDED")
			}
		}

		// Check per-user limit if userID is provided
		if userID != nil {
			userLimit := 1
			if coupon.PerUserLimit != nil && *coupon.PerUserLimit > 0 {
				userLimit = *coupon.PerUserLimit
			}
			userUsedCount, err := s.couponRepo.GetUserRedemptionCount(ctx, db, coupon.ID, *userID)
			if err != nil {
				return s.handleError(err, "Failed to check per-user coupon limit")
			}
			if userUsedCount >= userLimit {
				return apierr.BadRequestError(fmt.Sprintf("You have already used coupon '%s' the maximum allowable times", coupon.Code)).WithCode("COUPON_USER_LIMIT_EXCEEDED")
			}
		}

		res := &CouponValidationResult{
			ID:            coupon.ID,
			Code:          coupon.Code,
			DiscountType:  string(coupon.DiscountType),
			DiscountValue: coupon.DiscountValue,
			Message:       fmt.Sprintf("Coupon %s applied successfully!", coupon.Code),
		}

		if coupon.MinCartAmount != nil {
			res.MinCartAmount = coupon.MinCartAmount
		}

		var calculatedDiscount int64
		switch coupon.DiscountType {
		case model.CouponDiscountTypePercentage:
			pct := int(coupon.DiscountValue)
			res.DiscountPercent = &pct
			if grossTotal > 0 {
				calculatedDiscount = (grossTotal * coupon.DiscountValue) / 100
			}
		case model.CouponDiscountTypeFixedAmount:
			fixed := float64(coupon.DiscountValue) / 100.0
			res.FixedDiscount = &fixed
			calculatedDiscount = coupon.DiscountValue
		}

		if calculatedDiscount > grossTotal && grossTotal > 0 {
			calculatedDiscount = grossTotal
		}

		// Shipping follows the admin-configured rule: free at/above the configured
		// threshold, otherwise the configured flat rate.
		var shippingCost int64
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
		return nil, s.handleError(err, "Failed to validate coupon")
	}

	return result, nil
}
