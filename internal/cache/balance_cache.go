package cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mohmdsaalim/EngineX/internal/constants"
)


// GetBalance reads user's available balance from Redis.
// Returns balance in scaled int64 (100 = ₹1.00)
func (r *RedisClient) GetBalance(ctx context.Context, userID string) (int64, error){
	key := constants.RedisKeyBalance + userID 
	val, err := r.Get(ctx, key)
	if err != nil{
		return 0, fmt.Errorf("get balance: %w", err )
	}
	if val == ""{
		return 0, nil 
	}
	balance, err := strconv.ParseInt(val, 10, 64)
	if err != nil{
		return 0, fmt.Errorf("parse balance: %w", err)
	}
	return balance, nil
}

// set balace into Redis 
func (r *RedisClient) SetBalance(ctx context.Context, userID string, balance int64) error{
	key := constants.RedisKeyBalance + userID
	return r.Set(ctx, key, strconv.FormatInt(balance, 10), 0) // no TTL - permamanet
}