package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	"github.com/mohmdsaalim/EngineX/internal/config"
	kafkapkg "github.com/mohmdsaalim/EngineX/internal/kafka"
	"github.com/mohmdsaalim/EngineX/internal/settlement"
	"github.com/mohmdsaalim/EngineX/pkg/logger"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	log := logger.New("executor")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Postgres
	pool, err := config.NewPgxPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Error("postgres connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 2. Redis
	redisClient, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// 3. Kafka consumer
	consumer := kafkapkg.NewConsumer(
		cfg.KafkaBroker,
		"trades.executed",
		"executor-group",
	)
	defer consumer.Close()

	// 4. Executor
	executor := settlement.NewExecutor(pool, redisClient)

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info("executor shutting down...")
		cancel()
		pool.Close()
		redisClient.Close()
		consumer.Close()
	}()

	log.Info("executor started", "topic", "trades.executed")

	// 6. Consume loop
	for {
		select {
		case <-ctx.Done():
			log.Info("executor stopped")
			return
		default:
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				log.Error("read message failed", "error", err)
				continue
			}

			// Process trade - commit regardless of result to prevent infinite retry
			// The executor has idempotency protection in DB
			if err := executor.ProcessTrade(ctx, msg.Value); err != nil {
				log.Error("process trade failed", "error", err)
			}

			// Commit offset after each message to prevent infinite retry on failures
			// Idempotency ensures same trade won't be processed twice
			if err := consumer.CommitMessage(ctx, msg); err != nil {
				log.Error("commit message failed", "error", err)
			}
		}
	}
}