package coupon

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

type Router struct {
	handler      *CouponHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler *CouponHandler, sessionStore middleware.SessionStore) *Router {
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	admin := vgroup.Group("/admin/promotions")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		admin.GET("", r.handler.ListCoupons)
		admin.POST("", r.handler.CreateCoupon)
		admin.GET("/:id", r.handler.GetCouponByID)
		admin.PATCH("/:id/status", r.handler.UpdateStatus)
		admin.DELETE("/:id", r.handler.DeleteCoupon)
	}
}
