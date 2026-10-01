package wishlist

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type WishlistRouter struct {
	handler *WishlistHandler
	secrets config.Secrets
}

func NewRouter(
	handler *WishlistHandler,
	secrets config.Secrets,
) *WishlistRouter {
	return &WishlistRouter{
		handler: handler,
		secrets: secrets,
	}
}

func (r *WishlistRouter) MapRoutes(vgroup *gin.RouterGroup) {
	wishlistGroup := vgroup.Group("/wishlist")
	wishlistGroup.Use(middleware.Authenticate(r.secrets, false))
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
