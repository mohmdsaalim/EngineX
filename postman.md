# EngineX Postman API Test Collection

## Overview
This document contains all API endpoints for testing the EngineX trading engine using Postman. Test 10+ buy/sell orders to verify the matching algorithm works properly.

## Base URL
```
http://localhost:8080
```

---

## 1. Authentication Endpoints

### 1.1 Register User
**POST** `/api/v1/auth/register`

**Request Body:**
```json
{
    "email": "trader1@example.com",
    "password": "password123",
    "full_name": "Trader One"
}
```

**Expected Response (201):**
```json
{
    "user_id": "uuid-string",
    "email": "trader1@example.com"
}
```

---

### 1.2 Login
**POST** `/api/v1/auth/login`

**Request Body:**
```json
{
    "email": "trader1@example.com",
    "password": "password123"
}
```

**Expected Response (200):**
```json
{
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
}
```

**Note:** Save `access_token` for all subsequent API calls.

---

## 2. Order Endpoints (Protected - Requires JWT)

### 2.1 Submit BUY Order (LIMIT)
**POST** `/api/v1/orders`

**Headers:**
```
Authorization: Bearer <access_token>
Content-Type: application/json
```

**Request Body:**
```json
{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 100
}
```

**Expected Response (202):**
```json
{
    "order_id": "uuid-string",
    "status": "queued",
    "message": "order submitted successfully"
}
```

---

### 2.2 Submit SELL Order (LIMIT)
**POST** `/api/v1/orders`

**Request Body:**
```json
{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 151000,
    "quantity": 50
}
```

**Expected Response (202):**
```json
{
    "order_id": "uuid-string",
    "status": "queued",
    "message": "order submitted successfully"
}
```

---

### 2.3 Submit MARKET Order
**POST** `/api/v1/orders`

**Request Body:**
```json
{
    "symbol": "INFY",
    "side": "BUY",
    "type": "MARKET",
    "quantity": 100
}
```

**Note:** MARKET orders don't require price field.

---

### 2.4 Submit Order Without Auth (Should Fail)
**POST** `/api/v1/orders`

**Request Body:**
```json
{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 100
}
```

**Expected Response (401):**
```json
{
    "error": "unauthorized",
    "message": "missing or invalid authorization token"
}
```

---

### 2.5 Invalid Order Price (Should Fail)
**POST** `/api/v1/orders`

**Request Body:**
```json
{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 0,
    "quantity": 100
}
```

**Expected Response (400):**
```json
{
    "error": "invalid_input",
    "message": "price required for LIMIT order"
}
```

---

## 3. Orderbook Endpoints (Protected)

### 3.1 Get Orderbook Depth
**GET** `/api/v1/orderbook/{symbol}`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Example:** `GET /api/v1/orderbook/INFY`

**Expected Response (200):**
```json
{
    "symbol": "INFY",
    "bids": [
        {"price": 150000, "quantity": 100},
        {"price": 149500, "quantity": 200}
    ],
    "asks": [
        {"price": 151000, "quantity": 50},
        {"price": 151500, "quantity": 150}
    ]
}
```

**Note:** Returns top 5 price levels for both bids and asks.

---

## 4. Health Check

### 4.1 Health Check
**GET** `/health`

**Expected Response (200):**
```json
{
    "status": "ok"
}
```

**Note:** No authentication required.

---

## 5. Testing the Matching Algorithm - 10 Orders Test

### Step-by-Step Test Flow

#### Order 1: Buy Limit at 150000 (100 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 100
}
```
**Result:** Order goes to orderbook (no match, no asks at/below 150000)

---

#### Order 2: Sell Limit at 150000 (50 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 50
}
```
**Result:** MATCH! Buy order (150000) matches with Sell order (150000)
- Sell order: FILLED (50 qty)
- Buy order: PARTIAL (50 filled, 50 remaining)

---

#### Order 3: Sell Limit at 150000 (30 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 30
}
```
**Result:** MATCH! Remaining buy order (50) matches with sell (30)
- Sell order: FILLED (30 qty)
- Buy order: PARTIAL (80 filled, 20 remaining)

---

#### Order 4: Buy Limit at 151000 (100 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 151000,
    "quantity": 100
}
```
**Result:** No match (no asks at/above 150000). Order goes to orderbook.

---

#### Order 5: Sell Limit at 150500 (80 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 150500,
    "quantity": 80
}
```
**Result:** No match (150500 < 150000? No - wait, buy is at 151000)
- 150500 <= 151000? YES - MATCH!
- Buy order (151000) matches with sell (150500)
- Sell: FILLED (80 qty)
- Buy: FILLED (100 qty) - 20 extra filled from remaining buy

Wait, let's recalculate: Buy at 151000 wants 100, sell at 150500 offers 80.
- 150500 <= 151000 → MATCH
- Fill 80, buy remaining: 100-80=20
- Buy order: PARTIAL (80 filled, 20 remaining)

---

#### Order 6: Buy Limit at 150000 (200 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "BUY",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 200
}
```
**Result:** No match (no asks at/below 150000). Order goes to orderbook.

---

#### Order 7: Sell Limit at 149500 (100 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 149500,
    "quantity": 100
}
```
**Result:** 149500 <= 150000? YES - MATCH!
- Buy order at 150000 (remaining 20) matches first (FIFO)
- Fill 20, sell remaining: 100-20=80
- Sell: PARTIAL (20 filled, 80 remaining)
- Buy order: FILLED

---

#### Order 8: Sell Limit at 149500 (50 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 149500,
    "quantity": 50
}
```
**Result:** MATCH with remaining sell (80 at 149500)
- 149500 <= 150000 → YES
- Fill 50, sell remaining: 80-50=30
- Sell (80): FILLED (50 filled)
- New sell: PARTIAL (50 filled, 50 remaining)

