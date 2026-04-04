package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret string
	accessTTL time.Duration
}

// New jwt manager 
// secret came from config -> 
func NewJwtManager(secret string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret: secret,
		accessTTL: accessTTL,
	}
}

// Gen access token func
func (m *JWTManager) GenerateAccessToken(userID, email string)(string, error) {
	claims := Claims{
		UserID: userID,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.secret))
	if err != nil{
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

//GenerateRefreshToken
func (m *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(m.secret))
	if err != nil{
		return "", fmt.Errorf("sign refresh token: %w", err)
	}
	return signed, nil
}
// -> 
// token validation func 
func (m *JWTManager) ValidateToken(tokenStr string) (*Claims, error){
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
	  func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok{
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(m.secret), nil
	})
	if err != nil{
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid{
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}