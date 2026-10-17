package user

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
)

// AddressRouter configures HTTP routing and middleware pipelines for customer address management endpoints.
type AddressRouter struct {
	handler      *AddressHandler
	sessionStore middleware.SessionStore
}

// NewAddressRouter initializes an AddressRouter with required handler and session verification dependencies.
func NewAddressRouter(
	handler *AddressHandler,
	sessionStore middleware.SessionStore,
) *AddressRouter {
	return &AddressRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes registers RESTful address management routes under the /api/v1/user/addresses group.
// All endpoints are protected by server-side session authentication middleware (Authenticate),
// ensuring an active user session is present before any handler executes.
func (r *AddressRouter) MapRoutes(v1 *gin.RouterGroup) {
	group := v1.Group("/user/addresses")
	group.Use(middleware.Authenticate(r.sessionStore))
	{
		// GET /api/v1/user/addresses - List all saved addresses for the authenticated customer
		group.GET("", r.handler.ListAddresses)
		// POST /api/v1/user/addresses - Create a new saved address
		group.POST("", r.handler.CreateAddress)
		// GET /api/v1/user/addresses/:id - Retrieve address details by UUID
		group.GET("/:id", r.handler.GetAddress)
		// PUT /api/v1/user/addresses/:id - Update an existing address
		group.PUT("/:id", r.handler.UpdateAddress)
		// PATCH /api/v1/user/addresses/:id/default - Designate an address as the customer's default
		group.PATCH("/:id/default", r.handler.SetDefaultAddress)
		// DELETE /api/v1/user/addresses/:id - Soft-delete an address
		group.DELETE("/:id", r.handler.DeleteAddress)
	}
}
