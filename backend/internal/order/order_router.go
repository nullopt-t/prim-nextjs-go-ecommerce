// Package order provides models, persistence repositories, business logic services,
// and HTTP handlers managing the lifecycle of customer and administrator orders.
package order

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/internal/model"
)

// OrderRouter wires HTTP endpoint routes for customer order management
// and administrative order inspection/mutation to the Gin engine.
type OrderRouter struct {
	// handler dispatches incoming HTTP requests to business service methods.
	handler *OrderHandler
	// sessionStore validates and loads session authentication credentials for secured routes.
	sessionStore middleware.SessionStore
}

// NewRouter constructs a new OrderRouter initialized with the provided handler and session store.
func NewRouter(handler *OrderHandler, sessionStore middleware.SessionStore) *OrderRouter {
	return &OrderRouter{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

// MapRoutes configures all REST routes for the order domain onto the provided router group.
// It establishes access controls, mounting customer routes requiring authentication and
// backoffice admin routes requiring both authentication and AdminRole RBAC verification.
func (r *OrderRouter) MapRoutes(vgroup *gin.RouterGroup) {
	// Administrative order endpoints group: protected by authentication and AdminRole RBAC.
	admin := vgroup.Group("/admin/orders")
	admin.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		// GET /admin/orders - paginated order listing with optional customerId and status filters.
		admin.GET("", r.handler.AdminListOrders)
		// GET /admin/orders/:id - retrieve detailed view of any order by ID.
		admin.GET("/:id", r.handler.AdminGetOrderByID)
		// PATCH /admin/orders/:id/status - transition order status and record audit log.
		admin.PATCH("/:id/status", r.handler.UpdateOrderStatus)
	}

	// Administrative dashboard and payments endpoints group
	adminRoot := vgroup.Group("/admin")
	adminRoot.Use(
		middleware.Authenticate(r.sessionStore),
		middleware.Authorize(model.AdminRole),
	)
	{
		adminRoot.GET("/dashboard", r.handler.GetDashboard)
		adminRoot.GET("/payments", r.handler.ListPayments)
		adminRoot.PATCH("/payments/:id/status", r.handler.UpdatePaymentStatus)
		adminRoot.POST("/payments/:id/refund", r.handler.RefundPayment)
	}

	// Customer order endpoints group: protected by customer session authentication.
	ordersGroup := vgroup.Group("/orders")
	{
		// GET /orders - list all orders placed by the currently authenticated customer.
		ordersGroup.GET("", middleware.Authenticate(r.sessionStore), r.handler.GetOrders)
		// GET /orders/:id - get specific order details ensuring ownership access control.
		ordersGroup.GET("/:id", middleware.Authenticate(r.sessionStore), r.handler.GetOrderByID)
		// POST /orders/:id/cancel - request cancellation of a pending/confirmed order by customer.
		ordersGroup.POST("/:id/cancel", middleware.Authenticate(r.sessionStore), r.handler.CancelOrder)
	}
}
