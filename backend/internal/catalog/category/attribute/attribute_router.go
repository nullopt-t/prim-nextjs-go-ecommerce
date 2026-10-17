package attribute

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

type AttributeRouter struct {
	handler      *AttributeHandler
	sessionStore middleware.SessionStore
}

func NewRouter(h *AttributeHandler, sessionStore middleware.SessionStore) *AttributeRouter {
	return &AttributeRouter{
		handler:      h,
		sessionStore: sessionStore,
	}
}

func (r *AttributeRouter) MapRoutes(vgroup *gin.RouterGroup) {
	adminAttrs := vgroup.Group("/admin/attributes")
	adminAttrs.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		adminAttrs.GET("", r.handler.ListAttributes)
		adminAttrs.POST("", r.handler.CreateAttribute)
		adminAttrs.DELETE("/:id", r.handler.DeleteAttribute)
	}
}
