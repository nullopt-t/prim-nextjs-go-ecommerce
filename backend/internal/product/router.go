package product

import "github.com/gin-gonic/gin"

type ProductHandler interface {
	GetAllProducts(c *gin.Context)
	CreateProduct(c *gin.Context)
	GetProductByID(c *gin.Context)
}

type Router struct {
	handler ProductHandler
}

func NewRouter(handler ProductHandler) *Router {
	return &Router{
		handler: handler,
	}
}

func (r *Router) MapRoutes(vgroup *gin.RouterGroup) {
	products := vgroup.Group("/products")
	products.GET("", r.handler.GetAllProducts)
	products.POST("", r.handler.CreateProduct)
	products.GET("/:id", r.handler.GetProductByID)
}
