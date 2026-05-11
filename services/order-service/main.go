package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var productServiceURL = os.Getenv("PRODUCT_SERVICE_URL")
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}

func main() {
	if productServiceURL == "" {
		productServiceURL = "http://localhost:8002"
		log.Println("WARNING: PRODUCT_SERVICE_URL not set, using default:", productServiceURL)
	}
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-secret-key-change-in-production")
		log.Println("WARNING: Using default JWT secret")
	}

	initDB()

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/orders", createOrderHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/orders", getOrdersHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/orders/{id:[0-9]+}", getOrderHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/orders/{id:[0-9]+}/status", updateOrderStatusHandler).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/orders/{id:[0-9]+}/cancel", cancelOrderHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/health", healthHandler).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-User-ID", "X-User-Role"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	log.Printf("Order service starting on port %s", port)
	log.Printf("Product service URL: %s", productServiceURL)
	log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}