---

#### Order 9: Buy MARKET (100 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "BUY",
    "type": "MARKET",
    "quantity": 100
}
```
**Result:** MARKET order matches at BEST price (lowest ask)
- Best ask: 149500 (remaining 30) + another sell at 149500 (50)
- Fill 30 + 50 = 80
- Still need 20 more, next best ask...
- Market order: PARTIAL (80 filled, 20 remaining)

---

#### Order 10: Sell Limit at 150000 (25 qty)
```http
POST {{baseUrl}}/api/v1/orders
Authorization: Bearer {{accessToken}}
Content-Type: application/json

{
    "symbol": "INFY",
    "side": "SELL",
    "type": "LIMIT",
    "price": 150000,
    "quantity": 25
}
```
**Result:** MATCH with remaining buy at 150000 (200 qty)
- 150000 <= 150000? YES
- Fill 25, buy remaining: 200-25=175
- Sell: FILLED (25 qty)
- Buy: PARTIAL (25 filled, 175 remaining)

---

## 6. Expected Order Book State After All Orders

```
INFY Order Book:
====================
BIDS (Buy Orders - Descending Price)
Price: 150000, Qty: 175 (remaining from order 6)

ASKS (Sell Orders - Ascending Price)
Price: 149500, Qty: 30 (remaining from order 7)
```

---

## 7. Kafka Topics for Verification

### View Orders Submitted to Engine
```bash
docker exec -it engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic orders.submitted \
  --from-beginning \
  --property print.key=true \
  --property print.value=true
```

### View Executed Trades
```bash
docker exec -it engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic trades.executed \
  --from-beginning \
  --property print.key=true \
  --property print.value=true
```

### View Orderbook Updates (Real-time)
```bash
docker exec -it engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic orderbook.updates \
  --from-beginning \
  --property print.key=true \
  --property print.value=true
```

---

## 8. PostgreSQL Verification Queries

### Check All Orders
```sql
SELECT id, user_id, symbol, side, type, price, quantity, filled_qty, status, created_at
FROM orders
ORDER BY created_at DESC;
```

### Check All Trades
```sql
SELECT id, buy_order_id, sell_order_id, buyer_id, seller_id, symbol, price, quantity, created_at
FROM trades
ORDER BY created_at DESC;
```

### Check Balances
```sql
SELECT * FROM balances;
```

### Check Positions
```sql
SELECT * FROM positions;
```

---

## 9. Error Codes Reference

| HTTP Code | Error | Message |
|-----------|-------|---------|
| 400 | invalid_input | Validation error (invalid side, missing price for LIMIT) |
| 401 | unauthorized | Missing or invalid JWT token |
| 403 | forbidden | Risk check failed |
| 500 | internal | Server error |
| 503 | service_unavailable | External service (Kafka/Risk/Auth) unavailable |

---

## 10. WebSocket Real-time Updates

### Connect
```
ws://localhost:8081/ws?symbol=INFY
```

### Expected Message Format
```json
{
    "type": "depth",
    "symbol": "INFY",
    "data": {
        "symbol": "INFY",
        "bids": [
            {"price": 150000, "quantity": 175}
        ],
        "asks": [
            {"price": 149500, "quantity": 30}
        ]
    }
}
```

---

## 11. Quick Test Commands (cURL)

### Health Check
```bash
curl -s http://localhost:8080/health
```

### Register User
```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"trader1@example.com","password":"password123","full_name":"Trader One"}'
```

### Login
```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"trader1@example.com","password":"password123"}'
```

### Submit Order (replace TOKEN)
```bash
curl -s -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{"symbol":"INFY","side":"BUY","type":"LIMIT","price":150000,"quantity":100}'
```

### Get Orderbook
```bash
curl -s http://localhost:8080/api/v1/orderbook/INFY \
  -H "Authorization: Bearer TOKEN"
```

---

## 12. Postman Collection JSON

Create a new Postman collection and add these requests:

### Collection: EngineX Trading Engine

**Folder 1: Auth**
- Register
- Login

**Folder 2: Orders**
- Submit Buy Order (LIMIT)
- Submit Sell Order (LIMIT)
- Submit Market Order
- Submit Order (No Auth - should fail)
- Submit Order (Invalid Price - should fail)

**Folder 3: Orderbook**
- Get Orderbook Depth

**Folder 4: Tests**
- 10 Order Matching Test

---

## 13. Matching Algorithm Verification Checklist

- [ ] Order 1 (Buy 150000): Goes to book (no matching ask)
- [ ] Order 2 (Sell 150000): Matches at 150000, sell FILLED, buy PARTIAL
- [ ] Order 3 (Sell 150000): Matches remaining buy, sell FILLED, buy PARTIAL
- [ ] Order 4 (Buy 151000): Goes to book (no matching ask)
- [ ] Order 5 (Sell 150500): Matches buy at 151000, both PARTIAL
- [ ] Order 6 (Buy 150000): Goes to book (no matching ask)
- [ ] Order 7 (Sell 149500): Matches buy at 150000 (FIFO), both PARTIAL
- [ ] Order 8 (Sell 149500): Matches remaining sell at 149500 (FIFO), both PARTIAL
- [ ] Order 9 (Buy MARKET): Matches at best price (149500), PARTIAL
- [ ] Order 10 (Sell 150000): Matches remaining buy at 150000, both FILLED/PARTIAL

**Key Algorithm Rules Verified:**
- Price Priority: Better prices match first
- Time Priority: FIFO at same price level
- Market Order: Matches only at best price
- Limit Order: Matches all price levels that satisfy price condition