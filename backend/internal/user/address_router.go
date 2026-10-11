package user

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type AddressRouter struct {
	handler      *AddressHandler
	sessionStore middleware.SessionStore
}

func NewAddressRouter(
	handler *AddressHandler,
	sessionStore middleware.SessionStore,
) *AddressRouter {
	return &AddressRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *AddressRouter) MapRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/user/addresses")
	group.Use(middleware.Authenticate(r.sessionStore))
	{
		group.GET("", r.handler.ListAddresses)
		group.POST("", r.handler.CreateAddress)
		group.GET("/:id", r.handler.GetAddress)
		group.PUT("/:id", r.handler.UpdateAddress)
		group.PATCH("/:id/default", r.handler.SetDefaultAddress)
		group.DELETE("/:id", r.handler.DeleteAddress)
	}
}
