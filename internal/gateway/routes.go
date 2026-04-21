package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_auth"
)

func SetupRoutes(r *gin.Engine, h *Handler, authClient gRPCauth.AuthServiceClient)  {
	// Health - no Auth 
	r.GET("/health", h.Health)
	// login-register
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login",  h.Login)

	// Protected routes - JWT needed
	v1 := r.Group("api/v1")
	v1.Use(AuthMiddleware(authClient))
	{
		v1.POST("/orders", h.SubmitOrder)
		v1.GET("/orderbook/:symbol", h.GetOrderBook)
	}
}
// kafka order.submitted commant 
// docker exec engine_kafka /opt/kafka/bin/kafka-console-consumer.sh \
//   --bootstrap-server localhost:9092 \
//   --topic orders.submitted \
//   --from-beginning \
//   --property print.key=true \
//   --property print.timestamp=true