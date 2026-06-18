package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/security"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/log"
)

const (
	ResendOTPLimit = 5
	VerifyOTPLimit = 3

	maxEmailLength = 255
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type ForgetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyResetTokenRequest struct {
	Token    string `json:"token" binding:"required,token"`
	Password string `json:"password" binding:"required,password"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,max=999999"`
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UserRegisterResponse struct {
	Message string `json:"message"`
}

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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

func (h *Handler) handleValidationError(c *gin.Context, err error) {
	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok && ve != nil {
		fieldErrors := make([]api.FieldError, 0, len(ve))
		for _, e := range ve {
			fieldErrors = append(fieldErrors, api.FieldError{
				Field: e.Field(),
				Tags:  e.Tag(),
			})
		}
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"bad request data",
			err,
		).WithFields(fieldErrors))
		return
	}
	c.Error(
		security.NewSecureError(
			http.StatusBadRequest,
			security.CodeValidation,
			"bad request data",
			err,
		),
	)
}

// UserRegister godoc
// @Summary Register a new user
// @Description Register a new user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body RegisterUserRequest true "User Credentials"
// @Success 201 {object} api.SuccessResponse{data=UserRegisterResponse}
// @Failure 400 {object} api.ErrorResponse
// @Failure 404 {object} api.ErrorResponse
// @Router /auth [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := validation.IsValidEmail(req.Email); err != nil {
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"invalid email",
			err,
		))
		return
	}

	if err := validation.IsValidPassword(req.Password); err != nil {
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"invalid password",
			err,
		))
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
		api.SuccessResponse{
			Message: "If the email is valid, you will receive a verification email.",
		},
	)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := validation.IsValidEmail(req.Email); err != nil {
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"invalid email",
			err,
		))
		return
	}

	if err := validation.IsValidPassword(req.Password); err != nil {
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"invalid password",
			err,
		))
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: map[string]string{
				"access_token":  tokens.AccessToken,
				"refresh_token": tokens.RefreshToken,
			},
		},
	)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	accessToken, err := h.authService.RotateAccessToken(
		c.Request.Context(),
		req.RefreshToken,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Data: RefreshTokenResponse{
				AccessToken: accessToken,
			},
		},
	)
}

func (h *Handler) ForgetPassword(c *gin.Context) {
	var req ForgetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := validation.IsValidEmail(req.Email); err != nil {
		c.Error(
			security.NewSecureError(
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid email address",
				err,
			),
		)
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
		c.Error(err)
		return
	}
	if !allow {
		c.Error(http.ErrBodyNotAllowed)
		return
	}

	// generate reset token and send email
	// if the email exists, a reset token will be generated and sent to the user's email
	// if the email does not exist, no action will be taken
	// but no error will be returned, to prevent email enumeration
	err = h.authService.ForgetPassword(ctx, req.Email)
	if err != nil {
		h.logger.Error("failed to forget password", log.Meta{
			"email": req.Email,
			"error": err,
		})
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "If an account exists, a password reset email has been sent.",
		},
	)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req VerifyResetTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := validation.IsValidPassword(req.Password); err != nil {
		c.Error(
			security.NewSecureError(
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid password",
				err,
			),
		)
		return
	}

	if err := h.authService.ResetPassword(
		c.Request.Context(),
		req.Token,
		req.Password,
	); err != nil {
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "password updated successfully",
		},
	)
}

func (h *Handler) VerifyOTP(c *gin.Context) {
	ctx := c.Request.Context()
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := validation.IsValidEmail(req.Email); err != nil {
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			"INVALID_EMAIL",
			"invalid email",
			err,
		))
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
		c.Error(err)
		return
	}

	if !allowed {
		c.Error(
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
		c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "OK",
		},
	)
}

func (h *Handler) ResendOTP(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	if err := validation.IsValidEmail(req.Email); err != nil {
		c.Error(security.NewSecureError(
			http.StatusBadRequest,
			"INVALID_EMAIL",
			"invalid email",
			err,
		))
		return
	}

	key := "otp:verify:" + req.Email
	allowed, err := h.limiter.Allow(
		ctx,
		key,
		ResendOTPLimit,
		1*time.Hour,
	)
	if err != nil {
		c.Error(err)
		return
	}

	if !allowed {
		c.Error(
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
		api.SuccessResponse{
			Message: "go check your email !",
		},
	)
}
