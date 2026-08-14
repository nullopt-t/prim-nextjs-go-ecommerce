package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/redis/go-redis/v9"
)

type SessionType string

const (
	SessionTypeGuest         SessionType = "guest"
	SessionTypeAuthenticated SessionType = "authenticated"
)

type UserSession struct {
	ID                string      `json:"id"`
	UserID            uuid.UUID   `json:"user_id"`
	Type              SessionType `json:"type"`
	CurrentRefreshJTI string      `json:"current_refresh_jti"`
	UserAgent         string      `json:"user_agent"`
	IPAddress         string      `json:"ip_address"`
	CreatedAt         time.Time   `json:"created_at"`
	LastActive        time.Time   `json:"last_active"`
}

type SessionService struct {
	redisClient *redis.Client
	logger      log.Logger
	sessionTTL  time.Duration
}

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

func (s *SessionService) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	sessionType SessionType,
	userAgent string,
	ipAddress string,
) (*UserSession, error) {
	sessionID := fmt.Sprintf("sess_%s", uuid.New().String())
	now := time.Now().UTC()

	sess := &UserSession{
		ID:         sessionID,
		UserID:     userID,
		Type:       sessionType,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		CreatedAt:  now,
		LastActive: now,
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

	// Add session ID to user's Redis session set
	if err := s.redisClient.SAdd(ctx, userSetKey, sessionID).Err(); err != nil {
		return nil, err
	}

	return sess, nil
}

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
	return &sess, nil
}

func (s *SessionService) UpdateRefreshJTI(
	ctx context.Context,
	sessionID string,
	jti string,
) error {
	sess, err := s.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	sess.CurrentRefreshJTI = jti
	sess.LastActive = time.Now().UTC()
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	sessKey := fmt.Sprintf("auth:session:%s", sessionID)
	// Refresh TTL as well
	return s.redisClient.Set(ctx, sessKey, data, s.sessionTTL).Err()
}

func (s *SessionService) DeleteSessionByID(
	ctx context.Context,
	userID uuid.UUID,
	sessionID string,
) error {
	sessKey := fmt.Sprintf("auth:session:%s", sessionID)
	userSetKey := fmt.Sprintf("auth:user_sessions:%s", userID.String())

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
