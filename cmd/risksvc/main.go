package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
     
	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_risk"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	"github.com/mohmdsaalim/EngineX/internal/config"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"github.com/mohmdsaalim/EngineX/internal/risk"
	"github.com/mohmdsaalim/EngineX/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
    godotenv.Load()

    cfg := config.Load()
    ctx := context.Background()

    pool, err := config.NewPgxPool(ctx, cfg.PostgresDSN)
    if err != nil {
        log.Fatalf("db connect postgres: %v", err)
    }
    defer pool.Close()
    // redis 
    redisClient, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
    if err != nil{
        log.Fatalf("redis: %v", err)
    }
 // 3 depndcy injection
    queries := repository.New(pool)
    checker := risk.NewChecker(redisClient, queries)
    grpcServer := risk.NewGRPCServer(checker)

    // 4 Start gRPC
    lis, err := net.Listen("tcp", cfg.RiskGRPCPort)
    if err != nil{
        log.Fatalf("Listen: %v ", err)
    }

    syslog := logger.New("risksvc")
    srv := grpc.NewServer(
        grpc.UnaryInterceptor(logger.GRPCLoggerInterceptor(syslog)),
    )
    gRPC_risk.RegisterRiskServiceServer(srv, grpcServer)
    reflection.Register(srv)

    log.Printf("risksvc gRPC listening on %s", cfg.RiskGRPCPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("risksvc shutting down...")
		srv.GracefulStop()
	}()

	if err := srv.Serve(lis); err != nil{
		log.Fatalf("serve: %v", err)
	}
    
}