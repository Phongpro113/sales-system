# Admin Service

The **Admin Service** is a microservice built using the [Gin Web Framework](https://gin-gonic.com/) in Go. It is designed to handle administrative tasks, specifically product management, within the sales system.

## Project Structure

The project follows a clean architecture pattern to separate concerns and improve maintainability:

```text
admin-service/
├── config/             # Configuration management (env vars, constants)
├── controller/         # Request handlers (logic for processing requests)
├── database/           # Database connection and migration logic
├── middlewares/        # Custom Gin middlewares (e.g., Auth, Logging)
├── models/             # Data models and structures (GORM entities)
├── repositories/       # Data access layer (DB queries and operations)
├── route/              # API route definitions and router setup
└── main.go             # Entry point of the application
```

### Components Detail

- **`config/config.go`**: Loads environment variables like `PORT` and `DATABASE_URL`.
- **`models/product.go`**: Defines the `Product` struct with GORM tags for database mapping.
- **`database/db.go`**: Initializes the GORM connection and performs auto-migrations.
- **`repositories/product_repository.go`**: Contains the logic for CRUD operations on the `products` table. This layer decouples the database logic from the controllers.
- **`middlewares/auth_middleware.go`**: Intercepts requests to check for the `X-User-Role` header. If the role is not `admin`, the request is aborted with a `403 Forbidden` status.
- **`controller/product_controller.go`**: Receives requests from the router, uses the repository to fetch/modify data, and returns the appropriate JSON response.
- **`route/route.go`**: Configures the Gin engine, applies global middlewares, and defines the API endpoints.

## API Endpoints

All endpoints are prefixed with `/api/admin` and require an **admin role**.

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/admin/products` | List all products with pagination |
| `POST` | `/api/admin/products` | Create a new product |
| `PUT` | `/api/admin/products/:id` | Update an existing product |
| `GET` | `/api/admin/health` | Service health check |

## How to Run

### Locally (for development)
Ensure you have Go installed and a PostgreSQL instance running.
```bash
# Install dependencies
go mod tidy

# Run the service
go run main.go
```

### Via Docker Compose
The service is automatically managed by the root `docker-compose.yml`.
```bash
docker compose up -d --build admin-service
```

## Example Requests

### 1. List Products
```bash
curl -H "X-User-Role: admin" "http://localhost:8080/api/admin/products?page=1&limit=10"
```

### 2. Create Product
```bash
curl -X POST -H "Content-Type: application/json" -H "X-User-Role: admin" \
-d '{"name": "New Laptop", "price": 1200.50, "stock": 50, "sku": "LAP-001"}' \
http://localhost:8080/api/admin/products
```

### 3. Update Product
```bash
curl -X PUT -H "Content-Type: application/json" -H "X-User-Role: admin" \
-d '{"price": 1150.00}' \
http://localhost:8080/api/admin/products/1
```
