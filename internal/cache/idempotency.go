package cache

import (
	"context"
	"time"

	"github.com/mohmdsaalim/EngineX/internal/constants"
)

// IsTradeProcessed checks if trade was already settled.
// Prevents double-settlement on Kafka consumer restart.
func (r *RedisClient) IsTradeProcessed(ctx context.Context, tradeID string) (bool, error){
	key := constants.RedisKeyIdempotency + tradeID
	val, err := r.Get(ctx, key)
	if err != nil{
		return false, err
	}
	return val != "", nil
}

// MarkTradeProcessed marks trade as settled in Redis.
// TTL 24h — after that key expires automatically.
func (r *RedisClient) MarkTradeProcessed(ctx context.Context, tradeID string) error {
	key := constants.RedisKeyIdempotency + tradeID
	return r.Set(ctx, key, "1", 24*time.Hour)
}