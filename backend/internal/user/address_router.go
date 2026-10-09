package user

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type AddressRouter struct {
	handler *AddressHandler
	secrets config.Secrets
}

func NewAddressRouter(
	handler *AddressHandler,
	secrets config.Secrets,
) *AddressRouter {
	return &AddressRouter{
		handler: handler,
		secrets: secrets,
	}
}

func (r *AddressRouter) MapRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/user/addresses")
	group.Use(middleware.Authenticate(r.secrets, false))
	{
		group.GET("", r.handler.ListAddresses)
		group.POST("", r.handler.CreateAddress)
		group.GET("/:id", r.handler.GetAddress)
		group.PUT("/:id", r.handler.UpdateAddress)
		group.PATCH("/:id/default", r.handler.SetDefaultAddress)
		group.DELETE("/:id", r.handler.DeleteAddress)
	}
}
