package review

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/catalog/errcode"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

type ReviewService struct {
	dbRunner database.Runner
	repo     *ReviewRepository
}

func NewService(dbRunner database.Runner, repo *ReviewRepository) *ReviewService {
	return &ReviewService{
		dbRunner: dbRunner,
		repo:     repo,
	}
}

type CreateReviewInput struct {
	ProductID   uuid.UUID
	UserID      uuid.UUID
	OrderItemID uuid.UUID
	Rating      int16
	Title       *string
	Body        *string
}

type UpdateReviewInput struct {
	Rating *int16
	Title  *string
	Body   *string
}

func (s *ReviewService) CreateReview(
	ctx context.Context,
	in CreateReviewInput,
) (*model.Review, error) {
	if in.Rating < 1 || in.Rating > 5 {
		return nil, apierr.ErrValidationFailed("rating must be between 1 and 5")
	}

	rv := &model.Review{
		ID:          uuid.New(),
		ProductID:   in.ProductID,
		UserID:      in.UserID,
		OrderItemID: in.OrderItemID,
		Rating:      in.Rating,
		Title:       in.Title,
		Body:        in.Body,
		Status:      model.ReviewStatusPending,
	}

	var apiErr *apierr.APIError
	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		// 1. Verify purchase
		pv, err := s.repo.VerifyPurchase(ctx, tx, in.OrderItemID)
		if err != nil {
			return apierr.ErrInternalError("failed to verify purchase").Wrap(err)
		}
		if pv == nil {
			return apierr.ErrNotFound("order item not found")
		}
		if pv.CustomerID == nil || *pv.CustomerID != in.UserID {
			return apierr.ErrForbidden("you can only review products from your own orders")
		}
		if pv.ProductID != in.ProductID {
			return apierr.ErrBadRequest("order item does not belong to the specified product")
		}
		if pv.OrderStatus == model.OrderStatusCanceled || pv.OrderStatus == model.OrderStatusRefunded {
			return apierr.ErrBadRequest("cannot review items from canceled or refunded orders")
		}

		// 2. Check if already reviewed
		existing, err := s.repo.GetByOrderItemID(ctx, tx, in.OrderItemID)
		if err != nil {
			return apierr.ErrInternalError("failed to check existing review").Wrap(err)
		}
		if existing != nil {
			return apierr.ErrConflict("a review has already been submitted for this order item").
				WithCode(errcode.CodeReviewAlreadyExists)
		}

		// 3. Create review
		if err := s.repo.Create(ctx, tx, rv); err != nil {
			return apierr.ErrInternalError("failed to create review").Wrap(err)
		}
		return nil
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, apierr.ErrInternalError("failed to create review").Wrap(err)
	}

	return rv, nil
}

func (s *ReviewService) GetReviewByID(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	var rv *model.Review
	var err error
	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		rv, err = s.repo.GetByID(ctx, db, id)
		return err
	})
	if err != nil {
		return nil, apierr.ErrInternalError("failed to get review by id").Wrap(err)
	}
	if rv == nil {
		return nil, apierr.ErrNotFound("review not found")
	}
	return rv, nil
}

func (s *ReviewService) GetRatingSummary(ctx context.Context, productID uuid.UUID) (*model.RatingSummary, error) {
	var summary *model.RatingSummary
	var err error
	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		summary, err = s.repo.GetRatingSummaryByProductID(ctx, db, productID)
		return err
	})
	if err != nil {
		return nil, apierr.ErrInternalError("failed to get rating summary").Wrap(err)
	}
	return summary, nil
}

func (s *ReviewService) GetRatingSummariesByProductIDs(
	ctx context.Context,
	productIDs []uuid.UUID,
) (map[uuid.UUID]*model.RatingSummary, error) {
	var summaries map[uuid.UUID]*model.RatingSummary
	var err error
	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		summaries, err = s.repo.GetRatingSummariesByProductIDs(ctx, db, productIDs)
		return err
	})
	if err != nil {
		return nil, apierr.ErrInternalError("failed to get rating summaries").Wrap(err)
	}
	return summaries, nil
}

func (s *ReviewService) UpdateUserReview(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
	in UpdateReviewInput,
) (*model.Review, error) {
	if in.Rating != nil && (*in.Rating < 1 || *in.Rating > 5) {
		return nil, apierr.ErrValidationFailed("rating must be between 1 and 5")
	}

	var updatedReview *model.Review
	var apiErr *apierr.APIError

	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		rv, err := s.repo.GetByID(ctx, tx, id)
		if err != nil {
			return apierr.ErrInternalError("failed to get review by id").Wrap(err)
		}
		if rv == nil {
			return apierr.ErrNotFound("review not found")
		}
		if rv.UserID != userID {
			return apierr.ErrForbidden("you can only update your own review")
		}

		if in.Rating != nil {
			rv.Rating = *in.Rating
		}
		if in.Title != nil {
			rv.Title = in.Title
		}
		if in.Body != nil {
			rv.Body = in.Body
		}
		rv.Status = model.ReviewStatusPending

		if err := s.repo.Update(ctx, tx, rv); err != nil {
			return apierr.ErrInternalError("failed to update review").Wrap(err)
		}
		updatedReview = rv
		return nil
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			return nil, apiErr
		}
		return nil, apierr.ErrInternalError("failed to update review").Wrap(err)
	}

	return updatedReview, nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	var apiErr *apierr.APIError

	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		rv, err := s.repo.GetByID(ctx, tx, id)
		if err != nil {
			return apierr.ErrInternalError("failed to get review by id").Wrap(err)
		}
		if rv == nil {
			return apierr.ErrNotFound("review not found")
		}
		if !isAdmin && rv.UserID != userID {
			return apierr.ErrForbidden("you can only delete your own review")
		}

		err = s.repo.Delete(ctx, tx, id)
		if err != nil {
			return apierr.ErrInternalError("failed to delete review").Wrap(err)
		}
		return nil
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.ErrInternalError("failed to delete review").Wrap(err)
	}

	return nil
}

func (s *ReviewService) UpdateReviewStatus(
	ctx context.Context,
	id uuid.UUID,
	status model.ReviewStatus,
) error {
	var apiErr *apierr.APIError

	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		rv, err := s.repo.GetByID(ctx, tx, id)
		if err != nil {
			return apierr.ErrInternalError("failed to get review by id").Wrap(err)
		}
		if rv == nil {
			return apierr.ErrNotFound("review not found")
		}

		err = s.repo.UpdateStatus(ctx, tx, id, status)
		if err != nil {
			return apierr.ErrInternalError("failed to update review status").Wrap(err)
		}
		return nil
	})

	if err != nil {
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.ErrInternalError("failed to update review status").Wrap(err)
	}

	return nil
}

func (s *ReviewService) ListReviews(
	ctx context.Context,
	q *pagination.ListQuery,
	productID *uuid.UUID,
	userID *uuid.UUID,
	status *model.ReviewStatus,
) (*pagination.PagedResult[model.Review], error) {
	var reviews []model.Review
	var total int
	var err error

	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		reviews, total, err = s.repo.List(ctx, db, q, productID, userID, status)
		return err
	})

	if err != nil {
		return nil, apierr.ErrInternalError("failed to list reviews").Wrap(err)
	}

	reviewPtrs := make([]*model.Review, len(reviews))
	for i := range reviews {
		reviewPtrs[i] = &reviews[i]
	}

	page := pagination.NewPage(q.Page, q.PageSize, total)
	return pagination.NewPagedResult(reviewPtrs, page), nil
}
