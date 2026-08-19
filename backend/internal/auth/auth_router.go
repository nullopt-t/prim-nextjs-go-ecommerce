package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/redis/go-redis/v9"
)

type Router struct {
	authHandler  *Handler
	secrets      config.Secrets
	limiter      middleware.RateLimiter
	logger       log.Logger
	redisClient  *redis.Client
	rateLimitCfg config.RateLimitConfig
}

func NewRouter(
	ah *Handler,
	secrets config.Secrets,
	limiter middleware.RateLimiter,
	logger log.Logger,
	redisClient *redis.Client,
	rateLimitCfg config.RateLimitConfig,
) *Router {
	return &Router{
		authHandler:  ah,
		secrets:      secrets,
		limiter:      limiter,
		logger:       logger,
		redisClient:  redisClient,
		rateLimitCfg: rateLimitCfg,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	auth := vgroup.Group("/auth")
	challenge := auth.Group("/challenge")
	challenge.POST("/start", middleware.RateLimit(r.limiter, "rate_limit:auth:start", r.rateLimitCfg.AuthStartRequests, r.rateLimitCfg.AuthStartWindow, r.logger), r.authHandler.StartChallenge)
	challenge.POST("/resend", middleware.RateLimit(r.limiter, "rate_limit:auth:resend", r.rateLimitCfg.AuthResendRequests, r.rateLimitCfg.AuthResendWindow, r.logger), r.authHandler.ResendChallenge)
	challenge.POST("/verify", middleware.RateLimit(r.limiter, "rate_limit:auth:verify", r.rateLimitCfg.AuthVerifyRequests, r.rateLimitCfg.AuthVerifyWindow, r.logger), r.authHandler.VerifyChallenge)
	auth.POST("/refresh", r.authHandler.Refresh)

	// protected
	auth.Use(middleware.Authenticate(r.secrets))
	auth.Use(middleware.StrictSessionCheck(r.redisClient))
	auth.GET("/me", r.authHandler.GetMe)
	auth.GET("/sessions", r.authHandler.GetSessions)
	auth.DELETE("/sessions/:id", r.authHandler.DeleteSessionByID)
	auth.POST("/logout", r.authHandler.Logout)
}
