package wishlist

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/storage"
)

// ProductReader defines the subset of product catalog queries required by WishlistService.
type ProductReader interface {
	GetByID(ctx context.Context, qe database.QueryExecutor, id uuid.UUID) (*model.Product, error)
}

// WishlistService encapsulates use-case orchestration and business invariants for user wishlists.
type WishlistService struct {
	dbRunner        database.Runner
	repo            *WishlistRepository
	productReader   ProductReader
	storageProvider storage.StorageProvider
	logger          log.Logger
}

// NewService constructs a new WishlistService instance with decoupled repository and storage dependencies.
func NewService(
	dbRunner database.Runner,
	repo *WishlistRepository,
	productReader ProductReader,
	storageProvider storage.StorageProvider,
	logger log.Logger,
) *WishlistService {
	if dbRunner == nil {
		panic("wishlist: database runner is required")
	}
	if repo == nil {
		panic("wishlist: wishlist repository is required")
	}
	if productReader == nil {
		panic("wishlist: product reader is required")
	}
	if storageProvider == nil {
		panic("wishlist: storage provider is required")
	}
	if logger == nil {
		panic("wishlist: logger is required")
	}
	return &WishlistService{
		dbRunner:        dbRunner,
		repo:            repo,
		productReader:   productReader,
		storageProvider: storageProvider,
		logger:          logger,
	}
}

// handleError maps persistence and database errors to domain APIErrors.
func (s *WishlistService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mappedErr := database.MapError(err)
	switch {
	case errors.Is(err, pgx.ErrNoRows), errors.Is(mappedErr, database.ErrNotFound):
		return apierr.NotFoundError("Wishlist item not found").
			WithCode(errcode.CodeWishlistItemNotFound).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrConflict):
		return apierr.ConflictError("Product already in wishlist").
			WithCode(errcode.CodeWishlistItemAlreadyExists).
			Wrap(err)
	case errors.Is(mappedErr, database.ErrForeignKeyViolation):
		return apierr.NotFoundError("Product not found").
			WithCode(errcode.CodeProductNotFound).
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

// AddToWishlist saves a product or specific variant to the authenticated customer's wishlist.
// Verifies that the product exists, is published, and that any specified variant belongs to the product
// inside a single database transaction with zero connection hopping.
func (s *WishlistService) AddToWishlist(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) (*model.WishlistItem, error) {
	newItem := &model.WishlistItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		VariantID: variantID,
	}

	var createdItem *model.WishlistItem

	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		// 1. Validate product existence and publication status on tx
		prod, err := s.productReader.GetByID(ctx, tx, productID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) || errors.Is(database.MapError(err), database.ErrNotFound) {
				return apierr.NotFoundError("Product not found").
					WithCode(errcode.CodeProductNotFound)
			}
			var apiErr *apierr.APIError
			if errors.As(err, &apiErr) {
				return apiErr
			}
			return s.handleError(err, "Failed to verify product")
		}
		if prod == nil {
			return apierr.NotFoundError("Product not found").
				WithCode(errcode.CodeProductNotFound)
		}
		if prod.Status != model.PublicationStatusPublished {
			return apierr.BadRequestError("Product is not available for wishlisting").
				WithCode(apierr.CodeInvalidInput)
		}

		// 2. Validate variant association on tx
		if variantID != nil {
			variantProdID, err := s.repo.GetVariantProductID(ctx, tx, *variantID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) || errors.Is(database.MapError(err), database.ErrNotFound) {
					return apierr.NotFoundError("Variant not found").
						WithCode(errcode.CodeVariantNotFound)
				}
				return s.handleError(err, "Failed to verify variant")
			}
			if variantProdID != productID {
				return apierr.BadRequestError("Variant does not belong to specified product").
					WithCode(errcode.CodeVariantProductMismatch)
			}
		}

		// 3. Mutate (insert wishlist item)
		if err := s.repo.Create(ctx, tx, newItem); err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrConflict) {
				return apierr.ConflictError("Product already in wishlist").
					WithCode(errcode.CodeWishlistItemAlreadyExists)
			}
			if errors.Is(mappedErr, database.ErrForeignKeyViolation) {
				return apierr.NotFoundError("Product not found").
					WithCode(errcode.CodeProductNotFound)
			}
			return s.handleError(err, "Failed to add product to wishlist")
		}

		// 4. Reload post-mutation state on same tx
		item, err := s.repo.GetByID(ctx, tx, newItem.ID, userID)
		if err != nil {
			return s.handleError(err, "Failed to retrieve created wishlist item")
		}
		if item == nil {
			return apierr.NotFoundError("Wishlist item not found").
				WithCode(errcode.CodeWishlistItemNotFound)
		}

		s.resolveThumbnail(item)
		createdItem = item
		return nil
	})

	if err != nil {
		return nil, s.handleError(err, "Failed to add product to wishlist")
	}

	return createdItem, nil
}

