package user

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/api/pagination"
)

// NotificationHandler processes HTTP operations for in-app customer notifications:
// paginated list queries, unread count polling, marking individual/all notifications as read,
// and notification deletion.
type NotificationHandler struct {
	service *NotificationService
}

// NewNotificationHandler constructs a new NotificationHandler instance.
func NewNotificationHandler(service *NotificationService) *NotificationHandler {
	if service == nil {
		panic("notification: service is required")
	}
	return &NotificationHandler{service: service}
}

// NotificationItemResponse represents the localized public API response structure for an in-app notification.
type NotificationItemResponse struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Category  string          `json:"category"`
	ActionURL *string         `json:"actionUrl,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty" swaggertype:"object"`
	IsRead    bool            `json:"isRead"`
	ReadAt    *string         `json:"readAt,omitempty"`
	CreatedAt string          `json:"createdAt"`
}

// NotificationCountResponse represents the badge count payload returned for user unread alerts.
type NotificationCountResponse struct {
	Count int `json:"count"`
}

// mapNotificationResponse converts an internal UserNotification domain model entity into the
// localized API response DTO, applying locale translations to bilingual title and message fields.
func mapNotificationResponse(n *model.UserNotification, locale string) NotificationItemResponse {
	var readAtStr *string
	if n.ReadAt != nil {
		s := n.ReadAt.Format(time.RFC3339)
		readAtStr = &s
	}

	return NotificationItemResponse{
		ID:        strconv.FormatInt(n.ID, 10),
		Title:     n.LocalizedTitle(locale),
		Message:   n.LocalizedMessage(locale),
		Category:  n.Category,
		ActionURL: n.ActionURL,
		Metadata:  n.Metadata,
		IsRead:    n.ReadAt != nil,
		ReadAt:    readAtStr,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
	}
}

// GetMyNotifications godoc
//
//	@Summary		List user notifications
//	@Description	Returns a paginated list of in-app notifications localized to the active site language.
//	@Tags			Notifications
//	@Produce		json
//	@Param			q			query		pagination.ListQuery	true	"Pagination parameters"
//	@Param			unreadOnly	query		bool					false	"Filter by unread only"
//	@Success		200			{object}	api.PaginatedResponse{data=[]NotificationItemResponse}
//	@Failure		401			{object}	api.UnauthorizedErrorResponse
//	@Router			/user/notifications [get]
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	// Extract authenticated user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Bind and sanitize pagination query parameters
	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ValidationFailedError("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 5, MaxPageSize: 10})

	// Parse unread-only boolean query filter and active content language
	unreadOnly := c.Query("unreadOnly") == "true"
	locale := middleware.GetLocale(c)

	// Fetch paginated notification items
	result, err := h.service.GetMyNotifications(c.Request.Context(), userID, unreadOnly, q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Map and localize notification items
	responses := make([]NotificationItemResponse, len(result.Items))
	for i, item := range result.Items {
		responses[i] = mapNotificationResponse(item, locale)
	}

	c.JSON(http.StatusOK, api.PaginatedResponse{
		Data: responses,
		Meta: result.Page,
	})
}

// GetUnreadCount godoc
//
//	@Summary		Get unread notifications count
//	@Description	Returns total count of unread notifications for the user's navbar badge.
//	@Tags			Notifications
//	@Produce		json
//	@Success		200	{object}	api.DataResponse{data=NotificationCountResponse}
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Router			/user/notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	// Soft authorization fallback: if unauthenticated or anonymous guest, return count 0
	if _, ok := middleware.GetUserID(c); !ok {
		c.JSON(http.StatusOK, api.DataResponse{Data: NotificationCountResponse{Count: 0}})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, api.DataResponse{Data: NotificationCountResponse{Count: 0}})
		return
	}

	// Query unread count index
	count, err := h.service.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: NotificationCountResponse{Count: count}})
}

// MarkAsRead godoc
//
//	@Summary		Mark notification as read
//	@Tags			Notifications
//	@Produce		json
//	@Param			id	path		string	true	"Notification UUID"
//	@Success		200	{object}	api.MessageResponse
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Failure		404	{object}	api.NotFoundErrorResponse
//	@Router			/user/notifications/{id}/read [patch]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// Extract authenticated user ID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Parse notification id
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid notification id").WithCode(apierr.CodeInvalidInput))
		return
	}

	if err := h.service.MarkAsRead(c.Request.Context(), id, userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Notification marked as read"})
}

// MarkAllAsRead godoc
//
//	@Summary		Mark all notifications as read
//	@Tags			Notifications
//	@Produce		json
//	@Success		200	{object}	api.MessageResponse
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Router			/user/notifications/mark-all-read [patch]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// Extract caller's user UUID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Atomically mark all unread notifications for the user as read
	_, err = h.service.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "All notifications marked as read"})
}

// DeleteNotification godoc
//
//	@Summary		Delete notification
//	@Tags			Notifications
//	@Produce		json
//	@Param			id	path		string	true	"Notification UUID"
//	@Success		200	{object}	api.MessageResponse
//	@Failure		401	{object}	api.UnauthorizedErrorResponse
//	@Failure		404	{object}	api.NotFoundErrorResponse
//	@Router			/user/notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	// Extract caller's user UUID from context
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Parse target notification id
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid notification id").WithCode(apierr.CodeInvalidInput))
		return
	}

	// Delete notification with strict tenant ownership validation
	if err := h.service.DeleteNotification(c.Request.Context(), id, userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Notification deleted successfully"})
}

// getUserIDFromContext retrieves the authenticated user UUID from the session
// stored in Gin context by the authentication middleware.
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, apierr.UnauthorizedError("unauthorized")
	}
	return userID, nil
}
