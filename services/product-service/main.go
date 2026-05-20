package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	initDB()
	startKafkaConsumer()

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/products", getProductsHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/products/{id:[0-9]+}", getProductHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/products", createProductHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/products/{id:[0-9]+}", updateProductHandler).Methods("PUT", "OPTIONS")
	api.HandleFunc("/products/{id:[0-9]+}/stock", updateStockHandler).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/products/{id:[0-9]+}/validate-buy", validateBuyHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/products/{id:[0-9]+}", deleteProductHandler).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/categories", getCategoriesHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/health", healthHandler).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-User-ID"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	log.Printf("Product service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}
