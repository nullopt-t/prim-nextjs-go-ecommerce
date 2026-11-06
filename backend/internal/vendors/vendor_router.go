package vendors

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// Router maps vendor portal and admin routes.
type Router struct {
	handler      *Handler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *Handler, sessionStore middleware.SessionStore) *Router {
	if handler == nil {
		panic("vendor: handler is required")
	}
	if sessionStore == nil {
		panic("vendor: session store is required")
	}
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers:
//   - /vendor/*  → authenticated merchant portal. Capability is derived from
//     the caller's vendors row and lifecycle status, never from users.role.
//   - /admin/vendors/* → administrator surface
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	portal := vgroup.Group("/vendor")
	portal.Use(middleware.Authenticate(r.sessionStore))
	{
		portal.GET("/dashboard", r.handler.GetDashboard)
		portal.POST("/register", r.handler.RegisterVendor)
		portal.GET("/slug-availability", r.handler.CheckSlugAvailability)
		portal.GET("/profile", r.handler.GetMyVendor)
		portal.PATCH("/profile", r.handler.UpdateMyVendor)
		portal.POST("/payouts", r.handler.RequestPayout)
		portal.GET("/payouts", r.handler.ListMyPayouts)
		portal.GET("/orders", r.handler.ListMyOrders)
		portal.GET("/orders/:id", r.handler.GetMyOrder)

		// Merchant-owned catalog. Ownership is enforced per product; platform
		// (admin) products are not reachable through these routes.
		portal.GET("/products", r.handler.ListMyProducts)
		portal.POST("/products", r.handler.CreateProduct)
		portal.PATCH("/products/:id", r.handler.UpdateProduct)
		portal.DELETE("/products/:id", r.handler.DeleteProduct)
		portal.POST("/products/:id/publish", r.handler.PublishProduct)
		portal.POST("/products/:id/archive", r.handler.ArchiveProduct)

		// Variants and stock movements for merchant-owned products.
		portal.GET("/products/:id/variants", r.handler.ListProductVariants)
		portal.POST("/products/:id/variants", r.handler.CreateVariant)
		portal.PATCH("/variants/:id", r.handler.UpdateVariant)
		portal.DELETE("/variants/:id", r.handler.DeleteVariant)
		portal.POST("/variants/:id/stock", r.handler.AdjustStock)
	}

	admin := vgroup.Group("/admin/vendors")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole, model.SuperRole),
	)
	{
		admin.GET("", r.handler.AdminListVendors)
		admin.GET("/:id", r.handler.AdminGetVendor)
		admin.PATCH("/:id/status", r.handler.AdminChangeVendorStatus)
		admin.PATCH("/:id/commission", r.handler.AdminSetCommission)
		admin.GET("/payouts", r.handler.AdminListPayouts)
		admin.PATCH("/payouts/:id/transition", r.handler.AdminTransitionPayout)
	}
}
