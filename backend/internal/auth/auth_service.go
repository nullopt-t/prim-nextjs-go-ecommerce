package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/log"
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

type SessionAuthResult struct {
	SessionID string
	UserID    uuid.UUID
}

type AuthService struct {
	challengeService *ChallengeService
	userService      UserService
	sessionService   *SessionService
	logger           log.Logger
	notifier         Notifier
}

func NewAuthService(
	logger log.Logger,
	challengeService *ChallengeService,
	userService UserService,
	sessionService *SessionService,
	notifier Notifier,
) *AuthService {
	return &AuthService{
		challengeService: challengeService,
		userService:      userService,
		sessionService:   sessionService,
		logger:           logger,
		notifier:         notifier,
	}
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
) (*SessionAuthResult, error) {
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

	role := "customer"
	if user.Role != nil {
		role = string(*user.Role)
	}

	sess, sErr := s.sessionService.CreateSession(ctx, user.ID, role, SessionTypeAuthenticated, userAgent, ipAddress)
	sessionID := ""
	if sErr == nil && sess != nil {
		sessionID = sess.ID
	} else if sErr != nil {
		s.logger.Error("failed to create session", log.Meta{"error": sErr, "user_id": user.ID.String()})
	}

	s.logger.Info("user successfully authenticated", log.Meta{
		"user_id":    user.ID.String(),
		"ip_address": ipAddress,
		"session_id": sessionID,
	})

	return &SessionAuthResult{
		SessionID: sessionID,
		UserID:    user.ID,
	}, nil
}
