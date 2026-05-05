.PHONY: help build up down logs clean test

help:
	@echo "Available commands:"
	@echo "  make build    - Build all Docker images"
	@echo "  make up       - Start all services"
	@echo "  make down     - Stop all services"
	@echo "  make logs     - View logs"
	@echo "  make clean    - Clean up containers and volumes"
	@echo "  make test     - Run tests"
	@echo "  make dev-auth     - Run auth service locally"
	@echo "  make dev-product  - Run product service locally"
	@echo "  make dev-order    - Run order service locally"
	@echo "  make dev-gateway  - Run API gateway locally"

build:
	docker compose build

up:
	docker compose up -d
	@echo "Services started. API Gateway available at http://localhost:8080"

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v
	docker system prune -f

test:
	cd services/auth-service && go test -v ./...
	cd services/product-service && go test -v ./...
	cd services/order-service && go test -v ./...

dev-auth:
	cd services/auth-service && go run main.go

dev-product:
	cd services/product-service && go run main.go

dev-order:
	cd services/order-service && go run main.go

dev-gateway:
	cd api-gateway && go run main.go

# Create databases for local development
init-db:
	docker exec -i sales-system_postgres_1 psql -U postgres -c "CREATE DATABASE IF NOT EXISTS auth_db"
	docker exec -i sales-system_postgres_1 psql -U postgres -c "CREATE DATABASE IF NOT EXISTS product_db"
	docker exec -i sales-system_postgres_1 psql -U postgres -c "CREATE DATABASE IF NOT EXISTS order_db"