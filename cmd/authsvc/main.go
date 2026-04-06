package main

import (
	"context"
	"log"
	"net"

	"github.com/joho/godotenv"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPCauth"
	"github.com/mohmdsaalim/EngineX/internal/auth"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	"github.com/mohmdsaalim/EngineX/internal/config"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
    godotenv.Load()

    cfg := config.Load() //. config load
    ctx := context.Background() 

    pool, err := config.NewPgxPool(ctx, cfg.PostgresDSN) // postgres connct
    if err != nil {
        log.Fatalf("db connect: %v", err)
    }
    defer pool.Close()

    // Connect Redis
    redis, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword)
    if err != nil {
        log.Fatalf("redis connect failed %v", err)
    }
    // injecting dependencies -> 
    queries := repository.New(pool)
    jwtManager := auth.NewJwtManager(cfg.JWTSecret, cfg.JWTAccessTTL)
    authservice := auth.NewService(queries, redis, jwtManager)
    grpcServer := auth.NewGRPCServer(authservice)

    // start Grpc servcer
    lis, err := net.Listen("tcp", cfg.AuthGRPCPort)
    if err != nil{
        log.Fatalf("falied authgrpc listen: %v", err)
    }
    srv := grpc.NewServer()
    gRPCauth.RegisterAuthServiceServer(srv, grpcServer)
    reflection.Register(srv)
    
    log.Printf("auth service GRPC listening on %s", cfg.AuthGRPCPort)
    if err := srv.Serve(lis); err!= nil{
        log.Fatalf(" serve: %v", err)
    }
    // graceful shutdown pending............
}