#!/bin/bash

# Build and start all services
docker-compose up -d --build

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 10

# Check service status
docker-compose ps

# Show logs
docker-compose logs -f