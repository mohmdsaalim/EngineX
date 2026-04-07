package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_auth"
)

func SetupRoutes(r *gin.Engine, h *Handler, authClient gRPCauth.AuthServiceClient)  {
	// Health - no Auth 
	r.GET("/health", h.Health)
	// Login
	// register

	// Protected routes - JWT needed
	v1 := r.Group("api/v1")
	v1.Use(AuthMiddleware(authClient))
	{
		v1.POST("/orders", h.SubmitOrder)
		v1.GET("/orderbook/:symbol", h.GetOrderBook)
	}
}