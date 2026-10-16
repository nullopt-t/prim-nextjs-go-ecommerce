package checkout

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type CheckoutRouter struct {
	handler      *CheckoutHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *CheckoutHandler, sessionStore middleware.SessionStore) *CheckoutRouter {
	return &CheckoutRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *CheckoutRouter) MapRoutes(vgroup *gin.RouterGroup) {
	group := vgroup.Group("/checkout")
	group.Use(middleware.Authenticate(r.sessionStore))
	{
		group.POST("", r.handler.Checkout)
	}

	coupons := vgroup.Group("/coupons")
	coupons.Use(middleware.OptionalAuthenticate(r.sessionStore))
	{
		coupons.POST("/apply", r.handler.ApplyCoupon)
	}
}
