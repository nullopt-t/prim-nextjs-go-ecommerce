package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"

	"github.com/gin-gonic/gin"
)

var (
	ErrNoClaimsInContext  = errors.New("no claims found in the context")
	ErrInvalidUserSubject = errors.New("invalid user subject")
)

const prefix = "Bearer "

func Authorize(requiredRole model.RoleCode) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("role")
		if !ok {
			_ = c.Error(
				security.NewSecureError(
					http.StatusUnauthorized,
					"INVALID_USER",
					"invalid user",
					nil,
				),
			)
			c.Abort()
			return
		}
		if role.(string) != string(requiredRole) {
			_ = c.Error(
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
			_ = c.Error(
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
			_ = c.Error(
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
			_ = c.Error(
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

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.UserEmail)
		c.Set("userRole", claims.UserRole)
		c.Next()
	}
}
