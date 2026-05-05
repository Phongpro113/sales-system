FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service .

# Final stage
FROM golang:1.21-alpine

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/auth-service .

EXPOSE 8001

CMD ["./auth-service"]