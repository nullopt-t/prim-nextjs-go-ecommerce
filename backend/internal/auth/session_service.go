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
	// Generate unique session identifier with standard prefix
	sessionID := fmt.Sprintf("sess_%s", uuid.New().String())
	now := time.Now().UTC()

	sess := &UserSession{
		ID:         sessionID,
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

	data, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	sessKey := fmt.Sprintf("auth:session:%s", sessionID)
	userSetKey := fmt.Sprintf("auth:user_sessions:%s", userID.String())

	// Store session JSON with TTL in Redis
	if err := s.redisClient.Set(ctx, sessKey, data, s.sessionTTL).Err(); err != nil {
		return nil, err
	}

	// Add session ID to user's Redis session set if authenticated
	if userID != uuid.Nil {
		if err := s.redisClient.SAdd(ctx, userSetKey, sessionID).Err(); err != nil {
			return nil, err
		}
		// Set TTL on userSetKey to avoid accumulation of dead session keys
		_ = s.redisClient.Expire(ctx, userSetKey, s.sessionTTL*2).Err()
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
	userSetKey := fmt.Sprintf("auth:user_sessions:%s", userID.String())
	sessionIDs, err := s.redisClient.SMembers(ctx, userSetKey).Result()
	if err != nil {
		return nil, err
	}

	sessions := make([]UserSession, 0, len(sessionIDs))
	for _, sessID := range sessionIDs {
		sessKey := fmt.Sprintf("auth:session:%s", sessID)
		val, err := s.redisClient.Get(ctx, sessKey).Result()
		if err == redis.Nil {
			// Clean up expired session ID from user set
			s.redisClient.SRem(ctx, userSetKey, sessID)
			continue
		} else if err != nil {
			continue
		}

		var sess UserSession
		if err := json.Unmarshal([]byte(val), &sess); err == nil {
			sessions = append(sessions, sess)
		}
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
	sessKey := fmt.Sprintf("auth:session:%s", sessionID)
	val, err := s.redisClient.Get(ctx, sessKey).Result()
	if err == redis.Nil {
		return nil, apierr.New(http.StatusUnauthorized, "Session expired or invalid").WithCode(apierr.CodeUnauthorized)
	}
	if err != nil {
		return nil, err
	}

	var sess UserSession
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, err
	}

	// Auto-refresh session if last activity was more than half the session TTL ago (or at least 1 hour ago)
	refreshThreshold := s.sessionTTL / 2
	if refreshThreshold < time.Hour && s.sessionTTL >= time.Hour {
		refreshThreshold = time.Hour
	} else if refreshThreshold < time.Minute {
		refreshThreshold = 10 * time.Second
	}

	now := time.Now().UTC()
	if now.Sub(sess.LastActive) >= refreshThreshold {
		sess.LastActive = now
		if updatedData, err := json.Marshal(sess); err == nil {
			_ = s.redisClient.Set(ctx, sessKey, updatedData, s.sessionTTL).Err()
		}
	}

	// Refresh the user's session-set TTL on access. The set is only assigned a
	// TTL once at creation (sessionTTL*2); without this refresh it expires while
	// individually sliding session keys remain valid, making DeleteAllSessions
	// and DeleteSessionByID silent no-ops for long-lived sessions.
	if sess.UserID != uuid.Nil {
		userSetKey := fmt.Sprintf("auth:user_sessions:%s", sess.UserID.String())
		_ = s.redisClient.Expire(ctx, userSetKey, s.sessionTTL*2).Err()
	}

	return &sess, nil
}

// DeleteSessionByID deletes a specific active session key from Redis and removes it from the user's session set.
// Enforces ownership check to prevent IDOR attacks.
func (s *SessionService) DeleteSessionByID(
	ctx context.Context,
	userID uuid.UUID,
	sessionID string,
) error {
	userSetKey := fmt.Sprintf("auth:user_sessions:%s", userID.String())

	// IDOR check: verify that the session actually belongs to this user
	isMember, err := s.redisClient.SIsMember(ctx, userSetKey, sessionID).Result()
	if err != nil {
		s.logger.Error("failed to verify session ownership", log.Meta{"error": err, "session_id": sessionID, "user_id": userID.String()})
		return err
	}
	if !isMember {
		return apierr.New(http.StatusNotFound, "Session not found").WithCode(apierr.CodeNotFound)
	}

	sessKey := fmt.Sprintf("auth:session:%s", sessionID)

	if err := s.redisClient.Del(ctx, sessKey).Err(); err != nil {
		s.logger.Error("failed to delete session key", log.Meta{"error": err, "session_id": sessionID})
		return err
	}
	if err := s.redisClient.SRem(ctx, userSetKey, sessionID).Err(); err != nil {
		s.logger.Error("failed to remove session from user set", log.Meta{"error": err, "session_id": sessionID})
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
	userSetKey := fmt.Sprintf("auth:user_sessions:%s", userID.String())

	sessionIDs, err := s.redisClient.SMembers(ctx, userSetKey).Result()
	if err != nil {
		s.logger.Error("failed to get user sessions for deletion", log.Meta{"error": err, "user_id": userID.String()})
		return err
	}

	pipe := s.redisClient.Pipeline()
	for _, sid := range sessionIDs {
		sessKey := fmt.Sprintf("auth:session:%s", sid)
		pipe.Del(ctx, sessKey)
	}
	pipe.Del(ctx, userSetKey)

	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("failed to execute pipeline deleting all user sessions", log.Meta{"error": err, "user_id": userID.String()})
		return err
	}

	s.logger.Info("all user sessions successfully deleted", log.Meta{"user_id": userID.String(), "count": len(sessionIDs)})
	return nil
}
