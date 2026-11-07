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
	"github.com/nullopt-t/prim-backend/internal/catalog/coupon"
	"github.com/nullopt-t/prim-backend/internal/catalog/inventory"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/notifier"
	"github.com/nullopt-t/prim-backend/internal/order"
	"github.com/nullopt-t/prim-backend/internal/setting"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// CartReader is the subset of cart persistence checkout needs for reading and locking.
type CartReader interface {
	LockCartByUserID(ctx context.Context, userID uuid.UUID) error
	GetCartByUserID(ctx context.Context, userID uuid.UUID) (*model.Cart, error)
	GetCartBySessionID(ctx context.Context, sessionID string) (*model.Cart, error)
}

// CartWriter owns cart mutation.
type CartWriter interface {
	ClearItems(ctx context.Context, cartID uuid.UUID) error
}

// OrderWriter owns order creation and payment recording because it enforces order invariants
// (line-item validation, totals, status history, payment recording). It participates in the
// caller's transaction without opening a second connection.
type OrderWriter interface {
	PlaceOrder(ctx context.Context, in *order.CreateOrderInput) (*model.Order, error)
	RecordPayment(ctx context.Context, in *order.CreatePaymentInput) (*model.Payment, error)
}

// StockWriter owns stock reservation/commitment invariants and participates in
// the caller's transaction without opening a second connection.
type StockWriter interface {
	HoldStock(ctx context.Context, in inventory.ReserveStockInput) (*model.InventoryReservation, error)
	CommitReservationsForOrder(ctx context.Context, orderID uuid.UUID) error
}

// UserReader is the subset of user persistence checkout needs.
type UserReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// AddressReader is the subset of address persistence checkout needs.
type AddressReader interface {
	GetByIDAndUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.UserAddress, error)
}

// CouponService owns coupon eligibility and discount rules. Checkout supplies
// the authoritative gross total and the executor it already holds, so the
// rules live in one place without opening a second connection.
type CouponService interface {
	GetCodeByID(ctx context.Context, id uuid.UUID) (string, error)
	GetSellerIDByCode(ctx context.Context, code string) (*uuid.UUID, error)
	ApplyByCode(ctx context.Context, code string, grossTotal int64, userID *uuid.UUID) (*coupon.AppliedCoupon, error)
	RecordRedemption(ctx context.Context, couponID uuid.UUID, userID uuid.UUID, orderID uuid.UUID, discountAmount int64, currency string) error
}

// VariantReader is the subset of variant persistence checkout needs for
// authoritative pricing and order snapshots.
type VariantReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProductVariant, error)
}

// ProductReader is the subset of product persistence checkout needs for
// order snapshots and kit expansion.
type ProductReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	// GetBundleComposition resolves a product's kit composition. It reports
	// database.ErrNotFound for products that are not bundles.
	GetBundleComposition(ctx context.Context, productID uuid.UUID) (*model.BundleComposition, error)
}

// SellerReader is the subset of seller data checkout needs: standing and
// commission terms for the line's seller. Implemented by the seller
// repository; inactive sellers make their items unpurchasable.
type SellerReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Seller, error)
}

// ObjectReader is the subset of storage-object persistence checkout needs
// to resolve thumbnail public URLs.
type ObjectReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Object, error)
}

// SettingsProvider exposes the effective store configuration used to price
// orders (shipping, taxes, payment availability, reservation TTL).
type SettingsProvider interface {
	GetConfig(ctx context.Context) (setting.Config, error)
}

// OrderNotifier is the consumer-owned notification contract used after a
// checkout commits. Notification failures are logged and never roll back an
// already-committed order.
type OrderNotifier interface {
	NotifyOrderPlaced(ctx context.Context, in notifier.OrderPlacedPayload) error
}

// AdminOrderNotifier broadcasts committed order events to the admin live stream.
type AdminOrderNotifier interface {
	Publish(ctx context.Context, ev notifier.AdminOrderEvent) error
}

type CheckoutService struct {
	cartReader    CartReader
	cartWriter    CartWriter
	orderWriter   OrderWriter
	stockWriter   StockWriter
	userReader    UserReader
	addressReader AddressReader
	couponService CouponService
	variantReader VariantReader
	productReader ProductReader
	sellerReader  SellerReader
	objectReader  ObjectReader
	storage       storage.StorageProvider
	settings      SettingsProvider
	runner        database.Runner
	orderNotifier OrderNotifier
	adminNotifier AdminOrderNotifier
	logger        log.Logger
}

