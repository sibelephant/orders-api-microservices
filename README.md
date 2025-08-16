# Orders API Microservice

A high-performance RESTful API microservice for order management built with Go, Chi router, and Redis. This project implements clean architecture principles with proper separation of concerns, making it scalable and maintainable.

## 🏗️ Architecture

The application follows **Clean Architecture** patterns with clear layer separation:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Layer    │────│  Application    │────│   Repository    │
│   (handlers)    │    │   (business)    │    │   (data access) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Components

- **HTTP Handlers**: Request/response processing and validation
- **Application Layer**: Business logic and service orchestration
- **Repository Layer**: Data persistence with Redis
- **Models**: Domain entities and data structures

## 🚀 Features

- ✅ **CRUD Operations**: Create, Read, Update, Delete orders
- ✅ **Pagination**: Efficient cursor-based pagination for large datasets
- ✅ **Status Management**: Order lifecycle management (created → shipped → completed)
- ✅ **Clean Architecture**: Proper separation of concerns
- ✅ **Redis Integration**: High-performance data storage with atomic transactions
- ✅ **Error Handling**: Comprehensive error handling with proper HTTP status codes
- ✅ **Configuration**: Environment-based configuration
- ✅ **Graceful Shutdown**: Proper resource cleanup on application termination

## 🛠️ Tech Stack

- **Language**: Go 1.24+
- **HTTP Router**: Chi v5
- **Database**: Redis
- **UUID Generation**: Google UUID
- **Architecture**: Clean Architecture / Repository Pattern

## 📋 Prerequisites

Before running this application, make sure you have:

- **Go 1.24+** installed
- **Redis server** running (local or remote)
- **Git** for cloning the repository

## ⚡ Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/sibelephant/orders-api-microservices.git
cd orders-api-microservices
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Start Redis (if not already running)

```bash
# Using Docker
docker run -d -p 6379:6379 redis:latest

# Or using local installation
redis-server
```

### 4. Configure Environment (Optional)

```bash
export REDIS_ADDR="localhost:6379"
export SERVER_PORT="8080"
```

### 5. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080` 🎉

## 🔧 Configuration

The application supports configuration through environment variables:

| Variable      | Default          | Description          |
| ------------- | ---------------- | -------------------- |
| `REDIS_ADDR`  | `localhost:6379` | Redis server address |
| `SERVER_PORT` | `8080`           | HTTP server port     |

### Configuration Example

```bash
# Development
export REDIS_ADDR="localhost:6379"
export SERVER_PORT="8080"

# Production
export REDIS_ADDR="redis.production.com:6379"
export SERVER_PORT="80"
```

## 📚 API Documentation

### Base URL

```
http://localhost:8080
```

### Endpoints Overview

| Method   | Endpoint       | Description                 |
| -------- | -------------- | --------------------------- |
| `GET`    | `/`            | Health check                |
| `POST`   | `/orders`      | Create a new order          |
| `GET`    | `/orders`      | List all orders (paginated) |
| `GET`    | `/orders/{id}` | Get order by ID             |
| `PUT`    | `/orders/{id}` | Update order status         |
| `DELETE` | `/orders/{id}` | Delete order                |

---

### 🆕 Create Order

**POST** `/orders`

Create a new order with line items.

#### Request Body

```json
{
  "customer_id": "550e8400-e29b-41d4-a716-446655440000",
  "line_items": [
    {
      "item_id": "110e8400-e29b-41d4-a716-446655440001",
      "quantity": 2,
      "price": 1999
    }
  ]
}
```

#### Response

```json
{
  "order_id": 17426351878409116549,
  "customer_id": "550e8400-e29b-41d4-a716-446655440000",
  "lineitems": [
    {
      "item_id": "110e8400-e29b-41d4-a716-446655440001",
      "quantity": 2,
      "price": 1999
    }
  ],
  "created_at": "2025-08-16T10:30:00Z",
  "shipped_at": null,
  "completed_at": null
}
```

#### Status Codes

- `201 Created` - Order created successfully
- `400 Bad Request` - Invalid request body

---

### 📋 List Orders

**GET** `/orders?cursor={cursor}`

Retrieve a paginated list of orders.

#### Query Parameters

| Parameter | Type   | Default | Description       |
| --------- | ------ | ------- | ----------------- |
| `cursor`  | uint64 | 0       | Pagination cursor |

#### Response

```json
{
  "items": [
    {
      "order_id": 17426351878409116549,
      "customer_id": "550e8400-e29b-41d4-a716-446655440000",
      "lineitems": [...],
      "created_at": "2025-08-16T10:30:00Z",
      "shipped_at": null,
      "completed_at": null
    }
  ],
  "next": 12345
}
```

#### Status Codes

- `200 OK` - Orders retrieved successfully
- `400 Bad Request` - Invalid cursor parameter

---

### 🔍 Get Order by ID

**GET** `/orders/{id}`

Retrieve a specific order by its ID.

#### Path Parameters

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `id`      | uint64 | Order ID    |

#### Response

```json
{
  "order_id": 17426351878409116549,
  "customer_id": "550e8400-e29b-41d4-a716-446655440000",
  "lineitems": [...],
  "created_at": "2025-08-16T10:30:00Z",
  "shipped_at": "2025-08-16T11:00:00Z",
  "completed_at": null
}
```

#### Status Codes

- `200 OK` - Order found
- `400 Bad Request` - Invalid order ID format
- `404 Not Found` - Order not found

---

### ✏️ Update Order Status

**PUT** `/orders/{id}`

Update the status of an existing order (shipped/completed).

