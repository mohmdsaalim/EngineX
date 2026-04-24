# Executor Service Execution Flow

## Overview
The Executor is the settlement service that finalizes trades executed by the Engine. It consumes trade messages from Kafka, performs atomic database updates to settle balances and positions, and ensures exactly-once processing through idempotency.

## Architecture
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         EXECUTOR SERVICE                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐               │
│  │   Kafka      │─────▶│   Executor   │─────▶│  PostgreSQL  │               │
│  │ Consumer     │      │ (Settlement) │      │   (Atomic)   │               │
│  │ trades.exec  │      │              │      │              │               │
│  └──────────────┘      └──────────────┘      └──────────────┘               │
│                              │                        │                     │
│                              ▼                        │                     │
│                      ┌──────────────┐                 │                     │
│                      │    Redis     │◀────────────────┘                     │
│                      │ (Idempotency)│                                       │
│                      └──────────────┘                                       │
│                                                                             │
└───────────────────────────────────────────────────────────────--------──────┘
```

## Core Responsibilities

| Responsibility | Description |
|-----------------|-------------|
| Trade Settlement | Update buyer/seller balances atomically |
| Order Status | Mark orders as FILLED in database |
| Idempotency | Prevent double-settlement on restart |
| Exactly-Once | Ensure each trade settles exactly once |

## Execution Flow

### 1. Receive Trade from Kafka
```
Consumer.ReadMessage() → Get raw bytes from trades.executed topic
```

### 2. Deserialize Trade Message
```
json.Unmarshal(raw) → TradeMessage struct
    - ID, BuyOrderID, SellOrderID
    - BuyerID, SellerID
    - Symbol, Price, Quantity
```

### 3. Validate Trade
```
validateTrade(trade)
    ├── Check trade.ID not empty
    ├── Check BuyOrderID, SellOrderID not empty
    ├── Check BuyerID, SellerID not empty
    ├── Check Symbol not empty
    └── Check Price > 0 and Quantity > 0
```

### 4. Idempotency Check (Redis)
```
IsTradeProcessed(trade.ID)
    ├── Key: "trade:" + trade.ID
    ├── If exists: SKIP (already processed)
    └── If not exists: CONTINUE
```

### 5. Atomic Settlement (Postgres Transaction)
```
BEGIN TRANSACTION
    │
    ├──► 1. INSERT trade record
    │         CreateTrade(buy_order, sell_order, buyer, seller, symbol, price, qty)
    │
    ├──► 2. DEBIT buyer balance
    │         DebitBalance(buyerID, tradeValue)
    │         tradeValue = Price * Quantity
    │
    ├──► 3. CREDIT seller balance
    │         CreditBalance(sellerID, tradeValue, lockedRelease)
    │
    ├──► 4. UPDATE buy order status
    │         UpdateOrderStatus(buyOrderID, "FILLED", quantity)
    │
    ├──► 5. UPDATE sell order status
    │         UpdateOrderStatus(sellOrderID, "FILLED", quantity)
    │
    └──► COMMIT (or ROLLBACK on any error)
```

### 6. Mark as Processed (Redis)
```
MarkTradeProcessed(trade.ID)
    ├── Key: "trade:" + trade.ID
    ├── Value: "1"
    └── TTL: 24 hours
```

### 7. Commit Kafka Offset
```
Consumer.CommitMessage(msg)
    - Prevents re-processing on restart
```

## Transaction Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ BEGIN TX    │────▶│ All ops OK  │────▶│ COMMIT TX  │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │
       │                   │ (any error)
       ▼                   ▼
┌─────────────┐     ┌─────────────┐
│ ROLLBACK TX │     │ ROLLBACK TX  │
└─────────────┘     └─────────────┘
```

## Error Handling

| Error | Action |
|-------|--------|
| Unmarshal failed | Log error, commit offset (poison message) |
| Validation failed | Return error, commit offset |
| Idempotency check failed | Return error (retry) |
| DB operation failed | Rollback transaction, return error (retry) |
| Redis mark failed | Log error only, continue (trade settled in DB) |
| Commit offset failed | Log error, continue |

## Idempotency Protection

Two layers ensure exactly-once:

### Layer 1: Redis (Fast Path)
```
Before processing: Check Redis for "trade:{id}"
- If exists → Skip (already processed)
- If not exists → Continue and process
```

### Layer 2: PostgreSQL (Safe Path)
```
Transaction wraps all operations:
- If any DB operation fails → ROLLBACK
- Only COMMIT if ALL succeed
- Trade record exists in DB (permanent proof)
```

### Why Both?
- Redis is fast (in-memory)
- DB is permanent (survives crashes)
- Even if Redis key lost, DB has record
- On restart, idempotency check goes to DB level

## What Happens If Executor Fails

### Scenario 1: Unmarshal Fails
```
Cause: Corrupt message
Action: Log error, commit offset
Result: Message skipped, no retry (poison message)
```

### Scenario 2: Validation Fails
```
Cause: Missing required fields
Action: Log error, commit offset
Result: Message skipped (invalid trade)
```

### Scenario 3: Idempotency Check Fails (Redis Down)
```
Cause: Redis unavailable
Action: Return error, don't commit
Result: Message re-processed on restart
```

### Scenario 4: DB Transaction Fails
```
Cause: Insufficient balance, DB error, constraint violation
Action: Rollback, return error, don't commit
Result: Message re-processed on restart
```

### Scenario 5: Commit Offset Fails
```
Cause: Kafka broker issue
Action: Log error, continue
Result: Message may re-process (idempotency prevents double-settlement)
```

## Kafka Topics

| Topic | Direction | Purpose |
|-------|-----------|---------|
| trades.executed | In | Engine → Executor |
| (no output) | - | Settlement is internal (Postgres) |

## Database Operations

### Tables Modified
1. **trades** - Record of executed trade
2. **balances** - Buyer debited, Seller credited
3. **orders** - Both orders marked FILLED

### Balance Flow
```
Buyer:
  Before: ₹100,000
  Trade Value: ₹15,000
  After: ₹85,000 (debited)

Seller:
  Before: ₹50,000
  Trade Value: ₹15,000
  After: ₹65,000 (credited)
```

## Core Components

| Component | File | Role |
|----------|------|------|
| Executor | cmd/executor/main.go | Entry point, consumer loop |
| Executor | internal/settlement/executor.go | Core settlement logic |
| TradeMessage | internal/settlement/executor.go | Trade data structure |

## Flow Summary

```
Kafka Message → Deserialize → Validate → Idempotency Check
                                           │
                            ┌────────────────┴────────────────┐
                            ▼                                 ▼
                       Already processed                    Continue
                            │                                 │
                            ▼                                 ▼
                       Skip (return)                  Atomic DB Transaction
                                                     │
                                    ┌─────────────────┼─────────────────┐
                                    ▼                 ▼                 ▼
                               INSERT trade         Update balances    Update orders
                                    │                 │                 │
                                    └─────────────────┼─────────────────┘
                                                      ▼
                                                   COMMIT
                                                      │
                                    ┌─────────────────┼─────────────────┐
                                    ▼                 ▼                 ▼
                               Mark Redis        Log success       Commit Kafka
```