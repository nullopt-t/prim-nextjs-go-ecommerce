package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

type NotificationService struct {
	dbRunner database.Runner
	repo     *NotificationRepository
	logger   log.Logger
}

func NewNotificationService(
	dbRunner database.Runner,
	repo *NotificationRepository,
	logger log.Logger,
) *NotificationService {
	return &NotificationService{
		dbRunner: dbRunner,
		repo:     repo,
		logger:   logger,
	}
}

// CreateNotification saves a notification to PostgreSQL.
func (s *NotificationService) CreateNotification(
	ctx context.Context,
	n *model.UserNotification,
) error {
	return s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		return s.repo.Create(ctx, db, n)
	})
}

// GetMyNotifications returns a paginated list of notifications for the user.
func (s *NotificationService) GetMyNotifications(
	ctx context.Context,
	userID uuid.UUID,
	unreadOnly bool,
	q *pagination.ListQuery,
) (*pagination.PagedResult[model.UserNotification], error) {
	var items []model.UserNotification
	var total int

	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		items, total, err = s.repo.ListByUser(ctx, db, userID, unreadOnly, q)
		return err
	})
	if err != nil {
		return nil, apierr.ErrInternalError("Failed to fetch notifications").Wrap(err)
	}

	itemPtrs := make([]*model.UserNotification, len(items))
	for i := range items {
		itemPtrs[i] = &items[i]
	}

	page := pagination.NewPage(q.Page, q.PageSize, total)
	return pagination.NewPagedResult(itemPtrs, page), nil
}

// GetUnreadCount returns the number of unread alerts for the customer badge.
func (s *NotificationService) GetUnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	var count int
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		count, err = s.repo.CountUnreadByUser(ctx, db, userID)
		return err
	})
	if err != nil {
		return 0, apierr.ErrInternalError("Failed to fetch unread notification count").Wrap(err)
	}
	return count, nil
}

// MarkAsRead marks a single notification as read.
func (s *NotificationService) MarkAsRead(
	ctx context.Context,
	id, userID uuid.UUID,
) error {
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		err := s.repo.MarkAsRead(ctx, db, id, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apierr.ErrNotFound("Notification not found")
			}
			return err
		}
		return nil
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.ErrInternalError("Failed to mark notification as read").Wrap(err)
	}
	return nil
}

// MarkAllAsRead marks all user notifications as read.
func (s *NotificationService) MarkAllAsRead(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	var count int64
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		var err error
		count, err = s.repo.MarkAllAsRead(ctx, db, userID)
		return err
	})
	if err != nil {
		return 0, apierr.ErrInternalError("Failed to mark notifications as read").Wrap(err)
	}
	return count, nil
}

// DeleteNotification removes a notification.
func (s *NotificationService) DeleteNotification(
	ctx context.Context,
	id, userID uuid.UUID,
) error {
	err := s.dbRunner.WithDB(ctx, func(db database.QueryExecutor) error {
		err := s.repo.Delete(ctx, db, id, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apierr.ErrNotFound("Notification not found")
			}
			return err
		}
		return nil
	})
	if err != nil {
		var apiErr *apierr.APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return apierr.ErrInternalError("Failed to delete notification").Wrap(err)
	}
	return nil
}
