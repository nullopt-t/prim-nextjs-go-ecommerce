// Package cart provides HTTP handlers, services, repository layers, and route definitions
// for shopping cart operations, supporting both authenticated customer accounts and anonymous guest sessions.
package cart

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// CartRouter defines the HTTP route mappings and middleware bindings for all shopping cart operations.
// It bridges incoming Gin HTTP requests to the CartHandler endpoints.
type CartRouter struct {
	// handler encapsulates the request-response translation logic for cart interactions.
	handler *CartHandler
	// sessionStore manages session lifecycle, token decoding, and user identity resolution.
	sessionStore middleware.SessionStore
}

// NewRouter instantiates and returns a configured CartRouter.
// It accepts a CartHandler pointer and a session store interface implementation.
func NewRouter(handler *CartHandler, sessionStore middleware.SessionStore) *CartRouter {
	return &CartRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers the shopping cart REST API endpoints onto the provided Gin RouterGroup.
// All cart endpoints reside under the "/cart" prefix and utilize OptionalAuthenticate middleware.
// OptionalAuthenticate permits both unauthenticated guest visitors (identified by session cookies)
// and authenticated customers (identified by session cookies) to interact with their carts.
func (r *CartRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Initialize the /cart route sub-group from the base API router group.
	cartGroup := vgroup.Group("/cart")

	// Apply OptionalAuthenticate middleware: extracts userID into context if valid session credentials
	// exist, or allows execution to continue anonymously for guest users with session ID cookies.
	cartGroup.Use(middleware.OptionalAuthenticate(r.sessionStore))
	{
		// GET /cart: Retrieve current shopping cart contents, aggregated totals, and line item inventory statuses.
		cartGroup.GET("", r.handler.GetCart)

		// DELETE /cart: Clear all items from the active cart (hard or soft deletion).
		cartGroup.DELETE("", r.handler.ClearCart)

		// POST /cart/items: Add a product variant and quantity to the cart, or increment existing line item quantity.
		cartGroup.POST("/items", r.handler.AddItem)

		// PATCH /cart/items/:id: Update the target quantity for an existing cart item identified by item UUID.
		cartGroup.PATCH("/items/:id", r.handler.UpdateItemQuantity)

		// DELETE /cart/items/:id: Remove a specific item from the cart identified by item UUID.
		cartGroup.DELETE("/items/:id", r.handler.RemoveItem)
	}
}