func NewService(
	cartReader CartReader,
	cartWriter CartWriter,
	orderWriter OrderWriter,
	stockWriter StockWriter,
	userReader UserReader,
	addressReader AddressReader,
	couponService CouponService,
	variantReader VariantReader,
	productReader ProductReader,
	sellerReader SellerReader,
	objectReader ObjectReader,
	storageProvider storage.StorageProvider,
	settings SettingsProvider,
	runner database.Runner,
	orderNotifier OrderNotifier,
	adminNotifier AdminOrderNotifier,
	logger log.Logger,
) *CheckoutService {
	if cartReader == nil {
		panic("checkout: cart reader is required")
	}
	if cartWriter == nil {
		panic("checkout: cart writer is required")
	}
	if orderWriter == nil {
		panic("checkout: order writer is required")
	}
	if stockWriter == nil {
		panic("checkout: stock writer is required")
	}
	if userReader == nil {
		panic("checkout: user reader is required")
	}
	if addressReader == nil {
		panic("checkout: address reader is required")
	}
	if couponService == nil {
		panic("checkout: coupon service is required")
	}
	if variantReader == nil {
		panic("checkout: variant reader is required")
	}
	if productReader == nil {
		panic("checkout: product reader is required")
	}
	if sellerReader == nil {
		panic("checkout: seller reader is required")
	}
	if objectReader == nil {
		panic("checkout: object reader is required")
	}
	if storageProvider == nil {
		panic("checkout: storage provider is required")
	}
	if settings == nil {
		panic("checkout: settings provider is required")
	}
	if runner == nil {
		panic("checkout: database runner is required")
	}
	if orderNotifier == nil {
		panic("checkout: order notifier is required")
	}
	if adminNotifier == nil {
		panic("checkout: admin notifier is required")
	}
	if logger == nil {
		panic("checkout: logger is required")
	}
	return &CheckoutService{
		cartReader:    cartReader,
		cartWriter:    cartWriter,
		orderWriter:   orderWriter,
		stockWriter:   stockWriter,
		userReader:    userReader,
		addressReader: addressReader,
		couponService: couponService,
		variantReader: variantReader,
		productReader: productReader,
		sellerReader:  sellerReader,
		objectReader:  objectReader,
		storage:       storageProvider,
		settings:      settings,
		runner:        runner,
		orderNotifier: orderNotifier,
		adminNotifier: adminNotifier,
		logger:        logger,
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
	cfg, err := s.settings.GetConfig(ctx)
	if err != nil {
		return setting.DefaultConfig()
	}
	return cfg
}

// CheckoutInput carries everything a single checkout use case needs from the
// caller; all business rules live inside Checkout.
type CheckoutInput struct {
	CustomerID        uuid.UUID
	ShippingAddressID uuid.UUID
	CouponCode        *string
	CouponID          *uuid.UUID
	PaymentMethod     string
}

// pricedCart is the authoritative result of re-pricing a cart against live
// catalog data: order-ready line items and the server-computed gross total.
type pricedCart struct {
	items      []order.CreateOrderItemInput
	grossTotal int64
	// sellerTotals accumulates line totals per selling seller so merchant
	// coupons can be priced against the eligible subtotal only.
	sellerTotals map[uuid.UUID]int64
	currency     string
}

// appliedCoupon captures the resolved coupon state checkout needs while the
// order is being composed and redeemed.
type appliedCoupon struct {
	id       *uuid.UUID
	code     *string
	currency string
	discount int64
}

// resolvedPayment is the normalized payment-method decision derived from the
// request and the admin-configured availability.
type resolvedPayment struct {
	normalizedName string
	isCOD          bool
	codFee         int64
}

// Checkout places the customer's order end-to-end.
//
// One transaction covers cart lock, authoritative pricing, coupon validation
// and redemption, stock reservation, order creation, payment recording, and
// cart clearance. Any failure rolls everything back (including reservations),
// so there is no deducted-but-no-order or order-without-payment outcome.
// Each numbered responsibility below lives in its own named helper.
func (s *CheckoutService) Checkout(ctx context.Context, in *CheckoutInput) (*model.Order, error) {
	cfg := s.resolveConfig(ctx)

	var createdOrder *model.Order
	var customerName string

	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		// 1. Lock and load the cart so concurrent checkouts serialize.
		cart, err := s.lockAndLoadCart(ctx, in.CustomerID)
		if err != nil {
			return err
		}

		// 2. Resolve the buyer and shipping address that identify the order.
		buyer, err := s.resolveBuyer(ctx, in.CustomerID)
		if err != nil {
			return err
		}
		customerName = buyer.displayName

		shippingAddr, err := s.resolveShippingAddress(ctx, in.CustomerID, in.ShippingAddressID)
		if err != nil {
			return err
		}

		// 3. Authoritative pricing from live variant/product data.
		priced, err := s.priceCart(ctx, cart)
		if err != nil {
			return err
		}

		// 4. Validate and resolve the requested coupon before it is redeemed.
		coupon, err := s.resolveCoupon(ctx, in, priced)
		if err != nil {
			return err
		}

		// 5. Derive shipping, payment, and tax from the effective store config.
		shippingCost := s.shippingCost(cfg, priced.grossTotal)
		payment, err := s.resolvePaymentMethod(cfg, in.PaymentMethod)
		if err != nil {
			return err
		}
		taxAmount := s.taxAmount(cfg, priced.grossTotal, coupon.discount)

		// 6. Create the order, then 7. reserve stock against it, so each
		//    reservation is linked to the order id and can be committed (on
		//    payment) or released (on cancellation/expiry) deterministically.
		orderIn := &order.CreateOrderInput{
			CustomerID:      &in.CustomerID,
			CustomerEmail:   buyer.email,
			ShippingAddress: buildShippingAddress(shippingAddr),
			CouponID:        coupon.id,
			CouponCode:      coupon.code,
			DiscountAmount:  coupon.discount,
			ShippingAmount:  shippingCost,
			TaxAmount:       taxAmount,
			CODFee:          payment.codFee,
			PaymentMethod:   payment.normalizedName,
			Currency:        priced.currency,
			Locale:          localeFromContext(ctx),
			Items:           priced.items,
		}

		createdOrder, err = s.orderWriter.PlaceOrder(ctx, orderIn)
		if err != nil {
			return err
		}

		if err := s.reserveOrderStock(ctx, cart, createdOrder, priced.items, cfg); err != nil {
			return err
		}

		// Cash-on-delivery orders are confirmed the moment they are placed:
		// payment is collected on delivery, so there is no online payment
		// window to hold stock for. Commit the deduction immediately so the
		// units leave salable stock and the expiry reaper never cancels a
		// legitimate COD order. Card orders keep their hold until the
		// payment is captured.
		if payment.isCOD {
			if err := s.stockWriter.CommitReservationsForOrder(ctx, createdOrder.ID); err != nil {
				return err
			}
		}

		// 8. Record the coupon redemption so the transaction rolls back on failure.
		if err := s.recordCouponRedemption(ctx, coupon, in.CustomerID, createdOrder); err != nil {
			return err
		}

		// 9. Record the payment transaction against the created order.
		if err := s.recordOrderPayment(ctx, createdOrder, payment.normalizedName); err != nil {
			return err
		}

		// 10. Clear the cart inside the same transaction so checkout and cart
		//     clearance commit atomically.
		if err := s.cartWriter.ClearItems(ctx, cart.ID); err != nil {
			return s.handleError(err, "Failed to clear cart")
		}

		return nil
	})

	if err != nil {
		return nil, s.handleError(err, "Checkout failed")
	}

	// Notifications run only after the transaction commits: a Kafka/SMTP
	// failure must never turn a successful order into an error response.
	s.notifyOrderPlaced(ctx, cfg, createdOrder, customerName)

	return createdOrder, nil
}

