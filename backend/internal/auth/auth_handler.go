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
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"
)

type IdentifierType string

const (
	Email IdentifierType = "email"
	Phone IdentifierType = "phone"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,token"`
}

type TokensResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type MeResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Role      string    `json:"role,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero"`
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

type StartChallengeRequest struct {
	Identifier string `json:"identifier" binding:"required"`
}

type StartChallengeResponse struct {
	Identifier string `json:"identifier"`
	ExpiresAt  string `json:"expires_at"`
	Duration   int64  `json:"duration"`
}

type ResendChallengeRequest struct {
	Identifier string `json:"identifier" binding:"required"`
}

type VerifyChallengeRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Code       string `json:"code" binding:"required,len=6,numeric"`
}

// StartChallenge godoc
// @Summary Start an authentication challenge
// @Description Starts an authentication challenge by sending a verification code to the provided email or phone number.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body StartChallengeRequest true "Challenge Request"
// @Success 200 {object} api.DataResponse{data=StartChallengeResponse}
// @Failure 400 {object} api.BadReqResponse
// @Failure 429 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/challenge/start [post]
func (h *Handler) StartChallenge(c *gin.Context) {
	var req StartChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	var identifierType IdentifierType
	phone, err := utils.IsValidPhone(req.Identifier)
	if err != nil {
		email, err := utils.IsValidEmail(req.Identifier)
		if err != nil {
			_ = c.Error(
				security.NewSecureError(
					http.StatusBadRequest,
					security.CodeValidation,
					"invalid identifier format",
					err,
				),
			)
			return
		}

		req.Identifier = email
		identifierType = Email
	} else {
		identifierType = Phone
		req.Identifier = phone
	}

	ctx := c.Request.Context()
	challenge, err := h.authService.StartChallange(
		ctx,
		req.Identifier,
		identifierType,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{
		Data: StartChallengeResponse{
			Identifier: req.Identifier,
			ExpiresAt:  challenge.ExpiresAt.UTC().Format(time.RFC3339),
			Duration:   ChallengeTTL.Milliseconds(),
		},
	})
}

// VerifyChallenge godoc
// @Summary Verify an authentication challenge
// @Description Verifies the one-time code sent to the user's email or phone and returns an access token and refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body VerifyChallengeRequest true "Challenge Verification Request"
// @Success 200 {object} api.DataResponse{data=TokensResponse}
// @Failure 400 {object} api.BadReqResponse
// @Failure 401 {object} api.UnauthorizedResponse
// @Failure 429 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/challenge/verify [post]
func (h *Handler) VerifyChallenge(c *gin.Context) {
	var req VerifyChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	tokens, err := h.authService.VerifyChallange(
		ctx,
		req.Identifier,
		req.Code,
	)
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

// ResendChallenge godoc
// @Summary Resend an authentication challenge
// @Description Resends a new verification code to the provided email or phone number if allowed by the challenge policy.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResendChallengeRequest true "Resend Challenge Request"
// @Success 200 {object} api.MessageResponse
// @Failure 400 {object} api.BadReqResponse
// @Failure 429 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /auth/challenge/resend [post]
func (h *Handler) ResendChallenge(c *gin.Context) {
	var req ResendChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	err := h.authService.ResendChallange(
		ctx,
		req.Identifier,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "the code sent succeccfully",
	})
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

	user, err := h.authService.GetCurrentUser(
		c.Request.Context(),
		val.(uuid.UUID),
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: MeResponse{
				ID:     user.ID,
				Status: string(user.Status),
				Role:   string(*user.Role),
			},
		},
	)
}
