package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_auth"
	"github.com/mohmdsaalim/EngineX/internal/auth"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	"github.com/mohmdsaalim/EngineX/internal/config"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"github.com/mohmdsaalim/EngineX/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file not loaded (optional): %v", err)
	}

	cfg := config.Load()
	if cfg == nil {
		log.Fatal("config load failed: configuration is nil")
	}

	ctx := context.Background()

	pool, err := config.NewPgxPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	redis, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("redis connect failed %v", err)
	}
	defer redis.Close()

	queries := repository.New(pool)
	jwtManager := auth.NewJwtManager(cfg.JWTSecret, cfg.JWTAccessTTL)
	authservice := auth.NewService(queries, redis, jwtManager)
	grpcServer := auth.NewGRPCServer(authservice)

	lis, err := net.Listen("tcp", cfg.AuthGRPCPort)
	if err != nil {
		log.Fatalf("falied authgrpc listen: %v", err)
	}
	syslog := logger.New("authsvc")
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(logger.GRPCLoggerInterceptor(syslog)),
	)
	gRPCauth.RegisterAuthServiceServer(srv, grpcServer)
	reflection.Register(srv)

	log.Printf("auth service GRPC listening on %s", cfg.AuthGRPCPort)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Printf("serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down gracefully...")
	srv.GracefulStop()
	log.Println("server stopped")
}