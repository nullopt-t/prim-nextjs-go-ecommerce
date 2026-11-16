package checkout

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

type CheckoutRouter struct {
	handler      *CheckoutHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *CheckoutHandler, sessionStore middleware.SessionStore) *CheckoutRouter {
	if handler == nil {
		panic("checkout: handler is required")
	}
	if sessionStore == nil {
		panic("checkout: session store is required")
	}
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
		group.POST("/confirm", r.handler.ConfirmPayment)
	}

	// Provider webhook: unauthenticated by design; the HMAC signature is the
	// auth. Must NOT inherit the session Authenticate middleware above.
	webhook := vgroup.Group("/checkout/webhook")
	{
		webhook.POST("/payment", r.handler.HandlePaymentWebhook)
	}

	coupons := vgroup.Group("/coupons")
	coupons.Use(middleware.OptionalAuthenticate(r.sessionStore))
	{
		coupons.POST("/apply", r.handler.ApplyCoupon)
	}
}
