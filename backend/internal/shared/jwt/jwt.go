package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"

	_jwt "github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID    string `json:"user_id"`
	UserRole  string `json:"user_role"`
	UserEmail string `json:"user_email"`
	Type      string `json:"type"`
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

func (s *JWTManager) GenerateAccessToken(userID, email, userRole string) (string, error) {
	claims := UserClaims{
		UserID:    userID,
		UserEmail: email,
		Type:      "access_token",
		RegisteredClaims: _jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  _jwt.NewNumericDate(time.Now()),
			ExpiresAt: _jwt.NewNumericDate(time.Now().Add(AccessTokenExpiration)),
		},
	}
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secrets.JwtAccessTokenSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := UserClaims{
		UserID: userID,
		Type:   "refresh_token",
		RegisteredClaims: _jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  _jwt.NewNumericDate(time.Now()),
			ExpiresAt: _jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiration)),
		},
	}
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secrets.JwtRefreshTokenSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *JWTManager) GenerateTokenPair(userID, email, userRole string) (string, string, error) {
	accessToken, err := s.GenerateAccessToken(userID, email, userRole)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.GenerateRefreshToken(userID)
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
