package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

var (
	ErrNoClaimsInContext  = errors.New("no claims found in the context")
	ErrInvalidUserSubject = errors.New("invalid user subject")
)

const (
	// SessionCookieName is the canonical single session cookie name for pure server sessions.
	SessionCookieName = "ps"
)

type SessionStore interface {
	GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error)
}

// Authorize verifies that the authenticated user possesses one of the allowed roles.
// Users with SuperRole ("superadmin") automatically pass any staff-level check.
func Authorize(allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, ok := c.Get("userRole")
		if !ok {
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		userRole := strings.ToLower(roleVal.(string))

		// Superadmin has universal clearance across administrative checkpoints
		if userRole == string(model.SuperRole) {
			c.Next()
			return
		}

		for _, allowed := range allowedRoles {
			if userRole == strings.ToLower(string(allowed)) {
				c.Next()
				return
			}
		}

		_ = c.Error(apierr.ForbiddenError("Insufficient permissions"))
		c.Abort()
	}
}

// Authenticate provides pure server-side session authentication middleware.
// It looks up the session ID via the SessionStore, loads the user_id and userRole, and sets them in context.
func Authenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessionID string

		if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != "" {
			sessionID = cookie
		}

		if sessionID == "" {
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		session, err := sessionStore.GetSessionByID(c.Request.Context(), sessionID)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set("sessionID", session.ID)
		c.Set("userID", session.UserID)
		if session.UserRole != "" {
			c.Set("userRole", session.UserRole)
		} else {
			c.Set("userRole", "customer")
		}

		c.Next()
	}
}

// OptionalAuthenticate attempts to authenticate via the session cookie if present,
// setting user context without aborting or returning 401 if unauthenticated or absent.
func OptionalAuthenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessionID string

		if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != "" {
			sessionID = cookie
		}

		if sessionID != "" {
			if session, err := sessionStore.GetSessionByID(c.Request.Context(), sessionID); err == nil && session != nil {
				c.Set("sessionID", session.ID)
				c.Set("userID", session.UserID)
				if session.UserRole != "" {
					c.Set("userRole", session.UserRole)
				} else {
					c.Set("userRole", "customer")
				}
			}
		}

		c.Next()
	}
}

