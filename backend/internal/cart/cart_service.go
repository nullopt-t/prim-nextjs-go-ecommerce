package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type VariantService interface {
	GetVariantByID(ctx context.Context, variantID uuid.UUID) (*model.ProductVariant, error)
	// GetVariantByIDWithTx reads a variant (and its media) on the caller's
	// transaction executor so no second connection is checked out.
	GetVariantByIDWithTx(ctx context.Context, tx database.QueryExecutor, variantID uuid.UUID) (*model.ProductVariant, error)
	ListVariantMedia(ctx context.Context, variantID uuid.UUID) ([]*model.VariantMedia, error)
}

type ProductService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	// GetByIDWithTx reads a product on the caller's transaction executor so no
	// second connection is checked out.
	GetByIDWithTx(ctx context.Context, tx database.QueryExecutor, id uuid.UUID) (*model.Product, error)
}

type InventoryService interface {
	GetStockByVariantID(ctx context.Context, variantID uuid.UUID) (*model.InventoryStock, error)
	// GetStockByVariantIDWithTx reads stock on the caller's transaction executor
	// so a call made inside a transaction never checks out a second connection.
	GetStockByVariantIDWithTx(ctx context.Context, tx database.QueryExecutor, variantID uuid.UUID) (*model.InventoryStock, error)
	GetStockForVariants(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID]*model.InventoryStock, error)
	// GetStockForVariantsWithTx reads stock on the caller's transaction executor
	// so a call made inside a transaction never checks out a second connection.
	GetStockForVariantsWithTx(ctx context.Context, tx database.QueryExecutor, variantIDs []uuid.UUID) (map[uuid.UUID]*model.InventoryStock, error)
}

type CartService struct {
	dr               database.Runner
	cartRepo         *CartRepository
	variantService   VariantService
	productService   ProductService
	inventoryService InventoryService
}

func NewService(
	dr database.Runner,
	cartRepo *CartRepository,
	variantService VariantService,
	productService ProductService,
	inventoryService InventoryService,
) *CartService {
	return &CartService{
		dr:               dr,
		cartRepo:         cartRepo,
		variantService:   variantService,
		productService:   productService,
		inventoryService: inventoryService,
	}
}

// GetOrCreateCart resolves (or creates) the caller's active cart and
// hydrates its line items with variant, product and stock details.
//
// Only the locking cart lookup runs in a transaction; the hydrating reads
// run on the pool afterwards, so a request never holds a transaction open
// while asking the pool for another connection.
func (s *CartService) GetOrCreateCart(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
) (*model.Cart, error) {
	var cart *model.Cart

	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		var txErr error
		cart, txErr = s.lockCart(ctx, tx, userID, sessionID)
		return txErr
	})
	if err != nil {
		return nil, err
	}

	return s.enrichCart(ctx, cart)
}

// GetOrCreateCartWithTx resolves (or creates) the customer's cart within the
// caller's transaction, taking a FOR UPDATE row lock so concurrent checkouts
// for the same cart serialize instead of racing into duplicate orders.
//
// Line items are hydrated on that same transaction (checkout needs
// authoritative prices), so no second connection is ever checked out while
// the caller's transaction is open.
func (s *CartService) GetOrCreateCartWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	userID *uuid.UUID,
	sessionID *string,
) (*model.Cart, error) {
	cart, err := s.lockCart(ctx, tx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	return s.enrichCartWithTx(ctx, tx, cart)
}

// lockCart resolves the cart owned by the given user or session, creating it
// when it does not exist yet. The FOR UPDATE lock serializes concurrent
// writers for the same cart. It touches only cart rows, so it is safe to
// call inside a caller's transaction.
func (s *CartService) lockCart(
	ctx context.Context,
	tx database.QueryExecutor,
	userID *uuid.UUID,
	sessionID *string,
) (*model.Cart, error) {
	var cart *model.Cart
	var err error
	switch {
	case userID != nil:
		cart, err = s.cartRepo.GetCartByUserIDForUpdate(ctx, tx, *userID)
	case sessionID != nil:
		cart, err = s.cartRepo.GetCartBySessionIDForUpdate(ctx, tx, *sessionID)
	default:
		return nil, apierr.BadRequestError("User ID or Session ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			newCart := &model.Cart{
				ID:        uuid.New(),
				UserID:    userID,
				SessionID: sessionID,
			}
			if createErr := s.cartRepo.CreateCart(ctx, tx, newCart); createErr != nil {
				// Partial unique index on carts(user_id) lets one concurrent
				// creation win; fetch the existing active cart instead of a 500.
				if errors.Is(database.MapError(createErr), database.ErrConflict) && userID != nil {
					if existing, gErr := s.cartRepo.GetCartByUserIDForUpdate(ctx, tx, *userID); gErr == nil {
						cart = existing
					} else if gErr != nil {
						return nil, gErr
					}
				} else {
					return nil, createErr
				}
			} else {
				cart = newCart
			}
		} else {
			return nil, err
		}
	}

	items, itemsErr := s.cartRepo.GetCartItems(ctx, tx, cart.ID)
	if itemsErr != nil && !errors.Is(database.MapError(itemsErr), database.ErrNotFound) {
		return nil, itemsErr
	}

	cart.Items = items
	return cart, nil
}

