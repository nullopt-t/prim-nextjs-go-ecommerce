package review

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/errcode"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
)

type ReviewService struct {
	dbRunner database.Runner
	repo     *ReviewRepository
}

func NewService(
	dbRunner database.Runner,
	repo *ReviewRepository,
) *ReviewService {
	if dbRunner == nil {
		panic("review: database runner is required")
	}
	if repo == nil {
		panic("review: review repository is required")
	}
	return &ReviewService{
		dbRunner: dbRunner,
		repo:     repo,
	}
}

// handleError maps repository and database driver errors to typed APIErrors.
func (s *ReviewService) handleError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}

	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	mapped := database.MapError(err)
	switch {
	case errors.Is(mapped, database.ErrNotFound):
		return apierr.NotFoundError("Review not found").
			WithCode(errcode.CodeReviewNotFound).
			Wrap(err)
	case errors.Is(mapped, database.ErrConflict):
		return apierr.ConflictError("A review has already been submitted for this order item").
			WithCode(errcode.CodeReviewAlreadyExists).
			Wrap(err)
	case errors.Is(mapped, database.ErrForeignKeyViolation):
		return apierr.BadRequestError("Referenced record does not exist").
			WithCode(apierr.CodeInvalidInput).
			Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).
			WithCode(apierr.CodeInternalError).
			Wrap(err).
			WithStack()
	}
}

type CreateReviewInput struct {
	ProductID   uuid.UUID
	UserID      uuid.UUID
	OrderItemID *uuid.UUID
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
	if in.ProductID == uuid.Nil || in.UserID == uuid.Nil {
		return nil, apierr.BadRequestError("Product ID and User ID are required").
			WithCode(apierr.CodeInvalidInput)
	}
	if in.Rating < 1 || in.Rating > 5 {
		return nil, apierr.ValidationFailedError("rating must be between 1 and 5")
	}

	var targetOrderItemID uuid.UUID
	var createdReview *model.Review

	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		if in.OrderItemID != nil && *in.OrderItemID != uuid.Nil {
			targetOrderItemID = *in.OrderItemID
			// 1. Verify specific purchase
			pv, err := s.repo.VerifyPurchase(ctx, targetOrderItemID)
			if err != nil {
				return err
			}
			if pv == nil {
				return apierr.NotFoundError("order item not found")
			}
			if pv.CustomerID == nil || *pv.CustomerID != in.UserID {
				return apierr.ForbiddenError("you can only review products from your own orders").
					WithCode(errcode.CodeReviewNotEligible)
			}
			if pv.ProductID != in.ProductID {
				return apierr.BadRequestError("order item does not belong to the specified product").
					WithCode(errcode.CodeReviewNotEligible)
			}
			if pv.OrderStatus != model.OrderStatusDelivered {
				return apierr.ForbiddenError("you can only review products that have been delivered").
					WithCode(errcode.CodeReviewNotEligible)
			}
		} else {
			// Auto-resolve eligible delivered purchase for this user and product
			eligibleID, err := s.repo.FindEligibleOrderItem(ctx, in.UserID, in.ProductID)
			if err != nil {
				return err
			}
			if eligibleID == uuid.Nil {
				return apierr.ForbiddenError("you can only review products you have purchased and received (delivered)").
					WithCode(errcode.CodeReviewNotEligible)
			}
			targetOrderItemID = eligibleID
		}

		// 2. Check if already reviewed
		existing, err := s.repo.GetByOrderItemID(ctx, targetOrderItemID)
		if err != nil {
			return err
		}
		if existing != nil {
			return apierr.ConflictError("a review has already been submitted for this order item").
				WithCode(errcode.CodeReviewAlreadyExists)
		}

		// 3. Create review (initialize as pending for moderation)
		reviewID := uuid.New()
		rv := &model.Review{
			ID:          reviewID,
			ProductID:   in.ProductID,
			UserID:      in.UserID,
			OrderItemID: targetOrderItemID,
			Rating:      in.Rating,
			Title:       in.Title,
			Body:        in.Body,
			Status:      model.ReviewStatusPending,
		}

		if err := s.repo.Create(ctx, rv); err != nil {
			return err
		}

		createdReview = rv
		return nil
	})

	if err != nil {
		return nil, s.handleError(err, "Failed to create review")
	}

	return createdReview, nil
}

func (s *ReviewService) GetReviewByID(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	if id == uuid.Nil {
		return nil, apierr.BadRequestError("Review ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var rv *model.Review
	err := s.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		rv, err = s.repo.GetByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to get review by id")
	}
	if rv == nil {
		return nil, apierr.NotFoundError("Review not found").
			WithCode(errcode.CodeReviewNotFound)
	}

	return rv, nil
}

func (s *ReviewService) GetRatingSummary(ctx context.Context, productID uuid.UUID) (*model.RatingSummary, error) {
	if productID == uuid.Nil {
		return nil, apierr.BadRequestError("Product ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	var summary *model.RatingSummary
	err := s.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		summary, err = s.repo.GetRatingSummaryByProductID(ctx, productID)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to get rating summary")
	}
	return summary, nil
}

func (s *ReviewService) GetRatingSummariesByProductIDs(
	ctx context.Context,
	productIDs []uuid.UUID,
) (map[uuid.UUID]*model.RatingSummary, error) {
	if len(productIDs) == 0 {
		return map[uuid.UUID]*model.RatingSummary{}, nil
	}

	var summaries map[uuid.UUID]*model.RatingSummary
	err := s.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		summaries, err = s.repo.GetRatingSummariesByProductIDs(ctx, productIDs)
		return err
	})
	if err != nil {
		return nil, s.handleError(err, "Failed to get rating summaries")
	}
	return summaries, nil
}