#### Path Parameters

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `id`      | uint64 | Order ID    |

#### Request Body

```json
{
  "status": "shipped"
}
```

#### Valid Status Values

- `"shipped"` - Mark order as shipped
- `"completed"` - Mark order as completed (must be shipped first)

#### Response

```json
{
  "order_id": 17426351878409116549,
  "customer_id": "550e8400-e29b-41d4-a716-446655440000",
  "lineitems": [...],
  "created_at": "2025-08-16T10:30:00Z",
  "shipped_at": "2025-08-16T11:00:00Z",
  "completed_at": null
}
```

#### Status Codes

- `200 OK` - Order updated successfully
- `400 Bad Request` - Invalid request or status transition
- `404 Not Found` - Order not found

---

### 🗑️ Delete Order

**DELETE** `/orders/{id}`

Delete an order by its ID.

#### Path Parameters

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `id`      | uint64 | Order ID    |

#### Response

No response body.

#### Status Codes

- `204 No Content` - Order deleted successfully
- `400 Bad Request` - Invalid order ID format
- `404 Not Found` - Order not found

---

## 🧪 Testing with cURL

### Create an Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "550e8400-e29b-41d4-a716-446655440000",
    "line_items": [
      {
        "item_id": "110e8400-e29b-41d4-a716-446655440001",
        "quantity": 2,
        "price": 1999
      }
    ]
  }'
```

### List Orders

```bash
curl http://localhost:8080/orders
```

### Get Order by ID

```bash
curl http://localhost:8080/orders/17426351878409116549
```

### Update Order Status

```bash
curl -X PUT http://localhost:8080/orders/17426351878409116549 \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'
```

### Delete Order

```bash
curl -X DELETE http://localhost:8080/orders/17426351878409116549
```

## 🏗️ Project Structure

```
orders-api-microservices/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── README.md              # This file
├── application/           # Application layer
│   ├── app.go            # Main application logic
│   ├── config.go         # Configuration management
│   └── routes.go         # HTTP route definitions
├── handler/              # HTTP handlers
│   └── order.go         # Order-related HTTP handlers
├── model/               # Domain models
│   └── order.go        # Order and LineItem models
└── repository/         # Data access layer
    └── order/
        └── redis.go   # Redis repository implementation
```

## 🔄 Data Flow

### Creating an Order

1. **HTTP Request** → Handler receives POST `/orders`
2. **Validation** → Handler validates request body
3. **Business Logic** → Generate order ID and timestamps
4. **Repository** → Save to Redis with atomic transaction
5. **Response** → Return created order data

### Listing Orders

1. **HTTP Request** → Handler receives GET `/orders`
2. **Pagination** → Parse cursor parameter
3. **Repository** → Use SSCAN + MGET for efficient retrieval
4. **Response** → Return paginated results with next cursor

## 🚀 Performance Features

### Redis Optimizations

- **Atomic Transactions**: Ensures data consistency
- **Cursor-based Pagination**: Memory-efficient pagination
- **Batch Operations**: MGET for multiple key retrieval
- **Set Operations**: Efficient order listing

### Application Optimizations

- **Connection Pooling**: Reuse Redis connections
- **Graceful Shutdown**: Proper resource cleanup
- **Context Propagation**: Request timeout handling
- **Pre-allocated Slices**: Minimize garbage collection

## 🐳 Docker Support

### Using Docker Compose

Create a `docker-compose.yml`:

```yaml
version: "3.8"
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  orders-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
```

### Create Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o orders-api main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/orders-api .
EXPOSE 8080
CMD ["./orders-api"]
```

Run with:

```bash
docker-compose up -d
```

## 🧪 Development

### Running Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

### Linting

```bash
golangci-lint run
```

### Building

```bash
go build -o orders-api main.go
```

## 📊 Monitoring

### Health Check

```bash
curl http://localhost:8080/
```

### Redis Monitoring

```bash
redis-cli monitor
redis-cli info
```

## 🚨 Error Handling

The API returns appropriate HTTP status codes and error messages:

| Status Code                 | Description        | Example                    |
| --------------------------- | ------------------ | -------------------------- |
| `200 OK`                    | Request successful | Order retrieved            |
| `201 Created`               | Resource created   | Order created              |
| `204 No Content`            | Resource deleted   | Order deleted              |
| `400 Bad Request`           | Invalid request    | Malformed JSON             |
| `404 Not Found`             | Resource not found | Order doesn't exist        |
| `500 Internal Server Error` | Server error       | Database connection failed |

## 🔒 Security Considerations

- **Input Validation**: All inputs are validated before processing
- **Error Sanitization**: Internal errors are not exposed to clients
- **Resource Limits**: Pagination prevents large data dumps
- **Context Timeouts**: Requests have timeout protection

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit your changes: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin feature/new-feature`
5. Submit a pull request

### Development Guidelines

- Follow Go conventions and best practices
- Write tests for new functionality
- Update documentation for API changes
- Use meaningful commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

- **sibelephant** - _Initial work_ - [GitHub](https://github.com/sibelephant)

## 🙏 Acknowledgments

- [Chi Router](https://github.com/go-chi/chi) for the excellent HTTP router
- [Redis](https://redis.io/) for high-performance data storage
- [Go](https://golang.org/) for the fantastic programming language
- Clean Architecture principles by Robert C. Martin

---

## 📞 Support

If you have any questions or issues, please:

1. Check the [Issues](https://github.com/sibelephant/orders-api-microservices/issues) page
2. Create a new issue if your problem isn't already reported
3. Provide as much detail as possible including:
   - Go version
   - Redis version
   - Error messages
   - Steps to reproduce