// buyerContact is the customer identity checkout needs to own an order.
type buyerContact struct {
	email       string
	displayName string
}

// resolveBuyer loads the purchasing user and a display name that degrades
// gracefully when no full name is on file.
func (s *CheckoutService) resolveBuyer(ctx context.Context, userID uuid.UUID) (*buyerContact, error) {
	buyer, err := s.userReader.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.NotFoundError("User not found").WithCode(apierr.CodeNotFound)
		}
		return nil, s.handleError(err, "Failed to resolve user")
	}

	name := strings.TrimSpace(buyer.FullName)
	if name == "" {
		name = buyer.Identifier
	}
	return &buyerContact{email: buyer.Identifier, displayName: name}, nil
}

// lockAndLoadCart acquires the row lock and the cart in one place. Locking
// first serializes concurrent checkouts so two requests cannot both read the
// same items, acquire variant locks in opposite orders, and deadlock (or
// create duplicate orders) before either clears the cart. Any not-found
// outcome is an empty cart to the caller.
func (s *CheckoutService) lockAndLoadCart(ctx context.Context, userID uuid.UUID) (*model.Cart, error) {
	if err := s.cartReader.LockCartByUserID(ctx, userID); err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
		}
		return nil, s.handleError(err, "Failed to lock cart")
	}

	cart, err := s.cartReader.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
		}
		return nil, err
	}
	if cart == nil || cart.ID == uuid.Nil || len(cart.Items) == 0 {
		return nil, apierr.BadRequestError("Cart is empty").WithCode("EMPTY_CART")
	}
	return cart, nil
}

// resolveShippingAddress loads an address owned by the customer; a missing
// address is a failure, not a fallback, so orders never ship to a
// guessed location.
func (s *CheckoutService) resolveShippingAddress(ctx context.Context, userID uuid.UUID, addressID uuid.UUID) (*model.UserAddress, error) {
	addr, err := s.addressReader.GetByIDAndUser(ctx, addressID, userID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.NotFoundError("Shipping address not found").WithCode(apierr.CodeNotFound)
		}
		return nil, s.handleError(err, "Failed to resolve shipping address")
	}
	return addr, nil
}

