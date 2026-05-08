package route

import (
	"admin-service/controller"
	"admin-service/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter(productCtrl *controller.ProductController) *gin.Engine {
	r := gin.Default()

	// Auth middleware to check for admin role
	r.Use(middlewares.AdminAuthMiddleware())

	api := r.Group("/api/admin")
	{
		api.GET("/products", productCtrl.GetProducts)
		api.GET("/products/:id", productCtrl.GetProduct)
		api.POST("/products", productCtrl.CreateProduct)
		api.PUT("/products/:id", productCtrl.UpdateProduct)
		api.GET("/health", productCtrl.HealthCheck)
	}

	return r
}
