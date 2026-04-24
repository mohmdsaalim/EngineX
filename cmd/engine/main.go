package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/internal/config"
	"github.com/mohmdsaalim/EngineX/internal/engine"
	kafkapkg "github.com/mohmdsaalim/EngineX/internal/kafka"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	// 1. Kafka producer — publishes trades + snapshots
	producer := kafkapkg.NewProducer(cfg.KafkaBroker)
	defer producer.Close()

	// 2. Kafka consumer — reads from orders.submitted
	consumer := kafkapkg.NewConsumer(cfg.KafkaBroker, "orders.submitted", "engine-group")
	defer consumer.Close()

	// 3. Create engine
	eng := engine.NewEngine(producer)

	// 4. Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("engine shutting down...")
		cancel()
		producer.Close()
	}()

	log.Println("engine started — consuming orders.submitted")
	eng.Consume(ctx, consumer)
}