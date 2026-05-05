#!/bin/bash

API_URL="http://localhost:8080/api"

echo "Testing Auth Service..."

# Register
echo "Registering user..."
curl -X POST $API_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'

echo -e "\n\nLogin..."
# Login
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

echo "Token: $TOKEN"

echo -e "\n\nTesting Product Service..."
# Get products
curl -X GET $API_URL/products \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\nTesting Order Service..."
# Create order
curl -X POST $API_URL/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":1,"quantity":2}]}'

echo -e "\n\nGet orders..."
curl -X GET $API_URL/orders \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\nDone!"