// enrichCart hydrates a cart's line items using a pooled connection.
func (s *CartService) enrichCart(ctx context.Context, cart *model.Cart) (*model.Cart, error) {
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var enrichErr error
		cart, enrichErr = s.enrichCartWithTx(ctx, db, cart)
		return enrichErr
	})
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// enrichCartWithTx hydrates a cart's line items with variant, product and
// stock detail on the supplied executor, so a caller that holds a
// transaction never checks out a second connection. Lookup failures are
// non-fatal: the item is returned without the missing detail.
func (s *CartService) enrichCartWithTx(
	ctx context.Context,
	exec database.QueryExecutor,
	cart *model.Cart,
) (*model.Cart, error) {
	items := cart.Items

	variantIDs := make([]uuid.UUID, 0, len(items))
	for i := range items {
		v, vErr := s.variantService.GetVariantByIDWithTx(ctx, exec, items[i].VariantID)
		if vErr == nil {
			items[i].Variant = v
			variantIDs = append(variantIDs, v.ID)

			// Fetch the associated product
			p, pErr := s.productService.GetByIDWithTx(ctx, exec, v.ProductID)
			if pErr == nil {
				items[i].Product = p
			}

			// Thumbnail resolution: prefer variant thumbnail or variant media
			if v.Thumbnail != nil && v.Thumbnail.PublicURL != "" {
				items[i].ThumbnailURL = v.Thumbnail.PublicURL
			} else if len(v.Media) > 0 && v.Media[0].Object != nil {
				items[i].ThumbnailURL = v.Media[0].Object.PublicURL
			}
		}
	}

	if len(variantIDs) > 0 && s.inventoryService != nil {
		stocks, stockErr := s.inventoryService.GetStockForVariantsWithTx(ctx, exec, variantIDs)
		if stockErr == nil && stocks != nil {
			for i := range items {
				if st, ok := stocks[items[i].VariantID]; ok {
					items[i].AvailableStock = st.AvailableQuantity
					items[i].InStock = st.AvailableQuantity >= items[i].Quantity
				}
			}
		}
	}

	cart.Items = items
	return cart, nil
}

// ClearCartWithTx soft-deletes all active items of the customer's cart inside
// the caller's transaction, so cart clearance commits atomically with the
// order. It performs no hydration, so it needs no extra connection.
func (s *CartService) ClearCartWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	userID *uuid.UUID,
	sessionID *string,
) error {
	cart, err := s.lockCart(ctx, tx, userID, sessionID)
	if err != nil {
		return err
	}

	if err := s.cartRepo.ClearCart(ctx, tx, cart.ID); err != nil {
		return err
	}

	return nil
}

// MaxCartItemQuantity represents the maximum allowable quantity for any single line item in the cart.
const MaxCartItemQuantity = 999

