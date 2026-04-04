package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}
// Reddis Connection this func cnct -> rds
func NewRedisClient(addr, password string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr : addr,
		Password: password,
		DB: 0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil{
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisClient{client: client}, nil
}

// Set store value in redis
func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get retrivs value by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil{
		return "", nil
	}
	return val, err
}

// delete removes a key 
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}