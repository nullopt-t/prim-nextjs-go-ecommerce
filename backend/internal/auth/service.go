package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/config"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"

	"github.com/m-mahmoud-alsaid/prim-backend/internal/model"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/crypto"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/jwt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type OTPService interface {
	SendOTP(
		ctx context.Context,
		email string,
		purpose string,
	) error

	VerifyOTP(
		ctx context.Context,
		email string,
		purpose string,
		otp string,
	) error
}

type UserService interface {
	CreateUser(
		ctx context.Context,
		email string,
		password string,
	) (*model.User, error)

	GetUserByEmail(
		ctx context.Context,
		email string,
	) (*model.User, error)

	GetUserByID(
		ctx context.Context,
		userID uuid.UUID,
	) (*model.User, error)

	CheckUserPassword(
		user *model.User,
		password string,
	) error

	UpdateUserPassword(
		ctx context.Context,
		user *model.User,
		password string,
	) error

	MarkEmailVerified(
		ctx context.Context,
		email string,
	) error
}

type Notifier interface {
	NotifyResetPassword(ctx context.Context, email, token string) error
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	jwtService  *jwt.JWTManager
	otpService  OTPService
	userService UserService
	logger      log.Logger
	redisClient *redis.Client
	notifier    Notifier
	secrets     *config.Secrets
}

func NewAuthService(
	logger log.Logger,
	userService UserService,
	jwtService *jwt.JWTManager,
	otpService OTPService,
	redisClient *redis.Client,
	notifier Notifier,
	secrets *config.Secrets,
) *AuthService {
	return &AuthService{
		jwtService:  jwtService,
		otpService:  otpService,
		userService: userService,
		logger:      logger,
		redisClient: redisClient,
		notifier:    notifier,
		secrets:     secrets,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	req LoginUserRequest,
) (*Tokens, error) {
	user, err := s.userService.GetUserByEmail(
		ctx,
		req.Email,
	)
	if err != nil {
		return nil, err
	}

	err = s.userService.CheckUserPassword(
		user,
		req.Password,
	)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		user.ID.String(),
		user.Role.String(),
	)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Register(
	ctx context.Context,
	req RegisterUserRequest,
) error {
	_, err := s.userService.CreateUser(
		ctx,
		req.Email,
		req.Password,
	)
	if err != nil {
		return err
	}

	err = s.otpService.SendOTP(
		ctx,
		req.Email,
		"register",
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) VerifyOTP(
	ctx context.Context,
	req VerifyOTPRequest,
) error {
	err := s.otpService.VerifyOTP(
		ctx,
		req.Email,
		"register",
		req.Code,
	)
	if err != nil {
		return err
	}

	return s.userService.MarkEmailVerified(
		ctx,
		req.Email,
	)
}

func (s *AuthService) SendEmailOTP(
	ctx context.Context,
	email string,
) error {
	user, err := s.userService.GetUserByEmail(
		ctx,
		email,
	)
	if err != nil {
		return err
	}

	if user.EmailVerifiedAt != nil {
		return nil
	}

	return s.otpService.SendOTP(ctx,
		user.Email,
		"register",
	)
}

func (s *AuthService) ForgetPassword(
	ctx context.Context,
	email string,
) error {
	user, err := s.userService.GetUserByEmail(
		ctx,
		email,
	)
	if err != nil {
		return err
	}

	// generate random reset token
	resetToken, err := jwt.RandomToken()
	if err != nil {
		return err
	}

	tokenHash, err := crypto.Hash(resetToken)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("reset-token:%s", tokenHash)
	err = s.redisClient.Set(ctx, key, user.ID.String(), time.Hour*24).Err()
	if err != nil {
		return err
	}

	return s.notifier.NotifyResetPassword(
		ctx,
		user.Email,
		resetToken,
	)
}

func (s *AuthService) ResetPassword(
	ctx context.Context,
	token string,
	newPassword string,
) error {
	tokenHash, err := crypto.Hash(token)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("reset-token:%s", tokenHash)
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		return err
	}

	defer s.redisClient.Del(ctx, key)

	user, err := s.userService.GetUserByID(
		ctx,
		userID,
	)
	if err != nil {
		return err
	}

	err = s.userService.UpdateUserPassword(
		ctx,
		user,
		newPassword,
	)
	if err != nil {
		return err
	}

	return nil
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

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(
		claims.UserID,
		claims.UserRole,
	)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
