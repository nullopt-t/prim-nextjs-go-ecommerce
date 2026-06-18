package auth

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	VerifyEmail(c *gin.Context)
	ResendOTP(c *gin.Context)
}

type Router struct {
	authHandler AuthHandler
}

func NewRouter(ah AuthHandler) *Router {
	return &Router{
		authHandler: ah,
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
}
