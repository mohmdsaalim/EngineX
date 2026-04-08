package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	gRPCauth "github.com/mohmdsaalim/EngineX/api/gen/gRPC_auth"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_risk"
	"github.com/mohmdsaalim/EngineX/internal/config"
	"github.com/mohmdsaalim/EngineX/internal/gateway"
	"github.com/mohmdsaalim/EngineX/internal/kafka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    godotenv.Load()
    cfg := config.Load()
// connect to auth service via grpc
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

	// 5. Wire handler
	handler := gateway.NewHandler(riskClient, kafkaProducer)

	// 6. Setup Gin
	r := gin.Default()
	gateway.SetupRoutes(r, handler, authClient)

	// 7. Start HTTP server
	log.Printf("gateway listening on %s", cfg.GatewayPort)
	if err := r.Run(cfg.GatewayPort); err != nil {
		log.Fatalf("gateway: %v", err)
	}

}