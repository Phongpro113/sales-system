package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Email     string    `json:"email" gorm:"unique;not null"`
    Password  string    `json:"-"`
    Name      string    `json:"name"`
    Role      string    `json:"role" gorm:"default:customer"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Name     string `json:"name"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

var db *gorm.DB
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func initDB() {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "host=postgres user=postgres password=postgres dbname=auth_db port=5432 sslmode=disable"
    }
    
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Auto migrate
    if err := db.AutoMigrate(&User{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
    
    // Create admin user if not exists
    var adminCount int64
    db.Model(&User{}).Where("role = ?", "admin").Count(&adminCount)
    if adminCount == 0 {
        hashedPassword, _ := hashPassword("admin123")
        admin := User{
            Email:    "admin@example.com",
            Password: hashedPassword,
            Name:     "Administrator",
            Role:     "admin",
        }
        db.Create(&admin)
        log.Println("Admin user created: admin@example.com / admin123")
    }
}

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func generateToken(user User) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
        "name":    user.Name,
        "role":    user.Role,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
        "iat":     time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate input
    if req.Email == "" || req.Password == "" || req.Name == "" {
        http.Error(w, "Email, password, and name are required", http.StatusBadRequest)
        return
    }
    
    hashedPassword, err := hashPassword(req.Password)
    if err != nil {
        http.Error(w, "Failed to hash password", http.StatusInternalServerError)
        return
    }
    
    user := User{
        Email:    req.Email,
        Password: hashedPassword,
        Name:     req.Name,
        Role:     "customer",
    }
    
    if err := db.Create(&user).Error; err != nil {
        http.Error(w, "Email already exists", http.StatusConflict)
        return
    }
    
    token, err := generateToken(user)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(AuthResponse{
        Token: token,
        User:  user,
    })
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    var user User
    if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    if !checkPasswordHash(req.Password, user.Password) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    
    token, err := generateToken(user)
    if err != nil {
        http.Error(w, "Failed to generate token", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(AuthResponse{
        Token: token,
        User:  user,
    })
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
    tokenString := r.Header.Get("Authorization")
    if tokenString == "" {
        http.Error(w, "No token provided", http.StatusUnauthorized)
        return
    }
    
    // Remove "Bearer " prefix if present
    if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
        tokenString = tokenString[7:]
    }
    
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, jwt.ErrSignatureInvalid
        }
        return jwtSecret, nil
    })
    
    if err != nil || !token.Valid {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    
    claims := token.Claims.(jwt.MapClaims)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(claims)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "service": "auth-service",
    })
}

func main() {
    // Set default JWT secret if not provided
    if len(jwtSecret) == 0 {
        jwtSecret = []byte("default-secret-key-change-in-production")
        log.Println("WARNING: Using default JWT secret. Set JWT_SECRET environment variable for production!")
    }
    
    initDB()
    
    r := mux.NewRouter()
    
    // API routes
    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/register", registerHandler).Methods("POST", "OPTIONS")
    api.HandleFunc("/login", loginHandler).Methods("POST", "OPTIONS")
    api.HandleFunc("/validate", validateHandler).Methods("GET", "OPTIONS")
    api.HandleFunc("/health", healthHandler).Methods("GET")
    
    // CORS configuration
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type", "X-User-ID"},
        AllowCredentials: true,
        Debug:            false,
    })
    
    handler := c.Handler(r)
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8001"
    }
    
    log.Printf("Auth service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}