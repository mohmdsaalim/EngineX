package constants

// Kafka topic names — never hardcode these strings anywhere else
const (
	TopicOrdersSubmitted  = "orders.submitted"
	TopicTradesExecuted   = "trades.executed"
	TopicOrderbookUpdates = "orderbook.updates"
)

// Kafka consumer group names
const (
	GroupEngine   = "engine-group"
	GroupExecutor = "executor-group"
	GroupWSHub    = "hub-group"
)

// Redis key prefixes
const (
	RedisKeySession     = "session:"    // session:{userID}
	RedisKeyBalance     = "balance:"    // balance:{userID}
	RedisKeyIdempotency = "trade:"      // trade:{tradeID}
	RedisKeyRateLimit   = "ratelimit:"  // ratelimit:{userID}
)