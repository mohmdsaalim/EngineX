package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	gRPCauth "github.com/mohmdsaalim/EngineX/api/gen/gRPC_auth"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_risk"
	"github.com/mohmdsaalim/EngineX/internal/config"
	"github.com/mohmdsaalim/EngineX/internal/gateway"
	"github.com/mohmdsaalim/EngineX/internal/kafka"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    godotenv.Load()
    cfg := config.Load()
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 1. Postgres connection
    pool, err := config.NewPgxPool(ctx, cfg.PostgresDSN)
    if err != nil {
        log.Fatalf("postgres connection failed: %v", err)
    }
    defer pool.Close()

    // 2. connect to auth service via grpc
    authConn, err := grpc.NewClient(
        "localhost"+cfg.AuthGRPCPort,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )

    if err != nil{
        log.Fatalf("auth connect: %v" , err)
    }

    defer authConn.Close()
    
    // connect to risk service via gRPC
    riskConn, err := grpc.NewClient(
		"localhost"+cfg.RiskGRPCPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("risk connect: %v", err)
	}
	defer riskConn.Close()

    // 3. Create gRPC clients
	authClient := gRPCauth.NewAuthServiceClient(authConn)
	riskClient := gRPC_risk.NewRiskServiceClient(riskConn)

	// 4. Kafka producer
	baseProducer := kafka.NewProducer(cfg.KafkaBroker)
	defer baseProducer.Close()
	
	kafkaProducer := gateway.NewKafkaProducer(baseProducer)

	// 5. Repository
	q := repository.New(pool)

	// 6. Wire handler
	handler := gateway.NewHandler(riskClient, kafkaProducer, authClient, q)

	// 6. Setup Gin
	r := gin.Default()
	gateway.SetupRoutes(r, handler, authClient)

	srv := &http.Server{
		Addr:    cfg.GatewayPort,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("gateway shutting down...")
		srv.Shutdown(context.Background())
	}()

	log.Printf("gateway listening on %s", cfg.GatewayPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway: %v", err)
	}

}