// UpdateUserReview allows a customer to update the rating, headline, or narrative body of their own review.
// Automatically resets moderation status to 'pending' to require moderator re-approval if changes are made.
func (s *ReviewService) UpdateUserReview(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
	in UpdateReviewInput,
) (*model.Review, error) {
	if id == uuid.Nil || userID == uuid.Nil {
		return nil, apierr.BadRequestError("Review ID and User ID are required").
			WithCode(apierr.CodeInvalidInput)
	}
	if in.Rating != nil && (*in.Rating < 1 || *in.Rating > 5) {
		return nil, apierr.ValidationFailedError("rating must be between 1 and 5")
	}

	var updatedReview *model.Review
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		rv, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if rv == nil {
			return apierr.NotFoundError("Review not found").
				WithCode(errcode.CodeReviewNotFound)
		}
		if rv.UserID != userID {
			return apierr.ForbiddenError("You can only update your own review").
				WithCode(apierr.CodeForbidden)
		}

		changed := false
		if in.Rating != nil && *in.Rating != rv.Rating {
			rv.Rating = *in.Rating
			changed = true
		}
		if in.Title != nil && (rv.Title == nil || *in.Title != *rv.Title) {
			rv.Title = in.Title
			changed = true
		}
		if in.Body != nil && (rv.Body == nil || *in.Body != *rv.Body) {
			rv.Body = in.Body
			changed = true
		}

		if !changed {
			updatedReview = rv
			return nil
		}

		rv.Status = model.ReviewStatusPending
		if err := s.repo.Update(ctx, rv); err != nil {
			return err
		}

		updatedReview = rv
		return nil
	})

	if err != nil {
		return nil, s.handleError(err, "Failed to update review")
	}

	return updatedReview, nil
}

func (s *ReviewService) DeleteReview(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	if id == uuid.Nil {
		return apierr.BadRequestError("Review ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		rv, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if rv == nil {
			return apierr.NotFoundError("Review not found").
				WithCode(errcode.CodeReviewNotFound)
		}
		if !isAdmin && rv.UserID != userID {
			return apierr.ForbiddenError("You can only delete your own review").
				WithCode(apierr.CodeForbidden)
		}

		return s.repo.Delete(ctx, id)
	})

	if err != nil {
		return s.handleError(err, "Failed to delete review")
	}

	return nil
}

func (s *ReviewService) UpdateReviewStatus(
	ctx context.Context,
	id uuid.UUID,
	status model.ReviewStatus,
) error {
	if id == uuid.Nil {
		return apierr.BadRequestError("Review ID is required").
			WithCode(apierr.CodeInvalidInput)
	}

	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		rv, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if rv == nil {
			return apierr.NotFoundError("Review not found").
				WithCode(errcode.CodeReviewNotFound)
		}

		return s.repo.UpdateStatus(ctx, id, status)
	})

	if err != nil {
		return s.handleError(err, "Failed to update review status")
	}

	return nil
}

// ListReviews returns the reviews matching the filter, paged by cursor.
func (s *ReviewService) ListReviews(
	ctx context.Context,
	limit int,
	cursor *pagination.Keyset,
	filter model.ReviewFilter,
) ([]model.Review, *pagination.CursorPage, error) {
	var reviews []model.Review
	var page *pagination.CursorPage
	err := s.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		reviews, page, err = s.repo.List(ctx, limit, cursor, filter)
		return err
	})

	if err != nil {
		return nil, nil, s.handleError(err, "Failed to list reviews")
	}

	return reviews, page, nil
}

// ToggleHelpful toggles a helpful vote on a review by the given user.
func (s *ReviewService) ToggleHelpful(
	ctx context.Context,
	reviewID, userID uuid.UUID,
) (bool, int, error) {
	if reviewID == uuid.Nil {
		return false, 0, apierr.BadRequestError("Review ID is required")
	}
	if userID == uuid.Nil {
		return false, 0, apierr.UnauthorizedError()
	}

	var voted bool
	var count int
	err := s.dbRunner.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		voted, count, err = s.repo.ToggleVote(txCtx, reviewID, userID)
		return err
	})
	if err != nil {
		return false, 0, s.handleError(err, "Failed to toggle helpful vote")
	}

	return voted, count, nil
}

// CheckReviewEligibility verifies if a customer has a delivered, unreviewed order item for a product.
func (s *ReviewService) CheckReviewEligibility(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
) (bool, error) {
	if userID == uuid.Nil || productID == uuid.Nil {
		return false, apierr.BadRequestError("User ID and Product ID are required").
			WithCode(apierr.CodeInvalidInput)
	}

	var eligibleID uuid.UUID
	err := s.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		eligibleID, err = s.repo.FindEligibleOrderItem(ctx, userID, productID)
		return err
	})

	if err != nil {
		return false, s.handleError(err, "Failed to verify review eligibility")
	}

	return eligibleID != uuid.Nil, nil
}
