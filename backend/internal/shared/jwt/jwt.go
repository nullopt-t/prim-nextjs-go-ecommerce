package jwt

import (
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
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

func (s *JWTManager) GenerateAccessToken(userID, userRole string) (string, error) {
	claims := UserClaims{
		UserID: userID,
		Type:   "access_token",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secrets.JwtRefreshTokenSecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *JWTManager) GenerateTokenPair(userID, userRole string) (string, string, error) {
	accessToken, err := s.GenerateAccessToken(userID, userRole)
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
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
