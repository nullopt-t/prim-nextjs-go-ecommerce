package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/redis/go-redis/v9"
)

type Router struct {
	authHandler *Handler
	secrets     *config.Secrets
	limiter     middleware.RateLimiter
	logger      log.Logger
	redisClient *redis.Client
}

func NewRouter(
	ah *Handler,
	secrets *config.Secrets,
	limiter middleware.RateLimiter,
	logger log.Logger,
	redisClient *redis.Client,
) *Router {
	return &Router{
		authHandler: ah,
		secrets:     secrets,
		limiter:     limiter,
		logger:      logger,
		redisClient: redisClient,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	auth := vgroup.Group("/auth")
	challenge := auth.Group("/challenge")
	challenge.POST("/start", middleware.RateLimit(r.limiter, "rate_limit:auth:start", 5, time.Minute, r.logger), r.authHandler.StartChallenge)
	challenge.POST("/resend", middleware.RateLimit(r.limiter, "rate_limit:auth:resend", 3, time.Minute, r.logger), r.authHandler.ResendChallenge)
	challenge.POST("/verify", middleware.RateLimit(r.limiter, "rate_limit:auth:verify", 10, time.Minute, r.logger), r.authHandler.VerifyChallenge)
	auth.POST("/refresh", r.authHandler.Refresh)

	// protected
	auth.Use(middleware.Authenticate(r.secrets))
	auth.Use(middleware.StrictSessionCheck(r.redisClient))
	auth.GET("/me", r.authHandler.GetMe)
	auth.GET("/sessions", r.authHandler.GetSessions)
	auth.DELETE("/sessions/:id", r.authHandler.DeleteSessionByID)
	auth.POST("/logout", r.authHandler.Logout)
}
