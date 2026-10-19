package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

	// ErrStaleSession indicates that a session is structurally valid but points
	// at a user account that no longer exists (deleted, or removed out-of-band
	// by a database reseed). It is treated as an authentication failure so the
	// client re-authenticates instead of surfacing a downstream 500.
	ErrStaleSession = errors.New("session references a non-existent user")
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
			switch {
			case errors.Is(err, ErrStaleSession):
				// The session is well-formed but its user is gone. Treat it as an
				// authentication failure and clear the cookie so the client
				// re-authenticates instead of retrying a dead session.
				clearSessionCookie(c)
				_ = c.Error(apierr.UnauthorizedError("Session is no longer valid").
					WithCode(apierr.CodeUnauthorized))
			default:
				var ae *apierr.APIError
				if errors.As(err, &ae) {
					_ = c.Error(err)
				} else {
					_ = c.Error(apierr.UnauthorizedError())
				}
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

// OptionalAuthenticate attempts to authenticate via the session cookie if present,
// setting user context without aborting or returning 401 if unauthenticated or absent.
// This is used on public routes that can provide enriched responses when logged in.
func OptionalAuthenticate(sessionStore SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := extractSessionID(c)

		if sessionID != "" {
			session, err := sessionStore.GetSessionByID(c.Request.Context(), sessionID)
			if err != nil && errors.Is(err, ErrStaleSession) {
				// The cookie points at a session whose user is gone. Drop it so the
				// client stops presenting a dead credential, then continue as an
				// anonymous visitor (this middleware never aborts).
				clearSessionCookie(c)
			}
			if err == nil && session != nil {
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
