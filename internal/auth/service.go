package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohmdsaalim/EngineX/internal/cache"
	"github.com/mohmdsaalim/EngineX/internal/constants"
	repository "github.com/mohmdsaalim/EngineX/internal/repository/generated"
	"github.com/mohmdsaalim/EngineX/pkg/apperr"
)
//
type Service struct {
	queries 	*repository.Queries
	redis   	*cache.RedisClient
	JWTManager  *JWTManager
}
// func to inject 
func NewService(queries *repository.Queries,redis *cache.RedisClient, JWTManager *JWTManager)*Service {
	return &Service{
		queries: queries,
		redis: redis,
		JWTManager: JWTManager,
	}
}
//. Register -> create new user
func (s *Service) Register(ctx context.Context, email, password, fullName string)(*repository.User, error) {
	// checking user alreay exist 
	existing, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil && existing.ID.Valid{
		return nil, apperr.New(apperr.CodeConflict, "email already registered")
	}
	// hash password 
	hash, err := HashPassword(password)
	if err != nil{
		return nil, apperr.Wrap(apperr.CodeInternal, "failed to hash password", err)
	}

	// save the data into db
	user, err := s.queries.CreateUser(ctx, repository.CreateUserParams{
		Email: email,
		PasswordHash: hash,
		FullName: fullName,
	})
	if err != nil{
		return nil, apperr.Wrap(apperr.CodeInternal, " failed to create user", err)
	}
	return &user, nil
}


// Login -> validate token
func (s *Service) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	// Find user ny email
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil{
		return "", "", apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}
	// verify password
	if err := CheckPassword(password, user.PasswordHash); err != nil{
		return "", "", apperr.New(apperr.CodeUnauthorized, "invalid credentials")
	}

	// Gen access token (15min)
	 userIDStr := uuid.UUID(user.ID.Bytes).String()
	accessToken, err = s.JWTManager.GenerateAccessToken(userIDStr, user.Email)
	if err != nil{
		return "", "", apperr.Wrap(apperr.CodeInternal, " failed to generate token ", err)
	}

	// Gen refresh token (7 days)
	refreshToken, err = s.JWTManager.GenerateRefreshToken(userIDStr)
	if err != nil{
		return "", "", apperr.Wrap(apperr.CodeInternal, "failed to generate refresh token", err)
	}

	// Store refresh token in redis 
	key := constants.RedisKeyBalance + userIDStr
	if err := s.redis.Set(ctx, key, refreshToken, 7*24*time.Hour); err != nil{
		return "", "", apperr.Wrap(apperr.CodeInternal, "failed to store session", err)
	}
	return accessToken, refreshToken, nil
// this func
// Login verifies credentials and returns JWT tokens.
// Flow: find user → verify password → generate tokens → store refresh in Redis
}

// ValidateToken 
func (s *Service) ValidateToken(ctx context.Context, token string) (*Claims, error){
	claims, err := s. JWTManager.ValidateToken(token)
	if err != nil{
		return nil, apperr.New(apperr.CodeUnauthorized, " invalid token")
	}
	return claims, nil
}