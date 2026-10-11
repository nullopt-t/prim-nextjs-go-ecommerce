package user

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
)

type NotificationRouter struct {
	handler      *NotificationHandler
	sessionStore middleware.SessionStore
}

func NewNotificationRouter(
	handler *NotificationHandler,
	sessionStore middleware.SessionStore,
) *NotificationRouter {
	return &NotificationRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes maps the notification routes to the router.
func (r *NotificationRouter) MapRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/user/notifications")
	{
		group.GET("/unread-count", middleware.Authenticate(r.sessionStore), r.handler.GetUnreadCount)

		protected := group.Group("")
		protected.Use(middleware.Authenticate(r.sessionStore))
		{
			protected.GET("", r.handler.GetMyNotifications)
			protected.PATCH("/:id/read", r.handler.MarkAsRead)
			protected.PATCH("/mark-all-read", r.handler.MarkAllAsRead)
			protected.DELETE("/:id", r.handler.DeleteNotification)
		}
	}
}
