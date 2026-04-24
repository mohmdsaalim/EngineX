# WSHub Service Execution Flow

## Overview
WSHub is the WebSocket hub service that provides real-time orderbook updates to clients. It consumes orderbook snapshots from Kafka and broadcasts them to connected WebSocket clientssubscribed by symbol.

## Architecture
```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WSHUB SERVICE                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐               │
│  │   Kafka      │─────▶│   WSHub      │─────▶│  WebSocket   │               │
│  │ Consumer     │      │    Hub       │      │   Clients    │               │
│  │ orderbook.up │      │              │      │              │               │
│  └──────────────┘      └──────────────┘      └──────────────┘               │
│                              │                                              │
│                              ▼                                              │
│                     ┌──────────────┐                                        │
│                     │  DepthMessage│ ◀──── Wraps snapshot                   │
│                     │  Wrapper     │                                        │
│                     └──────────────┘                                        │
│                            │                                                │
│                   ┌────────-┴─────────┐                                     │
│                   ▼                   ▼                                     │
│              ┌─────────┐          ┌─────────┐                               │
│              │  Bids   │          │  Asks   │                               │
│              │ Client  │          │ Client  │                               │
│              │   1     │          │   2...n │                               │
│              └─────────┘          └─────────┘                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Previous Code Issues (Before Fix)

### Issue 1: No Origin Check
```go
// BEFORE (INSECURE)
CheckOrigin: func(r *http.Request) bool {
    return true // allow all origins — vulnerable to CSRF
},
```
- Allowed all origins, security risk in production

### Issue 2: No Client Limits
```go
// BEFORE (NO LIMIT)
func (h *Hub) register(symbol string, c *Client) {
    h.clients[symbol] = append(h.clients[symbol], c)
}
```
- No limit on clients per symbol - DoS vulnerability

### Issue 3: Commit Before Broadcast
```go
// BEFORE (WRONG ORDER)
h.Broadcast(symbol, msg.Value)
consumer.CommitMessage(ctx, msg)
```
- Message committed before broadcast - data loss if broadcast fails

### Issue 4: No Symbol Validation
```go
// BEFORE (NO VALIDATION)
symbol := string(msg.Key)
if symbol == "" {
    continue
}
```
- No validation of symbol format or length

### Issue 5: buildDepthMessage Not Used
```go
// BEFORE (UNUSED FUNCTION)
func buildDepthMessage(symbol string, raw []byte) []byte {
    // Function defined but never called
    var data interface{}
    json.Unmarshal(raw, &data)  // Error ignored
    ...
}
```
- Raw Kafka messages broadcast directly instead of wrapped

### Issue 6: Silent Write Errors
```go
// BEFORE (NO LOGGING)
if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
    return  // Silent failure
}
```
- Write errors silently ignored

### Issue 7: Graceful Shutdown Race
```go
// BEFORE (POTENTIAL RACE)
cancel()      // Cancel first
srv.Shutdown(ctx)
```
- Context cancelled before graceful shutdown completes

## Updated Code (After Fix)

### Fix 1: Restricted Origin Check
```go
// AFTER (SECURE)
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "" || origin == "http://localhost" || origin == "https://yourdomain.com"
},
```
- Restricts origins to known domains

### Fix 2: Max Clients Per Symbol
```go
// AFTER (WITH LIMIT)
const maxClientsPerSymbol = 1000

type Hub struct {
    mu               sync.RWMutex
    clients          map[string][]*Client
    log              *slog.Logger
}

