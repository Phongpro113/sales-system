package main

import (
	"admin-service/config"
	"admin-service/controller"
	"admin-service/database"
	"admin-service/repositories"
	"admin-service/route"
	"log"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db := database.InitDB(cfg.DatabaseURL)

	// Initialize repositories
	productRepo := repositories.NewProductRepository(db)

	// Initialize controllers
	productCtrl := controller.NewProductController(productRepo)

	// Setup router
	r := route.SetupRouter(productCtrl)

	log.Printf("Admin service starting on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
