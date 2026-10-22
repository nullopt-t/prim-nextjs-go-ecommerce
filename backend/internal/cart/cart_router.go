// Package cart provides HTTP handlers, services, repository layers, and route definitions
// for shopping cart operations, supporting both authenticated customer accounts and anonymous guest sessions.
package cart

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// SessionStore defines the contract for retrieving server-side session models
// by their unique session identifier string (e.g. backed by Redis or PostgreSQL).
// SessionStore defines the session capabilities required by the cart subsystem:
// resolving active sessions for middleware and creating guest sessions when an
// unauthenticated visitor adds an item.
type SessionStore interface {
	middleware.SessionStore
	sessionCreator
}

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
// If the provided sessionStore also implements sessionCreator, the handler is
// wired automatically to mint guest sessions.
func NewRouter(handler *CartHandler, sessionStore middleware.SessionStore) *CartRouter {
	if creator, ok := sessionStore.(sessionCreator); ok {
		handler.SetSessionCreator(creator)
	}
	return &CartRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers the shopping cart REST API endpoints onto the provided Gin RouterGroup.
// All cart endpoints reside under the "/cart" prefix and utilize OptionalAuthenticate middleware.
// OptionalAuthenticate exposes both guest and authenticated server-side sessions so the
// handler can derive cart ownership from one canonical session object.
func (r *CartRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Initialize the /cart route sub-group from the base API router group.
	cartGroup := vgroup.Group("/cart")

	// Resolve the canonical "ps" session once per request. Authenticated
	// sessions own carts through user_id; guest sessions through session_id.
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
