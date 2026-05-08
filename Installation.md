# Installation Guide

This document provides step-by-step instructions for setting up and running the Sales System project.

## Prerequisites

Before you begin, ensure you have the following installed on your system:

- **Docker & Docker Compose**: For containerization and service orchestration.
- **Go (1.21+)**: Required for local backend development.
- **Node.js (v18+) & npm**: Required for local frontend development.
- **Make**: (Optional) For using the provided shortcuts.

## Project Structure

- `services/`: Contains backend microservices (Auth, Product, Order, Admin).
- `api-gateway/`: Entry point for all client requests.
- `frontend/`: React-based web interface.
- `docker-compose.yml`: Orchestration for all services and databases.

---

## 1. Quick Start (Recommended)

The easiest way to get the system running is using Docker Compose.

### Step 1: Clone the repository
```bash
git clone <repository-url>
cd sales-system
```

### Step 2: Environment Setup
Copy the example environment file and adjust variables if necessary:
```bash
cp .env.example .env  # Or create a .env file based on the template below
```

### Step 3: Build and Start Services
Using the `Makefile`:
```bash
make build
make up
```
Or using `docker-compose` directly:
```bash
docker compose build
docker compose up -d
```

### Step 4: Initialize Databases
Once the Postgres container is healthy, create the necessary databases:
```bash
make init-db
```

---

## 2. Local Development Setup

If you want to run services individually without Docker for faster development cycles:

### Backend Services
Navigate to each service directory and run:
```bash
go mod tidy
go run main.go
```
Alternatively, use the Makefile commands:
- `make dev-auth`
- `make dev-product`
- `make dev-order`
- `make dev-gateway`

### Frontend Application
```bash
cd frontend
npm install
npm start
```
The frontend will be available at `http://localhost:3000`. It is configured to proxy requests to the API Gateway at `http://localhost:8080`.

---

## 3. Environment Variables (`.env`)

The system uses the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_USER` | Database username | `postgres` |
| `POSTGRES_PASSWORD` | Database password | `postgres` |
| `JWT_SECRET` | Secret key for JWT signing | `your-secret-key` |
| `GATEWAY_PORT` | Port for the API Gateway | `8080` |

---

## 4. Verification

After starting the services, you can verify they are running:

- **API Gateway**: [http://localhost:8080](http://localhost:8080)
- **Frontend**: [http://localhost:3000](http://localhost:3000)
- **Database**: `localhost:5432`

Run the included API test script to ensure services are communicating correctly:
```bash
./test-api.sh
```

---

## 5. Troubleshooting

- **Container Conflicts**: If ports are already in use, modify the `PORTS` in `.env` or `docker-compose.yml`.
- **Database Connection**: Ensure the `postgres` service is healthy before starting other services.
- **Frontend Node Version**: If you encounter build errors, ensure you are using Node 18 or higher.
- **Clean Start**: To reset everything:
  ```bash
  make clean
  ```
