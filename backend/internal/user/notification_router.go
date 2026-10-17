package user

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// NotificationRouter manages route definitions and middleware bindings for user notification endpoints.
type NotificationRouter struct {
	handler      *NotificationHandler
	sessionStore middleware.SessionStore
}

// NewNotificationRouter constructs a new NotificationRouter instance.
func NewNotificationRouter(
	handler *NotificationHandler,
	sessionStore middleware.SessionStore,
) *NotificationRouter {
	return &NotificationRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes maps the notification routes to the Gin router group.
// All endpoints require server-side session authentication (Authenticate middleware)
// to ensure only the authenticated user can access or mutate their notifications.
func (r *NotificationRouter) MapRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/user/notifications")
	{
		// GET /api/v1/user/notifications/unread-count - Fetch total unread notifications for badge counter
		group.GET("/unread-count", middleware.Authenticate(r.sessionStore), r.handler.GetUnreadCount)

		// Protected sub-group for paginated list and mutation endpoints
		protected := group.Group("")
		protected.Use(middleware.Authenticate(r.sessionStore))
		{
			// GET /api/v1/user/notifications - Retrieve paginated user notifications
			protected.GET("", r.handler.GetMyNotifications)
			// PATCH /api/v1/user/notifications/:id/read - Mark single notification as read
			protected.PATCH("/:id/read", r.handler.MarkAsRead)
			// PATCH /api/v1/user/notifications/mark-all-read - Mark all unread notifications as read
			protected.PATCH("/mark-all-read", r.handler.MarkAllAsRead)
			// DELETE /api/v1/user/notifications/:id - Permanently delete a notification
			protected.DELETE("/:id", r.handler.DeleteNotification)
		}
	}
}
