package service

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"server/internal/config"
	"server/internal/models"
	"time"
)

type tokenType string

const (
	accessTokenType  tokenType = "access"
	refreshTokenType tokenType = "refresh"
)

type TokenService struct {
	cfg *config.Config
}

func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{cfg: cfg}
}

type Claims struct {
	UserID uint        `json:"user_id"`
	Role   models.Role `json:"role"`
	Type   tokenType   `json:"type"`
	jwt.RegisteredClaims
}

func (s *TokenService) GenerateTokens(user *models.User) (string, string, error) {
	accessToken, err := s.generateToken(user, accessTokenType, s.cfg.JWTAccessSecret, time.Hour*24*21)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.generateToken(user, refreshTokenType, s.cfg.JWTRefreshSecret, time.Hour*24*30)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *TokenService) generateToken(user *models.User, tType tokenType, secret string, expiration time.Duration) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		Type:   tType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *TokenService) ValidateToken(tokenString string, tType tokenType, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.Type != tType {
			return nil, fmt.Errorf("invalid token type")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}