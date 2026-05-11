package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	userIDKey    contextKey = "user_id"
	userRoleKey  contextKey = "user_role"
	userEmailKey contextKey = "user_email"
)

var (
	authServiceURL    string
	productServiceURL string
	orderServiceURL   string
	paymentServiceURL string
	adminServiceURL   string
	fileServiceURL    string
	jwtSecret         []byte
)

func init() {
	authServiceURL = getEnv("AUTH_SERVICE_URL", "http://localhost:8001")
	productServiceURL = getEnv("PRODUCT_SERVICE_URL", "http://localhost:8002")
	orderServiceURL = getEnv("ORDER_SERVICE_URL", "http://localhost:8003")
	paymentServiceURL = getEnv("PAYMENT_SERVICE_URL", "http://localhost:8004")
	adminServiceURL = getEnv("ADMIN_SERVICE_URL", "http://localhost:8005")
	fileServiceURL = getEnv("FILE_SERVICE_URL", "http://localhost:8081")
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
	{"/api/payments/momo/ipn", []string{"POST"}},
	{"/uploads", []string{"GET"}},
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
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func proxy(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		remote, err := url.Parse(target)
		if err != nil {
			log.Printf("Failed to parse URL %s: %v", target, err)
			http.Error(w, "Bad gateway configuration", http.StatusBadGateway)
			return
		}

		// Determine the target path.
		// If target URL has a path (like /api/login), use it.
		// If not, preserve the original request path.
		targetPath := remote.Path
		if targetPath == "" || targetPath == "/" {
			targetPath = r.URL.Path
		}

		// Use a custom ReverseProxy to fully control the request transformation
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = remote.Scheme
				req.URL.Host = remote.Host
				req.URL.Path = targetPath
				req.Host = remote.Host

				// Preserve query strings
				if remote.RawQuery == "" || req.URL.RawQuery == "" {
					req.URL.RawQuery = remote.RawQuery + req.URL.RawQuery
				} else {
					req.URL.RawQuery = remote.RawQuery + "&" + req.URL.RawQuery
				}

				// Add headers for downstream services
				if userID := r.Context().Value(userIDKey); userID != nil {
					req.Header.Set("X-User-ID", toString(userID))
				}
				if role := r.Context().Value(userRoleKey); role != nil {
					req.Header.Set("X-User-Role", toString(role))
				}
			},
		}

		log.Printf("Proxying %s %s -> %s %s", r.Method, r.URL.Path, remote.Host, targetPath)
		proxy.ServeHTTP(w, r)
	}
}

func proxyWithContext(target string) http.HandlerFunc {
	// In this simplified version, proxy already handles context/path correctly
	return proxy(target)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "api-gateway",
		"timestamp": time.Now().Unix(),
	})
}

func main() {
	log.Println("API Gateway starting...")
	log.Printf("Auth Service URL: %s", authServiceURL)
	log.Printf("Product Service URL: %s", productServiceURL)
	log.Printf("Order Service URL: %s", orderServiceURL)
	log.Printf("Admin Service URL: %s", adminServiceURL)
	log.Printf("File Service URL: %s", fileServiceURL)

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
	api.HandleFunc("/products/{id:[0-9]+}/validate-buy", proxyWithContext(productServiceURL)).Methods("POST", "OPTIONS")
	api.HandleFunc("/categories", proxyWithContext(productServiceURL)).Methods("GET", "OPTIONS")

	// Admin service routes
	api.HandleFunc("/admin/products", proxyWithContext(adminServiceURL)).Methods("GET", "POST", "OPTIONS")
	api.HandleFunc("/admin/products/{id:[0-9]+}", proxyWithContext(adminServiceURL)).Methods("GET", "PUT", "OPTIONS")
	api.HandleFunc("/admin/upload", proxyWithContext(adminServiceURL)).Methods("POST", "OPTIONS")
	api.HandleFunc("/admin/health", proxyWithContext(adminServiceURL)).Methods("GET")

	// Static files from file service (uploads)
	r.HandleFunc("/uploads/{filename}", proxy(fileServiceURL+"/uploads")).Methods("GET")

	// Order service routes (all require auth)
	api.HandleFunc("/orders", proxyWithContext(orderServiceURL)).Methods("GET", "POST", "OPTIONS")
	api.HandleFunc("/orders/{id:[0-9]+}", proxyWithContext(orderServiceURL)).Methods("GET", "OPTIONS")
	api.HandleFunc("/orders/{id:[0-9]+}/status", proxyWithContext(orderServiceURL)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/orders/{id:[0-9]+}/cancel", proxyWithContext(orderServiceURL)).Methods("POST", "OPTIONS")

	// Payment service routes
	api.HandleFunc("/payments", proxyWithContext(paymentServiceURL)).Methods("POST", "OPTIONS")
	api.HandleFunc("/payments/{id:[0-9]+}", proxyWithContext(paymentServiceURL)).Methods("GET", "OPTIONS")
	api.HandleFunc("/payments/order/{orderId:[0-9]+}", proxyWithContext(paymentServiceURL)).Methods("GET", "OPTIONS")
	// MoMo IPN is called by MoMo server (no auth header) — public
	api.HandleFunc("/payments/momo/ipn", proxy(paymentServiceURL)).Methods("POST", "OPTIONS")

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
