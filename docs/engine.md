# Engine Service Execution Flow

## Overview
The Engine is the core matching engine that executes trades between BUY and SELL orders. It consumes orders from Kafka, matches them using price-time priority, and publishes executed trades back to Kafka.

## Architecture
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            ENGINE SERVICE                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐               │
│  │   Kafka      │─────▶│   Engine     │──-──▶│   Kafka      │               │
│  │ Consumer     │      │ (Matching)   │      │ Producer     │               │
│  │ orders.sub   │      │              │      │ trades.exec  │               │
│  └──────────────┘      └──────────────┘      └──────────────┘               │
│                              │                                              │
│                              ▼                                              │
│                     ┌──────────────┐                                        │
│                     │ OrderBook    │ ◀──── One per symbol                    │
│                     │ (B-Tree)     │                                        │
│                     └──────────────┘                                        │
│                            │                                                │
│                   ┌──────-─┴───────┐                                        │
│                   ▼                ▼                                        │
│              ┌─────────┐     ┌─────────┐                                    │
│              │  Bids   │     │  Asks   │                                    │
│              │ (Buy)   │     │ (Sell)  │                                    │
│              └─────────┘     └─────────┘                                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Core Flow

### 1. Order Receipt (Consume from Kafka)
```
Consumer.ReadMessage() → proto.Unmarshal() → ProcessOrder() → CommitMessage()
```

### 2. Order Processing
```
ProcessOrder(msg)
    │
    ├──► Validate order (orderID, userID, symbol, quantity)
    │
    ├──► getOrCreateBook(symbol) → get or create OrderBook for symbol
    │
    ├──► Convert proto → Internal Order
    │       - ID, UserID, Symbol, Side, Type, Price, Quantity
    │
    ├──► book.Match(order) → Execute matching algorithm
    │       - Returns []Trade (executed trades)
    │
    ├──► publishTradeWithRetry() for each trade
    │       - Publish to Kafka topic: trades.executed
    │       - Retry 3 times on failure
    │
    └──► publishSnapshot()
            - Publish to Kafka topic: orderbook.updated
```

## BUY Order Matching Flow

### matchBuy() - BUY order matches against ASKS (sellers)
```
1. Get best ask (lowest price) from Asks tree
2. For LIMIT order:
   - If ask.Price > buy.Price → STOP (no match)
   - If ask.Price <= buy.Price → Continue to match
3. For MARKET order:
   - Match only at best ask (first level), then STOP
4. Call fillFromLevel() to execute trade
5. Repeat until order filled or no more asks
```

### Example: BUY @ 150000
```
Asks: [145000, 150000, 155000]  ← best ask = 145000
                 ↓
            145000 <= 150000 ✓ MATCH
            Create trade @ 145000
```

## SELL Order Matching Flow

### matchSell() - SELL order matches against BIDS (buyers)
```
1. Get best bid (highest price) from Bids tree
2. For LIMIT order:
   - If bid.Price < sell.Price → STOP (no match)
   - If bid.Price >= sell.Price → Continue to match
3. For MARKET order:
   - Match only at best bid (first level), then STOP
4. Call fillFromLevel() to execute trade
5. Repeat until order filled or no more bids
```

### Example: SELL @ 150000
```
Bids: [155000, 150000, 145000]  ← best bid = 155000
                 ↓
            155000 >= 150000 ✓ MATCH
            Create trade @ 155000
```

## fillFromLevel() - Execute Trade
```
fillFromLevel(incoming, level, tree)
    │
    ├──► Get first resting order from price level (FIFO)
    │
    ├──► Self-match check:
    │     - If incoming.UserID == resting.UserID → STOP
    │
    ├──► Calculate fillQty = min(incoming.Remaining, resting.Remaining)
    │
    ├──► buildTrade(incoming, resting, level.Price, fillQty)
    │       - Generate unique tradeID
    │       - Set BuyOrderID, SellOrderID
    │       - Set BuyerID, SellerID
    │       - Set Price, Quantity, ExecutedAt
    │
    ├──► Update filled quantities:
    │     - incoming.Filled += fillQty
    │     - resting.Filled += fillQty
    │
    ├──► Update resting order status:
    │     - If fully filled → remove from level, delete from Orders map
    │     - If partially filled → update status to Partial
    │
    └──► Update incoming order status:
          - If fully filled → StatusFilled
```

## B-Tree Data Structure

### Why B-Tree?
- O(log n) insert/delete
- Sorted by price naturally
- Efficient range queries

### Bids Tree (Buy Orders)
```
Ordering: a.Price > b.Price  (descending)
Highest price at top = best bid

Price: 155000  ← best bid (highest)
       150000
       145000
```

### Asks Tree (Sell Orders)
```
Ordering: a.Price < b.Price  (ascending)
Lowest price at top = best ask

Price: 145000  ← best ask (lowest)
       150000
       155000
```

### PriceLevel - All orders at same price
```
PriceLevel { Price: 150000, Orders: [Order1, Order2, ...] }
                      ↑
                 FIFO queue (time priority)
```

## Order Status Flow
```
┌──────────┐     ┌───────────┐     ┌──────────┐     ┌────────────┐
│ New      │────▶│  Open     │────▶│ Partial  │────▶│ Filled     │
│ Order    │     │ (on book) │     │(some qty)│     │ (all qty)  │
└──────────┘     └───────────┘     └──────────┘     └────────────┘
                     │
                     ▼
              ┌────────────┐
              │ Cancelled  │
              │            │
              └────────────┘
```

## Kafka Topics

| Topic | Direction | Purpose |
|-------|-----------|---------|
| orders.submitted | In | Gateway → Engine |
| trades.executed | Out | Engine → Gateway |
| orderbook.updated | Out | Engine → Clients |

## Error Handling

| Error | Handling |
|-------|-----------|
| Invalid order (nil/missing fields) | Return error, commit anyway to prevent requeue |
| Unmarhsal failed | Log error, commit to prevent poison message |
| Process failed | Log error, still commit |
| Publish trade failed | Retry 3 times, log error |
| Publish snapshot failed | Log error, continue |

## Self-Match Prevention

Two levels of protection:
1. **Entry check** - Before matching, check if user has opposing order at same price
2. **Fill check** - During fillFromLevel(), skip if resting.UserID == incoming.UserID

## Core Components

| Component | File | Role |
|------------|------|------|
| Engine | engine.go | Main orchestrator, Kafka consume, publish |
| OrderBook | orderbook.go | Holds bids/asks per symbol, B-tree storage |
| Matching | matching.go | Price-time priority matching algorithm |
| PriceLevel | pricelevel.go | All orders at same price (FIFO queue) |
| Types | types.go | Order, Trade, Side, OrderType definitions |

## Flow Summary

```
Order Receive → Validate → Get/Create OrderBook → Match
                                            │
                           ┌────────────────┴────────────────┐
                           ▼                                 ▼
                     matchBuy()                          matchSell()
                     (vs Asks)                          (vs Bids)
                           │                                 │
                           └────────────────┬────────────────┘
                                            ▼
                                    fillFromLevel()
                                            │
                                            ▼
                                    buildTrade()
                                            │
                                            ▼
                    ┌───────────────────────┴───────────────────────┐
                    ▼                                               ▼
              publishTrade()                                  addToBook()
              (to Kafka)                                        (if unfilled)
```