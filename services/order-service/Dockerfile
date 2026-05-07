FROM golang:1.26.2-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o order-service .

EXPOSE 8003

CMD ["./order-service"]