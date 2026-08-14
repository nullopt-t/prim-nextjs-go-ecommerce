package review

import (
	"context"

	"github.com/google/uuid"
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

func (s *ReviewService) CreateReview(
	ctx context.Context,
	productID uuid.UUID,
	userID uuid.UUID,
	orderItemID uuid.UUID,
	rating int16,
	title *string,
	body *string,
) (*model.Review, error) {
	rv := &model.Review{
		ID:          uuid.New(),
		ProductID:   productID,
		UserID:      userID,
		OrderItemID: orderItemID,
		Rating:      rating,
		Title:       title,
		Body:        body,
		Status:      model.ReviewStatusPending,
	}

	var err error
	err = s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.Create(ctx, db, rv)
	})
	if err != nil {
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

func (s *ReviewService) UpdateReviewStatus(
	ctx context.Context,
	id uuid.UUID,
	status model.ReviewStatus,
) error {
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

	return err
}

func (s *ReviewService) DeleteReview(ctx context.Context, id uuid.UUID) error {
	err := s.dbRunner.WithTx(ctx, func(tx database.QueryExecutor) error {
		rv, err := s.repo.GetByID(ctx, tx, id)
		if err != nil {
			return apierr.ErrInternalError("failed to get review by id").Wrap(err)
		}
		if rv == nil {
			return apierr.ErrNotFound("review not found")
		}

		err = s.repo.Delete(ctx, tx, id)
		if err != nil {
			return apierr.ErrInternalError("failed to delete review").Wrap(err)
		}
		return nil
	})

	return err
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
