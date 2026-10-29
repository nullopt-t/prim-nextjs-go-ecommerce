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
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

type VariantReader interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.ProductVariant, error)
	ListMediaByVariantIDs(ctx context.Context, qe database.QueryExecutor, variantIDs []uuid.UUID) (map[uuid.UUID][]*model.VariantMedia, error)
}

type ProductReader interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Product, error)
}

type InventoryReader interface {
	GetStock(ctx context.Context, qe database.QueryExecutor, variantID uuid.UUID) (*model.InventoryStock, error)
	GetStockForVariants(ctx context.Context, qe database.QueryExecutor, variantIDs []uuid.UUID) (map[uuid.UUID]*model.InventoryStock, error)
}

type ObjectReader interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Object, error)
}

type CartService struct {
	dr              database.Runner
	cartRepo        *CartRepository
	variantReader   VariantReader
	productReader   ProductReader
	inventoryReader InventoryReader
	objectReader    ObjectReader
	storageProvider storage.StorageProvider
}

func NewService(
	dr database.Runner,
	cartRepo *CartRepository,
	variantReader VariantReader,
	productReader ProductReader,
	inventoryReader InventoryReader,
	objectReader ObjectReader,
	storageProvider storage.StorageProvider,
) *CartService {
	return &CartService{
		dr:              dr,
		cartRepo:        cartRepo,
		variantReader:   variantReader,
		productReader:   productReader,
		inventoryReader: inventoryReader,
		objectReader:    objectReader,
		storageProvider: storageProvider,
	}
}

