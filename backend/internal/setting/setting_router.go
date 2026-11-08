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
	if h == nil {
		panic("setting: handler is required")
	}
	if sessionStore == nil {
		panic("setting: session store is required")
	}
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
		middleware.Authorize(model.AdminRole, model.SuperRole),
	)
	{
		admin.GET("", r.handler.GetSettings)
		// PUT /admin/settings mutates global storefront configuration.
		admin.PUT("", r.handler.UpdateSettings)
	}
}

// MapPublicRoutes binds the unauthenticated storefront settings endpoint.
func (r *SettingRouter) MapPublicRoutes(vgroup *gin.RouterGroup) {
	public := vgroup.Group("/settings")
	// Storefront settings drive visible content and must reflect admin changes
	// immediately, so browsers and proxies must not cache this response.
	public.Use(middleware.NoCache())
	{
		public.GET("/public", r.handler.GetPublicSettings)
		public.GET("/currencies", r.handler.GetCurrencies)
	}
}
