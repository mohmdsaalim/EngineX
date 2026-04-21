# Project Analysis Report

Generated: Tue Apr 21 2026

## Project Overview
- **Architecture**: Event-driven microservices (6 services)
- **Language**: Go 1.25.5
- **Database**: PostgreSQL + Redis + Kafka
- **Services**: Gateway, Auth, Risk, Engine, Executor, WS Hub

---

## Issues Table

| # | Category | Issue | Location | Severity | Solution |
|---|----------|-------|----------|----------|----------|----------|
| 1 | **Missing Proto** | `gRPC_order/order.pb.go` not generated | `api/gen/gRPC_order/` | High | Run `make proto-order` to generate from `order.proto` |
| 2 | **Import Error** | Gateway imports `gRPC_order` but package doesn't exist | `internal/gateway/handler.go:11` | High | Generate proto file first, then fix import in handler |
| 3 | **Missing wshub** | `cmd/wshub/main.go` not found but Makefile references it | `Makefile:43` | Medium | Create WebSocket hub service or remove from Makefile |
| 4 | **Missing risksvc** | `cmd/risksvc/main.go` not found | `Makefile:31` | Medium | Create risk service entry point |
| 5 | **Graceful Shutdown** | Auth service lacks graceful shutdown | `cmd/authsvc/main.go` | Medium | Add signal handling like executor has at lines 47-55 |
| 6 | **Error Handling** | Redis errors silently ignored in GetBalance | `internal/cache/redis.go:37-40` | Medium | Return proper error instead of nil |
| 7 | **Balance Cache** | No cache invalidation after trade settlement | `internal/settlement/executor.go` | Medium | Invalidate `balance:*` Redis key after debit/credit |
| 8 | **Position Check** | Position check commented out in risk checker | `internal/risk/checker.go:51-55` | Medium | Uncomment and implement for SELL orders |
| 9 | **Hardcoded TTL** | TTL uses nanoseconds (24*60*60*1000000000) | `internal/risk/checker.go:124` | Low | Use `time.Duration` properly: `24 * time.Hour` |
| 10 | **GetOrderBook** | Returns empty placeholder | `internal/gateway/handler.go:154` | Medium | Connect to Kafka or Redis for real data |
| 11 | **No Connection Pool** | Risk service creates new DB connection per call | `internal/risk/` | Medium | Implement connection pooling |
| 12 | **Missing Tests** | No tests for auth, risk, settlement packages | `internal/` | Medium | Add integration tests |
| 13 | **JWT Validation** | Auth middleware calls gRPC for every protected request | `internal/gateway/middleware.go:28` | Medium | Cache token validation in Redis |
| 14 | **OrderBook Race** | Engine uses single mutex for all symbols | `internal/engine/engine.go:34-45` | Low | Consider per-symbol goroutines as design mentions |
| 15 | **Config Override** | No validation for missing env vars on startup | `internal/config/config.go` | Low | Fail fast if required vars missing |
| 16 | **Kafka Producer** | No partition key strategy in PublishOrder | `internal/gateway/kafka_producer.go` | Low | Use user_id or order_id as partition key |
| 17 | **Duplicate Code** | Redis key format inconsistent | `internal/risk/checker.go:113` uses `"order:seen"` vs `internal/cache/idempotency.go` | Low | Create constants for key prefixes |
| 18 | **Missing Logger** | Executor and gateway missing structured logging | Multiple files | Low | Add slog like engine has |
| 19 | **Docker** | No health checks for services in docker-compose | `docker-compose.yml` | Medium | Add healthcheck endpoints |
| 20 | **CI/CD** | No test job in make test-all | `Makefile:66` | Low | Add `-count=1` to avoid test cache |

---

## Critical Issues (Fix First)

### 1. Missing Proto Order File
**File**: `api/gen/gRPC_order/order.pb.go` does not exist
**Impact**: Gateway fails to compile due to missing import
**Fix**:
```bash
make proto-order
```

### 2. Missing Service Entry Points
**Files**: `cmd/wshub/main.go`, `cmd/risksvc/main.go`
**Impact**: Cannot run all services via Makefile
**Fix**: Create missing main.go files or update Makefile

---

## Performance Issues

| Issue | Current | Recommended |
|-------|---------|-------------|
| Token Validation | gRPC call per request | Cache in Redis with TTL |
| OrderBook Access | Empty placeholder | Read from Kafka/Redis |
| Risk Checks | DB call per order | Cache balance in Redis |
| Balance Updates | No invalidation | Invalidate cache after trade |

---

## Security Issues

| Issue | Location | Fix |
|-------|----------|-----|
| Hardcoded JWT Secret | `config.go` | Use environment variable |
| No rate limiting | Gateway | Add rate limit middleware |
| Insecure gRPC creds | `cmd/gateway/main.go:23-24` | Use TLS in production |

---

## Testing Coverage

| Package | Status |
|---------|--------|
| `internal/engine` | Has unit tests |
| `internal/auth` | No tests |
| `internal/risk` | No tests |
| `internal/settlement` | No tests |
| `internal/gateway` | No tests |

---

## Recommended Priority

1. **Week 1**: Generate proto file, create missing services
2. **Week 2**: Add token caching, fix balance cache invalidation
3. **Week 3**: Add tests for auth, risk, settlement
4. **Week 4**: Performance optimization (caching, connection pooling)