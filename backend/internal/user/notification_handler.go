package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/pagination"
)

type NotificationHandler struct {
	service *NotificationService
}

func NewNotificationHandler(service *NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

type NotificationItemResponse struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Category  string          `json:"category"`
	ActionURL *string         `json:"actionUrl,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	IsRead    bool            `json:"isRead"`
	ReadAt    *string         `json:"readAt,omitempty"`
	CreatedAt string          `json:"createdAt"`
}

type NotificationCountResponse struct {
	Count int `json:"count"`
}

func mapNotificationResponse(n *model.UserNotification, locale string) NotificationItemResponse {
	var readAtStr *string
	if n.ReadAt != nil {
		s := n.ReadAt.Format(time.RFC3339)
		readAtStr = &s
	}

	return NotificationItemResponse{
		ID:        n.ID.String(),
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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	q := &pagination.ListQuery{}
	if err := c.ShouldBindQuery(q); err != nil {
		_ = c.Error(apierr.ErrValidationFailed("invalid query parameters"))
		return
	}
	q.Process(pagination.QueryOptions{DefaultPageSize: 15, MaxPageSize: 50})

	unreadOnly := c.Query("unreadOnly") == "true"
	locale := middleware.GetLocale(c)

	result, err := h.service.GetMyNotifications(c.Request.Context(), userID, unreadOnly, q)
	if err != nil {
		_ = c.Error(err)
		return
	}

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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("invalid notification id").WithCode(apierr.CodeInvalidInput))
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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

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
	userID, err := getUserIDFromContext(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(apierr.ErrBadRequest("invalid notification id").WithCode(apierr.CodeInvalidInput))
		return
	}

	if err := h.service.DeleteNotification(c.Request.Context(), id, userID); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Notification deleted successfully"})
}

func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, apierr.ErrUnauthorized("unauthorized")
	}
	if id, ok := val.(uuid.UUID); ok {
		return id, nil
	}
	if idStr, ok := val.(string); ok {
		id, err := uuid.Parse(idStr)
		if err == nil {
			return id, nil
		}
	}
	return uuid.Nil, apierr.ErrUnauthorized("invalid user identifier in context")
}