func (s *CartService) AddItem(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
	variantID uuid.UUID,
	quantity int,
) (*model.Cart, error) {
	if quantity <= 0 || quantity > MaxCartItemQuantity {
		return nil, apierr.BadRequestError("Quantity must be between 1 and 999").
			WithCode(errcode.CodeInvalidQuantity)
	}

	v, err := s.variantService.GetVariantByID(ctx, variantID)
	if err != nil {
		return nil, apierr.NotFoundError("Variant not found").
			WithCode(errcode.CodeVariantNotFound).
			Wrap(err)
	}

	// Validate available inventory (fail-closed)
	if s.inventoryService != nil {
		stock, stockErr := s.inventoryService.GetStockByVariantID(ctx, v.ID)
		if stockErr != nil {
			var apiErr *apierr.APIError
			if errors.As(stockErr, &apiErr) {
				return nil, apiErr
			}
			return nil, apierr.InternalServerError("Failed to check inventory").
				WithCode(apierr.CodeInternalError).
				Wrap(stockErr)
		}
		if stock == nil || stock.AvailableQuantity < quantity {
			return nil, apierr.BadRequestError("Insufficient inventory available").
				WithCode(errcode.CodeInsufficientInventory)
		}
	}

	if v.Price == nil {
		return nil, apierr.BadRequestError("Variant has no defined price").
			WithCode(apierr.CodeInvalidInput)
	}
	price := *v.Price
	currency := "EGP"
	if v.Currency != nil && *v.Currency != "" {
		currency = *v.Currency
	}

	var cart *model.Cart

	txErr := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		var err error
		switch {
		case userID != nil:
			cart, err = s.cartRepo.GetCartByUserIDForUpdate(ctx, tx, *userID)
		case sessionID != nil:
			cart, err = s.cartRepo.GetCartBySessionIDForUpdate(ctx, tx, *sessionID)
		default:
			return apierr.BadRequestError("User ID or Session ID is required").
				WithCode(apierr.CodeInvalidInput)
		}

		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				newCart := &model.Cart{
					ID:        uuid.New(),
					UserID:    userID,
					SessionID: sessionID,
				}
				if createErr := s.cartRepo.CreateCart(ctx, tx, newCart); createErr != nil {
					return createErr
				}
				cart = newCart
			} else {
				return err
			}
		}

		existingItem, getErr := s.cartRepo.GetItemByVariantID(ctx, tx, cart.ID, v.ID)
		if getErr == nil && existingItem != nil {
			newQty := existingItem.Quantity + quantity
			if newQty > MaxCartItemQuantity {
				return apierr.BadRequestError("Total item quantity cannot exceed 999").
					WithCode(errcode.CodeInvalidQuantity)
			}
			if s.inventoryService != nil {
				stock, stockErr := s.inventoryService.GetStockByVariantIDWithTx(ctx, tx, v.ID)
				if stockErr != nil {
					var apiErr *apierr.APIError
					if errors.As(stockErr, &apiErr) {
						return apiErr
					}
					return apierr.InternalServerError("Failed to check inventory").
						WithCode(apierr.CodeInternalError).
						Wrap(stockErr)
				}
				if stock == nil || stock.AvailableQuantity < newQty {
					return apierr.BadRequestError("Insufficient inventory available").
						WithCode(errcode.CodeInsufficientInventory)
				}
			}
		} else if getErr != nil && !errors.Is(database.MapError(getErr), database.ErrNotFound) {
			return getErr
		}

		newItem := &model.CartItem{
			ID:              uuid.New(),
			CartID:          cart.ID,
			VariantID:       v.ID,
			Quantity:        quantity,
			PriceAtPurchase: price,
			Currency:        currency,
		}
		if err := s.cartRepo.AddItem(ctx, tx, newItem); err != nil {
			return err
		}

		if newItem.Quantity > MaxCartItemQuantity {
			return apierr.BadRequestError("Total item quantity cannot exceed 999").
				WithCode(errcode.CodeInvalidQuantity)
		}
		if s.inventoryService != nil {
			stock, stockErr := s.inventoryService.GetStockByVariantIDWithTx(ctx, tx, v.ID)
			if stockErr != nil {
				var apiErr *apierr.APIError
				if errors.As(stockErr, &apiErr) {
					return apiErr
				}
				return apierr.InternalServerError("Failed to check inventory").
					WithCode(apierr.CodeInternalError).
					Wrap(stockErr)
			}
			if stock == nil || stock.AvailableQuantity < newItem.Quantity {
				return apierr.BadRequestError("Insufficient inventory available").
					WithCode(errcode.CodeInsufficientInventory)
			}
		}
		return nil
	})

	if txErr != nil {
		var apiErr *apierr.APIError
		if errors.As(txErr, &apiErr) {
			return nil, apiErr
		}
		return nil, apierr.InternalServerError("Failed to add item to cart").
			WithCode(apierr.CodeInternalError).
			Wrap(txErr).
			WithStack()
	}

	return s.GetOrCreateCart(ctx, userID, sessionID)
}

