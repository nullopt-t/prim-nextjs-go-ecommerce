package auth

import (
	"context"
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
	sessionService   *SessionService
	logger           log.Logger
	redisClient      *redis.Client
	notifier         Notifier
	secrets          *config.Secrets
}

func NewAuthService(
	logger log.Logger,
	challengeService *ChallengeService,
	userService UserService,
	sessionService *SessionService,
	jwtService *jwt.JWTManager,
	redisClient *redis.Client,
	notifier Notifier,
	secrets *config.Secrets,
) *AuthService {
	return &AuthService{
		jwtService:       jwtService,
		challengeService: challengeService,
		userService:      userService,
		sessionService:   sessionService,
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
		return "", "", apierr.New(http.StatusUnauthorized, "Invalid refresh token").WithCode(apierr.CodeUnauthorized)
	}

	user, err := s.userService.GetUserByID(
		ctx,
		claims.UserID,
	)
	if err != nil {
		return "", "", err
	}

	if claims.SessionID != "" {
		session, err := s.sessionService.GetSessionByID(ctx, claims.SessionID)
		if err != nil {
			return "", "", apierr.New(http.StatusUnauthorized, "Session expired or revoked").WithCode(apierr.CodeUnauthorized)
		}
		
		if session.CurrentRefreshJTI != claims.ID {
			// Replay detected! Delete session to protect user.
			_ = s.sessionService.DeleteSessionByID(ctx, user.ID, claims.SessionID)
			s.logger.Warn("refresh token replay detected, session revoked", log.Meta{"user_id": user.ID.String(), "session_id": claims.SessionID})
			return "", "", apierr.New(http.StatusUnauthorized, "Invalid token replay detected").WithCode(apierr.CodeUnauthorized)
		}
	}

	newClaims := &jwt.UserClaims{
		UserID:    user.ID,
		UserRole:  (*string)(user.Role),
		SessionID: claims.SessionID,
	}

	accessToken, newRefreshToken, err := s.jwtService.GenerateTokenPair(
		newClaims,
	)
	if err != nil {
		s.logger.Error("failed to generate token pair during rotation", log.Meta{"error": err})
		return "", "", err
	}

	if claims.SessionID != "" {
		if err := s.sessionService.UpdateRefreshJTI(ctx, claims.SessionID, newClaims.ID); err != nil {
			s.logger.Error("failed to update session JTI", log.Meta{"error": err, "session_id": claims.SessionID})
		}
	}

	s.logger.Info("token rotated successfully", log.Meta{"user_id": user.ID.String()})

	return accessToken, newRefreshToken, nil
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
) (*model.Challenge, error) {
	challenge, err := s.challengeService.Create(
		ctx,
		identifier,
		"email",
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
		s.logger.Warn("invalid challenge verification attempt", log.Meta{"identifier": identifier, "ip_address": ipAddress})
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

	if user.Status == model.StatusDeleted || user.Status == model.StatusInactive {
		return nil, apierr.New(http.StatusForbidden, "Account is inactive").WithCode(apierr.CodeForbidden)
	}

	if user.Status == model.StatusSuspended {
		if user.SuspendedUntil == nil || time.Now().Before(*user.SuspendedUntil) {
			return nil, apierr.New(http.StatusForbidden, "Account is suspended").WithCode(apierr.CodeForbidden)
		}
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, apierr.New(http.StatusForbidden, "Account is temporarily locked").WithCode(apierr.CodeForbidden)
	}

	if err := s.challengeService.MarkVerified(
		ctx,
		challenge,
	); err != nil {
		return nil, err
	}

	sess, sErr := s.sessionService.CreateSession(ctx, user.ID, SessionTypeAuthenticated, userAgent, ipAddress)
	sessionID := ""
	if sErr == nil && sess != nil {
		sessionID = sess.ID
	} else if sErr != nil {
		s.logger.Error("failed to create session", log.Meta{"error": sErr, "user_id": user.ID.String()})
	}

	claims := &jwt.UserClaims{
		UserID:    user.ID,
		UserRole:  (*string)(user.Role),
		SessionID: sessionID,
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		claims,
	)
	if err != nil {
		return nil, err
	}

	if sessionID != "" {
		if err := s.sessionService.UpdateRefreshJTI(ctx, sessionID, claims.ID); err != nil {
			s.logger.Error("failed to update session JTI", log.Meta{"error": err, "session_id": sessionID})
		}
	}

	s.logger.Info("user successfully authenticated", log.Meta{
		"user_id":    user.ID.String(),
		"ip_address": ipAddress,
		"session_id": sessionID,
	})

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
		UserID:       user.ID,
	}, nil
}
