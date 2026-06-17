package middleware

import (
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"errors"
	"net/http"
	"strings"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"

	"github.com/gin-gonic/gin"
)

var (
	ErrNoClaimsInContext  = errors.New("no claims found in the context")
	ErrInvalidUserSubject = errors.New("invalid user subject")
)

const prefix = "Bearer "

func ClaimsWithContext(c *gin.Context) (*jwt.UserClaims, error) {
	claims, ok := c.Get("claims")
	if !ok {
		return nil, ErrNoClaimsInContext
	}
	tclaims, ok := claims.(*jwt.UserClaims)
	if !ok {
		return nil, ErrInvalidUserSubject
	}
	return tclaims, nil
}

func Authorize(requiredRole model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := ClaimsWithContext(c)
		if err != nil {
			c.Error(
				security.NewSecureError(
					http.StatusUnauthorized,
					"INVALID_USER",
					"invalid user",
					err,
				),
			)
			c.Abort()
			return
		}
		if claims.UserRole != string(requiredRole) {
			c.Error(
				security.NewSecureError(
					http.StatusUnauthorized,
					"FORBIDDEN",
					"no permission",
					nil,
				),
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

func Authanticate(secrets *config.Secrets) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.Error(
				security.NewSecureError(
					http.StatusUnauthorized,
					"MISSING_AUTH_HEADER",
					"authorization header is required",
					nil,
				),
			)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, prefix) {
			c.Error(
				security.NewSecureError(
					http.StatusUnauthorized,
					"INVALID_AUTH_FORMAT",
					"authorization header must use Bearer scheme",
					nil,
				),
			)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, prefix)

		jwt := jwt.NewJWTManager(secrets)
		claims, err := jwt.VerifyToken(
			tokenString,
			secrets.JwtAccessTokenSecretKey,
		)

		if err != nil {
			c.Error(
				security.NewSecureError(
					http.StatusUnauthorized,
					"INVALID_TOKEN",
					"invalid or expired token",
					err,
				),
			)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
