FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product-service .

FROM golang:1.26.2-alpine

WORKDIR /root/

COPY --from=builder /app/product-service .

EXPOSE 8002

CMD ["./product-service"]