// priceCart re-prices every cart item against live variant/product data and
// builds the order-line snapshots. Pricing decisions are never taken from the
// client subtotal: the gross total is computed here and thumbnails resolve
// directly from storage (bucket/key), never from SQL.
func (s *CheckoutService) priceCart(ctx context.Context, cart *model.Cart) (*pricedCart, error) {
	result := &pricedCart{currency: "EGP", sellerTotals: map[uuid.UUID]int64{}}
	sellers := make(map[uuid.UUID]*model.Seller)

	for _, item := range cart.Items {
		variant, err := s.variantReader.GetByID(ctx, item.VariantID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return nil, apierr.BadRequestError(
					fmt.Sprintf("Variant %s is no longer available", item.VariantID),
				).WithCode(apierr.CodeInvalidInput)
			}
			return nil, s.handleError(err, "Failed to load variant")
		}
		if variant.Price == nil {
			return nil, apierr.BadRequestError(
				fmt.Sprintf("Variant %s has no defined price", item.VariantID),
			).WithCode(apierr.CodeInvalidInput)
		}
		unitPrice := *variant.Price
		if variant.Currency != nil && *variant.Currency != "" {
			result.currency = *variant.Currency
		}

		product, err := s.productReader.GetByID(ctx, variant.ProductID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return nil, apierr.BadRequestError(
					fmt.Sprintf("Product for variant %s is no longer available", item.VariantID),
				).WithCode(apierr.CodeInvalidInput)
			}
			return nil, s.handleError(err, "Failed to load product")
		}

		thumbnailURL, err := s.resolveThumbnailURL(ctx, variant.ThumbnailObjectID, product.ThumbnailObjectID)
		if err != nil {
			return nil, err
		}

		qty := int64(item.Quantity)
		if unitPrice > 0 && qty > math.MaxInt64/unitPrice {
			return nil, apierr.BadRequestError("Order total exceeds maximum allowable limit")
		}
		itemTotal := unitPrice * qty
		if itemTotal > math.MaxInt64-result.grossTotal {
			return nil, apierr.BadRequestError("Order total exceeds maximum allowable limit")
		}
		result.grossTotal += itemTotal

		// Seller attribution: platform products (nil seller) are unattributed.
		// A seller-owned product is only purchasable while its seller is
		// active; a suspended seller (or soft-deleted profile) blocks checkout
		// so orders cannot be attributed to a merchant unable to fulfill.
		sellerID, sellerName, commissionBps, err := s.resolveSellerTerms(
			ctx, sellers, product.SellerID,
		)
		if err != nil {
			return nil, err
		}

		sellerEarning, commissionAmount := computeSellerSplit(itemTotal, sellerID, commissionBps)
		if sellerID != nil {
			result.sellerTotals[*sellerID] += itemTotal
		}
		result.items = append(result.items, order.CreateOrderItemInput{
			VariantID:         item.VariantID,
			Quantity:          item.Quantity,
			PriceAtPurchase:   unitPrice,
			ProductSnapshot:   buildProductSnapshot(item, variant, product, unitPrice, thumbnailURL, sellerName),
			SellerID:          sellerID,
			SellerEarning:     sellerEarning,
			CommissionRateBps: commissionBpsForLine(sellerID, commissionBps),
			CommissionAmount:  commissionAmount,
		})
	}

	return result, nil
}

// resolveSellerTerms resolves the seller's standing and commission for an
// item, memoizing seller lookups within one cart pricing pass. A
// seller-owned product is only purchasable while its seller exists and is
// active; anything else is a checkout failure (unfulfillable order).
// Returns (sellerID, storeName, commissionBps) for attributed lines.
func (s *CheckoutService) resolveSellerTerms(
	ctx context.Context,
	sellers map[uuid.UUID]*model.Seller,
	productSellerID *uuid.UUID,
) (*uuid.UUID, string, int, error) {
	if productSellerID == nil {
		return nil, "", 0, nil
	}

	seller, ok := sellers[*productSellerID]
	if !ok {
		var err error
		seller, err = s.sellerReader.GetByID(ctx, *productSellerID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return nil, "", 0, apierr.BadRequestError(
					fmt.Sprintf("Product seller %s is no longer available", *productSellerID),
				).WithCode(apierr.CodeInvalidInput)
			}
			return nil, "", 0, s.handleError(err, "Failed to load seller")
		}
		sellers[*productSellerID] = seller
	}

	if seller.Status != model.SellerStatusActive {
		return nil, "", 0, apierr.BadRequestError(
			fmt.Sprintf("Product seller '%s' is currently unavailable", seller.StoreName),
		).WithCode(apierr.CodeInvalidInput)
	}

	return productSellerID, seller.StoreName, seller.CommissionRateBps, nil
}

// computeSellerSplit freezes both sides of the platform commission: the
// seller share and the commission amount. Platform lines have no seller and
// therefore zero commission/earning.
func computeSellerSplit(itemTotal int64, sellerID *uuid.UUID, commissionBps int) (earning, commission int64) {
	if sellerID == nil {
		return 0, 0
	}
	commission = itemTotal * int64(commissionBps) / 10000
	earning = itemTotal - commission
	if earning < 0 {
		return 0, commission
	}
	return earning, commission
}

// computeSellerEarning is retained as the focused unit-test entry point for
// the seller share.
func computeSellerEarning(itemTotal int64, sellerID *uuid.UUID, commissionBps int) int64 {
	earning, _ := computeSellerSplit(itemTotal, sellerID, commissionBps)
	return earning
}

func commissionBpsForLine(sellerID *uuid.UUID, commissionBps int) int {
	if sellerID == nil {
		return 0
	}
	return commissionBps
}

// resolveThumbnailURL turns the variant's (or product's fallback) thumbnail
// object into its public URL. A missing thumbnail object is treated as absent
// and never fails the item.
func (s *CheckoutService) resolveThumbnailURL(ctx context.Context, variantThumbnailID, productThumbnailID *uuid.UUID) (string, error) {
	thumbnailID := variantThumbnailID
	if thumbnailID == nil {
		thumbnailID = productThumbnailID
	}
	if thumbnailID == nil {
		return "", nil
	}

	obj, err := s.objectReader.GetByID(ctx, *thumbnailID)
	if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
		return "", s.handleError(err, "Failed to resolve thumbnail")
	}
	if obj == nil {
		return "", nil
	}
	return s.storage.GetPublicURL(obj.Bucket, obj.Key), nil
}

