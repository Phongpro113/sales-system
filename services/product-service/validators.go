package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// validateBuyHandler validates whether a product can be purchased.
// Called by frontend when user clicks "Buy now" before adding to cart.
func validateBuyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req ValidateBuyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	var product Product
	if err := db.First(&product, id).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ValidateBuyResponse{
			Valid:   false,
			Message: "Product not found",
		})
		return
	}

	if product.Stock <= 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidateBuyResponse{
			Valid:   false,
			Message: "Product is out of stock",
			Stock:   product.Stock,
			Price:   product.Price,
		})
		return
	}

	if product.Stock < req.Quantity {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ValidateBuyResponse{
			Valid:   false,
			Message: "Not enough stock. Available: " + strconv.Itoa(product.Stock),
			Stock:   product.Stock,
			Price:   product.Price,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ValidateBuyResponse{
		Valid:  true,
		Stock:  product.Stock,
		Price:  product.Price,
	})
}
