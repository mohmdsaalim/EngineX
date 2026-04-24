package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/internal/config"
	kafkapkg "github.com/mohmdsaalim/EngineX/internal/kafka"
	wshub "github.com/mohmdsaalim/EngineX/internal/websocket"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Kafka consumer
	consumer := kafkapkg.NewConsumer(
		cfg.KafkaBroker,
		"orderbook.updates",
		"hub-group",
	)
	defer consumer.Close()

	// 2. Hub
	hub := wshub.NewHub()

	// 3. Start Kafka consumer in background
	go hub.Consume(ctx, consumer)

	// 4. HTTP server for WebSocket
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.HandleWS)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:    cfg.WSHubPort,
		Handler: mux,
	}

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("wshub shutting down...")
		srv.Shutdown(ctx)
		cancel()
	}()

	log.Printf("wshub listening on %s", cfg.WSHubPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("wshub: %v", err)
	}
}