package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/catalog/cart/errcode"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type VariantService interface {
	GetVariantByID(ctx context.Context, variantID uuid.UUID) (*model.ProductVariant, error)
	ListVariantMedia(ctx context.Context, variantID uuid.UUID) ([]*model.VariantMedia, error)
}

type ProductService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
}

type InventoryService interface {
	GetStockByVariantID(ctx context.Context, variantID uuid.UUID) (*model.InventoryStock, error)
	GetStockForVariants(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID]*model.InventoryStock, error)
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

func (s *CartService) GetOrCreateCart(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
) (*model.Cart, error) {
	var cart *model.Cart

	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		var err error
		cart, err = s.getOrCreateCartWithTx(ctx, tx, userID, sessionID)
		return err
	})

	if err != nil {
		return nil, err
	}

	return cart, nil
}

// GetOrCreateCartWithTx resolves (or creates) the customer's cart within the
// caller's transaction, taking a FOR UPDATE row lock so concurrent checkouts
// for the same cart serialize instead of racing into duplicate orders.
func (s *CartService) GetOrCreateCartWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	userID *uuid.UUID,
	sessionID *string,
) (*model.Cart, error) {
	return s.getOrCreateCartWithTx(ctx, tx, userID, sessionID)
}

func (s *CartService) getOrCreateCartWithTx(
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

	variantIDs := make([]uuid.UUID, 0, len(items))
	for i := range items {
		v, vErr := s.variantService.GetVariantByID(ctx, items[i].VariantID)
		if vErr == nil {
			items[i].Variant = v
			variantIDs = append(variantIDs, v.ID)

			// Fetch the associated product
			p, pErr := s.productService.GetByID(ctx, v.ProductID)
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
		stocks, stockErr := s.inventoryService.GetStockForVariants(ctx, variantIDs)
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
// the caller's transaction, so cart clearance commits atomically with the order.
func (s *CartService) ClearCartWithTx(
	ctx context.Context,
	tx database.QueryExecutor,
	userID *uuid.UUID,
	sessionID *string,
) error {
	cart, err := s.getOrCreateCartWithTx(ctx, tx, userID, sessionID)
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

	txErr := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
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
				stock, stockErr := s.inventoryService.GetStockByVariantID(ctx, v.ID)
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
			stock, stockErr := s.inventoryService.GetStockByVariantID(ctx, v.ID)
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

	err = s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		item, err := s.cartRepo.GetItemByID(ctx, tx, cart.ID, itemID)
		if err != nil {
			return err
		}

		if s.inventoryService != nil {
			stock, stockErr := s.inventoryService.GetStockByVariantID(ctx, item.VariantID)
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
	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
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
				stock, stockErr := s.inventoryService.GetStockByVariantID(ctx, item.VariantID)
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
