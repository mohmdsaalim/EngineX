package risk

import (
	"context"
	// "log"

	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_risk"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


type GRPCServer struct {
	gRPC_risk.UnimplementedRiskServiceServer
	checker *Checker
}

func NewGRPCServer(checker *Checker) *GRPCServer {
	return &GRPCServer{checker: checker}
}

// checkOrder - called by gateway before every order goes to kafka

func (g *GRPCServer) CheckOrder(ctx context.Context, req *gRPC_risk.CheckOrderRequest) (*gRPC_risk.CheckOrderResponse, error) {
	// basic validation
	if req.UserId == "" || req.OrderId == "" || req.Price <= 0 || req.Quantity <= 0{
		return &gRPC_risk.CheckOrderResponse{
			Approved: false,
			RejectReason: "missing requied fields",
		}, nil
	}

	sideStr := "BUY"
	// log.Printf("side received: %v", req.Side)
	if req.Side == gRPC_risk.Side_SIDE_SELL{
		sideStr = "SELL"
	} // logic is confusing need to rechek and study////////////////////////

	err := g.checker.CheckOrder(ctx, OrderRequest{
		UserID:   req.UserId,
		orderID:  req.OrderId,
		Symbol:   req.Symbol,
		Side:     sideStr,
		Price:    req.Price,
		Quantity: req.Quantity,
	})

	if err != nil{
		// approved  = false with reason no grpc err
		if appErr, ok := err.(*apperr.AppError); ok {
			return  &gRPC_risk.CheckOrderResponse{
				Approved: false,
				RejectReason: appErr.Message,
			}, nil
		}
		return nil, status.Error(codes.Internal, " risk check failed ")
	}
	return &gRPC_risk.CheckOrderResponse{
		Approved: true,
	}, nil
}