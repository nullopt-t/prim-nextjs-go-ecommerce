package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// notificationStore is the consumer-owned persistence contract for customer
// notifications. *NotificationRepository is the production implementation;
// tests provide an in-memory fake.
type notificationStore interface {
	Create(ctx context.Context, n *model.UserNotification) error
	ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit int, cursor *pagination.Keyset) ([]*model.UserNotification, *pagination.CursorPage, error)
	CountUnreadByUser(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, id int64, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id int64, userID uuid.UUID) error
}

// NotificationService coordinates business operations for customer notifications:
// creation, paginated list retrieval, unread count aggregation, status updates (read),
// and deletion.
type NotificationService struct {
	dbRunner database.Runner
	repo     notificationStore
	logger   log.Logger
}

// NewNotificationService constructs a new NotificationService instance.
func NewNotificationService(
	dbRunner database.Runner,
	repo notificationStore,
	logger log.Logger,
) *NotificationService {
	if dbRunner == nil {
		panic("notification: database runner is required")
	}
	if repo == nil {
		panic("notification: notification repository is required")
	}
	if logger == nil {
		panic("notification: logger is required")
	}
	return &NotificationService{
		dbRunner: dbRunner,
		repo:     repo,
		logger:   logger,
	}
}

// handleError maps persistence and database errors to domain APIErrors.
func (s *NotificationService) handleError(err error, fallbackMsg string) error {
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
		return apierr.NotFoundError("Notification not found").Wrap(err)
	default:
		return apierr.InternalServerError(fallbackMsg).Wrap(err).WithStack()
	}
}

// CreateNotification saves a notification to PostgreSQL.
func (s *NotificationService) CreateNotification(
	ctx context.Context,
	n *model.UserNotification,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		return s.repo.Create(ctx, n)
	})
	if err != nil {
		return s.handleError(err, "Failed to create notification")
	}
	return nil
}

// GetMyNotifications returns a cursor-paginated list of notifications for the user,
// supporting optional filtering for unread notifications only.
func (s *NotificationService) GetMyNotifications(
	ctx context.Context,
	userID uuid.UUID,
	unreadOnly bool,
	limit int,
	cursor *pagination.Keyset,
) ([]*model.UserNotification, *pagination.CursorPage, error) {
	var items []*model.UserNotification
	var page *pagination.CursorPage

	err := s.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		items, page, err = s.repo.ListByUser(ctx, userID, unreadOnly, limit, cursor)
		return err
	})
	if err != nil {
		return nil, nil, s.handleError(err, "Failed to fetch notifications")
	}

	return items, page, nil
}

// GetUnreadCount returns the number of unread alerts for the customer badge.
func (s *NotificationService) GetUnreadCount(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	var count int
	err := s.dbRunner.WithExecutor(ctx, func(ctx context.Context) error {
		var err error
		count, err = s.repo.CountUnreadByUser(ctx, userID)
		return err
	})
	if err != nil {
		return 0, s.handleError(err, "Failed to fetch unread notification count")
	}
	return count, nil
}

// MarkAsRead marks a single notification as read.
func (s *NotificationService) MarkAsRead(
	ctx context.Context,
	id int64,
	userID uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		return s.repo.MarkAsRead(ctx, id, userID)
	})
	if err != nil {
		return s.handleError(err, "Failed to mark notification as read")
	}
	return nil
}

// MarkAllAsRead marks all user notifications as read.
func (s *NotificationService) MarkAllAsRead(
	ctx context.Context,
	userID uuid.UUID,
) (int64, error) {
	var count int64
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		var err error
		count, err = s.repo.MarkAllAsRead(ctx, userID)
		return err
	})
	if err != nil {
		return 0, s.handleError(err, "Failed to mark notifications as read")
	}
	return count, nil
}

// DeleteNotification removes a notification belonging to the specified user.
func (s *NotificationService) DeleteNotification(
	ctx context.Context,
	id int64,
	userID uuid.UUID,
) error {
	err := s.dbRunner.WithTx(ctx, func(ctx context.Context) error {
		return s.repo.Delete(ctx, id, userID)
	})
	if err != nil {
		return s.handleError(err, "Failed to delete notification")
	}
	return nil
}
