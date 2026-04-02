package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/internal/config"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
)

func main() {
    godotenv.Load()

    cfg := config.Load()
    ctx := context.Background()

    pool, err := config.NewPgxPool(ctx, cfg.PostgresDSN)
    if err != nil {
        log.Fatalf("db connect: %v", err)
    }
    defer pool.Close()

    queries := repository.New(pool)
    _ = queries // wire into services next

    log.Println("authsvc started")
    // gRPC server goes here — Day 3 task
    select {}
}