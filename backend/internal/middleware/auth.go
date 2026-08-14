package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNoClaimsInContext  = errors.New("no claims found in the context")
	ErrInvalidUserSubject = errors.New("invalid user subject")
)

func Authorize(requiredRole model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("userRole")
		if !ok {
			_ = c.Error(apierr.New(http.StatusUnauthorized, "Unauthorized").WithCode(apierr.CodeUnauthorized))
			c.Abort()
			return
		}
		if role.(string) != strings.ToLower(string(requiredRole)) {
			_ = c.Error(apierr.ErrForbidden("Insufficient permissions").WithCode(apierr.CodeForbidden))
			c.Abort()
			return
		}
		c.Next()
	}
}

// Authenticate provides unified authentication middleware for routes.
// Pass optional=true for public/guest routes that can optionally attach user authentication if present.
func Authenticate(secrets *config.Secrets, optional ...bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		isOptional := len(optional) > 0 && optional[0]

		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			cookie, err := c.Cookie("access_token")
			if err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		if tokenString == "" {
			if isOptional {
				c.Next()
				return
			}
			_ = c.Error(apierr.New(http.StatusUnauthorized, "Unauthorized").WithCode(apierr.CodeUnauthorized))
			c.Abort()
			return
		}

		jwtManager := jwt.NewJWTManager(secrets)
		claims, err := jwtManager.VerifyToken(tokenString, secrets.JwtAccessTokenSecretKey)
		if err != nil {
			if isOptional {
				c.Next()
				return
			}
			_ = c.Error(apierr.New(http.StatusUnauthorized, "Invalid token").WithCode(apierr.CodeUnauthorized))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		if claims.SessionID != "" {
			c.Set("sessionID", claims.SessionID)
		}
		if claims.UserRole != nil {
			c.Set("userRole", *claims.UserRole)
		}
		c.Next()
	}
}

// StrictSessionCheck ensures the session ID from the token still exists in Redis.
// It must be placed AFTER Authenticate.
func StrictSessionCheck(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, exists := c.Get("sessionID")
		if !exists || sessionID == "" {
			c.Next()
			return
		}

		sessKey := fmt.Sprintf("auth:session:%s", sessionID)
		res, err := redisClient.Exists(c.Request.Context(), sessKey).Result()
		if err != nil || res == 0 {
			_ = c.Error(apierr.New(http.StatusUnauthorized, "Session expired or revoked").WithCode(apierr.CodeUnauthorized))
			c.Abort()
			return
		}

		c.Next()
	}
}
