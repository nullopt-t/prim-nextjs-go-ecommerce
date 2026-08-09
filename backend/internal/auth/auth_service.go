package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserService interface {
	CreateUser(
		ctx context.Context,
		identifier string,
	) (*model.User, error)

	GetUserByIdentifier(
		ctx context.Context,
		identifier string,
	) (*model.User, error)

	GetUserByID(
		ctx context.Context,
		userID uuid.UUID,
	) (*model.User, error)
}

type Notifier interface {
	NotifyOTP(
		ctx context.Context,
		channel,
		identifier,
		otp string,
	) error
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	SessionID    string
	UserID       uuid.UUID
}

type AuthService struct {
	jwtService       *jwt.JWTManager
	challengeService *ChallengeService
	userService      UserService
	logger           log.Logger
	redisClient      *redis.Client
	notifier         Notifier
	secrets          *config.Secrets
}

func NewAuthService(
	logger log.Logger,
	challengeService *ChallengeService,
	userService UserService,
	jwtService *jwt.JWTManager,
	redisClient *redis.Client,
	notifier Notifier,
	secrets *config.Secrets,
) *AuthService {
	return &AuthService{
		jwtService:       jwtService,
		challengeService: challengeService,
		userService:      userService,
		logger:           logger,
		redisClient:      redisClient,
		notifier:         notifier,
		secrets:          secrets,
	}
}

func (s *AuthService) RotateToken(
	ctx context.Context,
	refreshToken string,
) (string, string, error) {
	claims, err := s.jwtService.VerifyToken(
		refreshToken,
		s.secrets.JwtRefreshTokenSecretKey,
	)
	if err != nil {
		return "", "", err
	}

	user, err := s.userService.GetUserByID(
		ctx,
		claims.UserID,
	)
	if err != nil {
		return "", "", err
	}

	newClaims := &jwt.UserClaims{
		UserID:   user.ID,
		UserRole: (*string)(user.Role),
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		newClaims,
	)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.User, error) {
	return s.userService.GetUserByID(ctx, userID)
}

func (s *AuthService) StartChallenge(
	ctx context.Context,
	identifier string,
	identifierType IdentifierType,
) (*model.Challenge, error) {
	var channel string
	switch identifierType {
	case Email:
		channel = "email"
	case Phone:
		channel = "sms"
	}
	challenge, err := s.challengeService.Create(
		ctx,
		identifier,
		channel,
	)
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

func (s *AuthService) ResendChallenge(
	ctx context.Context,
	identifier string,
) error {
	challenge, err := s.challengeService.Get(
		ctx,
		identifier,
	)
	if err != nil {
		return err
	}

	err = s.challengeService.Resend(
		ctx,
		challenge,
	)

	return err
}

func (s *AuthService) VerifyChallenge(
	ctx context.Context,
	identifier,
	code string,
	userAgent string,
	ipAddress string,
) (*Tokens, error) {
	challenge, err := s.challengeService.Get(
		ctx,
		identifier,
	)
	if err != nil {
		return nil, err
	}

	ok, err := s.challengeService.Verify(
		ctx,
		challenge,
		code,
	)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, apierr.New(
			http.StatusUnauthorized,
			"Unauthorized",
		).WithCode(apierr.CodeUnauthorized)
	}

	user, err := s.userService.GetUserByIdentifier(
		ctx,
		challenge.Identifier,
	)
	if err != nil {
		user, err = s.userService.CreateUser(
			ctx,
			challenge.Identifier,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := s.challengeService.MarkVerified(
		ctx,
		challenge,
	); err != nil {
		return nil, err
	}

	claims := &jwt.UserClaims{
		UserID:   user.ID,
		UserRole: (*string)(user.Role),
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		claims,
	)
	if err != nil {
		return nil, err
	}

	sess, sErr := s.CreateSession(ctx, user.ID, SessionTypeAuthenticated, userAgent, ipAddress)
	sessionID := ""
	if sErr == nil && sess != nil {
		sessionID = sess.ID
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
		UserID:       user.ID,
	}, nil
}

type SessionType string

const (
	SessionTypeGuest         SessionType = "guest"
	SessionTypeAuthenticated SessionType = "authenticated"
)

type UserSession struct {
	ID         string      `json:"id"`
	UserID     uuid.UUID   `json:"user_id"`
	Type       SessionType `json:"type"`
	UserAgent  string      `json:"user_agent"`
	IPAddress  string      `json:"ip_address"`
	CreatedAt  time.Time   `json:"created_at"`
	LastActive time.Time   `json:"last_active"`
}

func (s *AuthService) CreateSession(
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

	// Store session JSON with 30-day TTL in Redis
	if err := s.redisClient.Set(ctx, sessKey, data, 30*24*time.Hour).Err(); err != nil {
		return nil, err
	}

	// Add session ID to user's Redis session set
	if err := s.redisClient.SAdd(ctx, userSetKey, sessionID).Err(); err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *AuthService) GetUserSessions(
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

func (s *AuthService) DeleteSessionByID(
	ctx context.Context,
	userID uuid.UUID,
	sessionID string,
) error {
	sessKey := fmt.Sprintf("auth:session:%s", sessionID)
	userSetKey := fmt.Sprintf("auth:user_sessions:%s", userID.String())

	if err := s.redisClient.Del(ctx, sessKey).Err(); err != nil {
		return err
	}
	if err := s.redisClient.SRem(ctx, userSetKey, sessionID).Err(); err != nil {
		return err
	}
	return nil
}
