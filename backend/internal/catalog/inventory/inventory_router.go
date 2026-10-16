package inventory

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

type Router struct {
	handler      *InventoryHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *InventoryHandler, sessionStore middleware.SessionStore) *Router {
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/inventory")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		admin.GET("", r.handler.ListInventoryOverview)
		admin.GET("/ledgers", r.handler.ListRecentLedgers)
		admin.POST("/variants/:variantId/adjust", r.handler.AdjustInventory)
		admin.POST("/products/:productId/adjust", r.handler.AdjustProductInventory)
	}
}
