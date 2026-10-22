package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/redis/go-redis/v9"
)

// SessionType defines the authorization scope classification for client sessions (guest or authenticated).
type SessionType = model.SessionType

const (
	// SessionTypeGuest indicates an unauthenticated anonymous browser session (e.g. guest shopping cart).
	SessionTypeGuest = model.SessionTypeGuest
	// SessionTypeAuthenticated indicates a verified customer or staff session bound to a specific user account.
	SessionTypeAuthenticated = model.SessionTypeAuthenticated

	sessionKeyPrefix        = "auth:session:"
	userSessionSetKeyPrefix = "auth:user_sessions:"
)

// UserSession alias to internal/model.UserSession representing session payload stored in Redis.
type UserSession = model.UserSession

// SessionService manages server-side user session lifecycles in Redis:
// creation, activity tracking, sliding window auto-refresh, active session listing,
// single session revocation, and global user session invalidation.
type SessionService struct {
	redisClient *redis.Client
	logger      log.Logger
	sessionTTL  time.Duration
}

// NewSessionService constructs a new SessionService with Redis client and session lifetime configuration.
func NewSessionService(
	redisClient *redis.Client,
	logger log.Logger,
	sessionTTL time.Duration,
) *SessionService {
	return &SessionService{
		redisClient: redisClient,
		logger:      logger,
		sessionTTL:  sessionTTL,
	}
}

// CreateSession creates a new server-side session object in Redis.
// It assigns a secure random session identifier with prefix "sess_", serializes session metadata,
// stores the session JSON with a TTL, and registers the session ID in the user's active session set.
func (s *SessionService) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	userRole string,
	sessionType SessionType,
	userAgent string,
	ipAddress string,
) (*UserSession, error) {
	sess, err := newUserSession(userID, userRole, sessionType, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	if err := s.storeSession(ctx, sess); err != nil {
		return nil, err
	}
	if err := s.registerUserSession(ctx, sess.UserID, sess.ID); err != nil {
		return nil, err
	}

	return sess, nil
}

// GetUserSessions lists all active sessions registered for the given user UUID.
// It iterates through the user's Redis set, fetches session data, prunes expired session IDs (lazy cleanup),
// and returns all currently active sessions.
func (s *SessionService) GetUserSessions(
	ctx context.Context,
	userID uuid.UUID,
) ([]UserSession, error) {
	userSetKey := userSessionSetKey(userID)
	sessionIDs, err := s.userSessionIDs(ctx, userSetKey)
	if err != nil {
		return nil, err
	}

	sessions := make([]UserSession, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		sess, err := s.readSession(ctx, sessionID)
		if err == redis.Nil {
			s.pruneExpiredUserSessionID(ctx, userSetKey, sessionID)
			continue
		}
		if err != nil {
			continue
		}
		sessions = append(sessions, *sess)
	}

	return sessions, nil
}

// GetSessionByID fetches session metadata from Redis by session ID.
// If the session has been active for more than half its TTL, it automatically refreshes
// LastActive and extends the Redis TTL (sliding expiration window).
func (s *SessionService) GetSessionByID(
	ctx context.Context,
	sessionID string,
) (*UserSession, error) {
	sess, err := s.readSession(ctx, sessionID)
	if err == redis.Nil {
		return nil, apierr.UnauthorizedError("Session expired or invalid")
	}
	if err != nil {
		return nil, err
	}

	s.rotateSession(ctx, sess)
	s.refreshUserSessionSetTTL(ctx, sess.UserID)

	return sess, nil
}

// DeleteSessionByID deletes a specific active session key from Redis and removes it from the user's session set.
// Enforces ownership check to prevent IDOR attacks.
func (s *SessionService) DeleteSessionByID(
	ctx context.Context,
	userID uuid.UUID,
	sessionID string,
) error {
	userSetKey := userSessionSetKey(userID)

	if err := s.verifySessionOwnership(ctx, userSetKey, userID, sessionID); err != nil {
		return err
	}
	if err := s.deleteSessionKey(ctx, sessionID); err != nil {
		return err
	}
	if err := s.removeSessionFromUserSet(ctx, userSetKey, sessionID); err != nil {
		return err
	}

	s.logger.Info("session successfully deleted", log.Meta{"user_id": userID.String(), "session_id": sessionID})
	return nil
}

// DeleteAllSessions invalidates all active sessions for the given user UUID by deleting all session keys
// and removing the user's session tracking set from Redis in a single atomic pipeline.
func (s *SessionService) DeleteAllSessions(
	ctx context.Context,
	userID uuid.UUID,
) error {
	userSetKey := userSessionSetKey(userID)

	sessionIDs, err := s.userSessionIDs(ctx, userSetKey)
	if err != nil {
		s.logger.Error("failed to get user sessions for deletion", log.Meta{"error": err, "user_id": userID.String()})
		return err
	}

	if err := s.deleteUserSessions(ctx, userSetKey, sessionIDs); err != nil {
		s.logger.Error("failed to execute pipeline deleting all user sessions", log.Meta{"error": err, "user_id": userID.String()})
		return err
	}

	s.logger.Info("all user sessions successfully deleted", log.Meta{"user_id": userID.String(), "count": len(sessionIDs)})
	return nil
}

// newUserSession builds and validates a session before it is persisted.
func newUserSession(
	userID uuid.UUID,
	userRole string,
	sessionType SessionType,
	userAgent string,
	ipAddress string,
) (*UserSession, error) {
	now := time.Now().UTC()
	sess := &UserSession{
		ID:         fmt.Sprintf("sess_%s", uuid.New().String()),
		UserID:     userID,
		UserRole:   model.UserRole(userRole),
		Type:       sessionType,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		CreatedAt:  now,
		LastActive: now,
	}
	if err := sess.Validate(); err != nil {
		return nil, err
	}
	return sess, nil
}

// storeSession serializes and persists a session under its Redis key.
func (s *SessionService) storeSession(ctx context.Context, sess *UserSession) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return s.redisClient.Set(ctx, sessionKey(sess.ID), data, s.sessionTTL).Err()
}

