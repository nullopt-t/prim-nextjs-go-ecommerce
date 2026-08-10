package cart

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type CartRouter struct {
	handler *CartHandler
	secrets *config.Secrets
}

func NewRouter(handler *CartHandler, secrets *config.Secrets) *CartRouter {
	return &CartRouter{
		handler: handler,
		secrets: secrets,
	}
}

func (r *CartRouter) MapRoutes(vgroup *gin.RouterGroup) {
	cartGroup := vgroup.Group("/cart")
	cartGroup.Use(middleware.Authenticate(r.secrets, true))
	{
		cartGroup.GET("", r.handler.GetCart)
		cartGroup.DELETE("", r.handler.ClearCart)
		cartGroup.POST("/items", r.handler.AddItem)
		cartGroup.PATCH("/items/:id", r.handler.UpdateItemQuantity)
		cartGroup.DELETE("/items/:id", r.handler.RemoveItem)
	}
}
