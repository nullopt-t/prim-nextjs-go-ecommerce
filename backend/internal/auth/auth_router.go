package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/middleware"
	"github.com/nullopt-t/prim-backend/pkg/config"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// Router registers and manages HTTP route bindings for authentication flows,
// public passwordless challenge handshakes, rate limiting policies, and protected session operations.
type Router struct {
	authHandler   *Handler
	secrets       config.Secrets
	limiter       middleware.RateLimiter
	logger        log.Logger
	sessionStore  middleware.SessionStore
	rateLimitCfg  config.RateLimitConfig
	userResolver  middleware.UserResolver
}

// NewRouter constructs a new authentication router with dependencies needed for routing,
// rate limiting enforcement, and session authentication middleware.
// userResolver, when non-nil, re-validates each user's account status and role
// from the database on every authenticated request.
func NewRouter(
	ah *Handler,
	secrets config.Secrets,
	limiter middleware.RateLimiter,
	logger log.Logger,
	sessionStore middleware.SessionStore,
	rateLimitCfg config.RateLimitConfig,
	userResolver ...middleware.UserResolver,
) *Router {
	r := &Router{
		authHandler:  ah,
		secrets:      secrets,
		limiter:      limiter,
		logger:       logger,
		sessionStore: sessionStore,
		rateLimitCfg: rateLimitCfg,
	}
	if len(userResolver) > 0 {
		r.userResolver = userResolver[0]
	}
	return r
}

// MapRoutes configures public challenge endpoints and protected session endpoints on the provided Gin router group.
// It applies IP-level rate limiting middleware to prevent brute-force and resource exhaustion attacks on challenge endpoints,
// and enforces session authentication middleware on sensitive user profile and session revocation endpoints.
func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	// Base group for authentication routes (/api/v1/auth)
	auth := vgroup.Group("/auth")

	// Challenge sub-group for public passwordless OTP authentication (/api/v1/auth/challenge)
	// Created before auth.Use(middleware.Authenticate) so these remain publicly accessible without a session cookie.
	challenge := auth.Group("/challenge")

	// POST /api/v1/auth/challenge/start
	// Rate-limited endpoint to initiate an OTP challenge for a provided email identifier.
	challenge.POST("/start",
		middleware.RateLimit(r.limiter, "auth:start",
			r.rateLimitCfg.AuthStartRequests,
			r.rateLimitCfg.AuthStartWindow,
		),
		r.authHandler.StartChallenge)

	// POST /api/v1/auth/challenge/resend
	// Rate-limited endpoint to resend an OTP verification code if the challenge is active.
	challenge.POST("/resend",
		middleware.RateLimit(r.limiter, "auth:resend",
			r.rateLimitCfg.AuthResendRequests,
			r.rateLimitCfg.AuthResendWindow,
		),
		r.authHandler.ResendChallenge)

	// POST /api/v1/auth/challenge/verify
	// Rate-limited endpoint to verify the one-time code and issue a new authenticated session cookie.
	challenge.POST("/verify",
		middleware.RateLimit(r.limiter, "auth:verify",
			r.rateLimitCfg.AuthVerifyRequests,
			r.rateLimitCfg.AuthVerifyWindow,
		),
		r.authHandler.VerifyChallenge)

	// Protected endpoints: enforce server-side session authentication middleware
	// A dedicated sub-group guarantees that protected middleware does not leak or depend on declaration order.
	protected := auth.Group("")
	protected.Use(middleware.Authenticate(r.sessionStore, r.userResolver))

	// GET /api/v1/auth/me - Retrieve current authenticated user profile
	protected.GET("/me", r.authHandler.GetMe)

	// GET /api/v1/auth/sessions - List all active sessions registered for the authenticated user
	protected.GET("/sessions", r.authHandler.GetSessions)

	// DELETE /api/v1/auth/sessions - Revoke all active sessions for the user and clear cookies
	protected.DELETE("/sessions", r.authHandler.DeleteAllSessions)

	// DELETE /api/v1/auth/sessions/:id - Revoke a specific active session by ID
	protected.DELETE("/sessions/:id", r.authHandler.DeleteSessionByID)

	// POST /api/v1/auth/logout - Revoke the current active session and clear auth cookies
	protected.POST("/logout", r.authHandler.Logout)
}