// buildProductSnapshot freezes the live catalog state into the JSON blob stored
// on the order line, so the order keeps what the customer saw even if the
// variant changes or disappears later. Seller identity (when attributed) is
// frozen alongside product data.
func buildProductSnapshot(item model.CartItem, variant *model.ProductVariant, product *model.Product, unitPrice int64, thumbnailURL, sellerName string) string {
	snapshotMap := make(map[string]any)
	snapshotMap["variantId"] = item.VariantID.String()
	snapshotMap["sku"] = variant.SKU
	snapshotMap["title"] = variant.Title
	if product.Title != "" && !strings.HasPrefix(variant.Title, product.Title) {
		snapshotMap["title"] = product.Title + " - " + variant.Title
	}
	snapshotMap["price"] = unitPrice
	if thumbnailURL != "" {
		snapshotMap["thumbnailUrl"] = thumbnailURL
		snapshotMap["image"] = thumbnailURL
	}
	snapshotMap["productId"] = product.ID.String()
	snapshotMap["productTitle"] = product.Title
	snapshotMap["slug"] = product.Slug
	if sellerName != "" {
		snapshotMap["sellerName"] = sellerName
	}

	snapshotBytes, _ := json.Marshal(snapshotMap)
	return string(snapshotBytes)
}

// resolveCoupon resolves the requested coupon (by direct code or by coupon id)
// and validates/prices it against the authoritative gross total. A coupon
// scoped to a merchant is priced against that merchant's eligible subtotal
// only, and requires at least one eligible item in the cart. An empty
// result with no coupon id means "no coupon requested".
func (s *CheckoutService) resolveCoupon(ctx context.Context, in *CheckoutInput, priced *pricedCart) (*appliedCoupon, error) {
	targetCode := ""
	if in.CouponCode != nil && strings.TrimSpace(*in.CouponCode) != "" {
		targetCode = strings.TrimSpace(*in.CouponCode)
	} else if in.CouponID != nil && *in.CouponID != uuid.Nil {
		resolvedCode, err := s.couponService.GetCodeByID(ctx, *in.CouponID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return nil, apierr.BadRequestError("Coupon not found").WithCode(errcode.CodeCouponNotFound)
			}
			return nil, s.handleError(err, "Failed to resolve coupon")
		}
		targetCode = resolvedCode
	}

	if targetCode == "" {
		return &appliedCoupon{}, nil
	}

	// A merchant coupon discounts only that merchant's items: price it
	// against the eligible subtotal, never the whole cart.
	pricedTotal := priced.grossTotal
	if sellerID, err := s.couponService.GetSellerIDByCode(ctx, targetCode); err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
			return nil, apierr.BadRequestError("Coupon not found").WithCode(errcode.CodeCouponNotFound)
		}
		return nil, s.handleError(err, "Failed to resolve coupon scope")
	} else if sellerID != nil {
		pricedTotal = priced.sellerTotals[*sellerID]
		if pricedTotal == 0 {
			return nil, apierr.BadRequestError("This coupon does not apply to the items in your cart").
				WithCode(errcode.CodeCouponNotFound)
		}
	}

	applied, err := s.couponService.ApplyByCode(ctx, targetCode, pricedTotal, &in.CustomerID)
	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.BadRequestError("Coupon not found").WithCode(errcode.CodeCouponNotFound)
		}
		return nil, s.handleError(err, "Failed to validate coupon")
	}

	return &appliedCoupon{
		id:       &applied.ID,
		code:     &applied.Code,
		currency: applied.Currency,
		discount: applied.DiscountAmount,
	}, nil
}

// shippingCost applies the admin-configured shipping rule: free at/above the
// configured threshold, otherwise the configured flat rate.
func (s *CheckoutService) shippingCost(cfg setting.Config, grossTotal int64) int64 {
	if threshold := cfg.FreeShippingThresholdMinorUnits(); grossTotal > 0 && grossTotal < threshold {
		return cfg.ShippingMinorUnits()
	}
	return 0
}

// resolvePaymentMethod validates the requested payment method against the
// admin-configured payment availability and applies the COD fee where it
// applies. An empty method means cash on delivery.
func (s *CheckoutService) resolvePaymentMethod(cfg setting.Config, requestedMethod string) (*resolvedPayment, error) {
	paymentMethod := strings.ToLower(strings.TrimSpace(requestedMethod))
	isCOD := paymentMethod == "" || paymentMethod == "cod" || paymentMethod == "cash_on_delivery"

	if isCOD && !cfg.AllowCOD {
		return nil, apierr.BadRequestError("Cash on delivery is currently unavailable").
			WithCode("PAYMENT_METHOD_UNAVAILABLE")
	}
	if !isCOD && !cfg.AllowCards {
		return nil, apierr.BadRequestError("Online card payments are currently unavailable").
			WithCode("PAYMENT_METHOD_UNAVAILABLE")
	}

	normalizedName := "cash_on_delivery"
	if !isCOD {
		normalizedName = paymentMethod
	}

	codFee := int64(0)
	if isCOD {
		codFee = cfg.CODFeeMinorUnits()
	}
	return &resolvedPayment{normalizedName: normalizedName, isCOD: isCOD, codFee: codFee}, nil
}

