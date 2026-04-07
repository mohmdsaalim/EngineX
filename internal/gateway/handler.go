package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_order"
	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_risk"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
	"github.com/mohmdsaalim/EngineX/pkg/response"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	riskClient gRPC_risk.RiskServiceClient
	KafkaProducer *KafkaProducer
}

func NewHandler(riskClient gRPC_risk.RiskServiceClient,KafkaProducer *KafkaProducer ) *Handler {
	return &Handler{
		riskClient: riskClient,
		KafkaProducer: KafkaProducer,
	}
}

// SubmitOrderRequest is the incoming HTTP request body.
type SubmitOrderRequest struct {
	Symbol   string `json:"symbol"    binding:"required"`
	Side     string `json:"side"      binding:"required,oneof=BUY SELL"`
	Type     string `json:"type"      binding:"required,oneof=LIMIT MARKET"`
	Price    int64  `json:"price"`
	Quantity int64  `json:"quantity"  binding:"required,gt=0"`
}

// OrderMessage is what gets published to Kafka.
type OrderMessage struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Type      string    `json:"type"`
	Price     int64     `json:"price"`
	Quantity  int64     `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

// submitOrder handles POST/api/v1/orders
// flow -> validate -> risk check -> publish to kafka -> 202
func (h *Handler) SubmitOrder(c *gin.Context)  {
	// parse and validate requet body
	var req SubmitOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil{
		response.Fail(c, apperr.New(apperr.CodeInvalidInput, err.Error()))
		return
	}

	// LIMIT order must have price > 0
	if req.Type == "LIMIT" && req.Price <= 0{
		response.Fail(c, apperr.New(apperr.CodeInvalidInput, " price required for LIMIT order"))
		return
	}

	// get userID from JWT midleware context
	userID := c.GetString("userID")

	// generate unique orderID
	orderID := uuid.New().String()

	// Risk check via gRPC 
	sideEnum := gRPC_risk.Side_SIDE_BUY
	if req.Side == "SELL"{
		sideEnum = gRPC_risk.Side_SIDE_SELL
	}

	riskResp, err := h.riskClient.CheckOrder(c.Request.Context(),
		&gRPC_risk.CheckOrderRequest{
			UserId: userID,
			OrderId: orderID,
			Symbol: req.Symbol,
			Side: sideEnum,
			Price: req.Price,
			Quantity: req.Quantity,
})
if err != nil{
	response.Fail(c, apperr.New(apperr.CodeInternal, "risk service unavailable"))
	return
}

if !riskResp.Approved{
	response.Fail(c, apperr.New(apperr.CodeForbidden, riskResp.RejectReason))
	return
}

// chnaged JSON to Protobufs
payload, err := proto.Marshal(&gRPC_order.OrderMessage{
	OrderId: orderID,
	UserId: userID,
	Symbol: req.Symbol,
	Side: req.Side,
	Type: req.Type,
	Price: req.Price,
	Quantity: req.Quantity,
	CreatedAt: time.Now().UnixNano(),
})

if err != nil{
	response.Fail(c, apperr.New(apperr.CodeInternal, "failed to serialize order"))
	return
}

// Publish to kafka orders.submitted
ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
defer cancel()

if err := h.KafkaProducer.PublishOrder(ctx, req.Symbol, payload); err !=nil{
	response.Fail(c, apperr.New(apperr.CodeInternal, "failed to queue order to kafka"))
	return
}

response.Accepted(c, gin.H{
	"order_id":orderID,
	"status": "queued",
	"message" : "order submitted successfully",
})

}

func (h *Handler) GetOrderBook(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		response.Fail(c, apperr.New(apperr.CodeInvalidInput, "symbol required"))
		return
	}
	// placeholder — engine will serve this in Day 6
	response.OK(c, gin.H{"symbol": symbol, "bids": []interface{}{}, "asks": []interface{}{}})
}

// Health handles GET /healthz
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}