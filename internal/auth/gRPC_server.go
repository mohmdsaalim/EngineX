package auth

import (
	"context"

	"github.com/mohmdsaalim/EngineX/api/gen/gRPCauth"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// this like contoller just accpt req and give responce n logic in here just push the execution into service of auth
// gRPC Authservice
type GRPCServer struct {
	gRPCauth.UnimplementedAuthServiceServer
	service *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{
		service: svc,
	}
}

// ValidateToken - called this func by Gateway on every request.
func (g *GRPCServer) ValidateToken(ctx context.Context, req *gRPCauth.ValidateTokenRequest) (*gRPCauth.ValidateTokenResponse, error){
	claims, err := g.service.ValidateToken(ctx, req.Token)// pushing the token into -> auth/service 
	if err != nil{
		return &gRPCauth.ValidateTokenResponse{Valid: false}, nil
	}
	return &gRPCauth.ValidateTokenResponse{
		Valid: true,
		UserId: claims.UserID,
		Email: claims.Email,
	},nil
}

//Register - creating New user via rpc
func (g *GRPCServer) Register(ctx context.Context, req *gRPCauth.RegisterRequest) (*gRPCauth.RegisterResponse, error){
	if req.Email == "" || req.Password == "" || req.FullName == ""{
		return nil, status.Error(codes.InvalidArgument, "email, password and full_name required")
	}

	user, err := g.service.Register(ctx, req.Email, req.Password, req.FullName)
	if err != nil{
		if appErr, ok := err.(*apperr.AppError); ok{
			if appErr.Code == apperr.CodeConflict{
				return nil, status.Error(codes.AlreadyExists, appErr.Message)
			}
		}
		return nil, status.Error(codes.Internal, "registration failed")
	}
	return &gRPCauth.RegisterResponse{
		UserId: user.ID.String(),
		Email: user.Email,
	}, nil
}

// Login - authenticate and return tokens via gRPC.
func (g *GRPCServer) Login(ctx context.Context, req *gRPCauth.LoginRequest) (*gRPCauth.LoginResponse, error) {
	if req.Email == "" || req.Password == ""{
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}

	access, refresh, err := g.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, " invalid credentials")
	}

	return &gRPCauth.LoginResponse{
		AccessToken: access,
		RefreshToken: refresh,
	}, nil
}
