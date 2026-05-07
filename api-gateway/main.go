package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "strings"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/gorilla/mux"
    "github.com/rs/cors"
)

type contextKey string

const (
    userIDKey contextKey = "user_id"
    userRoleKey contextKey = "user_role"
    userEmailKey contextKey = "user_email"
)

var (
    authServiceURL    string
    productServiceURL string
    orderServiceURL   string
    jwtSecret         []byte
)

func init() {
    authServiceURL = getEnv("AUTH_SERVICE_URL", "http://localhost:8001")
    productServiceURL = getEnv("PRODUCT_SERVICE_URL", "http://localhost:8002")
    orderServiceURL = getEnv("ORDER_SERVICE_URL", "http://localhost:8003")
    jwtSecret = []byte(getEnv("JWT_SECRET", "default-secret-key-change-in-production"))
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// publicPaths defines routes that don't require authentication
var publicPaths = []struct {
    path    string
    methods []string
}{
    {"/api/auth/register", []string{"POST"}},
    {"/api/auth/login", []string{"POST"}},
    {"/api/products", []string{"GET"}},
    {"/api/categories", []string{"GET"}},
}

// isPublicRoute checks if the request matches a public (no-auth) route
func isPublicRoute(r *http.Request) bool {
    if r.URL.Path == "/health" {
        return true
    }
    // Allow all OPTIONS (preflight) requests
    if r.Method == "OPTIONS" {
        return true
    }
    for _, p := range publicPaths {
        // Exact match or prefix match for paths like /api/products/123
        pathMatch := r.URL.Path == p.path || strings.HasPrefix(r.URL.Path, p.path+"/")
        if pathMatch {
            for _, m := range p.methods {
                if r.Method == m {
                    return true
                }
            }
        }
    }
    return false
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for public routes
        if isPublicRoute(r) {
            next.ServeHTTP(w, r)
            return
        }
        
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "No token provided", http.StatusUnauthorized)
            return
        }
        
        // Remove "Bearer " prefix
        tokenString = strings.TrimPrefix(tokenString, "Bearer ")
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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
        
        // Add user info to context
        ctx := context.WithValue(r.Context(), userIDKey, claims["user_id"])
        ctx = context.WithValue(ctx, userRoleKey, claims["role"])
        ctx = context.WithValue(ctx, userEmailKey, claims["email"])
        
        // Also add to headers for downstream services
        r.Header.Set("X-User-ID", toString(claims["user_id"]))
        r.Header.Set("X-User-Role", toString(claims["role"]))
        r.Header.Set("X-User-Email", toString(claims["email"]))
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func toString(v interface{}) string {
    if v == nil {
        return ""
    }
    return v.(string)
}

func proxy(target string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        remote, err := url.Parse(target)
        if err != nil {
            log.Printf("Failed to parse URL %s: %v", target, err)
            http.Error(w, "Bad gateway configuration", http.StatusBadGateway)
            return
        }
        
        proxy := httputil.NewSingleHostReverseProxy(remote)
        
        // Preserve the path
        r.URL.Host = remote.Host
        r.URL.Scheme = remote.Scheme
        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
        r.Host = remote.Host
        
        // Log the request
        log.Printf("Proxying %s %s -> %s", r.Method, r.URL.Path, target)
        
        proxy.ServeHTTP(w, r)
    }
}

func proxyWithContext(target string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Preserve the /api prefix so downstream services receive the routes
        // they already expose (for example, /api/products and /api/orders).
        originalPath := r.URL.Path
        
        remote, err := url.Parse(target)
        if err != nil {
            log.Printf("Failed to parse URL %s: %v", target, err)
            http.Error(w, "Bad gateway configuration", http.StatusBadGateway)
            return
        }
        
        proxy := httputil.NewSingleHostReverseProxy(remote)
        
        r.URL.Host = remote.Host
        r.URL.Scheme = remote.Scheme
        r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
        r.Host = remote.Host
        
        log.Printf("Proxying %s %s -> %s", r.Method, originalPath, target)
        
        proxy.ServeHTTP(w, r)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "service": "api-gateway",
        "timestamp": time.Now().Unix(),
    })
}

func main() {
    log.Println("API Gateway starting...")
    log.Printf("Auth Service URL: %s", authServiceURL)
    log.Printf("Product Service URL: %s", productServiceURL)
    log.Printf("Order Service URL: %s", orderServiceURL)
    
    r := mux.NewRouter()
    
    // Health check endpoint (no auth required)
    r.HandleFunc("/health", healthHandler).Methods("GET")
    
    // All API routes go through the subrouter with auth middleware.
    // Public routes are whitelisted in isPublicRoute() and skip auth.
    api := r.PathPrefix("/api").Subrouter()
    api.Use(authMiddleware)
    
    // Auth service routes
    api.HandleFunc("/auth/register", proxy(authServiceURL+"/api/register")).Methods("POST", "OPTIONS")
    api.HandleFunc("/auth/login", proxy(authServiceURL+"/api/login")).Methods("POST", "OPTIONS")
    api.HandleFunc("/auth/validate", proxy(authServiceURL+"/api/validate")).Methods("GET", "OPTIONS")
    
    // Product service routes (GET is public, POST/PUT/DELETE require auth)
    api.HandleFunc("/products", proxyWithContext(productServiceURL)).Methods("GET", "POST", "OPTIONS")
    api.HandleFunc("/products/{id:[0-9]+}", proxyWithContext(productServiceURL)).Methods("GET", "PUT", "DELETE", "OPTIONS")
    api.HandleFunc("/products/{id:[0-9]+}/stock", proxyWithContext(productServiceURL)).Methods("PATCH", "OPTIONS")
    api.HandleFunc("/categories", proxyWithContext(productServiceURL)).Methods("GET", "OPTIONS")

    // Order service routes (all require auth)
    api.HandleFunc("/orders", proxyWithContext(orderServiceURL)).Methods("GET", "POST", "OPTIONS")
    api.HandleFunc("/orders/{id:[0-9]+}", proxyWithContext(orderServiceURL)).Methods("GET", "OPTIONS")
    api.HandleFunc("/orders/{id:[0-9]+}/status", proxyWithContext(orderServiceURL)).Methods("PATCH", "OPTIONS")
    api.HandleFunc("/orders/{id:[0-9]+}/cancel", proxyWithContext(orderServiceURL)).Methods("POST", "OPTIONS")
    
    // CORS configuration
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type", "X-User-ID", "X-User-Role", "X-User-Email"},
        ExposedHeaders:   []string{"Content-Length", "Content-Type"},
        AllowCredentials: true,
        MaxAge:           300,
    })
    
    port := getEnv("PORT", "8080")
    handler := c.Handler(r)
    
    log.Printf("API Gateway listening on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}