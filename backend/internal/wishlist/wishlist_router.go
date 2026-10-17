package wishlist

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
)

// WishlistRouter maps incoming HTTP endpoints for customer wishlist management.
// All wishlist routes require customer authentication via middleware.Authenticate.
type WishlistRouter struct {
	handler      *WishlistHandler
	sessionStore middleware.SessionStore
}

// NewRouter constructs a WishlistRouter with its handler and sessionStore dependencies.
func NewRouter(
	handler *WishlistHandler,
	sessionStore middleware.SessionStore,
) *WishlistRouter {
	return &WishlistRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers the customer wishlist REST API endpoints onto the provided Gin RouterGroup.
// Every endpoint in this group requires a valid authenticated user session.
// Note on guest wishlist: Currently, wishlist endpoints require authenticated customer sessions.
// Guest wishlist persistence and merging can be integrated in future milestones alongside user registration flows.
func (r *WishlistRouter) MapRoutes(vgroup *gin.RouterGroup) {
	wishlistGroup := vgroup.Group("/wishlist")
	wishlistGroup.Use(middleware.Authenticate(r.sessionStore))
	{
		// POST /wishlist & POST /wishlist/items: Add product or variant to authenticated user's wishlist.
		wishlistGroup.POST("", r.handler.AddToWishlist)
		wishlistGroup.POST("/items", r.handler.AddToWishlist)

		// GET /wishlist: Retrieve a paginated list of items saved in the customer's wishlist.
		wishlistGroup.GET("", r.handler.GetMyWishlist)

		// GET /wishlist/count: Lightweight endpoint returning total item count in customer's wishlist.
		wishlistGroup.GET("/count", r.handler.GetWishlistCount)

		// GET /wishlist/check/:productId: Query whether a specific product or variant is saved.
		wishlistGroup.GET("/check/:productId", r.handler.CheckInWishlist)

		// DELETE /wishlist: Clears entire wishlist or selectively removes by ?productId / ?itemId.
		wishlistGroup.DELETE("", r.handler.ClearWishlist)

		// DELETE /wishlist/items/:id: Deletes a specific wishlist entry by record UUID.
		wishlistGroup.DELETE("/items/:id", r.handler.RemoveItem)

		// DELETE /wishlist/products/:productId: Deletes saved entry by catalog product UUID.
		wishlistGroup.DELETE("/products/:productId", r.handler.RemoveByProductID)
	}
}
