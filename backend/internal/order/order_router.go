package order

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

type OrderRouter struct {
	handler      *OrderHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *OrderHandler, sessionStore middleware.SessionStore) *OrderRouter {
	return &OrderRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *OrderRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/orders")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		admin.GET("", r.handler.AdminListOrders)
		admin.GET("/:id", r.handler.AdminGetOrderByID)
		admin.PATCH("/:id/status", r.handler.UpdateOrderStatus)
	}

	ordersGroup := vgroup.Group("/orders")
	{
		ordersGroup.GET("", middleware.Authenticate(r.sessionStore), r.handler.GetOrders)
		ordersGroup.GET("/:id", middleware.Authenticate(r.sessionStore), r.handler.GetOrderByID)
		ordersGroup.POST("/:id/cancel", middleware.Authenticate(r.sessionStore), r.handler.CancelOrder)
	}
}
