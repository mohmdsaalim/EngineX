package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	"github.com/mohmdsaalim/EngineX/internal/constants"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
)

type Service struct {
	queries    *repository.Queries
	redis      *cache.RedisClient
	JWTManager *JWTManager
}

func NewService(queries *repository.Queries, redis *cache.RedisClient, JWTManager *JWTManager) *Service {
	return &Service{
		queries:    queries,
		redis:      redis,
		JWTManager: JWTManager,
	}
}

func (s *Service) Register(ctx context.Context, email, password, fullName string) (*repository.User, error) {
	exists, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil && exists.ID.Valid {
		return nil, apperr.New(apperr.CodeConflict, "email already registered")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to hash password", err)
	}

	user, err := s.queries.CreateUser(ctx, repository.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		FullName:    fullName,
	})
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to create user", err)
	}
	return &user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}

	if !user.ID.Valid {
		return "", "", apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}

	if err := CheckPassword(password, user.PasswordHash); err != nil {
		return "", "", apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}

	userIDStr := uuid.UUID(user.ID.Bytes).String()
	accessToken, err = s.JWTManager.GenerateAccessToken(userIDStr, user.Email)
	if err != nil {
		return "", "", apperr.Wrap(apperr.CodeInternal, "failed to generate token", err)
	}

	refreshToken, err = s.JWTManager.GenerateRefreshToken(userIDStr, user.Email)
	if err != nil {
		return "", "", apperr.Wrap(apperr.CodeInternal, "failed to generate refresh token", err)
	}

	key := constants.RedisKeySession + userIDStr
	if err := s.redis.Set(ctx, key, refreshToken, 7*24*time.Hour); err != nil {
		return "", "", apperr.Wrap(apperr.CodeInternal, "failed to store session", err)
	}
	return accessToken, refreshToken, nil
}

func (s *Service) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}
	claims, err := s.JWTManager.ValidateToken(token)
	if err != nil {
		return nil, apperr.New(apperr.CodeUnauthorized, "invalid token")
	}
	if claims == nil || claims.UserID == "" {
		return nil, apperr.New(apperr.CodeUnauthorized, "invalid token claims")
	}
	return claims, nil
}