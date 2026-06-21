package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
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

type RoleService interface {
	UserRoles(
		ctx context.Context,
		userID uuid.UUID,
	) ([]*model.Role, error)
}

type Notifier interface {
	NotifyResetPassword(
		ctx context.Context,
		email,
		token string,
	) error
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	jwtService  *jwt.JWTManager
	otpService  OTPService
	userService UserService
	roleService RoleService
	logger      log.Logger
	redisClient *redis.Client
	notifier    Notifier
	secrets     *config.Secrets
}

func NewAuthService(
	logger log.Logger,
	userService UserService,
	roleService RoleService,
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
		roleService: roleService,
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

	err = user.VerifyPassword(
		req.Password,
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusUnauthorized,
			"INVALID_CREDENTIALS",
			"email or password is incorrect",
			err,
		)
	}

	roles, err := s.roleService.UserRoles(
		ctx,
		user.ID,
	)
	if err != nil {
		return nil, security.NewSecureError(
			http.StatusInternalServerError,
			security.CodeInternal,
			"failed to fetch user roles",
			err,
		)
	}

	claims := &jwt.UserClaims{
		UserID:    user.ID.String(),
		UserEmail: user.Email,
		UserRole:  model.RolesToStrings(roles),
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

func (s *AuthService) ForgotPassword(
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

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", err
	}

	user, err := s.userService.GetUserByID(
		ctx,
		userID,
	)
	if err != nil {
		return "", "", err
	}

	roles, err := s.roleService.UserRoles(
		ctx,
		user.ID,
	)
	if err != nil {
		return "", "", err
	}

	newClaims := &jwt.UserClaims{
		UserID:    user.ID.String(),
		UserEmail: user.Email,
		UserRole:  model.RolesToStrings(roles),
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
) (*model.User, []*model.Role, error) {
	user, err := s.userService.GetUserByID(
		ctx,
		userID,
	)
	if err != nil {
		return nil, nil, err
	}

	roles, err := s.roleService.UserRoles(
		ctx,
		user.ID,
	)
	if err != nil {
		return nil, nil, err
	}

	return user, roles, nil
}
