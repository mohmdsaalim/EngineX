.PHONY: up down migrate seed run-auth run-risk run-gateway run-engine run-executor run-wshub run test build vet clean

# ── Infrastructure ──────────────────────────────────────────
up:
	docker-compose up -d

down:
	docker-compose down -v

# ── Database ─────────────────────────────────────────────────
migrate:
	migrate \
		-path ./migrations \
		-database "postgres://engine_user:engine_pass@localhost:5432/engine_db?sslmode=disable" \
		up

migrate-down:
	migrate \
		-path ./migrations \
		-database "postgres://engine_user:engine_pass@localhost:5432/engine_db?sslmode=disable" \
		down

seed:
	go run scripts/seed.go

# ── Services ─────────────────────────────────────────────────
run-auth:
	go run cmd/authsvc/main.go

run-risk:
	go run cmd/risksvc/main.go

run-gateway:
	go run cmd/gateway/main.go

run-engine:
	go run cmd/engine/main.go

run-executor:
	go run cmd/executor/main.go

run-wshub:
	go run cmd/wshub/main.go

# ── Run all services (requires tmux or separate terminals) ───
run: up migrate seed
	@echo "Start each service manually:"
	@echo "  make run-auth"
	@echo "  make run-risk"
	@echo "  make run-gateway"
	@echo "  make run-engine"
	@echo "  make run-executor"
	@echo "  make run-wshub"

# ── Code Quality ─────────────────────────────────────────────
build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./internal/engine/... -v -race -cover

test-all:
	go test ./... -v -race

# ── sqlc ─────────────────────────────────────────────────────
generate:
	sqlc generate

# ── Proto ────────────────────────────────────────────────────
proto-auth:
	protoc \
		--go_out=. \
		--go_opt=module=github.com/mohmdsaalim/EngineX \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/mohmdsaalim/EngineX \
		api/proto/auth.proto

proto-risk:
	protoc \
		--go_out=. \
		--go_opt=module=github.com/mohmdsaalim/EngineX \
		--go-grpc_out=. \
		--go-grpc_opt=module=github.com/mohmdsaalim/EngineX \
		api/proto/risk.proto

proto-order:
	protoc \
		--go_out=. \
		--go_opt=module=github.com/mohmdsaalim/EngineX \
		api/proto/order.proto

proto: proto-auth proto-risk proto-order

# ── Docker Builds ────────────────────────────────────────────
docker-build:
	docker build -f deployments/docker/authsvc.Dockerfile    -t enginex-authsvc:latest    .
	docker build -f deployments/docker/gateway.Dockerfile    -t enginex-gateway:latest    .
	docker build -f deployments/docker/engine.Dockerfile     -t enginex-engine:latest     .
	docker build -f deployments/docker/executor.Dockerfile   -t enginex-executor:latest   .
	docker build -f deployments/docker/wshub.Dockerfile      -t enginex-wshub:latest      .
	docker build -f deployments/docker/risksvc.Dockerfile    -t enginex-risksvc:latest    .

# ── Kafka Topics ─────────────────────────────────────────────
kafka-topics:
	docker exec engine_kafka /opt/kafka/bin/kafka-topics.sh \
		--bootstrap-server localhost:9092 --list

kafka-orders:
	docker exec engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--topic orders.submitted --from-beginning

kafka-trades:
	docker exec engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--topic trades.executed --from-beginning

kafka-orderbook:
	docker exec engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--topic orderbook.updates --from-beginning

# ── DB Inspect ───────────────────────────────────────────────
db-users:
	docker exec -it engine_postgres psql -U engine_user -d engine_db \
		-c "SELECT id, email, full_name, created_at FROM users;"

db-orders:
	docker exec -it engine_postgres psql -U engine_user -d engine_db \
		-c "SELECT id, user_id, symbol, side, type, price, quantity, filled_qty, status FROM orders;"

db-trades:
	docker exec -it engine_postgres psql -U engine_user -d engine_db \
		-c "SELECT id, symbol, price, quantity, created_at FROM trades;"

db-balances:
	docker exec -it engine_postgres psql -U engine_user -d engine_db \
		-c "SELECT user_id, available, locked FROM balances;"

db-positions:
	docker exec -it engine_postgres psql -U engine_user -d engine_db \
		-c "SELECT user_id, symbol, quantity, locked_qty FROM positions;"

# ── Redis Inspect ────────────────────────────────────────────
redis-sessions:
	docker exec engine_redis redis-cli keys "session:*"

redis-balances:
	docker exec engine_redis redis-cli keys "balance:*"

redis-trades:
	docker exec engine_redis redis-cli keys "trade:*"

redis-orders:
	docker exec engine_redis redis-cli keys "order:seen:*"

# ── Clean ────────────────────────────────────────────────────
clean:
	go clean ./...
	docker-compose down -v
	docker volume prune -f

# ── Full Reset ───────────────────────────────────────────────
reset: clean up migrate seed
	@echo "✅ Full reset complete — ready to run services"