// taxAmount applies tax unless prices are already tax-inclusive. Tax is
// computed on the discounted subtotal so it is never charged on discounts.
func (s *CheckoutService) taxAmount(cfg setting.Config, grossTotal, totalDiscount int64) int64 {
	if cfg.PricesIncludeTax || cfg.TaxRatePercent <= 0 {
		return 0
	}
	taxable := grossTotal - totalDiscount
	if taxable < 0 {
		taxable = 0
	}
	return int64(math.Round(float64(taxable) * cfg.TaxRatePercent / 100.0))
}

// buildShippingAddress projects the customer's stored address onto the order
// address shape, filling optional fields only when present.
func buildShippingAddress(shippingAddr *model.UserAddress) *model.Address {
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
	return addr
}

// localeFromContext returns the request locale, normalized to one of the two
// supported notification locales.
func localeFromContext(ctx context.Context) string {
	locale := middleware.GetLocaleFromContext(ctx)
	if locale != "ar" && locale != "en" {
		locale = "en"
	}
	return locale
}

// reserveOrderStock holds stock for every order line, linked to the order. The
// hold is committed into a sale when payment is captured and released if the
// order is cancelled or the hold expires. All reservations live in the caller's
// transaction and roll back with it on any failure.
//
// Reservations happen in a deterministic (sorted) order so two concurrent
// checkouts that share variants cannot deadlock by acquiring the same row
// locks in opposite orders (mirrors CommitReservationsForOrder).
func (s *CheckoutService) reserveOrderStock(ctx context.Context, cart *model.Cart, createdOrder *model.Order, orderItems []order.CreateOrderItemInput, cfg setting.Config) error {
	holds, err := s.stockHolds(ctx, orderItems)
	if err != nil {
		return err
	}

	for _, hold := range holds {
		if _, err := s.stockWriter.HoldStock(ctx, inventory.ReserveStockInput{
			VariantID: hold.variantID,
			CartID:    &cart.ID,
			OrderID:   &createdOrder.ID,
			Quantity:  hold.quantity,
			Duration:  time.Duration(cfg.ReservationTTLMinutes()) * time.Minute,
		}); err != nil {
			return err
		}
	}
	return nil
}

// stockHold is one variant quantity checkout must reserve.
type stockHold struct {
	variantID uuid.UUID
	quantity  int
}

// stockHolds maps order lines onto the stock they consume. A kit line holds its
// components (line quantity × component quantity) because a bundle carries no
// inventory of its own; an ordinary line holds its own variant. Holds are merged
// per variant and returned sorted by variant UUID so concurrent checkouts that
// share variants acquire row locks in the same order and cannot deadlock.
func (s *CheckoutService) stockHolds(
	ctx context.Context,
	orderItems []order.CreateOrderItemInput,
) ([]stockHold, error) {
	quantities := make(map[uuid.UUID]int, len(orderItems))
	for _, item := range orderItems {
		variant, err := s.variantReader.GetByID(ctx, item.VariantID)
		if err != nil {
			return nil, apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(err)
		}

		composition, err := s.productReader.GetBundleComposition(ctx, variant.ProductID)
		if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, fmt.Errorf("load bundle composition for variant %s: %w", item.VariantID, err)
		}
		if composition == nil || len(composition.Items) == 0 {
			quantities[item.VariantID] += item.Quantity
			continue
		}

		for _, component := range composition.Items {
			if component.StockVariantID == nil || component.Quantity <= 0 {
				return nil, apierr.BadRequestError("A bundle item is no longer available").
					WithCode(errcode.CodeInsufficientInventory)
			}
			quantities[*component.StockVariantID] += item.Quantity * component.Quantity
		}
	}

	holds := make([]stockHold, 0, len(quantities))
	for variantID, quantity := range quantities {
		holds = append(holds, stockHold{variantID: variantID, quantity: quantity})
	}
	sort.Slice(holds, func(i, j int) bool {
		return holds[i].variantID.String() < holds[j].variantID.String()
	})
	return holds, nil
}

// recordCouponRedemption persists the coupon redemption against the created
// order. Errors propagate so the transaction rolls back.
func (s *CheckoutService) recordCouponRedemption(ctx context.Context, coupon *appliedCoupon, userID uuid.UUID, createdOrder *model.Order) error {
	if coupon.id == nil || coupon.discount <= 0 {
		return nil
	}

	redemptionCurrency := coupon.currency
	if redemptionCurrency == "" {
		redemptionCurrency = createdOrder.Currency
	}
	return s.couponService.RecordRedemption(
		ctx, *coupon.id, userID, createdOrder.ID, coupon.discount, redemptionCurrency,
	)
}

