package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"

	_jwt "github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	UserRole *string   `json:"user_role,omitempty"`
	Type     string    `json:"type"`
	_jwt.RegisteredClaims
}

type JWTManager struct {
	secrets *config.Secrets
}

func NewJWTManager(
	secrets *config.Secrets,
) *JWTManager {
	return &JWTManager{
		secrets: secrets,
	}
}

const (
	AccessTokenExpiration  = 15 * time.Minute
	RefreshTokenExpiration = 24 * time.Hour
)

func RandomToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func (s *JWTManager) GenerateAccessToken(claims *UserClaims) (string, error) {
	claims.Type = "access_token"
	claims.RegisteredClaims = _jwt.RegisteredClaims{
		Subject:   claims.UserID.String(),
		IssuedAt:  _jwt.NewNumericDate(time.Now()),
		ExpiresAt: _jwt.NewNumericDate(time.Now().Add(AccessTokenExpiration)),
	}
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secrets.JwtAccessTokenSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *JWTManager) GenerateRefreshToken(claims *UserClaims) (string, error) {
	claims.Type = "refresh_token"
	claims.RegisteredClaims = _jwt.RegisteredClaims{
		Subject:   claims.UserID.String(),
		IssuedAt:  _jwt.NewNumericDate(time.Now()),
		ExpiresAt: _jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiration)),
	}
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secrets.JwtRefreshTokenSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *JWTManager) GenerateTokenPair(claims *UserClaims) (string, string, error) {
	accessToken, err := s.GenerateAccessToken(claims)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.GenerateRefreshToken(claims)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *JWTManager) VerifyToken(tokenString string, secretKey string) (*UserClaims, error) {
	claims := &UserClaims{}
	_, err := _jwt.ParseWithClaims(tokenString, claims, func(token *_jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
