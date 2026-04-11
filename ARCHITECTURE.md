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


// need to update the arch note