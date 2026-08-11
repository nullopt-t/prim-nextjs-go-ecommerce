package cart

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/cart/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type VariantService interface {
	GetVariantByID(ctx context.Context, variantID uuid.UUID) (*model.ProductVariant, error)
	ListVariantMedia(ctx context.Context, variantID uuid.UUID) ([]*model.VariantMedia, error)
}

type ProductService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
}

type CartService struct {
	dr             database.Runner
	cartRepo       *CartRepository
	variantService VariantService
	productService ProductService
}

func NewService(
	dr database.Runner,
	cartRepo *CartRepository,
	variantService VariantService,
	productService ProductService,
) *CartService {
	return &CartService{
		dr:             dr,
		cartRepo:       cartRepo,
		variantService: variantService,
		productService: productService,
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
		switch {
		case userID != nil:
			cart, err = s.cartRepo.GetCartByUserID(ctx, tx, *userID)
		case sessionID != nil:
			cart, err = s.cartRepo.GetCartBySessionID(ctx, tx, *sessionID)
		default:
			return apierr.ErrBadRequest("User ID or Session ID is required").
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

		items, itemsErr := s.cartRepo.GetCartItems(ctx, tx, cart.ID)
		if itemsErr != nil && !errors.Is(database.MapError(itemsErr), database.ErrNotFound) {
			return itemsErr
		}

		for i := range items {
			v, vErr := s.variantService.GetVariantByID(ctx, items[i].VariantID)
			if vErr == nil {
				items[i].Variant = v
				
				// Fetch the associated product
				p, pErr := s.productService.GetByID(ctx, v.ProductID)
				if pErr == nil {
					items[i].Product = p
				}

				// Fetch variant media for the thumbnail
				media, mediaErr := s.variantService.ListVariantMedia(ctx, v.ID)
				if mediaErr == nil && len(media) > 0 && media[0].Object != nil {
					items[i].ThumbnailURL = media[0].Object.PublicURL
				}
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

func (s *CartService) AddItem(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
	variantPublicID uuid.UUID,
	quantity int,
) (*model.Cart, error) {
	v, err := s.variantService.GetVariantByID(ctx, variantPublicID)
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
		switch {
		case userID != nil:
			cart, err = s.cartRepo.GetCartByUserID(ctx, tx, *userID)
		case sessionID != nil:
			cart, err = s.cartRepo.GetCartBySessionID(ctx, tx, *sessionID)
		default:
			return apierr.ErrBadRequest("User ID or Session ID is required").
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
		if getErr == nil {
			// Item exists, update quantity
			err = s.cartRepo.UpdateItemQuantity(ctx, tx, cart.ID, existingItem.ID, existingItem.Quantity+quantity)
			return err
		} else if !errors.Is(database.MapError(getErr), database.ErrNotFound) {
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
