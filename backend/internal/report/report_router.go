package report

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// Router registers the admin reporting endpoints.
type Router struct {
	handler      *Handler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a report router.
func NewRouter(handler *Handler, sessionStore middleware.SessionStore) *Router {
	return &Router{handler: handler, sessionStore: sessionStore}
}

// MapRoutes binds /admin/reports behind authentication and AdminRole RBAC.
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/reports")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole, model.SuperRole),
	)
	{
		admin.GET("/sales", r.handler.GetSalesReport)
		admin.GET("/export", r.handler.ExportReport)
	}
}
