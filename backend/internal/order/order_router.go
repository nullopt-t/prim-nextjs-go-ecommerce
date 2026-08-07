package order

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type OrderRouter struct {
	handler *OrderHandler
	secrets *config.Secrets
}

func NewRouter(handler *OrderHandler, secrets *config.Secrets) *OrderRouter {
	return &OrderRouter{
		handler: handler,
		secrets: secrets,
	}
}

func (r *OrderRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/orders")
	admin.Use(
		middleware.Authenticate(r.secrets, true),
		middleware.Authorize(model.AdminRole),
	)
	{
		admin.PATCH("/:id/status", r.handler.UpdateOrderStatus)
	}

	ordersGroup := vgroup.Group("/orders")
	{
		ordersGroup.GET("", middleware.Authenticate(r.secrets, true), r.handler.GetOrders)
		ordersGroup.GET("/:id", middleware.Authenticate(r.secrets, true), r.handler.GetOrderByID)
	}
}
