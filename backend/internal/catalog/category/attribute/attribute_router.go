package attribute

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

type AttributeRouter struct {
	handler      *AttributeHandler
	sessionStore middleware.SessionStore
}

func NewRouter(h *AttributeHandler, sessionStore middleware.SessionStore) *AttributeRouter {
	if h == nil {
		panic("attribute: handler is required")
	}
	if sessionStore == nil {
		panic("attribute: session store is required")
	}
	return &AttributeRouter{
		handler:      h,
		sessionStore: sessionStore,
	}
}

func (r *AttributeRouter) MapRoutes(vgroup *gin.RouterGroup) {
	adminAttrs := vgroup.Group("/admin/attributes")
	adminAttrs.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole, model.SuperRole),
	)
	{
		adminAttrs.GET("", r.handler.ListAttributes)
		adminAttrs.POST("", r.handler.CreateAttribute)
		adminAttrs.PATCH("/:id", r.handler.UpdateAttribute)
		adminAttrs.DELETE("/:id", r.handler.DeleteAttribute)
	}
}
