package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	"github.com/mohmdsaalim/EngineX/internal/config"
	kafkapkg "github.com/mohmdsaalim/EngineX/internal/kafka"
	"github.com/mohmdsaalim/EngineX/internal/settlement"
)

func main() {
	godotenv.Load()
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Postgres
	pool, err := config.NewPgxPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	// 2. Redis
	redisClient, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

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
		log.Println("executor shutting down...")
		cancel()
	}()

	log.Println("executor started — consuming trades.executed")

	// 6. Consume loop
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}

			if err := executor.ProcessTrade(ctx, msg.Value); err != nil {
				log.Printf("process trade error: %v", err)
				// Do NOT commit offset — retry on restart
				continue
			}

			// Only commit AFTER successful settlement
			if err := consumer.CommitMessage(ctx, msg); err != nil {
				log.Printf("commit error: %v", err)
			}
		}
	}
}