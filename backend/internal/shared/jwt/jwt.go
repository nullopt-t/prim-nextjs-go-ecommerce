// Package jwt manages the issuance, claims serialization, and verification of JSON Web Tokens.
package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/pkg/config"

	_jwt "github.com/golang-jwt/jwt/v5"
)

// UserClaims embeds standard registered claims alongside application-specific identity attributes.
type UserClaims struct {
	// UserID identifies the user subject
	UserID    uuid.UUID `json:"user_id"`
	// SessionID ties the token to an active session record in Redis
	SessionID string    `json:"session_id"`
	// UserRole specifies authorization role (admin, customer, etc.)
	UserRole  *string   `json:"user_role,omitempty"`
	// Type distinguishes between "access_token" and "refresh_token"
	Type     string    `json:"type"`
	// RegisteredClaims contains standard JWT claims (exp, iat, jti, sub)
	_jwt.RegisteredClaims
}

// JWTManager handles signing and validating authentication tokens.
type JWTManager struct {
	secrets config.Secrets
}

// NewJWTManager creates an instance of JWTManager configured with secrets.
func NewJWTManager(
	secrets config.Secrets,
) *JWTManager {
	return &JWTManager{
		secrets: secrets,
	}
}

const (
	// AccessTokenExpiration defines the lifespan of a short-lived access token (2 hours).
	AccessTokenExpiration  = 2 * time.Hour
	// RefreshTokenExpiration defines the lifespan of a persistent refresh token (30 days).
	RefreshTokenExpiration = 30 * 24 * time.Hour
)

// RandomToken generates a cryptographically secure 256-bit random hex string.
func RandomToken() (string, error) {
	b := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// GenerateAccessToken generates an HMAC-SHA256 signed access token.
func (s *JWTManager) GenerateAccessToken(claims *UserClaims) (string, error) {
	claims.Type = "access_token"
	claims.RegisteredClaims = _jwt.RegisteredClaims{
		ID:        uuid.New().String(),
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

// GenerateRefreshToken generates an HMAC-SHA256 signed refresh token.
func (s *JWTManager) GenerateRefreshToken(claims *UserClaims) (string, error) {
	claims.Type = "refresh_token"
	claims.RegisteredClaims = _jwt.RegisteredClaims{
		ID:        uuid.New().String(),
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

// GenerateTokenPair signs both access and refresh tokens for a user session.
// Clones the claims struct to ensure access and refresh tokens receive distinct, uncorrupted claims.
func (s *JWTManager) GenerateTokenPair(claims *UserClaims) (string, string, error) {
	if claims == nil {
		return "", "", errors.New("claims cannot be nil")
	}

	// Clone claims for access token
	accessClaims := *claims
	accessToken, err := s.GenerateAccessToken(&accessClaims)
	if err != nil {
		return "", "", err
	}

	// Clone claims for refresh token
	refreshClaims := *claims
	refreshToken, err := s.GenerateRefreshToken(&refreshClaims)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// VerifyToken parses and validates a JWT token string against the given secret key.
// Validates HMAC signing method to prevent algorithm confusion attacks and checks token validity.
func (s *JWTManager) VerifyToken(tokenString string, secretKey string) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := _jwt.ParseWithClaims(tokenString, claims, func(token *_jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*_jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