func (h *Hub) register(symbol string, c *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if len(h.clients[symbol]) >= maxClientsPerSymbol {
        h.log.Error("max clients limit reached", "symbol", symbol)
        return
    }
    h.clients[symbol] = append(h.clients[symbol], c)
}
```
- Limits clients to 1000 per symbol

### Fix 3: Commit Before Broadcast
```go
// AFTER (CORRECT ORDER)
consumer.CommitMessage(ctx, msg)
wrappedMsg := buildDepthMessage(symbol, msg.Value)
if wrappedMsg != nil {
    h.Broadcast(symbol, wrappedMsg)
}
```
- Commits message before broadcast

### Fix 4: Symbol Validation
```go
// AFTER (WITH VALIDATION)
if len(symbol) < 1 || len(symbol) > 10 {
    h.log.Error("invalid symbol format", "symbol", symbol)
    consumer.CommitMessage(ctx, msg)
    continue
}
```
- Validates symbol length (1-10 chars)

### Fix 5: Use buildDepthMessage
```go
// AFTER (WRAPS MESSAGE)
wrappedMsg := buildDepthMessage(symbol, msg.Value)
if wrappedMsg != nil {
    h.Broadcast(symbol, wrappedMsg)
}
```
- Wraps raw Kafka data in DepthMessage format

### Fix 6: Log Write Errors
```go
// AFTER (WITH LOGGING)
if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
    c.log.Error("write failed", "error", err)
    return
}
```
- Logs write errors

### Fix 7: Shutdown Order Fixed
```go
// AFTER (CORRECT ORDER)
go func() {
    <-quit
    log.Println("wshub shutting down...")
    srv.Shutdown(ctx)
    cancel()
}()
```
- Shutdown completes before cancel

## Core Flow

### 1. Kafka Consumer Flow
```
Consume(ctx, consumer)
    │
    ├──► ReadMessage() from Kafka topic: orderbook.updates
    │
    ├──► Extract symbol from message key
    │
    ├──► Validate symbol format (length 1-10)
    │
    ├──► CommitMessage() to Kafka
    │
    ├──► buildDepthMessage() - wrap in DepthMessage format
    │
    └──► Broadcast() to all clients subscribed to symbol
```

### 2. Client Connection Flow
```
HandleWS(w, r)
    │
    ├──► Get symbol from URL query param: ?symbol=INFY
    │
    ���─��► Validate symbol required
    │
    ├──► Upgrade HTTP to WebSocket
    │
    ├──► Create client with sendCh buffer
    │
    ├──► register() client to symbol
    │
    └──► Start WritePump() goroutine
```

### 3. WritePump (Client Message Writer)
```
WritePump(ctx, onDone)
    │
    ├──► Setup ping ticker (every 30s)
    │
    ├──► select:
    │     ├──► ctx.Done() → close connection
    │     │
    │     ├──► sendCh msg → WriteMessage() to WS
    │     │
    │     └──► ticker.C → send PingMessage()
    │
    └──► On exit: close connection & call onDone()
```

## DepthMessage Format

### Message Structure
```json
{
  "type": "depth",
  "symbol": "INFY",
  "data": {
    "symbol": "INFY",
    "bids": [
      { "price": 150000, "quantity": 100 },
      { "price": 149000, "quantity": 50 }
    ],
    "asks": [
      { "price": 151000, "quantity": 25 },
      { "price": 152000, "quantity": 75 }
    ]
  }
}
```

## Client Duties

### WebSocket Client Responsibilities
1. **Connect** - Establish WebSocket connection with symbol query param
2. **Listen** - Receive depth updates in real-time
3. **Ping/Pong** - Maintain connection via keepalive (30s interval)
4. **Handle Drop** - Reconnect on disconnect

### Client Connection URL
```
ws://host:port/ws?symbol=INFY
```

## Kafka Topics

| Topic | Direction | Purpose |
|-------|-----------|---------|
| orderbook.updates | In | Engine → WSHub (depth snapshots) |

## Error Handling

| Error | Handling |
|-------|-----------|
| Invalid symbol | Log error, commit message, skip |
| Symbol validation fail | Log error, commit, skip |
| JSON unmarshal fail | Return nil, don't broadcast, log error |
| Write to client fail | Log error, close connection |
| Max clients reached | Log error, reject new connection |
| Kafka read error | Log error, continue |

## Core Components

| Component | File | Role |
|------------|------|------|
| Hub | hub.go | Manages client connections, consumes Kafka, broadcasts |
| Client | client.go | Individual WebSocket connection, send buffer |
| DepthMessage | hub.go | Wrapper struct for depth updates |

## Flow Summary

```
Kafka Message (orderbook.updates)
         │
         ▼
   ReadMessage()
         │
         ▼
  Extract Symbol
         │
         ▼
 Validate Symbol
         │
         ▼
 CommitMessage()
         │
         ▼
buildDepthMessage()
         │
         ▼
  Broadcast()
         │
         ▼
   Send to all clients
         │
    ┌────┴────┐
    ▼         ▼
Client 1  Client 2...
```

## REST Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| /ws | GET | WebSocket upgrade |
| /healthz | GET | Health check |

## Configuration

| Setting | Value | Description |
|----------|-------|-------------|
| maxClientsPerSymbol | 1000 | Max clients per symbol |
| pingInterval | 30s | Ping interval for keepalive |
| writeTimeout | 10s | Write timeout |
| sendBufferSize | 256 | Client send buffer |