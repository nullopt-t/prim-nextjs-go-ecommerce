package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/model"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/database"
	"github.com/nullopt-t/prim-backend/pkg/log"
)

// UserService defines the account persistence and user identity contract
// required by the authentication domain for identity retrieval and JIT (just-in-time) user provisioning.
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

// Notifier defines the external dispatch contract for delivering one-time passwords (OTPs)
// to users across supported communication channels (e.g., SMTP email, SMS gateways).
type Notifier interface {
	NotifyOTP(
		ctx context.Context,
		channel,
		identifier,
		otp string,
	) error
}

// SessionAuthResult encapsulates the resulting session credentials and verified user ID
// upon successful completion of a challenge-response authentication handshake.
type SessionAuthResult struct {
	SessionID string
	UserID    uuid.UUID
}

// AuthService encapsulates domain business logic for passwordless authentication flows,
// challenge initiation, OTP verification, user status validation (active/suspended/locked),
// and authenticated session establishment.
type AuthService struct {
	challengeService *ChallengeService
	userService      UserService
	sessionService   *SessionService
	logger           log.Logger
	notifier         Notifier
}

// NewAuthService constructs a new AuthService domain component with required dependencies.
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

// GetUserByID retrieves a user record by its primary UUID key from the underlying user service.
func (s *AuthService) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.User, error) {
	return s.userService.GetUserByID(ctx, userID)
}

// StartChallenge initiates a new passwordless OTP verification challenge for the given user identifier (email).
// It generates a cryptographically secure 6-digit OTP, securely hashes it before Redis storage,
// sets an expiration TTL, and dispatches the raw OTP via the configured notifier channel.
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

// ResendChallenge regenerates and re-dispatches an OTP code for an active pending challenge,
// enforcing strict rate limits on maximum resend attempts to prevent communication flooding and spam.
func (s *AuthService) ResendChallenge(
	ctx context.Context,
	identifier string,
) error {
	// Fetch active challenge from Redis
	challenge, err := s.challengeService.Get(
		ctx,
		identifier,
	)
	if err != nil {
		return err
	}

	// Reissue and transmit new OTP
	err = s.challengeService.Resend(
		ctx,
		challenge,
	)

	return err
}

// VerifyChallenge completes the challenge-response authentication state machine:
// 1. Fetches challenge from Redis and verifies the submitted OTP against the stored cryptographic hash.
// 2. Enforces attempt caps to prevent OTP brute-forcing.
// 3. Resolves or just-in-time (JIT) provisions the user account in PostgreSQL.
// 4. Verifies account status (rejects deleted, inactive, suspended, or locked accounts).
// 5. Atomically transitions the challenge state to 'verified' in Redis to prevent replay attacks.
// 6. Creates a server-side session in Redis with IP and User-Agent binding.
func (s *AuthService) VerifyChallenge(
	ctx context.Context,
	identifier,
	code string,
	userAgent string,
	ipAddress string,
) (*SessionAuthResult, error) {
	// Retrieve active challenge from Redis
	challenge, err := s.challengeService.Get(
		ctx,
		identifier,
	)
	if err != nil {
		return nil, err
	}

	// Verify submitted OTP against hashed secret and check maximum attempt limits
	ok, err := s.challengeService.Verify(
		ctx,
		challenge,
		code,
	)
	if err != nil {
		return nil, err
	}

	// Unauthorized if code is invalid or maximum verification attempts exceeded
	if !ok {
		s.logger.Warn("invalid challenge verification attempt", log.Meta{"identifier": identifier, "ip_address": ipAddress})
		return nil, apierr.New(
			http.StatusUnauthorized,
			"Unauthorized",
		).WithCode(apierr.CodeUnauthorized)
	}

	// Resolve existing user account or provision a new customer account (Just-In-Time signup)
	user, err := s.userService.GetUserByIdentifier(
		ctx,
		challenge.Identifier,
	)
	if err != nil {
		var ae *apierr.APIError
		isNotFound := errors.Is(err, database.ErrNotFound) || (errors.As(err, &ae) && (ae.Status == http.StatusNotFound || ae.Code == apierr.CodeNotFound))
		if !isNotFound {
			return nil, err
		}

		user, err = s.userService.CreateUser(
			ctx,
			challenge.Identifier,
		)
		if err != nil {
			return nil, err
		}
	}

	// Enforce lifecycle access controls: inactive and soft-deleted accounts are forbidden
	if user.Status == model.StatusDeleted || user.Status == model.StatusInactive {
		return nil, apierr.New(http.StatusForbidden, "Account is inactive").WithCode(apierr.CodeForbidden)
	}

	// Enforce suspension window: permanently suspended or active suspension period blocks authentication
	if user.Status == model.StatusSuspended {
		if user.SuspendedUntil == nil || time.Now().Before(*user.SuspendedUntil) {
			return nil, apierr.New(http.StatusForbidden, "Account is suspended").WithCode(apierr.CodeForbidden)
		}
	}

	// Enforce temporary security lock window (e.g. from failed security events)
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, apierr.New(http.StatusForbidden, "Account is temporarily locked").WithCode(apierr.CodeForbidden)
	}

	// Atomically mark the challenge as verified in Redis via Lua script to prevent OTP reuse
	if err := s.challengeService.MarkVerified(
		ctx,
		challenge,
	); err != nil {
		return nil, err
	}

	// Resolve default role for session token context
	role := "customer"
	if user.Role != "" {
		role = string(user.Role)
	}

	// Establish server-side session in Redis
	sess, sErr := s.sessionService.CreateSession(ctx, user.ID, role, SessionTypeAuthenticated, userAgent, ipAddress)
	if sErr != nil {
		s.logger.Error("failed to create session", log.Meta{"error": sErr, "user_id": user.ID.String()})
		return nil, apierr.InternalServerError("Failed to establish session").Wrap(sErr)
	}

	sessionID := sess.ID

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
