# EngineX - Running the Project for Review

This guide provides step-by-step instructions to run the EngineX trading engine and demonstrate that all core components are working, including order matching.

---

## Prerequisites

Ensure you have:
- **Go 1.23+** installed (`go version`)
- **Docker & Docker Compose** installed
- **PostgreSQL** not running on port 5432 (or stop it first)
- **Redis** not running on port 6379

---

## Step 1: Start Infrastructure (Docker)

```bash
docker-compose up -d
```

**Wait for all services to be healthy (1-2 minutes):**

```bash
docker ps
```

Expected output:
| CONTAINER | IMAGE | STATUS |
|-----------|-------|--------|
| engine_kafka | apache/kafka:3.7.0 | Up |
| engine_postgres | postgres:16-alpine | Up |
| engine_redis | redis:7-alpine | Up |

**Verify Kafka is ready:**

```bash
docker exec engine_kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```

Expected output: (empty initially is fine - topics auto-create)

---

## Step 2: Run Database Migrations

```bash
make migrate
```

Expected output: Migrations applied (users, orders, trades, balances, positions tables)

**Verify tables created:**

```bash
make db-users
```

---

## Step 3: Seed Test Data (Optional but Recommended)

```bash
make seed
```

This creates:
- Test users with accounts
- Initial balances for trading

**Verify data:**

```bash
make db-balances
```

---

## Step 4: Start All Services

You need to run each service in a separate terminal (or use tmux/screen). Start them in this order:

### Terminal 1: Auth Service (gRPC)

```bash
make run-auth
```
- Port: **:9091**
- Status: Should see "Auth service listening on :9091"

### Terminal 2: Risk Service (gRPC)

```bash
make run-risk
```
- Port: **:9093** (note: 9092 is used by Kafka, so risk runs on 9093)
- Status: Should see "Risk service listening on :9093"

### Terminal 3: Gateway (REST API)

```bash
make run-gateway
```
- Port: **:8080**
- Status: Should see "Gateway listening on :8080"

### Terminal 4: Engine (Matching Engine)

```bash
make run-engine
```
- Status: Should see "Engine started, waiting for orders..."
- This is the core order matching engine

### Terminal 5: Executor (Trade Settlement)

```bash
make run-executor
```
- Status: Should see "Executor started, listening for trades..."

### Terminal 6: WebSocket Hub

```bash
make run-wshub
```
- Port: **:8081**
- Status: Should see "WebSocket hub listening on :8081"

---

## Step 5: Verify All Services Are Running

### Health Checks:

```bash
# Gateway health
curl http://localhost:8080/health
```

Expected: `{"status":"ok"}`

```bash
# Redis check
docker exec engine_redis redis-cli ping
```

Expected: `PONG`

```bash
# Kafka check
docker exec engine_kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```

Expected: Shows topics like `orders.submitted`, `trades.executed`, `orderbook.updates`

---

## Step 6: Prove Order Matching is Working

### 6.1: Register a Test User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "trader@test.com",
    "password": "password123",
    "full_name": "Test Trader"
  }'
```

Expected: `{"token":"eyJ..."}`

### 6.2: Login to Get Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "trader@test.com",
    "password": "password123"
  }'
```

Save the token returned.

### 6.3: Submit a BUY Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 100
  }'
```

Expected: Response with order ID, status "OPEN"

### 6.4: Submit a Matching SELL Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -d '{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 100
  }'
```

**Expected: Trade executed!**

- Order is filled at price 150000
- A trade is generated

---

## Step 7: Verify Trade Was Executed

### Check Trades in Database:

```bash
make db-trades
```

Expected output should show:
- A trade with symbol "INFY"
- Price: 150000
- Quantity: 100

### Watch Kafka Trade Topic (Real-time):

```bash
docker exec engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic trades.executed \
  --from-beginning \
  --property print.key=true \
  --property print.timestamp=true
```

You should see trade JSON messages being produced when orders match.

### Watch Orders Submitted Topic:

```bash
docker exec engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic orders.submitted \
  --from-beginning \
  --property print.key=true
```

---

## Step 8: Get Order Book Depth

```bash
curl -X GET http://localhost:8080/api/v1/orderbook/INFY \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

Expected: JSON with bids and asks (should be empty after trade - fully matched)

---

## Step 9: Test Unit Tests

To prove the core engine logic works:

```bash
make test
```

This runs unit tests for the engine:
- Order book tests
- Matching algorithm tests
- Order status tests

Expected: All tests PASS

---

## Quick Verification Checklist

Use this when showing your reviewer:

| # | Check | Command | Expected Result |
|---|-------|---------|------------------|
| 1 | Docker running | `docker ps` | 3 containers (kafka, postgres, redis) |
| 2 | Gateway up | `curl http://localhost:8080/health` | `{"status":"ok"}` |
| 3 | Engine running | Check terminal with engine | "Engine started" log |
| 4 | Database has trades | `make db-trades` | At least 1 trade |
| 5 | Unit tests pass | `make test` | All tests PASS |

---

## How the Order Matching Works (Core Process)

### Understanding the Flow:

1. **Order Submission**:
   - Client sends order to Gateway (REST API)
   - Gateway publishes to Kafka topic `orders.submitted`

2. **Order Matching (Engine)**:
   - Engine reads from Kafka (`orders.submitted` topic)
   - B-Tree order book matches BUY orders against SELL orders
   - Price-Time Priority: matches best price first, then FIFO at same price

3. **Trade Execution**:
   - When prices cross (BUY price >= SELL price), a trade is executed
   - Engine produces to `trades.executed` topic
   - Engine produces to `orderbook.updates` topic

4. **Trade Settlement**:
   - Executor reads from `trades.executed`
   - Persists trade to PostgreSQL

5. **Real-time Updates**:
   - WSHub reads `orderbook.updates`
   - Pushes to WebSocket clients

---

## Troubleshooting

### Services not starting?
- Check ports: `lsof -i :8080 -i :9091 -i :9093`
- Stop conflicting services

### Kafka not ready?
- Wait 60 seconds after `docker-compose up`
- Check logs: `docker logs engine_kafka`

### Orders not matching?
- Ensure engine is running and reading from Kafka
- Check engine terminal for matching logs

### Database connection errors?
- Wait for postgres to be healthy: `docker ps`
- Retry migration: `make migrate-down && make migrate`

---

## Demo Script (for Reviewer)

1. Start everything as above
2. Show `docker ps` - all 3 containers running
3. Show gateway health: `curl http://localhost:8080/health`
4. Submit BUY order: Show order ID returned
5. Submit SELL order at same price: Show order filled
6. Show trades in DB: `make db-trades`
7. Show unit tests: `make test`
8. Open orderbook: Show depth

This proves:
- Gateway accepts orders
- Kafka message flow works
- Engine matches orders
- Trades are persisted
- All core components functional


docker exec -it engine_postgres psql -U engine_user -d engine_db \
  -c "SELECT symbol, price, quantity FROM trades;"