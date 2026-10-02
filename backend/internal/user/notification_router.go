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

func (r *NotificationRouter) MapRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/user/notifications")
	group.Use(middleware.Authenticate(r.secrets, false))
	{
		group.GET("", r.handler.GetMyNotifications)
		group.GET("/unread-count", r.handler.GetUnreadCount)
		group.PATCH("/:id/read", r.handler.MarkAsRead)
		group.PATCH("/mark-all-read", r.handler.MarkAllAsRead)
		group.DELETE("/:id", r.handler.DeleteNotification)
	}
}
