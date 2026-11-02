package vendor

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
//   - /vendor/*  → vendor portal (vendor + admin roles for management;
//     profile reads for any authenticated vendor)
//   - /admin/vendors/* → administrator surface
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	// Vendor portal: authenticated users only; role owns portal access.
	portal := vgroup.Group("/vendor")
	portal.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.VendorRole, model.AdminRole, model.SuperRole),
	)
	{
		portal.POST("/register", r.handler.RegisterVendor)
		portal.GET("/profile", r.handler.GetMyVendor)
		portal.PATCH("/profile", r.handler.UpdateMyVendor)
		portal.POST("/payouts", r.handler.RequestPayout)
		portal.GET("/payouts", r.handler.ListMyPayouts)
	}

	admin := vgroup.Group("/admin/vendors")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole, model.SuperRole),
	)
	{
		admin.GET("/:id", r.handler.AdminGetVendor)
		admin.PATCH("/:id/status", r.handler.AdminChangeVendorStatus)
		admin.PATCH("/:id/commission", r.handler.AdminSetCommission)
		admin.GET("/payouts", r.handler.AdminListPayouts)
		admin.PATCH("/payouts/:id/transition", r.handler.AdminTransitionPayout)
	}
}
