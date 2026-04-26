# EngineX - High-Performance Trading Engine

A production-ready trading engine built with Go, implementing event-driven microservices architecture for high-frequency order matching.

## Overview

EngineX is a real-time order matching engine capable of processing 100k+ orders/second. It uses a B-tree based order book for efficient price-time priority matching, with all components communicating via Kafka for durability and scalability.

## Architecture

EngineX follows an event-driven microservices architecture where services communicate primarily through Apache Kafka for loose coupling, durability, and scalability.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              EngineX System                                         │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐          │
│  │    Users     │   │    Risk      │   │    Auth      │   │   Market     │          │
│  │  (Clients)   │   │   Service    │   │   Service    │   │    Data      │          │
│  └──────┬───────┘   └──────┬───────┘   └──────┬───────┘   └──────┬───────┘          │
│         │                  │                  │                  │                  │
│         └──────────────────┴────────┬─────-───┴──────────────────┘                  │
│                                     │                                               │
│                                     ▼                                               │
│                            ┌────────────────┐                                       │
│                            │    Gateway     │                                       │
│                            │  REST / gRPC   │                                       │
│                            └────────┬───────┘                                       │
│                                     │                                               │
│                                     │ HTTP/gRPC                                     │
│                                     ▼                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                         Apache Kafka                                        │    │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐              │    │
│  │  │ orders.submitted│  │trades.executed  │  │orderbook.updated│              │    │
│  │  │   (Topic)       │  │   (Topic)       │  │   (Topic)       │              │    │
│  │  └────────┬────--──┘  └──────┬──--───--─┘  └────┬────------──┘              │    │
│  │           │                  │                  │                           │    │
│  └───────────┼──────────────────┼──────────────────┼─────────────────────---───┘    │
│              │                  │                  │                                │
│              ▼                  ▼                  │                                │
│  ┌────────────────┐  ┌────────────────┐            │                                │
│  │    Engine      │  │   Executor     │            │                                │
│  │  (Matching)    │  │  (Settlement)  │            │                                │
│  │  Consumer      │  │   Consumer     │            │                                │
│  └───────┬──────-─┘  └───────┬──────-─┘            │                                │
│          │                  │                      │                                │
│          │                  │                      │                                │
│          ▼                  ▼                      │                                │
│  ┌────────────────┐  ┌──────────────┐              │                                │
│  │   Postgres     │  │  Postgres    │◀──────--─---─┘                                │
│  │  (Orders)      │  │  (Trades)    │                                               │
│  └────────────────┘  └──────────────┘                                               │                                                                           │
│          │                                                                          │
│          └───────────────────────┐                                                  │
│                                  │                                                  │
│                                  ▼                                                  │
│                       ┌────────────────┐                                            │
│                       │     WSHub      │                                            │
│                       │  (WebSocket)   │                                            │
│                       └───────┬────────┘                                            │
│                               │                                                     │
│                               ▼                                                     │
│                       ┌────────────────┐                                            │
│                       │     Redis      │                                            │
│                       │  (Cache/PubSub)│                                            │
│                       └────────────────┘                                            │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. Order Submission:
   Client → Gateway (REST) → Kafka (orders.submitted) → Engine

2. Order Matching:
   Engine reads from Kafka → Matches against Order Book → 
   Produces trades.executed → Executor
   Produces orderbook.updated → WSHub

3. Trade Settlement:
   Executor reads trades.executed → Postgres (trades table)

4. Real-time Updates:
   WSHub reads orderbook.updated → Redis PubSub → WebSocket Clients
```

### Component Responsibilities

| Component | Type | Description |
|------------|------|-------------|
| Gateway | Service | HTTP/gRPC entry point, validates and publishes orders to Kafka |
| Auth Service | gRPC | User authentication, JWT token issuance and validation |
| Risk Service | gRPC | Pre-trade risk checks (position limits, exposure) |
| Engine | Kafka Consumer | B-Tree order book, price-time priority matching |
| Executor | Kafka Consumer | Trade settlement, persists trades to database |
| WSHub | Service | Real-time market depth via WebSocket |

### Key Design Decisions

- **Kafka as Backbone**: All inter-service communication flows through Kafka for durability and ability to replay
- **B-Tree Order Book**: O(log n) lookup for efficient price-time priority matching
- **Event-Driven**: Services are decoupled and can scale independently
- **Redis for PubSub**: Low-latency real-time updates to WebSocket clients

## Services

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| gateway | 8080 | REST | HTTP entry point, order submission |
| authsvc | 9091 | gRPC | User authentication, JWT |
| risksvc | 9092 | gRPC | Risk checks, position limits |
| engine | - | Kafka | Order matching, trade execution |
| executor | - | Kafka | Trade settlement, DB persistence |
| wshub | 8081 | WebSocket | Real-time depth updates |

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 14+
- Kafka 3.5+
- Redis 7+

### 1. Start Infrastructure

```bash
docker compose up -d
```

### 2. Run Database Migrations

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -path internal/repository/migrations -database "postgres://postgres:postgres@localhost:5432/enginex?sslmode=disable" up
```

