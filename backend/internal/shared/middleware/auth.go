package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
	ErrInvalidUserSubject = errors.New("invalid subject")

	// ErrMissingSessionCookie indicates that the request did not carry the
	// canonical session cookie. Callers decide whether that is fatal (authenticated
	// routes) or simply means the request is anonymous (optional routes).
	ErrMissingSessionCookie = errors.New("session cookie is missing")
)

const (
	// SessionCookieName is the canonical single session cookie name for pure server sessions.
	// In the browser, this cookie is expected to be marked HttpOnly, Secure, and SameSite.
	SessionCookieName = "ps"

	// SessionCookieMaxAge is the browser cookie lifetime in seconds (30 days).
	SessionCookieMaxAge = 30 * 24 * 3600

	// sessionContextKey is the single Gin context key populated by the auth
	// middleware. Handlers read the session through GetSession (or the small
	// accessors below) instead of depending on duplicated primitive values.
	sessionContextKey = "session"
)

// SessionStore defines the contract for retrieving server-side session models
// by their unique session identifier string (e.g. backed by Redis or PostgreSQL).
type SessionStore interface {
	// GetSessionByID retrieves a UserSession by its unique identifier.
	// It returns the session pointer or an error if the lookup fails or the session does not exist.
	GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error)
}

// GetSession returns the server-side session stored by Authenticate or
// OptionalAuthenticate. OptionalAuthenticate exposes both authenticated and
// guest sessions; use GetUserID when only a real user is relevant.
func GetSession(c *gin.Context) (*model.UserSession, bool) {
	value, exists := c.Get(sessionContextKey)
	if !exists {
		return nil, false
	}

	session, ok := value.(*model.UserSession)
	if !ok || session == nil {
		return nil, false
	}
	return session, true
}

// GetUserID returns the authenticated user's ID from the session context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	session, ok := GetSession(c)
	if !ok || session.UserID == uuid.Nil {
		return uuid.Nil, false
	}
	return session.UserID, true
}

// GetUserRole returns the authenticated user's role from the session context.
// Ad-hoc sessions created without an explicit role retain the historical
// customer default.
func GetUserRole(c *gin.Context) (model.UserRole, bool) {
	session, ok := GetSession(c)
	if !ok || session.Type == model.SessionTypeGuest {
		return "", false
	}
	if session.UserRole == "" {
		return model.CustomerRole, true
	}
	return session.UserRole, true
}

// Authorize verifies that the authenticated user possesses one of the allowed
// roles. Role grants are explicit at the route layer: no role receives an
// implicit bypass here. If the session or role is missing, or does not match
// the allowlist, the request is aborted with a 401 Unauthorized or 403 Forbidden
// error response respectively.
func Authorize(allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := GetUserRole(c)
		if !ok {
			// No authenticated session in context -> user has not been authenticated
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		// Check if the user's role matches any of the explicitly allowed roles
		for _, allowed := range allowedRoles {
			if userRole == allowed {
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
func extractSessionID(c *gin.Context) (string, error) {
	cookie, err := c.Cookie(SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrMissingSessionCookie
		}
		return "", fmt.Errorf("read session cookie: %w", err)
	}
	if cookie == "" {
		return "", ErrMissingSessionCookie
	}
	return cookie, nil
}

// SetSessionCookie writes the canonical session cookie with the standard
// lifetime and HttpOnly/SameSite protections.
func SetSessionCookie(c *gin.Context, sessionID string, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, sessionID, SessionCookieMaxAge, "/", "", secure, true)
}

// ClearSessionCookie expires the canonical session cookie so a client holding a
// revoked or stale session stops presenting it. Attributes are omitted because
// a browser matches a deletion cookie by name, domain and path, not by its
// Secure/HttpOnly flags.
func ClearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}

// isAuthenticatedSession reports whether the session represents a real user.
// Guest sessions and sessions with a nil user ID are never allowed to populate
// authenticated context, preventing privilege escalation.
func isAuthenticatedSession(session *model.UserSession) bool {
	return session != nil &&
		session.Type != model.SessionTypeGuest &&
		session.UserID != uuid.Nil
}

// loadSession extracts the cookie and resolves any valid server-side session,
// including guest sessions.
func loadSession(c *gin.Context, sessionStore SessionStore) (*model.UserSession, error) {
	sessionID, err := extractSessionID(c)
	if err != nil {
		return nil, err
	}

	return sessionStore.GetSessionByID(c.Request.Context(), sessionID)
}

// loadAuthenticatedSession resolves a session and filters out guests. It
// returns (nil, nil) when the cookie points at a valid but anonymous session;
// callers decide whether that is acceptable.
func loadAuthenticatedSession(c *gin.Context, sessionStore SessionStore) (*model.UserSession, error) {
	session, err := loadSession(c, sessionStore)
	if err != nil {
		return nil, err
	}
	if !isAuthenticatedSession(session) {
		return nil, nil
	}

	return session, nil
}

// abortAuthentication normalizes session lookup failures into an HTTP error.
func abortAuthentication(c *gin.Context, err error) {
	var ae *apierr.APIError
	if errors.As(err, &ae) {
		_ = c.Error(err)
	} else {
		_ = c.Error(apierr.UnauthorizedError())
	}
	c.Abort()
}

// Authenticate provides pure server-side session authentication middleware.
// It looks up the session via the SessionStore and stores only the loaded
// session object in the Gin context. If the session cookie is missing, invalid,
// or expired, it halts the middleware chain and aborts with an authentication error.
// The session snapshot is trusted as the source of truth for role and status.
func Authenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := loadAuthenticatedSession(c, sessionStore)
		if err != nil {
			abortAuthentication(c, err)
			return
		}
		if session == nil {
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		c.Set(sessionContextKey, session)
		c.Next()
	}
}

// OptionalAuthenticate resolves the session cookie when present and stores the
// session context without aborting or returning 401 when the request is
// anonymous. Both authenticated and guest sessions are exposed so public
// routes (for example the cart) can work with either identity. Invalid or
// missing cookies are treated as anonymous.
func OptionalAuthenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := loadSession(c, sessionStore)
		if err != nil {
			c.Next()
			return
		}

		if session != nil {
			c.Set(sessionContextKey, session)
		}
		c.Next()
	}
}
