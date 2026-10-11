package wishlist

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

type WishlistRouter struct {
	handler      *WishlistHandler
	sessionStore middleware.SessionStore
}

func NewRouter(
	handler *WishlistHandler,
	sessionStore middleware.SessionStore,
) *WishlistRouter {
	return &WishlistRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *WishlistRouter) MapRoutes(vgroup *gin.RouterGroup) {
	wishlistGroup := vgroup.Group("/wishlist")
	wishlistGroup.Use(middleware.Authenticate(r.sessionStore))
	{
		wishlistGroup.POST("", r.handler.AddToWishlist)
		wishlistGroup.POST("/items", r.handler.AddToWishlist)
		wishlistGroup.GET("", r.handler.GetMyWishlist)
		wishlistGroup.GET("/count", r.handler.GetWishlistCount)
		wishlistGroup.GET("/check/:productId", r.handler.CheckInWishlist)
		wishlistGroup.DELETE("", r.handler.ClearWishlist)
		wishlistGroup.DELETE("/items/:id", r.handler.RemoveItem)
		wishlistGroup.DELETE("/products/:productId", r.handler.RemoveByProductID)
	}
}
