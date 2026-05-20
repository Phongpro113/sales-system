package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "Unauthorized: no user ID", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		http.Error(w, "Order must have at least one item", http.StatusBadRequest)
		return
	}

	var total float64
	var orderItems []OrderItem

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			http.Error(w, fmt.Sprintf("Invalid quantity for product %d", item.ProductID), http.StatusBadRequest)
			return
		}

		product, err := getProduct(item.ProductID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Product %d not found: %v", item.ProductID, err), http.StatusBadRequest)
			return
		}

		if product.Stock < item.Quantity {
			http.Error(w, fmt.Sprintf("Insufficient stock for product %s. Available: %d", product.Name, product.Stock), http.StatusBadRequest)
			return
		}

		total += product.Price * float64(item.Quantity)

		orderItems = append(orderItems, OrderItem{
			ProductID:   item.ProductID,
			ProductName: product.Name,
			Quantity:    item.Quantity,
			Price:       product.Price,
		})
	}

	order := Order{
		OrderNumber:     generateOrderNumber(),
		UserID:          uint(userID),
		Total:           total,
		Status:          "pending",
		PaymentMethod:   req.PaymentMethod,
		ShippingAddress: req.ShippingAddress,
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create order: %v", err), http.StatusInternalServerError)
		return
	}

	for i := range orderItems {
		orderItems[i].Subtotal = orderItems[i].Price * float64(orderItems[i].Quantity)
	}
	order.Items = orderItems

	publishOrderCreated(order, orderItems)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	role := r.Header.Get("X-User-Role")
	query := db.Preload("Items")
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	query.Model(&Order{}).Count(&total)

	var orders []Order
	query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders)

	for i := range orders {
		orders[i].Items = enrichItemsWithProductNames(orders[i].Items)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders":      orders,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

func getOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)
	role := r.Header.Get("X-User-Role")

	query := db.Preload("Items")
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	var order Order
	if err := query.First(&order, id).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	order.Items = enrichItemsWithProductNames(order.Items)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func updateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validStatuses := map[string]bool{
		"pending": true, "confirmed": true, "paid": true,
		"shipped": true, "delivered": true, "cancelled": true,
	}
	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	var order Order
	if err := db.First(&order, id).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	order.Status = req.Status
	order.UpdatedAt = time.Now()

	if err := db.Save(&order).Error; err != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func cancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var order Order
	if err := db.First(&order, id).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	if order.Status != "pending" && order.Status != "confirmed" {
		http.Error(w, "Only pending or confirmed orders can be cancelled", http.StatusBadRequest)
		return
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()
	db.Save(&order)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "order-service",
	})
}
