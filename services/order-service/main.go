package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"
    
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Order struct {
    ID         uint        `json:"id" gorm:"primaryKey"`
    OrderNumber string     `json:"order_number" gorm:"unique;not null"`
    UserID     uint        `json:"user_id" gorm:"not null;index"`
    Total      float64     `json:"total"`
    Status     string      `json:"status" gorm:"default:pending"` // pending, confirmed, paid, shipped, delivered, cancelled
    PaymentMethod string   `json:"payment_method"`
    ShippingAddress string `json:"shipping_address"`
    Notes       string      `json:"notes"`
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
    Items       []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
    ID          uint    `json:"id" gorm:"primaryKey"`
    OrderID     uint    `json:"order_id" gorm:"index"`
    ProductID   uint    `json:"product_id" gorm:"not null"`
    ProductName string  `json:"product_name" gorm:"-"`
    Quantity    int     `json:"quantity" gorm:"not null"`
    Price       float64 `json:"price" gorm:"not null"`
    Subtotal    float64 `json:"subtotal" gorm:"-"`
}

type Product struct {
    ID    uint    `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
    Stock int     `json:"stock"`
}

type CreateOrderRequest struct {
    Items []struct {
        ProductID uint `json:"product_id"`
        Quantity  int  `json:"quantity"`
    } `json:"items"`
    PaymentMethod   string `json:"payment_method"`
    ShippingAddress string `json:"shipping_address"`
    Notes          string `json:"notes"`
}

var db *gorm.DB
var productServiceURL = os.Getenv("PRODUCT_SERVICE_URL")
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func initDB() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=postgres user=postgres password=postgres dbname=order_db port=5432 sslmode=disable"
    }
    
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    if err := db.AutoMigrate(&Order{}, &OrderItem{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
}

func enrichItemsWithProductNames(items []OrderItem) []OrderItem {
    for i := range items {
        if items[i].ProductName != "" {
            items[i].Subtotal = items[i].Price * float64(items[i].Quantity)
            continue
        }
        product, err := getProduct(items[i].ProductID)
        if err == nil {
            items[i].ProductName = product.Name
        } else {
            items[i].ProductName = fmt.Sprintf("Product #%d", items[i].ProductID)
        }
        items[i].Subtotal = items[i].Price * float64(items[i].Quantity)
    }
    return items
}

func generateOrderNumber() string {
    return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}

func getProduct(id uint) (*Product, error) {
    url := fmt.Sprintf("%s/api/products/%d", productServiceURL, id)
    
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to call product service: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("product service returned %d: %s", resp.StatusCode, string(body))
    }
    
    var product Product
    if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
        return nil, fmt.Errorf("failed to decode product: %w", err)
    }
    
    return &product, nil
}

func updateProductStock(productID uint, quantity int) error {
    url := fmt.Sprintf("%s/api/products/%d/stock", productServiceURL, productID)
    
    reqBody := map[string]int{"quantity": -quantity}
    jsonData, _ := json.Marshal(reqBody)
    
    req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("stock update failed: %s", string(body))
    }
    
    return nil
}

func createOrderHandler(w http.ResponseWriter, r *http.Request) {
    // Get user ID from header (set by API gateway)
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
    
    // Validate products and calculate totals
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
        
        subtotal := product.Price * float64(item.Quantity)
        total += subtotal
        
        orderItems = append(orderItems, OrderItem{
            ProductID:   item.ProductID,
            ProductName: product.Name,
            Quantity:    item.Quantity,
            Price:       product.Price,
        })
    }
    
    // Create order
    order := Order{
        OrderNumber:     generateOrderNumber(),
        UserID:         uint(userID),
        Total:          total,
        Status:         "pending",
        PaymentMethod:  req.PaymentMethod,
        ShippingAddress: req.ShippingAddress,
        Notes:          req.Notes,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
    }
    
    // Use transaction
    err = db.Transaction(func(tx *gorm.DB) error {
        // Create order
        if err := tx.Create(&order).Error; err != nil {
            return err
        }
        
        // Create order items
        for i := range orderItems {
            orderItems[i].OrderID = order.ID
        }
        if err := tx.Create(&orderItems).Error; err != nil {
            return err
        }
        
        // Update stock for each product (call external service)
        for _, item := range req.Items {
            if err := updateProductStock(item.ProductID, item.Quantity); err != nil {
                return fmt.Errorf("failed to update stock for product %d: %w", item.ProductID, err)
            }
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
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(order)
}

func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
    // Get user ID from header
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
    
    // Check if admin (from header)
    role := r.Header.Get("X-User-Role")
    
    var orders []Order
    query := db.Preload("Items")
    
    if role != "admin" {
        query = query.Where("user_id = ?", userID)
    }
    
    // Pagination
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
    query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders)
    
    for i := range orders {
        orders[i].Items = enrichItemsWithProductNames(orders[i].Items)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "orders": orders,
        "total":  total,
        "page":   page,
        "limit":  limit,
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
    
    var order Order
    query := db.Preload("Items")
    
    if role != "admin" {
        query = query.Where("user_id = ?", userID)
    }
    
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
    
    // Restore stock for cancelled order
    var items []OrderItem
    db.Where("order_id = ?", id).Find(&items)
    
    for _, item := range items {
        if err := updateProductStock(item.ProductID, -item.Quantity); err != nil {
            log.Printf("Warning: Failed to restore stock for product %d: %v", item.ProductID, err)
        }
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
    
    // API routes
    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/orders", createOrderHandler).Methods("POST", "OPTIONS")
    api.HandleFunc("/orders", getOrdersHandler).Methods("GET", "OPTIONS")
    api.HandleFunc("/orders/{id:[0-9]+}", getOrderHandler).Methods("GET", "OPTIONS")
    api.HandleFunc("/orders/{id:[0-9]+}/status", updateOrderStatusHandler).Methods("PATCH", "OPTIONS")
    api.HandleFunc("/orders/{id:[0-9]+}/cancel", cancelOrderHandler).Methods("POST", "OPTIONS")
    api.HandleFunc("/health", healthHandler).Methods("GET")
    
    // CORS configuration
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