func (s *CartService) populateCartDetails(
	ctx context.Context,
	qe database.QueryExecutor,
	cart *model.Cart,
) error {
	if cart == nil || len(cart.Items) == 0 {
		return nil
	}

	variantIDs := make([]uuid.UUID, 0, len(cart.Items))
	for i := range cart.Items {
		item := &cart.Items[i]
		variantIDs = append(variantIDs, item.VariantID)

		if s.variantReader != nil {
			v, err := s.variantReader.GetByID(ctx, qe, item.VariantID)
			if err == nil && v != nil {
				item.Variant = v

				if s.objectReader != nil && v.ThumbnailObjectID != nil && *v.ThumbnailObjectID != uuid.Nil {
					obj, oErr := s.objectReader.GetByID(ctx, qe, *v.ThumbnailObjectID)
					if oErr == nil && obj != nil {
						v.Thumbnail = obj
						if s.storageProvider != nil {
							item.ThumbnailURL = s.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
							obj.PublicURL = item.ThumbnailURL
						}
					}
				}

				if item.ThumbnailURL == "" && v.Thumbnail != nil {
					if v.Thumbnail.PublicURL != "" {
						item.ThumbnailURL = v.Thumbnail.PublicURL
					} else if s.storageProvider != nil && v.Thumbnail.Bucket != "" && v.Thumbnail.Key != "" {
						item.ThumbnailURL = s.storageProvider.GetPublicURL(v.Thumbnail.Bucket, v.Thumbnail.Key)
						v.Thumbnail.PublicURL = item.ThumbnailURL
					}
				}

				if s.productReader != nil && v.ProductID != uuid.Nil {
					p, pErr := s.productReader.GetByID(ctx, qe, v.ProductID)
					if pErr == nil && p != nil {
						item.Product = p

						if item.ThumbnailURL == "" {
							if s.objectReader != nil && p.ThumbnailObjectID != nil && *p.ThumbnailObjectID != uuid.Nil {
								obj, oErr := s.objectReader.GetByID(ctx, qe, *p.ThumbnailObjectID)
								if oErr == nil && obj != nil {
									p.Thumbnail = obj
									if s.storageProvider != nil {
										item.ThumbnailURL = s.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
										obj.PublicURL = item.ThumbnailURL
									}
								}
							}

							if item.ThumbnailURL == "" && p.Thumbnail != nil {
								if p.Thumbnail.PublicURL != "" {
									item.ThumbnailURL = p.Thumbnail.PublicURL
								} else if s.storageProvider != nil && p.Thumbnail.Bucket != "" && p.Thumbnail.Key != "" {
									item.ThumbnailURL = s.storageProvider.GetPublicURL(p.Thumbnail.Bucket, p.Thumbnail.Key)
									p.Thumbnail.PublicURL = item.ThumbnailURL
								}
							}
						}
					}
				}
			}
		}
	}

	if s.variantReader != nil && len(variantIDs) > 0 {
		var needsMedia []uuid.UUID
		for i := range cart.Items {
			if cart.Items[i].ThumbnailURL == "" {
				needsMedia = append(needsMedia, cart.Items[i].VariantID)
			}
		}
		if len(needsMedia) > 0 {
			mediaMap, mErr := s.variantReader.ListMediaByVariantIDs(ctx, qe, needsMedia)
			if mErr == nil && mediaMap != nil {
				for i := range cart.Items {
					if cart.Items[i].ThumbnailURL == "" {
						if list, ok := mediaMap[cart.Items[i].VariantID]; ok && len(list) > 0 {
							if list[0].ObjectID != uuid.Nil && s.objectReader != nil {
								obj, oErr := s.objectReader.GetByID(ctx, qe, list[0].ObjectID)
								if oErr == nil && obj != nil {
									list[0].Object = obj
									if s.storageProvider != nil {
										cart.Items[i].ThumbnailURL = s.storageProvider.GetPublicURL(obj.Bucket, obj.Key)
										obj.PublicURL = cart.Items[i].ThumbnailURL
									}
								}
							}
							if cart.Items[i].ThumbnailURL == "" && list[0].Object != nil {
								if list[0].Object.PublicURL != "" {
									cart.Items[i].ThumbnailURL = list[0].Object.PublicURL
								} else if s.storageProvider != nil && list[0].Object.Bucket != "" && list[0].Object.Key != "" {
									cart.Items[i].ThumbnailURL = s.storageProvider.GetPublicURL(list[0].Object.Bucket, list[0].Object.Key)
									list[0].Object.PublicURL = cart.Items[i].ThumbnailURL
								}
							}
						}
					}
				}
			}
		}
	}

	if s.inventoryReader != nil && len(variantIDs) > 0 {
		stocks, err := s.inventoryReader.GetStockForVariants(ctx, qe, variantIDs)
		if err == nil && stocks != nil {
			for i := range cart.Items {
				if st, ok := stocks[cart.Items[i].VariantID]; ok && st != nil {
					cart.Items[i].AvailableStock = st.AvailableQuantity
					cart.Items[i].InStock = st.AvailableQuantity >= cart.Items[i].Quantity
				}
			}
		}
	}

	return nil
}

func (s *CartService) GetOrCreateCart(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
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
		return s.populateCartDetails(ctx, tx, cart)
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetCartByID(ctx context.Context, id uuid.UUID) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		cart, err = s.cartRepo.GetCartByID(ctx, exec, id)
		if err != nil {
			return err
		}
		return s.populateCartDetails(ctx, exec, cart)
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetCartByUserID(ctx context.Context, userID uuid.UUID) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		cart, err = s.cartRepo.GetCartByUserID(ctx, exec, userID)
		if err != nil {
			return err
		}
		return s.populateCartDetails(ctx, exec, cart)
	})
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GetCartBySessionID(ctx context.Context, sessionID string) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		cart, err = s.cartRepo.GetCartBySessionID(ctx, exec, sessionID)
		if err != nil {
			return err
		}
		return s.populateCartDetails(ctx, exec, cart)
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

	var cart *model.Cart
	txErr := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		v, getErr := s.variantReader.GetByID(ctx, tx, variantID)
		if getErr != nil {
			return apierr.NotFoundError("Variant not found").
				WithCode(errcode.CodeVariantNotFound).
				Wrap(getErr)
		}

		if s.inventoryReader != nil {
			stock, stockErr := s.inventoryReader.GetStock(ctx, tx, v.ID)
			if stockErr != nil {
				var apiErr *apierr.APIError
				if errors.As(stockErr, &apiErr) {
					return apiErr
				}
				return apierr.InternalServerError("Failed to check inventory").Wrap(stockErr)
			}
			if stock == nil || stock.AvailableQuantity < quantity {
				return apierr.BadRequestError("Insufficient inventory available").
					WithCode(errcode.CodeInsufficientInventory)
			}
		}

		price := int64(0)
		if v.Price != nil {
			price = *v.Price
		}
		currency := "EGP"
		if v.Currency != nil && *v.Currency != "" {
			currency = *v.Currency
		}

		if err := s.cartRepo.AddItem(ctx, tx, &model.CartItem{
			CartID:          cartID,
			VariantID:       v.ID,
			Quantity:        quantity,
			PriceAtPurchase: price,
			Currency:        currency,
		}); err != nil {
			return err
		}

		var err error
		cart, err = s.cartRepo.GetCartByID(ctx, tx, cartID)
		if err != nil {
			return err
		}
		return s.populateCartDetails(ctx, tx, cart)
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
	return cart, nil
}

