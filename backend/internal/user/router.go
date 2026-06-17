package user

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetUserByID(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetAllUsers(c *gin.Context)
	UserRegister(c *gin.Context)
	UserLogin(c *gin.Context)
	RefreshToken(c *gin.Context)
	ForgetPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	GetMe(c *gin.Context)
	VerifyOTP(c *gin.Context)
	ResendOTP(c *gin.Context)
}

type Router struct {
	handler UserHandler
	config  *config.Config
}

func NewRouter(handler UserHandler, config *config.Config) *Router {
	return &Router{
		handler: handler,
		config:  config,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	auth := vgroup.Group("/auth")
	auth.POST("/register", r.handler.UserRegister)
	auth.POST("/login", r.handler.UserLogin)
	auth.POST("/refresh", r.handler.RefreshToken)
	auth.POST("/forget-password", r.handler.ForgetPassword)
	auth.POST("/reset-password", r.handler.ResetPassword)
	auth.POST("/verify-otp", r.handler.VerifyOTP)
	auth.POST("/resend-otp", r.handler.ResendOTP)

	// protected auth routes
	auth.Use(
		middleware.Authanticate(r.config.KeysCfg),
	)

	users := vgroup.Group("/users")
	users.Use(
		middleware.Authanticate(r.config.KeysCfg),
	)

	users.GET("/me", r.handler.GetMe)
	users.GET("", r.handler.GetAllUsers)
	users.GET("/:id", r.handler.GetUserByID)
	users.DELETE("/:id", r.handler.DeleteUser)
}