// registerUserSession adds an authenticated user's session ID to their tracking set.
func (s *SessionService) registerUserSession(ctx context.Context, userID uuid.UUID, sessionID string) error {
	if userID == uuid.Nil {
		return nil
	}

	userSetKey := userSessionSetKey(userID)
	if err := s.redisClient.SAdd(ctx, userSetKey, sessionID).Err(); err != nil {
		return err
	}

	// Set TTL on the set to avoid accumulation of dead session keys.
	_ = s.redisClient.Expire(ctx, userSetKey, s.sessionTTL*2).Err()
	return nil
}

// readSession fetches and decodes one session without applying sliding-expiry logic.
func (s *SessionService) readSession(ctx context.Context, sessionID string) (*UserSession, error) {
	val, err := s.redisClient.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		return nil, err
	}

	var sess UserSession
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// rotateSession extends the sliding expiration window when the session has been
// idle long enough. Persistence failures preserve the previous behavior: the
// session is still returned to the caller, and the next request can retry.
func (s *SessionService) rotateSession(ctx context.Context, sess *UserSession) {
	now := time.Now().UTC()
	if now.Sub(sess.LastActive) < s.rotationThreshold() {
		return
	}

	sess.LastActive = now
	data, err := json.Marshal(sess)
	if err != nil {
		return
	}
	_ = s.redisClient.Set(ctx, sessionKey(sess.ID), data, s.sessionTTL).Err()
}

// rotationThreshold limits how frequently a session is written back to Redis.
func (s *SessionService) rotationThreshold() time.Duration {
	threshold := s.sessionTTL / 2
	switch {
	case threshold < time.Minute:
		return 10 * time.Second
	case threshold < time.Hour && s.sessionTTL >= time.Hour:
		return time.Hour
	default:
		return threshold
	}
}

// refreshUserSessionSetTTL keeps the user's tracking set alive for as long as
// any of its individual session keys can remain valid. Without this refresh the
// set expires first, making DeleteAllSessions and DeleteSessionByID silent no-ops
// for long-lived sessions.
func (s *SessionService) refreshUserSessionSetTTL(ctx context.Context, userID uuid.UUID) {
	if userID == uuid.Nil {
		return
	}
	_ = s.redisClient.Expire(ctx, userSessionSetKey(userID), s.sessionTTL*2).Err()
}

// userSessionIDs returns the session IDs tracked for a user.
func (s *SessionService) userSessionIDs(ctx context.Context, userSetKey string) ([]string, error) {
	return s.redisClient.SMembers(ctx, userSetKey).Result()
}

// pruneExpiredUserSessionID removes a session ID whose session key has expired.
func (s *SessionService) pruneExpiredUserSessionID(ctx context.Context, userSetKey, sessionID string) {
	s.redisClient.SRem(ctx, userSetKey, sessionID)
}

// verifySessionOwnership rejects revocation attempts for sessions the user does not own.
func (s *SessionService) verifySessionOwnership(
	ctx context.Context,
	userSetKey string,
	userID uuid.UUID,
	sessionID string,
) error {
	isMember, err := s.redisClient.SIsMember(ctx, userSetKey, sessionID).Result()
	if err != nil {
		s.logger.Error("failed to verify session ownership", log.Meta{"error": err, "session_id": sessionID, "user_id": userID.String()})
		return err
	}
	if !isMember {
		return apierr.New(http.StatusNotFound, "Session not found").WithCode(apierr.CodeNotFound)
	}
	return nil
}

// deleteSessionKey removes the session payload from Redis.
func (s *SessionService) deleteSessionKey(ctx context.Context, sessionID string) error {
	if err := s.redisClient.Del(ctx, sessionKey(sessionID)).Err(); err != nil {
		s.logger.Error("failed to delete session key", log.Meta{"error": err, "session_id": sessionID})
		return err
	}
	return nil
}

// removeSessionFromUserSet unlinks a revoked session from its user's tracking set.
func (s *SessionService) removeSessionFromUserSet(ctx context.Context, userSetKey, sessionID string) error {
	if err := s.redisClient.SRem(ctx, userSetKey, sessionID).Err(); err != nil {
		s.logger.Error("failed to remove session from user set", log.Meta{"error": err, "session_id": sessionID})
		return err
	}
	return nil
}

// deleteUserSessions atomically deletes all session payloads and the tracking set.
func (s *SessionService) deleteUserSessions(ctx context.Context, userSetKey string, sessionIDs []string) error {
	pipe := s.redisClient.Pipeline()
	for _, sessionID := range sessionIDs {
		pipe.Del(ctx, sessionKey(sessionID))
	}
	pipe.Del(ctx, userSetKey)
	_, err := pipe.Exec(ctx)
	return err
}

func sessionKey(sessionID string) string {
	return sessionKeyPrefix + sessionID
}

func userSessionSetKey(userID uuid.UUID) string {
	return userSessionSetKeyPrefix + userID.String()
}