// UpdateCartItemQuantity updates a line item by its database identity.
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

	var cart *model.Cart
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		// Validate the target item and the requested quantity against stock on
		// the same transaction before mutating, so an over-limit update is
		// rejected (INSUFFICIENT_INVENTORY) instead of silently persisted.
		current, err := s.cartRepo.GetCartByID(ctx, tx, cartID)
		if err != nil {
			return err
		}

		var variantID uuid.UUID
		for i := range current.Items {
			if current.Items[i].ID == itemID {
				variantID = current.Items[i].VariantID
				break
			}
		}
		if variantID == uuid.Nil {
			return database.ErrNotFound
		}

		if s.inventoryReader != nil {
			stock, stockErr := s.inventoryReader.GetStock(ctx, tx, variantID)
			if stockErr != nil {
				var apiErr *apierr.APIError
				if errors.As(stockErr, &apiErr) {
					return apiErr
				}
				return apierr.InternalServerError("Failed to check inventory").Wrap(stockErr)
			}
			if stock == nil || stock.AvailableQuantity < quantity {
				return apierr.BadRequestError("Insufficient inventory available").
					WithCode(errcode.CodeInsufficientInventory)
			}
		}

		if err := s.cartRepo.UpdateItemQuantity(ctx, tx, cartID, itemID, quantity); err != nil {
			return err
		}

		cart, err = s.cartRepo.GetCartByID(ctx, tx, cartID)
		if err != nil {
			return err
		}
		return s.populateCartDetails(ctx, tx, cart)
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

	return cart, nil
}

func (s *CartService) RemoveCartItem(
	ctx context.Context,
	cartID uuid.UUID,
	itemID int64,
) (*model.Cart, error) {
	var cart *model.Cart
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		if err := s.cartRepo.RemoveItem(ctx, tx, cartID, itemID); err != nil {
			return err
		}

		var err error
		cart, err = s.cartRepo.GetCartByID(ctx, tx, cartID)
		if err != nil {
			return err
		}
		return s.populateCartDetails(ctx, tx, cart)
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

	return cart, nil
}

// ClearItems empties the cart's line items on the caller's executor. Checkout
// calls this inside its own transaction so cart clearance commits with order
// creation instead of opening a second connection.
func (s *CartService) ClearItems(
	ctx context.Context,
	tx *database.Tx,
	cartID uuid.UUID,
) error {
	return s.cartRepo.ClearCart(ctx, tx, cartID)
}

func (s *CartService) ClearCart(
	ctx context.Context,
	cartID uuid.UUID,
) error {
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		return s.ClearItems(ctx, tx, cartID)
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
	err := s.dr.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
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
				Items:     slices.Clone(guestCart.Items),
			}
			uErr := s.cartRepo.CreateCart(ctx, tx, userCart)
			if uErr != nil {
				return uErr
			}
			return nil
		}

		for _, guestItem := range slices.Clone(guestCart.Items) {
			guestItem.CartID = userCart.ID
			aErr := s.cartRepo.AddItem(ctx, tx, &guestItem)
			if aErr != nil {
				return aErr
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
