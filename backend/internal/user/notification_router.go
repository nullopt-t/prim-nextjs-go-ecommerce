package user

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type NotificationRouter struct {
	handler *NotificationHandler
	secrets config.Secrets
}

func NewNotificationRouter(
	handler *NotificationHandler,
	secrets config.Secrets,
) *NotificationRouter {
	return &NotificationRouter{
		handler: handler,
		secrets: secrets,
	}
}

// MapRoutes maps the notification routes to the router.
func (r *NotificationRouter) MapRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/user/notifications")
	{
		group.GET("/unread-count", middleware.Authenticate(r.secrets, true), r.handler.GetUnreadCount)

		protected := group.Group("")
		protected.Use(middleware.Authenticate(r.secrets, false))
		{
			protected.GET("", r.handler.GetMyNotifications)
			protected.PATCH("/:id/read", r.handler.MarkAsRead)
			protected.PATCH("/mark-all-read", r.handler.MarkAllAsRead)
			protected.DELETE("/:id", r.handler.DeleteNotification)
		}
	}
}
