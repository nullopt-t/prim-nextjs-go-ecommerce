package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	VerifyEmail(c *gin.Context)
	ResendOTP(c *gin.Context)
	ChangePassword(c *gin.Context)
	GetMe(c *gin.Context)
	EmailStatus(c *gin.Context)
	GetSessions(c *gin.Context)
	DeleteSessionByID(c *gin.Context)
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
	auth.POST("/register", r.authHandler.Register)
	auth.POST("/login", r.authHandler.Login)
	auth.POST("/refresh", r.authHandler.Refresh)
	auth.POST("/forgot-password", r.authHandler.ForgotPassword)
	auth.POST("/reset-password", r.authHandler.ResetPassword)
	auth.POST("/verify-email", r.authHandler.VerifyEmail)
	auth.POST("/resend-otp", r.authHandler.ResendOTP)
	auth.POST("/change-password", r.authHandler.ChangePassword)

	// protected
	auth.Use(middleware.Authanticate(r.secrets))
	auth.GET("/me", r.authHandler.GetMe)
	auth.GET("/email-status", r.authHandler.EmailStatus)
	auth.GET("/sessions", r.authHandler.GetSessions)
	auth.DELETE("/sessions/:id", r.authHandler.DeleteSessionByID)
}
