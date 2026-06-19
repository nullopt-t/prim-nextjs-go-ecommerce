package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/types"
)

const (
	ResendOTPLimit = 5
	VerifyOTPLimit = 3
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,token"`
}

type TokensResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
}

type ResetTokenRequest struct {
	Token    string `json:"token" binding:"required,token"`
	Password string `json:"password" binding:"required,password,min=8,max=70"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
	Code  string `json:"code" binding:"required,len=6,numeric"`
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
}

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=70"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=70"`
}

type MeResponse struct {
	ID            uuid.UUID `json:"id,omitempty"`
	Email         string    `json:"email,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	Role          string    `json:"role,omitempty"`
	Status        string    `json:"status,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitzero"`
}

type Handler struct {
	authService *AuthService
	limiter     *security.RateLimiter
	logger      log.Logger
}

func NewAuthHandler(
	authService *AuthService,
	limiter *security.RateLimiter,
	logger log.Logger,
) *Handler {
	return &Handler{
		authService: authService,
		limiter:     limiter,
		logger:      logger,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body RegisterUserRequest true "User Credentials"
// @Success 201 {object} api.MessageResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	err := h.authService.Register(
		c.Request.Context(),
		req,
	)
	if err != nil {
		h.logger.Error(
			"user register",
			log.Meta{
				"Error": err,
			},
		)
	}

	c.JSON(
		http.StatusCreated,
		api.MessageResponse{
			Message: "If the email is valid, you will receive a verification email.",
		},
	)
}

// Login godoc
// @Summary Login user by email and password
// @Description Login user by email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginUserRequest true "User Credentials"
// @Success 200 {object} api.DataResponse{data=TokensResponse}
// @Failure 400 {object} api.BadReqResponse
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: TokensResponse{
				AccessToken:  tokens.AccessToken,
				RefreshToken: tokens.RefreshToken,
			},
		},
	)
}

// Refresh godoc
// @Summary Rotate refresh token and issue new access and refresh tokens.
// @Description Rotate refresh token and issue new access and refresh tokens.
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh_token body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} api.DataResponse{data=TokensResponse}
// @Failure 400 {object} api.BadReqResponse
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 429 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	accessToken, refreshToken, err := h.authService.RotateToken(
		c.Request.Context(),
		req.RefreshToken,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: TokensResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
		},
	)
}

// ForgetPassword godoc
// @Summary Email reset password link
// @Description Email reset password link
// @Tags Auth
// @Accept json
// @Produce json
// @Param email body ForgotPasswordRequest true "User Email"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 429 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/forget-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	allow, err := h.limiter.Allow(
		ctx,
		"forget_password:"+req.Email,
		10,
		time.Minute,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if !allow {
		_ = c.Error(security.NewSecureError(
			http.StatusTooManyRequests,
			security.CodeRateLimit,
			"too many requests",
			nil,
		))
		return
	}

	// generate reset token and send email
	// if the email exists, a reset token will be generated and sent to the user's email
	// if the email does not exist, no action will be taken
	// but no error will be returned, to prevent email enumeration
	err = h.authService.ForgotPassword(ctx, req.Email)
	if err != nil {
		h.logger.Error("failed to forget password", log.Meta{
			"email": req.Email,
			"error": err,
		})
	}

	c.JSON(
		http.StatusOK,
		api.MessageResponse{
			Message: "If an account exists, a password reset email has been sent.",
		},
	)
}

// ResetPassword godoc
// @Summary Reset password using a reset token
// @Description Reset password using a reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body ResetTokenRequest true "Password and Reset Token"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if err := h.authService.ResetPassword(
		c.Request.Context(),
		req.Token,
		req.Password,
	); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.MessageResponse{
			Message: "password updated successfully",
		},
	)
}

// VerifyOTP godoc
// @Summary Verify a user's email address
// @Description Verify a user's email using a one-time verification code.
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body VerifyOTPRequest true "Email and OTP"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 429 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	key := "otp:verify:" + req.Email
	allowed, err := h.limiter.Allow(
		ctx,
		key,
		VerifyOTPLimit,
		10*time.Minute,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if !allowed {
		_ = c.Error(
			security.NewSecureError(
				http.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"too many verification attempts",
				nil,
			),
		)
		return
	}

	if err := h.authService.VerifyOTP(ctx, req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.MessageResponse{
			Message: "Email verified successfully.",
		},
	)
}

// ResendOTP godoc
// @Summary resend email otp
// @Description resend email otp
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body ResendOTPRequest true "Email and OTP"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 429 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/resend-otp [post]
func (h *Handler) ResendOTP(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	key := "otp:send:" + req.Email
	allowed, err := h.limiter.Allow(
		ctx,
		key,
		ResendOTPLimit,
		1*time.Hour,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if !allowed {
		_ = c.Error(
			security.NewSecureError(
				http.StatusTooManyRequests,
				"RATE_LIMIT_EXCEEDED",
				"too many otp resend requests",
				nil,
			),
		)
		return
	}

	// we should not return an error here, as the email may not exist
	// if the email does not exist, we will simply not send an OTP
	err = h.authService.SendEmailOTP(ctx, req.Email)
	if err != nil {
		h.logger.Error(
			"send otp email",
			log.Meta{
				"Error": err,
			},
		)
	}

	c.JSON(
		http.StatusOK,
		api.MessageResponse{
			Message: "If the email exists, a new verification code has been sent.",
		},
	)
}

// GetMe godoc
// @Summary fetch session user data
// @Description fetch session user data
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} api.DataResponse{data=MeResponse}
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		_ = c.Error(security.NewSecureError(
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"no session for this user",
			nil,
		))
		return
	}

	userID, err := uuid.Parse(val.(string))
	if err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.authService.GetCurrentUser(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: MeResponse{
				ID:            user.ID,
				Email:         user.Email,
				EmailVerified: types.BoolFromPtr(user.EmailVerifiedAt),
				Role:          user.Role.String(),
				Status:        string(user.Status),
			},
		},
	)
}