// GetMyWishlist returns a paginated list of wishlist items for the authenticated user.
func (s *WishlistService) GetMyWishlist(
	ctx context.Context,
	userID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.WishlistItem], error) {
	var items []model.WishlistItem
	var total int

	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		items, total, err = s.repo.ListByUser(ctx, exec, userID, q)
		return err
	})

	if err != nil {
		return nil, s.handleError(err, "Failed to retrieve wishlist")
	}

	itemPtrs := make([]*model.WishlistItem, len(items))
	for i := range items {
		s.resolveThumbnail(&items[i])
		itemPtrs[i] = &items[i]
	}

	page := pagination.NewPage(q.Page, q.PageSize, total)
	return pagination.NewPagedResult(itemPtrs, page), nil
}

// GetWishlistCount returns the total number of items in the user's wishlist.
func (s *WishlistService) GetWishlistCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	var count int

	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		count, err = s.repo.CountByUser(ctx, exec, userID)
		return err
	})

	if err != nil {
		return 0, s.handleError(err, "Failed to count wishlist items")
	}

	return count, nil
}

// CheckInWishlist checks whether a product or variant is already present in the user's wishlist.
func (s *WishlistService) CheckInWishlist(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) (bool, *uuid.UUID, error) {
	var exists bool
	var itemID *uuid.UUID

	err := s.dbRunner.WithExecutor(ctx, func(exec database.QueryExecutor) error {
		var err error
		exists, itemID, err = s.repo.Exists(ctx, exec, userID, productID, variantID)
		return err
	})

	if err != nil {
		return false, nil, s.handleError(err, "Failed to check wishlist status")
	}

	return exists, itemID, nil
}

// RemoveItem removes a specific wishlist entry by its ID for the authenticated user.
func (s *WishlistService) RemoveItem(
	ctx context.Context,
	userID uuid.UUID,
	itemID uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		err := s.repo.DeleteByID(ctx, tx, itemID, userID)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return apierr.NotFoundError("Wishlist item not found").
					WithCode(errcode.CodeWishlistItemNotFound)
			}
			return s.handleError(err, "Failed to remove item from wishlist")
		}
		return nil
	})

	if err != nil {
		return s.handleError(err, "Failed to remove item from wishlist")
	}

	return nil
}

// RemoveByProductID removes a product or product/variant combination from the user's wishlist.
func (s *WishlistService) RemoveByProductID(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, tx *database.Tx) error {
		err := s.repo.DeleteByProduct(ctx, tx, userID, productID, variantID)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return apierr.NotFoundError("Product not found in wishlist").
					WithCode(errcode.CodeWishlistItemNotFound)
			}
			return s.handleError(err, "Failed to remove product from wishlist")
		}
		return nil
	})

	if err != nil {
		return s.handleError(err, "Failed to remove product from wishlist")
	}

	return nil
}

// ClearWishlist empties all wishlist items belonging to the authenticated user.
func (s *WishlistService) ClearWishlist(
	ctx context.Context,
	userID uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context, exec *database.Tx) error {
		return s.repo.ClearByUser(ctx, exec, userID)
	})

	if err != nil {
		return s.handleError(err, "Failed to clear wishlist")
	}

	return nil
}

// resolveThumbnail populates the public CDN URL using the injected storage provider.
func (s *WishlistService) resolveThumbnail(item *model.WishlistItem) {
	if item.ThumbnailBucket != nil && item.ThumbnailKey != nil && *item.ThumbnailBucket != "" && *item.ThumbnailKey != "" {
		url := s.storageProvider.GetPublicURL(*item.ThumbnailBucket, *item.ThumbnailKey)
		item.ThumbnailURL = &url
	}
}
