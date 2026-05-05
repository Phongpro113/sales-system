package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
    
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Product struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Description string    `json:"description"`
    Price       float64   `json:"price" gorm:"not null"`
    Stock       int       `json:"stock" gorm:"not null;default:0"`
    Category    string    `json:"category"`
    ImageURL    string    `json:"image_url"`
    SKU         string    `json:"sku" gorm:"unique"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateStockRequest struct {
    Quantity int `json:"quantity"` // Can be positive or negative
}

var db *gorm.DB

func initDB() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=postgres user=postgres password=postgres dbname=product_db port=5432 sslmode=disable"
    }
    
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Auto migrate
    if err := db.AutoMigrate(&Product{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
    
    // Seed products if database is empty
    var count int64
    db.Model(&Product{}).Count(&count)
    if count == 0 {
        seedProducts()
    }
}

func seedProducts() {
    products := []Product{
        {
            Name:        "MacBook Pro 14",
            Description: "Apple M3 Pro chip, 16GB RAM, 512GB SSD",
            Price:       1999.99,
            Stock:       25,
            Category:    "Electronics",
            ImageURL:    "https://images.unsplash.com/photo-1517336714731-489689fd1ca8",
            SKU:         "MBP-14-M3-16-512",
        },
        {
            Name:        "Sony WH-1000XM5",
            Description: "Wireless Noise Canceling Headphones",
            Price:       399.99,
            Stock:       50,
            Category:    "Electronics",
            ImageURL:    "https://images.unsplash.com/photo-1618366712010-f4ae9c647dcb",
            SKU:         "SNY-XM5-BLK",
        },
        {
            Name:        "Nike Air Max",
            Description: "Running Shoes, breathable mesh upper",
            Price:       129.99,
            Stock:       100,
            Category:    "Clothing",
            ImageURL:    "https://images.unsplash.com/photo-1542291026-7eec264c27ff",
            SKU:         "NKE-AM-BLK",
        },
        {
            Name:        "The Go Programming Language",
            Description: "By Alan A. A. Donovan & Brian W. Kernighan",
            Price:       49.99,
            Stock:       75,
            Category:    "Books",
            ImageURL:    "https://images.unsplash.com/photo-1589998059171-988d887df646",
            SKU:         "BOOK-GO-001",
        },
        {
            Name:        "Mechanical Keyboard",
            Description: "RGB Backlit Mechanical Gaming Keyboard",
            Price:       89.99,
            Stock:       150,
            Category:    "Electronics",
            ImageURL:    "https://images.unsplash.com/photo-1595225476474-87563907a212",
            SKU:         "KEY-MECH-RGB",
        },
        {
            Name:        "Coffee Mug",
            Description: "Ceramic Coffee Mug, 15oz",
            Price:       14.99,
            Stock:       200,
            Category:    "Home",
            ImageURL:    "https://images.unsplash.com/photo-1514228742587-6b1558fcca3d",
            SKU:         "MUG-CER-BLK",
        },
    }
    
    for _, product := range products {
        db.Create(&product)
    }
    log.Println("Seeded database with sample products")
}

func getProductsHandler(w http.ResponseWriter, r *http.Request) {
    var products []Product
    query := db.Model(&Product{})
    
    // Filter by category
    if category := r.URL.Query().Get("category"); category != "" {
        query = query.Where("LOWER(category) = LOWER(?)", category)
    }
    
    // Search by name or description
    if search := r.URL.Query().Get("search"); search != "" {
        query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
    }
    
    // Filter by min price
    if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
        if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
            query = query.Where("price >= ?", price)
        }
    }
    
    // Filter by max price
    if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
        if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
            query = query.Where("price <= ?", price)
        }
    }
    
    // Filter by in stock
    if inStock := r.URL.Query().Get("in_stock"); inStock == "true" {
        query = query.Where("stock > 0")
    }
    
    // Sorting
    if sort := r.URL.Query().Get("sort"); sort != "" {
        switch sort {
        case "price_asc":
            query = query.Order("price ASC")
        case "price_desc":
            query = query.Order("price DESC")
        case "name_asc":
            query = query.Order("name ASC")
        case "name_desc":
            query = query.Order("name DESC")
        case "newest":
            query = query.Order("created_at DESC")
        default:
            query = query.Order("id DESC")
        }
    } else {
        query = query.Order("id DESC")
    }
    
    // Pagination
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit < 1 || limit > 100 {
        limit = 20
    }
    offset := (page - 1) * limit
    
    var total int64
    query.Model(&Product{}).Count(&total)
    query.Offset(offset).Limit(limit).Find(&products)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "products": products,
        "total":    total,
        "page":     page,
        "limit":    limit,
        "total_pages": (total + int64(limit) - 1) / int64(limit),
    })
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }
    
    var product Product
    if err := db.First(&product, id).Error; err != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

func createProductHandler(w http.ResponseWriter, r *http.Request) {
    var product Product
    if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate required fields
    if product.Name == "" || product.Price <= 0 {
        http.Error(w, "Name and positive price are required", http.StatusBadRequest)
        return
    }
    
    if err := db.Create(&product).Error; err != nil {
        if strings.Contains(err.Error(), "duplicate key") {
            http.Error(w, "SKU already exists", http.StatusConflict)
        } else {
            http.Error(w, "Failed to create product", http.StatusInternalServerError)
        }
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(product)
}

func updateProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }
    
    var product Product
    if err := db.First(&product, id).Error; err != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }
    
    var updates map[string]interface{}
    if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Don't allow updating ID
    delete(updates, "id")
    delete(updates, "created_at")
    
    if err := db.Model(&product).Updates(updates).Error; err != nil {
        http.Error(w, "Failed to update product", http.StatusInternalServerError)
        return
    }
    
    db.First(&product, id)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

func updateStockHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }
    
    var req UpdateStockRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    var product Product
    if err := db.First(&product, id).Error; err != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }
    
    newStock := product.Stock + req.Quantity
    if newStock < 0 {
        http.Error(w, "Insufficient stock", http.StatusBadRequest)
        return
    }
    
    product.Stock = newStock
    db.Save(&product)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

func deleteProductHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseUint(vars["id"], 10, 32)
    if err != nil {
        http.Error(w, "Invalid product ID", http.StatusBadRequest)
        return
    }
    
    result := db.Delete(&Product{}, id)
    if result.Error != nil {
        http.Error(w, "Failed to delete product", http.StatusInternalServerError)
        return
    }
    
    if result.RowsAffected == 0 {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}

func getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
    var categories []string
    db.Model(&Product{}).Distinct("category").Pluck("category", &categories)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(categories)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "healthy",
        "service": "product-service",
    })
}

func main() {
    initDB()
    
    r := mux.NewRouter()
    
    // API routes
    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/products", getProductsHandler).Methods("GET", "OPTIONS")
    api.HandleFunc("/products/{id:[0-9]+}", getProductHandler).Methods("GET", "OPTIONS")
    api.HandleFunc("/products", createProductHandler).Methods("POST", "OPTIONS")
    api.HandleFunc("/products/{id:[0-9]+}", updateProductHandler).Methods("PUT", "OPTIONS")
    api.HandleFunc("/products/{id:[0-9]+}/stock", updateStockHandler).Methods("PATCH", "OPTIONS")
    api.HandleFunc("/products/{id:[0-9]+}", deleteProductHandler).Methods("DELETE", "OPTIONS")
    api.HandleFunc("/categories", getCategoriesHandler).Methods("GET", "OPTIONS")
    api.HandleFunc("/health", healthHandler).Methods("GET")
    
    // CORS configuration
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