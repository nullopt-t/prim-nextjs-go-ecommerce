package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// Package middleware provides HTTP middleware handlers for the Gin web framework,
// including authentication, authorization, caching, CORS, error normalization,
// localization, rate limiting, and security headers.

var (
	// ErrNoClaimsInContext indicates that required claims could not be extracted from the request context.
	ErrNoClaimsInContext = errors.New("no claims found in the context")

	// ErrInvalidUserSubject indicates that the user subject extracted from context or token is malformed.
	ErrInvalidUserSubject = errors.New("invalid user subject")
)

const (
	// SessionCookieName is the canonical single session cookie name for pure server sessions.
	// In the browser, this cookie is expected to be marked HttpOnly, Secure, and SameSite.
	SessionCookieName = "ps"
)

// SessionStore defines the contract for retrieving server-side session models
// by their unique session identifier string (e.g. backed by Redis or PostgreSQL).
type SessionStore interface {
	// GetSessionByID retrieves a UserSession by its unique identifier.
	// It returns the session pointer or an error if the lookup fails or the session does not exist.
	GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error)
}

// Authorize verifies that the authenticated user possesses one of the allowed roles.
// Users with SuperRole ("superadmin") automatically pass any staff-level check.
// If the role is missing or does not match the allowlist, the request is aborted
// with a 401 Unauthorized or 403 Forbidden error response respectively.
func Authorize(allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attempt to extract the role string set by preceding authentication middleware
		roleVal, ok := c.Get("userRole")
		if !ok {
			// No user role found in Gin context -> user has not been authenticated
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		// Safe type assertion supporting string, model.UserRole, and fmt.Stringer
		var userRole string
		switch r := roleVal.(type) {
		case string:
			userRole = strings.ToLower(r)
		case model.UserRole:
			userRole = strings.ToLower(string(r))
		default:
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		// Superadmin has universal clearance across administrative checkpoints
		if userRole == string(model.SuperRole) {
			c.Next()
			return
		}

		// Check if the user's role matches any of the explicitly allowed roles
		for _, allowed := range allowedRoles {
			if userRole == strings.ToLower(string(allowed)) {
				c.Next()
				return
			}
		}

		// Role was found but is not permitted to access this endpoint -> 403 Forbidden
		_ = c.Error(apierr.ForbiddenError("Insufficient permissions"))
		c.Abort()
	}
}

// extractSessionID extracts the session identifier from the session cookie.
func extractSessionID(c *gin.Context) string {
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

// Authenticate provides pure server-side session authentication middleware.
// It looks up the session ID via the SessionStore, loads the user_id and userRole,
// and sets them in the Gin context. If the session cookie is missing, invalid, or expired,
// it halts the middleware chain and aborts with an authentication error.
// The session snapshot is trusted as the source of truth for role and status.
func Authenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := extractSessionID(c)

		// If no session ID could be extracted, reject request immediately with 401 Unauthorized
		if sessionID == "" {
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		// Retrieve session record from session storage using the request context for tracing/cancellation
		session, err := sessionStore.GetSessionByID(c.Request.Context(), sessionID)
		if err != nil {
			var ae *apierr.APIError
			if errors.As(err, &ae) {
				_ = c.Error(err)
			} else {
				_ = c.Error(apierr.UnauthorizedError())
			}
			c.Abort()
			return
		}

		if session == nil {
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		// Reject guest sessions or sessions with empty/nil UserID to prevent privilege escalation
		if session.Type == model.SessionTypeGuest || session.UserID == uuid.Nil {
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		c.Set("sessionID", session.ID)
		c.Set("userID", session.UserID)
		if session.UserRole != "" {
			c.Set("userRole", string(session.UserRole))
		} else {
			c.Set("userRole", "customer")
		}

		// Proceed to subsequent handlers in the Gin middleware chain
		c.Next()
	}
}

// OptionalAuthenticate attempts to authenticate via the session cookie or Bearer header if present,
// setting user context without aborting or returning 401 if unauthenticated or absent.
// This is used on public routes that can provide enriched responses when logged in.
func OptionalAuthenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := extractSessionID(c)

		if sessionID != "" {
			if session, err := sessionStore.GetSessionByID(c.Request.Context(), sessionID); err == nil && session != nil {
				// Prevent guest session privilege escalation in optional authentication
				if session.Type != model.SessionTypeGuest && session.UserID != uuid.Nil {
					c.Set("sessionID", session.ID)
					c.Set("userID", session.UserID)
					if session.UserRole != "" {
						c.Set("userRole", string(session.UserRole))
					} else {
						c.Set("userRole", "customer")
					}
				}
			}
		}

		// Always continue execution regardless of authentication outcome
		c.Next()
	}
}
