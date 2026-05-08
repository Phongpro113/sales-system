package route

import (
	"admin-service/controller"
	"admin-service/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter(productCtrl *controller.ProductController) *gin.Engine {
	r := gin.Default()

	r.Static("/uploads", "./uploads")
	api := r.Group("/api/admin")
	api.Use(middlewares.AdminAuthMiddleware())
	{
		api.GET("/products", productCtrl.GetProducts)
		api.GET("/products/:id", productCtrl.GetProduct)
		api.POST("/products", productCtrl.CreateProduct)
		api.PUT("/products/:id", productCtrl.UpdateProduct)
		api.POST("/upload", productCtrl.UploadImage)
		api.GET("/health", productCtrl.HealthCheck)
	}

	return r
}
