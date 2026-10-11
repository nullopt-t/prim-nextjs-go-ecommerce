package user

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetUserByID(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetAllUsers(c *gin.Context)
}

type Router struct {
	handler      UserHandler
	sessionStore middleware.SessionStore
}

func NewRouter(handler UserHandler, sessionStore middleware.SessionStore) *Router {
	return &Router{
		handler:      handler,
		sessionStore: sessionStore,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	users := vgroup.Group("/users")
	users.Use(
		middleware.Authenticate(r.sessionStore),
	)

	users.GET("", r.handler.GetAllUsers)
	users.GET("/:id", r.handler.GetUserByID)
	users.DELETE("/:id", r.handler.DeleteUser)
}
