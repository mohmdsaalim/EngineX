package auth

import (
	"context"
	"log"
	"regexp"

	"github.com/mohmdsaalim/EngineX/api/gen/gRPC_auth"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

type GRPCServer struct {
	gRPCauth.UnimplementedAuthServiceServer
	service *Service
}

func NewGRPCServer(svc *Service) *GRPCServer {
	return &GRPCServer{
		service: svc,
	}
}

func (g *GRPCServer) ValidateToken(ctx context.Context, req *gRPCauth.ValidateTokenRequest) (*gRPCauth.ValidateTokenResponse, error) {
	if req.Token == "" {
		return &gRPCauth.ValidateTokenResponse{Valid: false}, nil
	}
	claims, err := g.service.ValidateToken(ctx, req.Token)
	if err != nil {
		log.Printf("validate token error: %v", err)
		return &gRPCauth.ValidateTokenResponse{Valid: false}, nil
	}
	if claims == nil {
		return &gRPCauth.ValidateTokenResponse{Valid: false}, nil
	}
	return &gRPCauth.ValidateTokenResponse{
		Valid:  true,
		UserId: claims.UserID,
		Email:  claims.Email,
	}, nil
}

func (g *GRPCServer) Register(ctx context.Context, req *gRPCauth.RegisterRequest) (*gRPCauth.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return nil, status.Error(codes.InvalidArgument, "email, password and full_name required")
	}

	if !validateEmail(req.Email) {
		return nil, status.Error(codes.InvalidArgument, "invalid email format")
	}

	if !validatePassword(req.Password) {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 8 characters with uppercase, lowercase, and digit")
	}

	user, err := g.service.Register(ctx, req.Email, req.Password, req.FullName)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			if appErr.Code == apperr.CodeConflict {
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

func (g *GRPCServer) Login(ctx context.Context, req *gRPCauth.LoginRequest) (*gRPCauth.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}

	access, refresh, err := g.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return &gRPCauth.LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}