package cart

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type CartRouter struct {
	handler      *CartHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *CartHandler, sessionStore middleware.SessionStore) *CartRouter {
	return &CartRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *CartRouter) MapRoutes(vgroup *gin.RouterGroup) {
	cartGroup := vgroup.Group("/cart")
	cartGroup.Use(middleware.OptionalAuthenticate(r.sessionStore))
	{
		cartGroup.GET("", r.handler.GetCart)
		cartGroup.DELETE("", r.handler.ClearCart)
		cartGroup.POST("/items", r.handler.AddItem)
		cartGroup.PATCH("/items/:id", r.handler.UpdateItemQuantity)
		cartGroup.DELETE("/items/:id", r.handler.RemoveItem)
	}
}
