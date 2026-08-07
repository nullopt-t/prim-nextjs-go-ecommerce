package checkout

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type CheckoutRouter struct {
	handler *CheckoutHandler
	secrets *config.Secrets
}

func NewRouter(handler *CheckoutHandler, secrets *config.Secrets) *CheckoutRouter {
	return &CheckoutRouter{
		handler: handler,
		secrets: secrets,
	}
}

func (r *CheckoutRouter) MapRoutes(vgroup *gin.RouterGroup) {
	group := vgroup.Group("/checkout")
	group.Use(middleware.Authenticate(r.secrets, false))
	{
		group.POST("", r.handler.Checkout)
	}
}
