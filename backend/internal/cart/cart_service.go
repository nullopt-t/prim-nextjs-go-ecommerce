package cart

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/cart/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/variant"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type CartService struct {
	dr             database.Runner
	cartRepo       *CartRepository
	variantService *variant.VariantService
}

func NewService(
	dr database.Runner,
	cartRepo *CartRepository,
	variantService *variant.VariantService,
) *CartService {
	return &CartService{
		dr:             dr,
		cartRepo:       cartRepo,
		variantService: variantService,
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
		if userID != nil {
			cart, err = s.cartRepo.GetCartByUserID(ctx, tx, *userID)
		} else if sessionID != nil {
			cart, err = s.cartRepo.GetCartBySessionID(ctx, tx, *sessionID)
		} else {
			return apierr.ErrBadRequest("User ID or Session ID is required").
				WithCode(apierr.CodeInvalidInput)
		}

		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				now := time.Now().UTC()
				newCart := &model.Cart{
					ID:        uuid.New(),
					UserID:    userID,
					SessionID: sessionID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if createErr := s.cartRepo.CreateCart(ctx, tx, newCart); createErr != nil {
					return createErr
				}
				cart = newCart
			} else {
				return err
			}
		}

		items, itemsErr := s.cartRepo.GetCartItems(ctx, tx, cart.ID)
		if itemsErr != nil && !errors.Is(database.MapError(itemsErr), database.ErrNotFound) {
			return itemsErr
		}

		for i := range items {
			v, vErr := s.variantService.GetVariantByID(ctx, items[i].VariantID)
			if vErr == nil {
				items[i].Variant = v
			}
		}

		cart.Items = items
		return nil
	})

	if err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) AddItemToCart(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
	variantID uuid.UUID,
	quantity int,
) (*model.Cart, error) {
	v, err := s.variantService.GetVariantByID(ctx, variantID)
	if err != nil {
		return nil, apierr.ErrNotFound("Variant not found").
			WithCode(errcode.CodeVariantNotFound).
			Wrap(err)
	}

	price := int64(0)
	if v.Price != nil {
		price = *v.Price
	}
	currency := "USD"
	if v.Currency != nil {
		currency = *v.Currency
	}

	var cart *model.Cart

	txErr := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		var err error
		if userID != nil {
			cart, err = s.cartRepo.GetCartByUserID(ctx, tx, *userID)
		} else if sessionID != nil {
			cart, err = s.cartRepo.GetCartBySessionID(ctx, tx, *sessionID)
		} else {
			return apierr.ErrBadRequest("User ID or Session ID is required").
				WithCode(apierr.CodeInvalidInput)
		}

		if err != nil {
			if errors.Is(database.MapError(err), database.ErrNotFound) {
				now := time.Now().UTC()
				newCart := &model.Cart{
					ID:        uuid.New(),
					UserID:    userID,
					SessionID: sessionID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if createErr := s.cartRepo.CreateCart(ctx, tx, newCart); createErr != nil {
					return createErr
				}
				cart = newCart
			} else {
				return err
			}
		}

		existingItem, getErr := s.cartRepo.GetItemByVariantID(ctx, tx, cart.ID, variantID)
		if getErr != nil && !errors.Is(database.MapError(getErr), database.ErrNotFound) {
			return getErr
		}

		if existingItem != nil {
			newQty := existingItem.Quantity + quantity
			return s.cartRepo.UpdateItemQuantity(ctx, tx, cart.ID, existingItem.ID, newQty)
		}

		newItem := &model.CartItem{
			ID:              uuid.New(),
			CartID:          cart.ID,
			VariantID:       variantID,
			Quantity:        quantity,
			PriceAtPurchase: price,
			Currency:        currency,
			CartedAt:        time.Now().UTC(),
		}
		return s.cartRepo.AddItem(ctx, tx, newItem)
	})

	if txErr != nil {
		return nil, apierr.ErrInternalError("Failed to add item to cart").
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
	cart, err := s.GetOrCreateCart(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	err = s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.cartRepo.UpdateItemQuantity(ctx, db, cart.ID, itemID, quantity)
	})

	if err != nil {
		if errors.Is(database.MapError(err), database.ErrNotFound) {
			return nil, apierr.ErrNotFound("Cart item not found").
				WithCode(errcode.CodeCartItemNotFound).
				Wrap(err)
		}
		return nil, apierr.ErrInternalError("Failed to update cart item quantity").
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
			return nil, apierr.ErrNotFound("Cart item not found").
				WithCode(errcode.CodeCartItemNotFound).
				Wrap(err)
		}
		return nil, apierr.ErrInternalError("Failed to remove cart item").
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
		return apierr.ErrInternalError("Failed to clear cart").
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
) (*model.Cart, error) {
	err := s.dr.WithTx(ctx, func(tx database.QueryExecutor) error {
		guestCart, gErr := s.cartRepo.GetCartBySessionID(ctx, tx, sessionID)
		if gErr != nil {
			if errors.Is(database.MapError(gErr), database.ErrNotFound) {
				return nil
			}
			return gErr
		}

		userCart, uErr := s.cartRepo.GetCartByUserID(ctx, tx, userID)
		if uErr != nil {
			if errors.Is(database.MapError(uErr), database.ErrNotFound) {
				now := time.Now().UTC()
				newCart := &model.Cart{
					ID:        uuid.New(),
					UserID:    &userID,
					CreatedAt: now,
					UpdatedAt: now,
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
			if existing != nil {
				if err := s.cartRepo.UpdateItemQuantity(ctx, tx, userCart.ID, existing.ID, existing.Quantity+item.Quantity); err != nil {
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
		return nil, apierr.ErrInternalError("Failed to merge guest cart").
			WithCode(errcode.CodeMergeCartFailed).
			Wrap(err).
			WithStack()
	}

	return s.GetOrCreateCart(ctx, &userID, nil)
}
