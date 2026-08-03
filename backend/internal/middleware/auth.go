package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

var (
	ErrNoClaimsInContext  = errors.New("no claims found in the context")
	ErrInvalidUserSubject = errors.New("invalid user subject")
)

const prefix = "Bearer "

func Authorize(requiredRole model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("userRole")
		if !ok {
			_ = c.Error(
				apierr.New(
					http.StatusUnauthorized,
					"Unauthorized",
				).WithCode(apierr.CodeUnauthorized),
			)
			c.Abort()
			return
		}
		if role.(string) != strings.ToLower(string(requiredRole)) {
			_ = c.Error(
				apierr.New(
					http.StatusUnauthorized,
					"Unauthorized",
				).WithCode(apierr.CodeUnauthorized),
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
				apierr.New(
					http.StatusUnauthorized,
					"Unauthorized",
				).WithCode(apierr.CodeUnauthorized),
			)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, prefix) {
			_ = c.Error(
				apierr.New(
					http.StatusUnauthorized,
					"Unauthorized",
				).WithCode(apierr.CodeUnauthorized),
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
				apierr.New(
					http.StatusUnauthorized,
					"Unauthorized",
				).WithCode(apierr.CodeUnauthorized),
			)
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.UserRole)
		c.Next()
	}
}
