package cart

import (
	"context"
	"errors"
	"slices"

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

func (s *CartService) GetOrCreateCart(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		var err error
		if userID != nil {
			cart, err = s.cartRepo.GetCartByUserID(ctx, tx, *userID)
			if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
				return err
			}
		} else if sessionID != nil {
			cart, err = s.cartRepo.GetCartBySessionID(ctx, tx, *sessionID)
			if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
				return err
			}
		}

		if cart == nil || cart.ID == uuid.Nil {
			cart = &model.Cart{
				ID:        uuid.New(),
				UserID:    userID,
				SessionID: sessionID,
			}
			if err := s.cartRepo.CreateCart(ctx, tx, cart); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetCartByID(ctx context.Context, id uuid.UUID) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		cart, err = s.cartRepo.GetCartByID(ctx, db, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		cart, err = s.cartRepo.GetCartByUserID(ctx, db, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetCartBySessionID(ctx context.Context, sessionID string) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		cart, err = s.cartRepo.GetCartBySessionID(ctx, db, sessionID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) AddItem(
	ctx context.Context,
	cartID uuid.UUID,
	variantID uuid.UUID,
	quantity int,
) (*model.Cart, error) {
	if quantity <= 0 {
		return nil, apierr.BadRequestError("Quantity must be greater than 0").
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
			return nil, apierr.InternalServerError("Failed to check inventory").Wrap(stockErr)
		}
		if stock == nil || stock.AvailableQuantity < quantity {
			return nil, apierr.BadRequestError("Insufficient inventory available").WithCode(errcode.CodeInsufficientInventory)
		}
	}

	price := *v.Price
	currency := "EGP"
	if v.Currency != nil && *v.Currency != "" {
		currency = *v.Currency
	}

	txErr := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return s.cartRepo.AddItem(ctx, tx.(*database.Tx), &model.CartItem{
			CartID:          cartID,
			VariantID:       v.ID,
			Quantity:        quantity,
			PriceAtPurchase: price,
			Currency:        currency,
		})
	})
	if txErr != nil {
		return nil, apierr.InternalServerError("Failed to add item to cart").
			WithCode(apierr.CodeInternalError).
			Wrap(txErr).
			WithStack()
	}
	return s.GetCartByID(ctx, cartID)
}

// UpdateCartItemQuantity updates a line item by its database identity. The
// variant is read from the stored cart item rather than trusted from the client.
func (s *CartService) UpdateCartItemQuantity(
	ctx context.Context,
	cartID uuid.UUID,
	itemID int64,
	quantity int,
) (*model.Cart, error) {
	if quantity <= 0 {
		return nil, apierr.BadRequestError("Quantity must be greater than 0").
			WithCode(errcode.CodeInvalidQuantity)
	}

	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return s.cartRepo.UpdateItemQuantity(ctx, tx.(*database.Tx), cartID, itemID, quantity)
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

	return s.GetCartByID(ctx, cartID)
}

func (s *CartService) RemoveCartItem(
	ctx context.Context,
	cartID uuid.UUID,
	itemID int64,
) (*model.Cart, error) {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return s.cartRepo.RemoveItem(ctx, tx.(*database.Tx), cartID, itemID)
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

	return s.GetCartByID(ctx, cartID)
}

func (s *CartService) ClearCart(
	ctx context.Context,
	cartID uuid.UUID,
) error {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx database.QueryExecutor) error {
		return s.cartRepo.ClearCart(ctx, tx, cartID)
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
		guestCart, gErr := s.cartRepo.GetCartBySessionID(ctx, tx, sessionID)
		if gErr != nil {
			if errors.Is(database.MapError(gErr), database.ErrNotFound) {
				return nil
			}
			return gErr
		}

		userCart, err := s.cartRepo.GetCartByUserID(ctx, tx, userID)
		if err != nil && !errors.Is(database.MapError(err), database.ErrNotFound) {
			return err
		}

		if userCart == nil {
			userCart = &model.Cart{
				UserID:    &userID,
				SessionID: nil,
				Items: slices.Clone(guestCart.Items),
			}
			uErr := s.cartRepo.CreateCart(ctx, tx, userCart)
			if uErr != nil {
				return uErr
			}
			return nil
		}

		for _, guestItem := range slices.Clone(guestCart.Items) {
			guestItem.CartID = userCart.ID
			aErr := s.cartRepo.AddItem(ctx, tx.(*database.Tx),& guestItem)
			if aErr != nil {
				return aErr
			}
		}

		return s.cartRepo.DeleteCart(ctx, tx.(*database.Tx), guestCart.ID)
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
