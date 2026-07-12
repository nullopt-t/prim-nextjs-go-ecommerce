package product

import "github.com/gin-gonic/gin"

type ProductRouter struct {
	ph *ProductHandler
}

func NewRouter(
	ph *ProductHandler,
) *ProductRouter {
	return &ProductRouter{
		ph: ph,
	}
}

func (r *ProductRouter) MapRoutes(vgroup *gin.RouterGroup) {
	products := vgroup.Group("/products")
	products.GET("", r.ph.GetAllProducts)
	products.POST("", r.ph.CreateProduct)
	products.GET("/:id", r.ph.GetProductByID)
	products.GET("/slug/:slug", r.ph.GetProductBySlug)
}