func (s *CartService) UpdateCartItemQuantity(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
	itemID uuid.UUID,
	quantity int,
) (*model.Cart, error) {
	if quantity <= 0 || quantity > MaxCartItemQuantity {
		return nil, apierr.BadRequestError("Quantity must be between 1 and 999").
			WithCode(errcode.CodeInvalidQuantity)
	}

	cart, err := s.GetOrCreateCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	err = s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		item, err := s.cartRepo.GetItemByID(ctx, tx, cart.ID, itemID)
		if err != nil {
			return err
		}

		if s.inventoryService != nil {
			stock, stockErr := s.inventoryService.GetStockByVariantIDWithTx(ctx, tx, item.VariantID)
			if stockErr != nil {
				var apiErr *apierr.APIError
				if errors.As(stockErr, &apiErr) {
					return apiErr
				}
				return apierr.InternalServerError("Failed to check inventory").
					WithCode(apierr.CodeInternalError).
					Wrap(stockErr)
			}
			if stock == nil || stock.AvailableQuantity < quantity {
				return apierr.BadRequestError("Insufficient inventory available").
					WithCode(errcode.CodeInsufficientInventory)
			}
		}

		return s.cartRepo.UpdateItemQuantity(ctx, tx, cart.ID, itemID, quantity)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.NotFoundError("Cart item not found").
				WithCode(errcode.CodeCartItemNotFound).
				Wrap(err)
		}
		return nil, apierr.InternalServerError("Failed to update cart item quantity").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return s.GetOrCreateCart(ctx, userID, sessionID)
}

func (s *CartService) RemoveCartItem(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
	itemID uuid.UUID,
) (*model.Cart, error) {
	cart, err := s.GetOrCreateCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	err = s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.cartRepo.RemoveItem(ctx, db, cart.ID, itemID)
	})

	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.NotFoundError("Cart item not found").
				WithCode(errcode.CodeCartItemNotFound).
				Wrap(err)
		}
		return nil, apierr.InternalServerError("Failed to remove cart item").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return s.GetOrCreateCart(ctx, userID, sessionID)
}

func (s *CartService) ClearCart(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
) error {
	cart, err := s.GetOrCreateCart(ctx, userID, sessionID)
	if err != nil {
		return err
	}

	err = s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.cartRepo.ClearCart(ctx, db, cart.ID)
	})

	if err != nil {
		return apierr.InternalServerError("Failed to clear cart").
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}

	return nil
}

func (s *CartService) MergeGuestCart(
	ctx context.Context,
	sessionID string,
	userID uuid.UUID,
) error {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		guestCart, gErr := s.cartRepo.GetCartBySessionIDForUpdate(ctx, tx, sessionID)
		if gErr != nil {
			if errors.Is(database.MapError(gErr), database.ErrNotFound) {
				return nil
			}
			return gErr
		}

		userCart, uErr := s.cartRepo.GetCartByUserIDForUpdate(ctx, tx, userID)
		if uErr != nil {
			if errors.Is(database.MapError(uErr), database.ErrNotFound) {
				newCart := &model.Cart{
					ID:     uuid.New(),
					UserID: &userID,
				}
				if cErr := s.cartRepo.CreateCart(ctx, tx, newCart); cErr != nil {
					return cErr
				}
				userCart = newCart
			} else {
				return uErr
			}
		}

		guestItems, err := s.cartRepo.GetCartItems(ctx, tx, guestCart.ID)
		if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
			return err
		}

		for _, item := range guestItems {
			existing, err := s.cartRepo.GetItemByVariantID(ctx, tx, userCart.ID, item.VariantID)
			if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
				return err
			}

			targetQty := item.Quantity
			if existing != nil {
				targetQty = existing.Quantity + item.Quantity
			}
			if targetQty > MaxCartItemQuantity {
				targetQty = MaxCartItemQuantity
			}

			if s.inventoryService != nil {
				stock, stockErr := s.inventoryService.GetStockByVariantIDWithTx(ctx, tx, item.VariantID)
				if stockErr != nil {
					var apiErr *apierr.APIError
					if errors.As(stockErr, &apiErr) {
						return apiErr
					}
					return apierr.InternalServerError("Failed to check inventory").
						WithCode(apierr.CodeInternalError).
						Wrap(stockErr)
				}
				if stock == nil || stock.AvailableQuantity < targetQty {
					return apierr.BadRequestError("Insufficient inventory available for variant").
						WithCode(errcode.CodeInsufficientInventory)
				}
			}

			if existing != nil {
				if err := s.cartRepo.UpdateItemQuantity(ctx, tx, userCart.ID, existing.ID, targetQty); err != nil {
					return err
				}
			} else {
				newItem := &model.CartItem{
					ID:              uuid.New(),
					CartID:          userCart.ID,
					VariantID:       item.VariantID,
					Quantity:        item.Quantity,
					PriceAtPurchase: item.PriceAtPurchase,
					Currency:        item.Currency,
					CartedAt:        item.CartedAt,
				}
				if err := s.cartRepo.AddItem(ctx, tx, newItem); err != nil {
					return err
				}
			}
		}

		return s.cartRepo.DeleteCart(ctx, tx, guestCart.ID)
	})

	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.InternalServerError("Failed to merge guest cart").
			WithCode(errcode.CodeMergeCartFailed).
			Wrap(err).
			WithStack()
	}

	return nil
}
