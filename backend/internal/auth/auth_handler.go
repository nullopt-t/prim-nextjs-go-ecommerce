package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mahmoud-alsaid/prim-backend/internal/shared/validation"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api"
	"github.com/m-mahmoud-alsaid/prim-backend/pkg/api/apierr"
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
	RefreshToken string `json:"refresh_token"`
}

type MeResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Role      string    `json:"role,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero"`
}

type Handler struct {
	authService  *AuthService
	limiter      *security.RateLimiter
	logger       log.Logger
	isProduction bool
}

func NewAuthHandler(
	authService *AuthService,
	limiter *security.RateLimiter,
	logger log.Logger,
	isProduction bool,
) *Handler {
	return &Handler{
		authService:  authService,
		limiter:      limiter,
		logger:       logger,
		isProduction: isProduction,
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
//
//	@Summary		Start an authentication challenge
//	@Description	Starts an authentication challenge by sending a verification code to the provided email or phone number.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		StartChallengeRequest	true	"Challenge Request"
//	@Success		200		{object}	api.DataResponse{data=StartChallengeResponse}
//	@Failure		400		{object}	api.BadReqResponse
//	@Failure		429		{object}	api.ErrorResponse
//	@Failure		500		{object}	api.InternalServerErrorResponse
//	@Router			/auth/challenge/start [post]
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
				apierr.New(
					http.StatusBadRequest,
					"invalid identifier",
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
	challenge, err := h.authService.StartChallenge(
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

func (h *Handler) setAuthCookies(c *gin.Context, accessToken, refreshToken, sessionID string) {
	c.SetSameSite(http.SameSiteLaxMode)
	if accessToken != "" {
		c.SetCookie("access_token", accessToken, 900, "/", "", h.isProduction, true)
	}
	if refreshToken != "" {
		c.SetCookie("refresh_token", refreshToken, 30*24*3600, "/", "", h.isProduction, true)
	}
	if sessionID != "" {
		c.SetCookie("session_id", sessionID, 30*24*3600, "/", "", h.isProduction, true)
	}
}

// VerifyChallenge godoc
//
//	@Summary		Verify an authentication challenge
//	@Description	Verifies the one-time code sent to the user's email or phone and sets auth cookies.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		VerifyChallengeRequest	true	"Challenge Verification Request"
//	@Success		200		{object}	api.SuccessResponse
//	@Failure		400		{object}	api.BadReqResponse
//	@Failure		401		{object}	api.UnauthorizedResponse
//	@Failure		429		{object}	api.ErrorResponse
//	@Failure		500		{object}	api.InternalServerErrorResponse
//	@Router			/auth/challenge/verify [post]
func (h *Handler) VerifyChallenge(c *gin.Context) {
	var req VerifyChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	tokens, err := h.authService.VerifyChallenge(
		ctx,
		req.Identifier,
		req.Code,
		userAgent,
		ipAddress,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	h.setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken, tokens.SessionID)

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "Verification successful",
		},
	)
}

// ResendChallenge godoc
//
//	@Summary		Resend an authentication challenge
//	@Description	Resends a new verification code to the provided email or phone number if allowed by the challenge policy.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ResendChallengeRequest	true	"Resend Challenge Request"
//	@Success		200		{object}	api.MessageResponse
//	@Failure		400		{object}	api.BadReqResponse
//	@Failure		429		{object}	api.ErrorResponse
//	@Failure		500		{object}	api.InternalServerErrorResponse
//	@Router			/auth/challenge/resend [post]
func (h *Handler) ResendChallenge(c *gin.Context) {
	var req ResendChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	ctx := c.Request.Context()
	err := h.authService.ResendChallenge(
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
//
//	@Summary		Rotate refresh token and issue new access and refresh tokens.
//	@Description	Rotate refresh token and issue new access and refresh tokens.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			refresh_token	body		RefreshTokenRequest	false	"Refresh Token"
//	@Success		200				{object}	api.SuccessResponse
//	@Failure		400				{object}	api.BadReqResponse
//	@Failure		401				{object}	api.UnauthorizedResponse
//	@Failure		429				{object}	api.ErrorResponse
//	@Failure		500				{object}	api.InternalServerErrorResponse
//	@Router			/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	refreshToken := ""
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		refreshToken = req.RefreshToken
	} else if cookie, err := c.Cookie("refresh_token"); err == nil && cookie != "" {
		refreshToken = cookie
	}

	if refreshToken == "" {
		_ = c.Error(apierr.ErrBadRequest("Refresh token is required").WithCode(apierr.CodeInvalidInput))
		return
	}

	accessToken, newRefreshToken, err := h.authService.RotateToken(
		c.Request.Context(),
		refreshToken,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	h.setAuthCookies(c, accessToken, newRefreshToken, "")

	c.JSON(
		http.StatusOK,
		api.SuccessResponse{
			Message: "Token refreshed successfully",
		},
	)
}

// GetMe godoc
//
//	@Summary		fetch session user data
//	@Description	fetch session user data
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	api.DataResponse{data=MeResponse}
//	@Failure		500	{object}	api.InternalServerErrorResponse
//	@Router			/auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
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

// GetSessions godoc
//
//	@Summary		List active user sessions
//	@Description	Fetches all active login sessions stored in Redis for the current user.
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	api.DataResponse{data=[]UserSession}
//	@Failure		401	{object}	api.UnauthorizedResponse
//	@Failure		500	{object}	api.InternalServerErrorResponse
//	@Router			/auth/sessions [get]
func (h *Handler) GetSessions(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	sessions, err := h.authService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: sessions})
}

// DeleteSessionByID godoc
//
//	@Summary		Revoke session
//	@Description	Revokes and deletes a specific active session from Redis.
//	@Tags			Auth
//	@Produce		json
//	@Param			id	path		string	true	"Session ID to revoke"
//	@Success		200	{object}	api.MessageResponse
//	@Failure		401	{object}	api.UnauthorizedResponse
//	@Failure		500	{object}	api.InternalServerErrorResponse
//	@Router			/auth/sessions/{id} [delete]
func (h *Handler) DeleteSessionByID(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	sessionID := c.Param("id")

	err := h.authService.DeleteSessionByID(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
