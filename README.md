# Sales System

Hệ thống bán hàng microservices gồm frontend React và các backend service viết bằng Go, Java Spring Boot, giao tiếp qua REST API và Kafka.

## Kiến trúc

```
Frontend (React :3000)
        │
        ▼
API Gateway (:8080)
        │
   ┌────┴─────────────────────────┐
   │                              │
Auth     Product    Order    Payment    Admin
(:8001)  (:8002)   (:8003)  (:8004)   (:8005)
                      │
                   Kafka
                      │
                   Product (consume order.created → giảm stock)
```

## Services

| Service | Ngôn ngữ | Port | Database |
|---|---|---|---|
| api-gateway | Go | 8080 | — |
| auth-service | Go | 8001 | auth_db |
| product-service | Go | 8002 | product_db |
| order-service | Go | 8003 | order_db |
| payment-service | Java Spring Boot | 8004 | payment_db |
| admin-service | Go (Gin) | 8005 | product_db |
| frontend | React | 3000 | — |
| kafka | Confluent Kafka | 9092 | — |
| postgres | PostgreSQL 15 | 5432 | — |

## Kafka Events

| Topic | Publisher | Consumer | Mô tả |
|---|---|---|---|
| `order.created` | order-service | product-service | Giảm stock sau khi tạo đơn hàng |

## Payment Methods

- **COD** — Thanh toán khi nhận hàng
- **Bank Transfer** — Chuyển khoản ngân hàng
- **MoMo** — Thanh toán qua MoMo API (QR code chính thức)

## Yêu cầu

- Docker & Docker Compose
- Node.js 18+ (để chạy frontend local)

## Chạy project

```bash
# Copy env
cp .env.example .env   # Điền MOMO_PARTNER_CODE, MOMO_ACCESS_KEY, MOMO_SECRET_KEY

# Build và chạy tất cả services
docker compose up -d --build

# Chạy frontend (development)
cd frontend && npm install && npm start
```

## Lần đầu chạy

PostgreSQL tự tạo database từ `scripts/init-db.sql` khi volume chưa tồn tại.  
Nếu database đã tồn tại mà thiếu `payment_db`:

```bash
docker exec sales-system-postgres-1 psql -U postgres -c "CREATE DATABASE payment_db;"
docker compose restart payment-service
```

## Xem logs

```bash
# Tất cả services
docker compose logs -f

# Từng service
docker compose logs -f order-service
docker compose logs -f product-service
docker compose logs -f payment-service
```

## Cấu trúc thư mục

```
sales-system/
├── api-gateway/          # Go — reverse proxy + JWT auth
├── frontend/             # React — UI
├── services/
│   ├── auth-service/     # Go — đăng ký, đăng nhập, JWT
│   ├── product-service/  # Go — sản phẩm, stock, Kafka consumer
│   ├── order-service/    # Go — đơn hàng, Kafka producer
│   ├── payment-service/  # Java Spring Boot — thanh toán MoMo
│   └── admin-service/    # Go (Gin) — quản lý sản phẩm
├── scripts/
│   └── init-db.sql       # Tạo databases lần đầu
├── uploads/              # File upload (ảnh sản phẩm)
└── docker-compose.yml
```

## Biến môi trường

Xem `.env.example` để biết danh sách đầy đủ. Các biến quan trọng:

| Biến | Mô tả |
|---|---|
| `JWT_SECRET` | Secret key cho JWT |
| `MOMO_PARTNER_CODE` | MoMo partner code |
| `MOMO_ACCESS_KEY` | MoMo access key |
| `MOMO_SECRET_KEY` | MoMo secret key |
| `MOMO_ENDPOINT` | MoMo API endpoint (sandbox/production) |
