package role

import "github.com/gin-gonic/gin"

type Handler interface {
	GetAll(c *gin.Context)
}

type RoleRouter struct {
	handler Handler
}

func NewRoleRouter(roleHandler Handler) *RoleRouter {
	return &RoleRouter{
		handler: roleHandler,
	}
}

func (r *RoleRouter) MapRoutes(vgroup *gin.RouterGroup) {
	roles := vgroup.Group("/roles")
	roles.GET("", r.handler.GetAll)
}
