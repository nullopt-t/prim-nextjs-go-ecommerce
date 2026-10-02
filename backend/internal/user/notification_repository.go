package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/database"
)

// NotificationRepository performs SQL queries on user_notifications.
type NotificationRepository struct{}

// NewNotificationRepository constructs a NotificationRepository.
func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{}
}

// Create inserts a new notification record into PostgreSQL.
func (r *NotificationRepository) Create(
	ctx context.Context,
	qe database.QueryExecutor,
	n *model.UserNotification,
) error {
	query := `
		INSERT INTO user_notifications (
			id,
			user_id,
			title,
			message,
			category,
			action_url,
			metadata,
			read_at,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		RETURNING created_at
	`
	return qe.QueryRow(
		ctx,
		query,
		n.ID,
		n.UserID,
		n.Title,
		n.Message,
		n.Category,
		n.ActionURL,
		n.Metadata,
		n.ReadAt,
	).Scan(&n.CreatedAt)
}

// ListByUser retrieves paginated user notifications, optionally filtered to unread items only.
func (r *NotificationRepository) ListByUser(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
	unreadOnly bool,
	q *pagination.ListQuery,
) ([]model.UserNotification, int, error) {
	var countQuery string
	if unreadOnly {
		countQuery = `SELECT COUNT(*) FROM user_notifications WHERE user_id = $1 AND read_at IS NULL`
	} else {
		countQuery = `SELECT COUNT(*) FROM user_notifications WHERE user_id = $1`
	}

	var total int
	if err := qe.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user notifications: %w", err)
	}

	if total == 0 {
		return []model.UserNotification{}, 0, nil
	}

	var selectQuery string
	if unreadOnly {
		selectQuery = `
			SELECT id, user_id, title, message, category, action_url, metadata, read_at, created_at
			FROM user_notifications
			WHERE user_id = $1 AND read_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
	} else {
		selectQuery = `
			SELECT id, user_id, title, message, category, action_url, metadata, read_at, created_at
			FROM user_notifications
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
	}

	rows, err := qe.Query(ctx, selectQuery, userID, q.PageSize, q.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list user notifications: %w", err)
	}
	defer rows.Close()

	items := make([]model.UserNotification, 0, q.PageSize)
	for rows.Next() {
		var item model.UserNotification
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Message,
			&item.Category,
			&item.ActionURL,
			&item.Metadata,
			&item.ReadAt,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user notification: %w", err)
		}
		items = append(items, item)
	}

	return items, total, rows.Err()
}

// CountUnreadByUser queries the index on unread notifications.
func (r *NotificationRepository) CountUnreadByUser(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (int, error) {
	query := `SELECT COUNT(*) FROM user_notifications WHERE user_id = $1 AND read_at IS NULL`
	var count int
	if err := qe.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread user notifications: %w", err)
	}
	return count, nil
}

// MarkAsRead marks a single notification as read if it belongs to the user.
func (r *NotificationRepository) MarkAsRead(
	ctx context.Context,
	qe database.QueryExecutor,
	id, userID uuid.UUID,
) error {
	query := `
		UPDATE user_notifications
		SET read_at = now()
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL
	`
	cmd, err := qe.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// MarkAllAsRead marks all unread notifications for a user as read.
func (r *NotificationRepository) MarkAllAsRead(
	ctx context.Context,
	qe database.QueryExecutor,
	userID uuid.UUID,
) (int64, error) {
	query := `
		UPDATE user_notifications
		SET read_at = now()
		WHERE user_id = $1 AND read_at IS NULL
	`
	cmd, err := qe.Exec(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("mark all notifications as read: %w", err)
	}
	return cmd.RowsAffected(), nil
}

// Delete removes a notification by id.
func (r *NotificationRepository) Delete(
	ctx context.Context,
	qe database.QueryExecutor,
	id, userID uuid.UUID,
) error {
	query := `DELETE FROM user_notifications WHERE id = $1 AND user_id = $2`
	cmd, err := qe.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete user notification: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
