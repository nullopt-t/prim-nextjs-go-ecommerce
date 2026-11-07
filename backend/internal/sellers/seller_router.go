package sellers

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// Router maps seller portal and admin routes.
type Router struct {
	handler      *Handler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *Handler, sessionStore middleware.SessionStore) *Router {
	if handler == nil {
		panic("seller: handler is required")
	}
	if sessionStore == nil {
		panic("seller: session store is required")
	}
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers:
//   - /seller/*  → authenticated merchant portal. Capability is derived from
//     the caller's sellers row and lifecycle status, never from users.role.
//   - /admin/sellers/* → administrator surface
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	portal := vgroup.Group("/seller")
	portal.Use(middleware.Authenticate(r.sessionStore))
	{
		portal.GET("/dashboard", r.handler.GetDashboard)
		portal.POST("/register", r.handler.RegisterSeller)
		portal.GET("/slug-availability", r.handler.CheckSlugAvailability)
		portal.GET("/profile", r.handler.GetMySeller)
		portal.PATCH("/profile", r.handler.UpdateMySeller)
		portal.POST("/payouts", r.handler.RequestPayout)
		portal.GET("/payouts", r.handler.ListMyPayouts)
		portal.GET("/orders", r.handler.ListMyOrders)
		portal.GET("/orders/:id", r.handler.GetMyOrder)

		// Merchant-owned catalog. Ownership is enforced per product; platform
		// (admin) products are not reachable through these routes.
		portal.POST("/media", r.handler.UploadMedia)
		portal.GET("/attributes", r.handler.ListSellerAttributes)
		portal.POST("/attributes", r.handler.CreateSellerAttribute)
		portal.PATCH("/attributes/:id", r.handler.UpdateSellerAttribute)
		portal.DELETE("/attributes/:id", r.handler.DeleteSellerAttribute)
		portal.GET("/promotions", r.handler.ListPromotions)
		portal.POST("/promotions", r.handler.CreatePromotion)
		portal.PATCH("/promotions/:id/status", r.handler.SetPromotionStatus)
		portal.DELETE("/promotions/:id", r.handler.DeletePromotion)
		portal.GET("/bundles", r.handler.ListSellerBundles)
		portal.POST("/bundles", r.handler.CreateSellerBundle)
		portal.PATCH("/bundles/:id", r.handler.UpdateSellerBundle)
		portal.DELETE("/bundles/:id", r.handler.DeleteSellerBundle)
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

	admin := vgroup.Group("/admin/sellers")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole, model.SuperRole),
	)
	{
		admin.GET("", r.handler.AdminListSellers)
		admin.GET("/:id", r.handler.AdminGetSeller)
		admin.PATCH("/:id/status", r.handler.AdminChangeSellerStatus)
		admin.PATCH("/:id/commission", r.handler.AdminSetCommission)
		admin.GET("/payouts", r.handler.AdminListPayouts)
		admin.PATCH("/payouts/:id/transition", r.handler.AdminTransitionPayout)
	}
}
