#!/bin/bash

echo "Starting Sales System..."

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cat > .env << EOF
JWT_SECRET=$(openssl rand -base64 32)
POSTGRES_PASSWORD=$(openssl rand -base64 16)
EOF
fi

# Build and start services
docker compose up -d --build

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 15

# Check service status
echo "Service Status:"
docker compose ps

# Show access information
echo ""
echo "========================================="
echo "Sales System is running!"
echo "========================================="
echo "API Gateway: http://localhost:8080"
echo "Auth Service: http://localhost:8001"
echo "Product Service: http://localhost:8002"
echo "Order Service: http://localhost:8003"
echo ""
echo "Test credentials:"
echo "Admin: admin@example.com / admin123"
echo "Customer: test@example.com / password123 (after registration)"
echo "========================================="
echo ""
echo "View logs: docker compose logs -f"
echo "Stop services: docker compose down"