package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
)

var (
	ErrNoClaimsInContext  = errors.New("no claims found in the context")
	ErrInvalidUserSubject = errors.New("invalid user subject")
)

func Authorize(requiredRole model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := c.Get("userRole")
		if !ok {
			// No role in context means Authenticate didn't run — treat as unauthenticated
			_ = c.Error(apierr.New(http.StatusUnauthorized, "Unauthorized").WithCode(apierr.CodeUnauthorized))
			c.Abort()
			return
		}
		if role.(string) != strings.ToLower(string(requiredRole)) {
			// User is authenticated but lacks the required role → 403 Forbidden
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
		// TEMP: Turn off auth for testing
		c.Set("userID", uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		c.Set("userRole", "admin")
		c.Next()
	}
}