// recordOrderPayment records the pending payment transaction tied to the order
// before checkout commits. Errors propagate so the transaction rolls back.
func (s *CheckoutService) recordOrderPayment(ctx context.Context, createdOrder *model.Order, paymentMethod string) error {
	providerTxID := fmt.Sprintf("COD-%s", createdOrder.ID.String()[:8])
	if paymentMethod != "cash_on_delivery" {
		providerTxID = fmt.Sprintf("PAY-%s", createdOrder.ID.String()[:8])
	}

	_, err := s.orderWriter.RecordPayment(ctx, &order.CreatePaymentInput{
		OrderID:               createdOrder.ID,
		Amount:                createdOrder.TotalAmount,
		Currency:              createdOrder.Currency,
		Status:                model.StatusPending,
		Type:                  paymentMethod,
		Provider:              paymentMethod,
		ProviderTransactionID: &providerTxID,
	})
	if err != nil {
		return s.handleError(err, "Failed to record payment")
	}
	return nil
}

// notifyOrderPlaced publishes the order confirmation after the transaction
// commits. A Kafka/SMTP failure is logged and never turns a successful order
// into an error response.
func (s *CheckoutService) notifyOrderPlaced(ctx context.Context, cfg setting.Config, createdOrder *model.Order, customerName string) {
	if !cfg.NotifyOrderConfirmation {
		return
	}

	notifyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if notifyErr := s.orderNotifier.NotifyOrderPlaced(
		notifyCtx,
		buildOrderPlacedPayload(createdOrder, customerName),
	); notifyErr != nil {
		s.logger.Warn("failed to publish order placed notification", log.Meta{
			"orderId": createdOrder.ID.String(),
			"error":   notifyErr.Error(),
		})
	}

	orderNum := notifier.OrderNumber(createdOrder.ID)
	if err := s.adminNotifier.Publish(notifyCtx, notifier.AdminOrderEvent{
		ID:            createdOrder.ID,
		OrderNumber:   orderNum,
		CustomerEmail: createdOrder.CustomerEmail,
		CustomerName:  customerName,
		TotalAmount:   createdOrder.TotalAmount,
		Currency:      createdOrder.Currency,
		Status:        string(createdOrder.Status),
		CreatedAt:     createdOrder.CreatedAt,
	}); err != nil {
		s.logger.Warn("failed to broadcast order to admin stream", log.Meta{
			"orderId": createdOrder.ID.String(),
			"error":   err.Error(),
		})
	}
}

// buildOrderPlacedPayload projects the committed order into the notification
// contract. Snapshot parsing is best-effort: a malformed historical snapshot
// must not block an already-committed order confirmation.
func buildOrderPlacedPayload(o *model.Order, customerName string) notifier.OrderPlacedPayload {
	orderNumber := notifier.OrderNumber(o.ID)

	items := make([]notifier.OrderItemSummary, 0, len(o.Items))
	for _, item := range o.Items {
		title := "Item"
		sku := ""
		var snapshot struct {
			Title string `json:"title"`
			SKU   string `json:"sku"`
		}
		if err := json.Unmarshal([]byte(item.ProductSnapshot), &snapshot); err == nil {
			if snapshot.Title != "" {
				title = snapshot.Title
			}
			sku = snapshot.SKU
		}

		items = append(items, notifier.OrderItemSummary{
			Title:     title,
			SKU:       sku,
			Quantity:  item.Quantity,
			UnitPrice: item.PriceAtPurchase,
			LineTotal: item.PriceAtPurchase * int64(item.Quantity),
		})
	}

	var address notifier.OrderAddressSummary
	if o.ShippingAddress != nil {
		address = notifier.OrderAddressSummary{
			Street:     o.ShippingAddress.Street,
			City:       o.ShippingAddress.City,
			State:      o.ShippingAddress.State,
			PostalCode: o.ShippingAddress.PostalCode,
			Country:    o.ShippingAddress.Country,
		}
	}

	return notifier.OrderPlacedPayload{
		OrderID:         o.ID,
		OrderNumber:     orderNumber,
		CustomerID:      o.CustomerID,
		CustomerEmail:   o.CustomerEmail,
		CustomerName:    customerName,
		TotalAmount:     o.TotalAmount,
		Currency:        o.Currency,
		PaymentMethod:   o.PaymentMethod,
		Locale:          o.Locale,
		Items:           items,
		ShippingAddress: address,
	}
}

type CouponValidationResult struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	DiscountType    string    `json:"discountType"`
	DiscountValue   int64     `json:"discountValue"`
	Currency        string    `json:"currency"`
	DiscountPercent *int      `json:"discountPercent,omitempty"`
	FixedDiscount   *float64  `json:"fixedDiscount,omitempty"`
	MinCartAmount   *int64    `json:"minCartAmount,omitempty"`
	DiscountAmount  int64     `json:"discountAmount"`
	Subtotal        int64     `json:"subtotal"`
	Shipping        int64     `json:"shipping"`
	Total           int64     `json:"total"`
}

