package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

type AdminRouter struct {
	handler      *AdminHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *AdminHandler, sessionStore middleware.SessionStore) *AdminRouter {
	return &AdminRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *AdminRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		// Dashboard Overview
		admin.GET("/dashboard", r.handler.GetDashboard)

		// Promotions / Coupons
		admin.GET("/promotions", r.handler.ListPromotions)
		admin.POST("/promotions", r.handler.CreatePromotion)
		admin.PATCH("/promotions/:id/status", r.handler.UpdatePromotionStatus)
		admin.DELETE("/promotions/:id", r.handler.DeletePromotion)

		// Payments
		admin.GET("/payments", r.handler.ListPayments)
		admin.PATCH("/payments/:id/status", r.handler.UpdatePaymentStatus)
		admin.POST("/payments/:id/refund", r.handler.RefundPayment)

		// Media Library
		admin.GET("/media", r.handler.ListMedia)
		admin.POST("/media", r.handler.UploadMedia)

		// Attributes
		admin.GET("/attributes", r.handler.ListAttributes)
		admin.POST("/attributes", r.handler.CreateAttribute)
		admin.DELETE("/attributes/:id", r.handler.DeleteAttribute)

		// Inventory
		admin.GET("/inventory", r.handler.ListInventoryOverview)
		admin.GET("/inventory/ledgers", r.handler.ListRecentLedgers)
		admin.POST("/inventory/variants/:variantId/adjust", r.handler.AdjustInventory)
		admin.POST("/inventory/products/:productId/adjust", r.handler.AdjustProductInventory)

		// Settings
		admin.GET("/settings", r.handler.GetSettings)
		admin.PUT("/settings", r.handler.UpdateSettings)
	}
}
