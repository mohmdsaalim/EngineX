## Why Go?
Goroutines cost ~2KB stack vs 2MB for OS threads. One goroutine per
symbol matching loop. GC pause <1ms. Benchmark: 500k+ orders/sec.

## Why Kafka and not direct function calls?
Durability (survives crashes), replay (rebuild state), decoupling
(engine and executor scale independently).

## Why gRPC between Gateway and Risk/Auth?
Synchronous request-response needed. Typed contracts via Protobuf.
HTTP/2 multiplexing.

## Why 6 services and not a monolith?
Each scales independently. WS Hub scales on connections, Gateway
on CPU, Executor on Kafka lag. Separate failure domains.

## Why B-tree for the order book?
O(log n) insert/delete with natural sort order. Ascend() for asks,
Descend() for bids. Eliminates auxiliary sorted slice.

## Why int64 for prices, not float64?
IEEE 754 floating-point comparison bugs. 1500.0 == 1500.0 may fail.
Integer comparison is always exact.
// completed the engine need to write the test file and check it 
// tommarrow need fincsh the all oending in cloude tasks and start the freelance and leetcode basics revising deadline
// studying about the freelance working building a team for the works





✅ Step 1  — Topics exist with 6 partitions
✅ Step 2  — Kafka consumer terminal open
✅ Step 5  — SELL order returns 202
✅ Step 6  — Binary protobuf message in orders.submitted
✅ Step 7  — Partition offset incremented
✅ Step 8  — Engine log shows order processed, trades:0
✅ Step 9  — orderbook.updates has JSON snapshot
✅ Step 10 — No unmarshal errors in engine
✅ Step 11 — BUY order returns 202
✅ Step 12 — trades.executed has trade message
✅ Step 13 — Trade in Postgres, balances updated, orders FILLED
✅ Step 14 — Redis has trade:* and order:seen:* keys
