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
	users := vgroup.Group("/users")
	users.Use(
		middleware.Authanticate(r.config.KeysCfg),
	)

	users.GET("", r.handler.GetAllUsers)
	users.GET("/:id", r.handler.GetUserByID)
	users.DELETE("/:id", r.handler.DeleteUser)
}
