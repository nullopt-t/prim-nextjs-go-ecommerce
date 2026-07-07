package auth

import (
	"context"
	"net/http"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
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
		UserRole: string(*user.Role),
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		newClaims,
	)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) GetCurrentUser(
	ctx context.Context,
	userID uuid.UUID,
) (*model.User, error) {
	user, err := s.userService.GetUserByID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) StartChallange(
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

func (s *AuthService) ResendChallange(
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

func (s *AuthService) VerifyChallange(
	ctx context.Context,
	identifier,
	code string,
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
		return nil, security.NewSecureError(
			http.StatusUnauthorized,
			security.CodeUnauthorized,
			"incorrect otp",
			nil,
		)
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
		UserRole: string(*user.Role),
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		claims,
	)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
