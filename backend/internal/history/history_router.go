package history

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/pkg/config"
)

// Router maps the authenticated browsing-history endpoints.
type Router struct {
	handler      *Handler
	sessionStore middleware.SessionStore
	limiter      middleware.RateLimiter
	rateLimitCfg config.RateLimitConfig
}

// NewRouter constructs a browsing-history router.
func NewRouter(
	handler *Handler,
	sessionStore middleware.SessionStore,
	limiter middleware.RateLimiter,
	rateLimitCfg config.RateLimitConfig,
) *Router {
	if handler == nil {
		panic("history: handler is required")
	}
	if sessionStore == nil {
		panic("history: session store is required")
	}
	if limiter == nil {
		panic("history: rate limiter is required")
	}
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
		limiter:      limiter,
		rateLimitCfg: rateLimitCfg,
	}
}

// MapRoutes registers the customer history API under /user/history.
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	group := vgroup.Group("/user/history")
	group.Use(middleware.Authenticate(r.sessionStore))

	group.GET("/products", r.handler.GetHistory)

	writeLimit := middleware.RateLimit(
		r.limiter,
		"history:write",
		r.rateLimitCfg.HistoryWriteRequests,
		r.rateLimitCfg.HistoryWriteWindow,
	)
	// Register the static sync path before the parameterized record path.
	group.POST("/products/sync", writeLimit, r.handler.SyncHistory)
	group.POST("/products/:id", writeLimit, r.handler.RecordView)

	group.DELETE("/products/:id", r.handler.DeleteHistoryItem)
	group.DELETE("/products", r.handler.ClearHistory)
}