### 3. Start All Services

```bash
make run-all
```

Or individually:

```bash
make run-auth      # Auth service on :9091
make run-risk    # Risk service on :9092
make run-gateway # Gateway on :8080
make run-engine  # Engine (Kafka consumer)
make run-wshub  # WebSocket hub on :8081
make run-executor # Trade settlement
```

### 4. Submit Test Order

```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 100
  }'
```

## API Reference

### Gateway REST API

| Endpoint | Method | Description |
|----------|--------|------------|
| /api/orders | POST | Submit new order |
| /api/orders/:id | GET | Get order status |
| /api/orders/:id | DELETE | Cancel order |
| /api/positions | GET | Get user positions |
| /api/healthz | GET | Health check |

### Auth Service (gRPC)

```bash
# Register
grpcurl -plaintext -d '{"email":"user@test.com","password":"pass123","full_name":"Test User"}' \
  localhost:9091 auth.v1.AuthService/Register

# Login
grpcurl -plaintext -d '{"email":"user@test.com","password":"pass123"}' \
  localhost:9091 auth.v1.AuthService/Login

# Validate Token
grpcurl -plaintext -d '{"token":"<jwt_token>"}' \
  localhost:9091 auth.v1.AuthService/ValidateToken
```

### Risk Service (gRPC)

```bash
# Check Order Risk
grpcurl -plaintext -d '{
  "user_id": "user123",
  "symbol": "INFY",
  "side": "BUY",
  "quantity": 100,
  "price": 150000
}' localhost:9092 risk.v1.RiskService/CheckOrder
```

## Kafka Topics

| Topic | Partitions | Purpose |
|-------|-----------|---------|
| orders.submitted | 6 | Gateway → Engine |
| trades.executed | 6 | Engine → Executor |
| orderbook.updated | 6 | Engine → WSHub |
| orderbook.updates | 6 | Engine → WSHub |

## Order Book Algorithm

- **Data Structure**: B-Tree (degree 32) for O(log n) operations
- **Matching**: Price-Time Priority (FIFO at same price)
- **Self-Match Prevention**: Skip if incoming.UserID == resting.UserID

### Matching Rules

```
BUY Order:
  - Matches against lowest ASK first
  - Stop if ASK.price > BUY.price
  - Price-Time priority (earliest order wins)

SELL Order:
  - Matches against highest BID first
  - Stop if BID.price < SELL.price
  - Price-Time priority (earliest order wins)
```

## WebSocket API

Connect to receive real-time depth updates:

```javascript
const ws = new WebSocket('ws://localhost:8081/ws?symbol=INFY');

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(msg); // { type: "depth", symbol: "INFY", data: {...} }
};
```

## Configuration

Environment variables (see `.env.example`):

| Variable | Default | Description |
|----------|---------|-------------|
| POSTGRES_DSN | postgres://postgres:@localhost:5432/enginex | PostgreSQL connection |
| REDIS_ADDR | localhost:6379 | Redis address |
| KAFKA_BROKER | localhost:9092 | Kafka broker |
| JWT_SECRET | secret | JWT signing key |
| GATEWAY_PORT | :8080 | Gateway HTTP port |
| AUTH_PORT | :9091 | Auth gRPC port |
| RISK_PORT | :9092 | Risk gRPC port |
| WSHUB_PORT | :8081 | WebSocket port |

## Running Tests

```bash
# Unit tests
go test ./...

# With race detector
go test -race ./...

# Specific service
go test -v ./internal/engine/...
```

## Project Structure

```
EngineX/
├── api/gen/           # Generated protobuf code
├── cmd/              # Service entry points
│   ├── authsvc/      # Authentication service
│   ├── engine/      # Matching engine
│   ├── executor/    # Trade settlement
│   ├── gateway/     # HTTP gateway
│   ├── risksvc/     # Risk management
│   └── wshub/      # WebSocket hub
├── internal/
│   ├── auth/        # Auth service logic
│   ├── engine/      # Order matching
│   ├── kafka/      # Kafka producer/consumer
│   ├── repository/ # Database queries
│   └── websocket/  # WebSocket hub
├── docs/            # Service documentation
├── Makefile        # Build commands
├── docker-compose.yml
└── README.md
```

## Performance

- **Throughput**: 100k+ orders/second (single instance)
- **Latency**: <1ms order processing
- **Data Structure**: B-Tree O(log n) matching
- **Storage**: PostgreSQL for persistence, Redis for caching

## Monitoring (Coming Soon)

- Prometheus metrics endpoint
- Grafana dashboards
- K8s deployment with Helm charts

## Tech Stack

- **Language**: Go 1.23
- **Message Queue**: Kafka 3.5
- **Database**: PostgreSQL 14
- **Cache**: Redis 7
- **Web Framework**: Gin
- **WebSocket**: gorilla/websocket
- **Protocol**: gRPC + Protobuf

## Contributing

1. Fork the repository
2. Create feature branch
3. Write tests for new functionality
4. Ensure `go vet` and `go test` pass
5. Submit pull request

## License

MIT License