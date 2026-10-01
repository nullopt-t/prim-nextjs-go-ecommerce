package wishlist

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

type ProductService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
}

type ObjectService interface {
	GetObjectURL(ctx context.Context, bucket, key string) string
}

type WishlistService struct {
	dbRunner       database.Runner
	repo           *WishlistRepository
	productService ProductService
	objectService  ObjectService
	logger         log.Logger
}

func NewService(
	dbRunner database.Runner,
	repo *WishlistRepository,
	productService ProductService,
	objectService ObjectService,
	logger log.Logger,
) *WishlistService {
	return &WishlistService{
		dbRunner:       dbRunner,
		repo:           repo,
		productService: productService,
		objectService:  objectService,
		logger:         logger,
	}
}

func (s *WishlistService) AddToWishlist(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) (*model.WishlistItem, error) {
	if s.productService != nil {
		prod, err := s.productService.GetByID(ctx, productID)
		if err != nil {
			var apiErr *apierr.APIError
			if errors.As(err, &apiErr) {
				return nil, apiErr
			}
			return nil, apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound).
				Wrap(err)
		}
		if prod == nil {
			return nil, apierr.ErrNotFound("Product not found").
				WithCode(errcode.CodeProductNotFound)
		}
		if prod.Status != model.PublicationStatusPublished {
			return nil, apierr.ErrBadRequest("Product is not available for wishlisting").
				WithCode(apierr.CodeInvalidInput)
		}
	}

	newItem := &model.WishlistItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		VariantID: variantID,
	}

	var createdItem *model.WishlistItem
	var apiErr *apierr.APIError

	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		if err := s.repo.Create(ctx, tx, newItem); err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrConflict) {
				return apierr.ErrConflict("Product already in wishlist").
					WithCode(errcode.CodeWishlistItemAlreadyExists)
			}
			if errors.Is(mappedErr, database.ErrForeignKeyViolation) {
				return apierr.ErrNotFound("Product not found").
					WithCode(errcode.CodeProductNotFound)
			}
			return apierr.ErrInternalError("Failed to add product to wishlist").
				Wrap(err)
		}

		item, err := s.repo.GetByID(ctx, tx, newItem.ID, userID)
		if err != nil {
			return apierr.ErrInternalError("Failed to retrieve created wishlist item").
				Wrap(err)
		}
		if item == nil {
			return apierr.ErrNotFound("Wishlist item not found").
				WithCode(errcode.CodeWishlistItemNotFound)
		}

		s.resolveThumbnail(ctx, item)
		createdItem = item
		return nil
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, apierr.ErrInternalError("Failed to add product to wishlist").
			Wrap(err)
	}

	return createdItem, nil
}

func (s *WishlistService) GetMyWishlist(
	ctx context.Context,
	userID uuid.UUID,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.WishlistItem], error) {
	var items []model.WishlistItem
	var total int
	var err error

	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		items, total, err = s.repo.ListByUser(ctx, db, userID, q)
		return err
	})

	if err != nil {
		return nil, apierr.ErrInternalError("Failed to retrieve wishlist").Wrap(err)
	}

	itemPtrs := make([]*model.WishlistItem, len(items))
	for i := range items {
		s.resolveThumbnail(ctx, &items[i])
		itemPtrs[i] = &items[i]
	}

	page := pagination.NewPage(q.Page, q.PageSize, total)
	return pagination.NewPagedResult(itemPtrs, page), nil
}

func (s *WishlistService) GetWishlistCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	var count int
	var err error

	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		count, err = s.repo.CountByUser(ctx, db, userID)
		return err
	})

	if err != nil {
		return 0, apierr.ErrInternalError("Failed to count wishlist items").Wrap(err)
	}

	return count, nil
}

func (s *WishlistService) CheckInWishlist(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) (bool, *uuid.UUID, error) {
	var exists bool
	var itemID *uuid.UUID
	var err error

	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		exists, itemID, err = s.repo.Exists(ctx, db, userID, productID, variantID)
		return err
	})

	if err != nil {
		return false, nil, apierr.ErrInternalError("Failed to check wishlist status").Wrap(err)
	}

	return exists, itemID, nil
}

func (s *WishlistService) RemoveItem(
	ctx context.Context,
	userID uuid.UUID,
	itemID uuid.UUID,
) error {
	var apiErr *apierr.APIError

	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		err := s.repo.DeleteByID(ctx, tx, itemID, userID)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return apierr.ErrNotFound("Wishlist item not found").
					WithCode(errcode.CodeWishlistItemNotFound)
			}
			return apierr.ErrInternalError("Failed to remove item from wishlist").Wrap(err)
		}
		return nil
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.ErrInternalError("Failed to remove item from wishlist").Wrap(err)
	}

	return nil
}

func (s *WishlistService) RemoveByProductID(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	variantID *uuid.UUID,
) error {
	var apiErr *apierr.APIError

	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		err := s.repo.DeleteByProduct(ctx, tx, userID, productID, variantID)
		if err != nil {
			mappedErr := database.MapError(err)
			if errors.Is(mappedErr, database.ErrNotFound) {
				return apierr.ErrNotFound("Product not found in wishlist").
					WithCode(errcode.CodeWishlistItemNotFound)
			}
			return apierr.ErrInternalError("Failed to remove product from wishlist").Wrap(err)
		}
		return nil
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.ErrInternalError("Failed to remove product from wishlist").Wrap(err)
	}

	return nil
}

func (s *WishlistService) ClearWishlist(
	ctx context.Context,
	userID uuid.UUID,
) error {
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.ClearByUser(ctx, db, userID)
	})

	if err != nil {
		return apierr.ErrInternalError("Failed to clear wishlist").Wrap(err)
	}

	return nil
}

func (s *WishlistService) resolveThumbnail(ctx context.Context, item *model.WishlistItem) {
	if item.ThumbnailBucket != nil && item.ThumbnailKey != nil && *item.ThumbnailBucket != "" && *item.ThumbnailKey != "" && s.objectService != nil {
		url := s.objectService.GetObjectURL(ctx, *item.ThumbnailBucket, *item.ThumbnailKey)
		item.ThumbnailURL = &url
	}
}
