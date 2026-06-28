package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type AuthHandler interface {
	StartChallenge(c *gin.Context)
	ResendChallenge(c *gin.Context)
	VerifyChallenge(c *gin.Context)
	Refresh(c *gin.Context)
	GetMe(c *gin.Context)
	// GetSessions(c *gin.Context)
	// DeleteSessionByID(c *gin.Context)
}

type Router struct {
	authHandler AuthHandler
	secrets     *config.Secrets
}

func NewRouter(
	ah AuthHandler,
	secrets *config.Secrets,
) *Router {
	return &Router{
		authHandler: ah,
		secrets:     secrets,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	auth := vgroup.Group("/auth")
	challenge := auth.Group("/challenge")
	challenge.POST("/start", r.authHandler.StartChallenge)
	challenge.POST("/resend", r.authHandler.ResendChallenge)
	challenge.POST("/verify", r.authHandler.VerifyChallenge)
	auth.POST("/refresh", r.authHandler.Refresh)

	// protected
	auth.Use(middleware.Authanticate(r.secrets))
	auth.GET("/me", r.authHandler.GetMe)

	// auth.GET("/email-status", r.authHandler.EmailStatus)
	// auth.GET("/sessions", r.authHandler.GetSessions)
	// auth.DELETE("/sessions/:id", r.authHandler.DeleteSessionByID)
}