func (s *CheckoutService) ValidateCoupon(ctx context.Context, code string, userID *uuid.UUID, sessionID *string) (*CouponValidationResult, error) {
	cfg := s.resolveConfig(ctx)

	var result *CouponValidationResult

	// One read connection for the whole use case: cart pricing, coupon lookup and
	// redemption counts all run on the same QueryExecutor (no connection hopping).
	err := s.runner.WithTx(ctx, func(ctx context.Context) error {
		// The cart is optional here: validating a coupon without a cart still
		// works and prices against a zero subtotal.
		var grossTotal int64
		var cartData *model.Cart
		if userID != nil || sessionID != nil {
			var err error
			cartData, err = s.loadCartForOwner(ctx, userID, sessionID)
			if err != nil {
				return err
			}
			grossTotal, err = s.lenientCartTotal(ctx, cartData)
			if err != nil {
				return err
			}
		}

		// A merchant coupon prices against that merchant's eligible
		// subtotal, mirroring checkout.
		if sellerID, err := s.couponService.GetSellerIDByCode(ctx, code); err != nil {
			var apiErr *apierr.APIError
			if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
				return apierr.BadRequestError("Coupon not found").WithCode(errcode.CodeCouponNotFound)
			}
			return err
		} else if sellerID != nil {
			eligible, err := s.lenientSellerTotal(ctx, cartData, *sellerID)
			if err != nil {
				return err
			}
			if eligible == 0 {
				return apierr.BadRequestError("This coupon does not apply to the items in your cart").
					WithCode(errcode.CodeCouponNotFound)
			}
			grossTotal = eligible
		}

		// Client-supplied subtotals are never trusted for pricing decisions.
		applied, err := s.couponService.ApplyByCode(ctx, code, grossTotal, userID)
		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.BadRequestError(fmt.Sprintf("Coupon '%s' not found", code)).WithCode(errcode.CodeCouponNotFound)
			}
			return s.handleError(err, "Failed to validate coupon")
		}

		shipping := s.shippingCost(cfg, grossTotal)

		finalTotal := grossTotal - applied.DiscountAmount + shipping
		if finalTotal < 0 {
			finalTotal = 0
		}

		result = &CouponValidationResult{
			ID:              applied.ID,
			Code:            applied.Code,
			DiscountType:    applied.DiscountType,
			DiscountValue:   applied.DiscountValue,
			Currency:        applied.Currency,
			DiscountPercent: applied.DiscountPercent,
			FixedDiscount:   applied.FixedDiscount,
			MinCartAmount:   applied.MinCartAmount,
			DiscountAmount:  applied.DiscountAmount,
			Subtotal:        grossTotal,
			Shipping:        shipping,
			Total:           finalTotal,
		}
		return nil
	})

	if err != nil {
		return nil, s.handleError(err, "Failed to validate coupon")
	}

	return result, nil
}

// loadCartForOwner loads the cart belonging to the logged-in user or the guest
// session. A missing cart is treated as absent (nil) because coupon validation
// still works against a zero subtotal.
func (s *CheckoutService) loadCartForOwner(ctx context.Context, userID *uuid.UUID, sessionID *string) (*model.Cart, error) {
	var (
		cartData *model.Cart
		err      error
	)
	if userID != nil {
		cartData, err = s.cartReader.GetCartByUserID(ctx, *userID)
	} else {
		cartData, err = s.cartReader.GetCartBySessionID(ctx, *sessionID)
	}
	if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
		return nil, s.handleError(err, "Failed to load cart")
	}
	if err != nil || cartData == nil || cartData.ID == uuid.Nil {
		return nil, nil
	}
	return cartData, nil
}

// lenientCartTotal prices the cart for display purposes (coupon preview).
// Unlike priceCart it does not hard-fail: the snapshot price is a fallback
// when a variant has no live price left, and missing variants are skipped,
// so stale carts still get their coupon preview.
func (s *CheckoutService) lenientCartTotal(ctx context.Context, cart *model.Cart) (int64, error) {
	if cart == nil {
		return 0, nil
	}

	var grossTotal int64
	for _, item := range cart.Items {
		v, vErr := s.variantReader.GetByID(ctx, item.VariantID)
		if vErr == nil && v.Price != nil {
			grossTotal += int64(item.Quantity) * *v.Price
			continue
		}
		if vErr != nil && !errors.Is(database.MapError(vErr), database.ErrNotFound) {
			return 0, s.handleError(vErr, "Failed to load variant price")
		}
		grossTotal += int64(item.Quantity) * item.PriceAtPurchase
	}
	return grossTotal, nil
}

// lenientSellerTotal mirrors lenientCartTotal but only sums lines belonging
// to the given merchant, for pricing seller-scoped coupons in previews.
func (s *CheckoutService) lenientSellerTotal(ctx context.Context, cart *model.Cart, sellerID uuid.UUID) (int64, error) {
	if cart == nil {
		return 0, nil
	}

	var total int64
	for _, item := range cart.Items {
		v, vErr := s.variantReader.GetByID(ctx, item.VariantID)
		if vErr != nil {
			if !errors.Is(database.MapError(vErr), database.ErrNotFound) {
				return 0, s.handleError(vErr, "Failed to load variant price")
			}
			continue
		}
		product, pErr := s.productReader.GetByID(ctx, v.ProductID)
		if pErr != nil {
			if !errors.Is(database.MapError(pErr), database.ErrNotFound) {
				return 0, s.handleError(pErr, "Failed to load product")
			}
			continue
		}
		if product.SellerID == nil || *product.SellerID != sellerID {
			continue
		}
		if v.Price != nil {
			total += int64(item.Quantity) * *v.Price
		} else {
			total += int64(item.Quantity) * item.PriceAtPurchase
		}
	}
	return total, nil
}
