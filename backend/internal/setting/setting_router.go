package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// SettingRouter registers the administrative settings endpoints.
type SettingRouter struct {
	handler      *Handler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a settings router.
func NewRouter(h *Handler, sessionStore middleware.SessionStore) *SettingRouter {
	return &SettingRouter{
		handler:      h,
		sessionStore: sessionStore,
	}
}

// MapRoutes binds /admin/settings routes behind authentication and AdminRole RBAC.
func (r *SettingRouter) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/settings")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		admin.GET("", r.handler.GetSettings)
		admin.PUT("", r.handler.UpdateSettings)
	}
}

// MapPublicRoutes binds the unauthenticated storefront settings endpoint.
func (r *SettingRouter) MapPublicRoutes(vgroup *gin.RouterGroup) {
	public := vgroup.Group("/settings")
	public.Use(middleware.PublicCache(60))
	{
		public.GET("/public", r.handler.GetPublicSettings)
	}
}
