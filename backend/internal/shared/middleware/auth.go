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

	// ErrStaleSession indicates that a session is structurally valid but points
	// at a user account that no longer exists (deleted, or removed out-of-band
	// by a database reseed). It is treated as an authentication failure so the
	// client re-authenticates instead of surfacing a downstream 500.
	ErrStaleSession = errors.New("session references a non-existent user")

	// ErrMissingSessionCookie indicates that the request did not carry the
	// canonical session cookie. Callers decide whether that is fatal (authenticated
	// routes) or simply means the request is anonymous (optional routes).
	ErrMissingSessionCookie = errors.New("session cookie is missing")
)

// UserValidator reports whether a session's principal still exists. It is used
// to reject a stale session (for example after a database reseed) before any
// handler tries to write a row that references the missing user.
type UserValidator interface {
	// UserExists reports whether the user still exists and is not soft-deleted.
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
}

// validatingSessionStore wraps a SessionStore and rejects authenticated sessions
// whose user no longer exists. Guest sessions and sessions with a nil user are
// passed through untouched. Infrastructure failures while checking existence are
// surfaced as errors (mapped to 5xx), never silently accepted.
type validatingSessionStore struct {
	store     SessionStore
	validator UserValidator
}

// NewValidatingSessionStore decorates store with user-existence validation. If
// either argument is nil the original store is returned unchanged, so callers
// that have no user validator (tests, minimal deployments) are unaffected.
func NewValidatingSessionStore(store SessionStore, validator UserValidator) SessionStore {
	if store == nil || validator == nil {
		return store
	}
	return &validatingSessionStore{store: store, validator: validator}
}

// GetSessionByID loads the session and, for authenticated sessions, verifies the
// referenced user still exists. A missing user yields ErrStaleSession; a lookup
// failure yields a 5xx API error so it is not mistaken for a bad credential.
func (v *validatingSessionStore) GetSessionByID(ctx context.Context, sessionID string) (*model.UserSession, error) {
	session, err := v.store.GetSessionByID(ctx, sessionID)
	if err != nil || session == nil {
		return session, err
	}

	// Anonymous sessions carry no principal to validate.
	if session.Type == model.SessionTypeGuest || session.UserID == uuid.Nil {
		return session, nil
	}

	exists, err := v.validator.UserExists(ctx, session.UserID)
	if err != nil {
		return nil, apierr.InternalServerError("Failed to validate session").Wrap(err)
	}
	if !exists {
		return nil, fmt.Errorf("%w: user %s", ErrStaleSession, session.UserID)
	}

	return session, nil
}

const (
	// SessionCookieName is the canonical single session cookie name for pure server sessions.
	// In the browser, this cookie is expected to be marked HttpOnly, Secure, and SameSite.
	SessionCookieName = "ps"

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

// GetSession returns the authenticated session stored by Authenticate or
// OptionalAuthenticate. The boolean is false when the request is anonymous.
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
	if !ok {
		return "", false
	}
	if session.UserRole == "" {
		return model.CustomerRole, true
	}
	return session.UserRole, true
}

// Authorize verifies that the authenticated user possesses one of the allowed roles.
// Users with SuperRole ("superadmin") automatically pass any staff-level check.
// If the session or role is missing, or does not match the allowlist, the request
// is aborted with a 401 Unauthorized or 403 Forbidden error response respectively.
func Authorize(allowedRoles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := GetUserRole(c)
		if !ok {
			// No authenticated session in context -> user has not been authenticated
			_ = c.Error(apierr.UnauthorizedError())
			c.Abort()
			return
		}

		// Superadmin has universal clearance across administrative checkpoints
		if userRole == model.SuperRole {
			c.Next()
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

// clearSessionCookie expires the canonical session cookie (and any legacy auth
// cookies) so a client holding a revoked or stale session stops presenting it.
// Attributes are omitted because a browser matches a deletion cookie by name,
// domain and path, not by its Secure/HttpOnly flags.
func clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}

// isAuthenticatedSession reports whether the session represents a real user.
// Guest sessions and sessions with a nil user ID are never allowed to populate
// authenticated context, preventing privilege escalation.
func isAuthenticatedSession(session *model.UserSession) bool {
	return session != nil &&
		session.Type != model.SessionTypeGuest &&
		session.UserID != uuid.Nil
}

// loadAuthenticatedSession extracts the cookie, fetches the session, and applies
// all authentication validation in one place. It returns (nil, nil) when the
// cookie points at a valid but anonymous session; callers decide whether that
// is acceptable.
func loadAuthenticatedSession(c *gin.Context, sessionStore SessionStore) (*model.UserSession, error) {
	sessionID, err := extractSessionID(c)
	if err != nil {
		return nil, err
	}

	session, err := sessionStore.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		return nil, err
	}
	if !isAuthenticatedSession(session) {
		return nil, nil
	}

	return session, nil
}

// abortAuthentication normalizes session lookup failures into an HTTP error.
// Stale sessions are distinct: their cookie is cleared so the client does not
// keep retrying a session whose user has been removed.
func abortAuthentication(c *gin.Context, err error) {
	if errors.Is(err, ErrStaleSession) {
		clearSessionCookie(c)
		_ = c.Error(apierr.UnauthorizedError("Session is no longer valid").
			WithCode(apierr.CodeUnauthorized))
	} else {
		var ae *apierr.APIError
		if errors.As(err, &ae) {
			_ = c.Error(err)
		} else {
			_ = c.Error(apierr.UnauthorizedError())
		}
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

// OptionalAuthenticate attempts to authenticate via the session cookie if present,
// setting the session context without aborting or returning 401 when the request
// is anonymous. Stale sessions are cleared and treated as guests so public routes
// remain usable. This is used on public routes that can provide enriched
// responses when logged in.
func OptionalAuthenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := loadAuthenticatedSession(c, sessionStore)
		if err != nil {
			// A stale cookie should not be sent back on every subsequent request.
			// Other lookup failures (including infrastructure errors and missing
			// cookies) must not break public routes.
			if errors.Is(err, ErrStaleSession) {
				clearSessionCookie(c)
			}
			c.Next()
			return
		}

		if session != nil {
			c.Set(sessionContextKey, session)
		}
		c.Next()
	}
}
