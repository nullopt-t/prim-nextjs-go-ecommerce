package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

type Router struct {
	authHandler     *Handler
	secrets         config.Secrets
	limiter         middleware.RateLimiter
	logger          log.Logger
	sessionStore middleware.SessionStore
	rateLimitCfg config.RateLimitConfig
}

func NewRouter(
	ah *Handler,
	secrets config.Secrets,
	limiter middleware.RateLimiter,
	logger log.Logger,
	sessionStore middleware.SessionStore,
	rateLimitCfg config.RateLimitConfig,
) *Router {
	return &Router{
		authHandler:  ah,
		secrets:      secrets,
		limiter:      limiter,
		logger:       logger,
		sessionStore: sessionStore,
		rateLimitCfg: rateLimitCfg,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	auth := vgroup.Group("/auth")
	challenge := auth.Group("/challenge")

	challenge.POST("/start",
		middleware.RateLimit(r.limiter, "auth:start",
			r.rateLimitCfg.AuthStartRequests,
			r.rateLimitCfg.AuthStartWindow,
		),
		r.authHandler.StartChallenge)

	challenge.POST("/resend",
		middleware.RateLimit(r.limiter, "auth:resend",
			r.rateLimitCfg.AuthResendRequests,
			r.rateLimitCfg.AuthResendWindow,
		),
		r.authHandler.ResendChallenge)

	challenge.POST("/verify",
		middleware.RateLimit(r.limiter, "auth:verify",
			r.rateLimitCfg.AuthVerifyRequests,
			r.rateLimitCfg.AuthVerifyWindow,
		),
		r.authHandler.VerifyChallenge)


	// protected
	auth.Use(middleware.Authenticate(r.sessionStore))
	auth.GET("/me", r.authHandler.GetMe)
	auth.GET("/sessions", r.authHandler.GetSessions)
	auth.DELETE("/sessions", r.authHandler.DeleteAllSessions)
	auth.DELETE("/sessions/:id", r.authHandler.DeleteSessionByID)
	auth.POST("/logout", r.authHandler.Logout)
}
