package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nullopt-t/prim-backend/internal/shared/middleware"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
	"github.com/nullopt-t/prim-backend/pkg/log"
	"github.com/nullopt-t/prim-backend/pkg/utils"
)

// MeResponse encapsulates the sanitized current user profile returned by the /auth/me endpoint.
// It exposes identity, administrative role tier, account status, and registration timestamp
// without leaking sensitive credential material or internal metadata.
type MeResponse struct {
	ID         uuid.UUID `json:"id,omitempty"`
	Identifier string    `json:"identifier,omitempty"`
	Role       string    `json:"role,omitempty"`
	Status     string    `json:"status,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitzero"`
}

// Handler orchestrates HTTP transport layer operations for authentication,
// challenge lifecycles (passwordless OTP via email/SMS), session management,
// and cart unification between guest and authenticated states.
type Handler struct {
	authService    *AuthService
	sessionService *SessionService
	logger         log.Logger
	isProduction   bool
	cartMerger     CartMerger
	challengeTTL   time.Duration
}

// CartMerger defines the contract for transferring and consolidating cart line items
// accumulated under an anonymous guest session into the authenticated user's persistent cart.
type CartMerger interface {
	MergeGuestCart(ctx context.Context, sessionID string, userID uuid.UUID) error
}

// NewAuthHandler constructs a new Handler instance with required dependencies
// including the domain auth service, session storage manager, logger, and configuration flags.
func NewAuthHandler(
	authService *AuthService,
	sessionService *SessionService,
	logger log.Logger,
	isProduction bool,
	cartMerger CartMerger,
	challengeTTL time.Duration,
) *Handler {
	return &Handler{
		authService:    authService,
		sessionService: sessionService,
		logger:         logger,
		isProduction:   isProduction,
		cartMerger:     cartMerger,
		challengeTTL:   challengeTTL,
	}
}

// StartChallengeRequest represents the inbound JSON body required to initiate an OTP challenge.
type StartChallengeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// StartChallengeResponse represents the JSON response providing the client with challenge metadata,
// including expiration timestamp and remaining validity duration in milliseconds.
type StartChallengeResponse struct {
	Email     string `json:"email"`
	ExpiresAt string `json:"expiresAt"`
	Duration  int64  `json:"duration"`
}

// ResendChallengeRequest represents the inbound JSON body to trigger a new OTP dispatch.
type ResendChallengeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyChallengeRequest represents the inbound payload to validate an OTP challenge code.
// Optionally contains a flag indicating whether the client requests merging guest cart items.
type VerifyChallengeRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Code      string `json:"code" binding:"required,len=6,numeric"`
	MergeCart *bool  `json:"merge_cart,omitempty"`
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
	// Parse and validate the incoming JSON request payload against struct tags
	var req StartChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Canonicalize and strictly validate the email address format
	email, err := utils.IsValidEmail(req.Email)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid email"))
		return
	}

	// Delegate challenge creation, OTP generation, hashing, and notification dispatch to the domain service
	ctx := c.Request.Context()
	challenge, err := h.authService.StartChallenge(
		ctx,
		email,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Return successful challenge response containing the expiration date and duration in milliseconds
	c.JSON(http.StatusOK, api.DataResponse{
		Data: StartChallengeResponse{
			Email:     email,
			ExpiresAt: challenge.ExpiresAt.UTC().Format(time.RFC3339),
			Duration:  h.challengeTTL.Milliseconds(),
		},
	})
}

const (
	// SessionCookieMaxAge specifies the browser cookie lifetime in seconds (30 days).
	SessionCookieMaxAge = 30 * 24 * 3600 // 30 days
)

// setAuthCookies issues the HTTP-only server-side session cookie to the client browser.
// Uses SameSite=Lax and restricts transmission to HTTPS in production environments.
func (h *Handler) setAuthCookies(c *gin.Context, sessionID string) {
	c.SetSameSite(http.SameSiteLaxMode)
	if sessionID != "" {
		c.SetCookie(middleware.SessionCookieName, sessionID, SessionCookieMaxAge, "/", "", h.isProduction, true)
	}
}

// clearAuthCookies removes active and legacy authentication cookies from the client browser
// by setting their MaxAge to -1 and overwriting values with empty strings.
func (h *Handler) clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(middleware.SessionCookieName, "", -1, "/", "", h.isProduction, true)
	// Clear legacy cookies if they linger in the browser
	c.SetCookie("session_id", "", -1, "/", "", h.isProduction, true)
	c.SetCookie("access_token", "", -1, "/", "", h.isProduction, true)
	c.SetCookie("refresh_token", "", -1, "/", "", h.isProduction, true)
}

// VerifyChallenge godoc
//
//	@Summary		Verify an authentication challenge
//	@Description	Verifies the one-time code sent to the user's email or phone and sets auth cookies.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		VerifyChallengeRequest	true	"Challenge Verification Request"
//	@Success		204		"No Content"
//	@Failure		400		{object}	api.BadReqResponse
//	@Failure		401		{object}	api.UnauthorizedResponse
//	@Failure		429		{object}	api.ErrorResponse
//	@Failure		500		{object}	api.InternalServerErrorResponse
//	@Router			/auth/challenge/verify [post]
func (h *Handler) VerifyChallenge(c *gin.Context) {
	// Parse and validate the incoming verification payload (email, 6-digit numeric OTP, optional merge flag)
	var req VerifyChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Canonicalize and strictly validate the email address so /verify uses the
	// same identifier as /start (otherwise challenge lookups miss on case/tags)
	email, err := utils.IsValidEmail(req.Email)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid email"))
		return
	}

	// Capture client fingerprinting metadata (User-Agent and Client IP) for session binding
	ctx := c.Request.Context()
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	// Verify the OTP challenge code and transition the user account state
	authResult, err := h.authService.VerifyChallenge(
		ctx,
		email,
		req.Code,
		userAgent,
		ipAddress,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if authResult == nil || authResult.SessionID == "" {
		_ = c.Error(apierr.InternalServerError("Failed to establish user session"))
		return
	}

	// Merge guest cart if a guest session cookie is present and merging is enabled (defaults to true)
	shouldMerge := req.MergeCart == nil || *req.MergeCart
	guestSessionID, err := c.Cookie("session_id")
	if err == nil && guestSessionID != "" {
		if shouldMerge && h.cartMerger != nil {
			// Transfer items from guest session cart into user account cart
			if mergeErr := h.cartMerger.MergeGuestCart(c.Request.Context(), guestSessionID, authResult.UserID); mergeErr != nil {
				h.logger.Error("Failed to merge guest cart", log.Meta{"error": mergeErr, "session_id": guestSessionID, "user_id": authResult.UserID})
			}
			// Always clear guest session cookie after merge attempt so the guest cookie is not orphaned
			c.SetCookie("session_id", "", -1, "/", "", h.isProduction, true)
		} else if !shouldMerge {
			// User opted out of merging: clear guest session cookie to keep account cart pristine
			c.SetCookie("session_id", "", -1, "/", "", h.isProduction, true)
		}
	}

	// Set server-side session cookie for the authenticated user session
	h.setAuthCookies(c, authResult.SessionID)

	// Return 204 No Content indicating successful verification and cookie assignment
	c.Status(http.StatusNoContent)
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
	// Parse and validate the resend challenge request body
	var req ResendChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validation.ValidationError(c, err)
		return
	}

	// Canonicalize and strictly validate the email address so /resend uses the
	// same identifier as /start
	email, err := utils.IsValidEmail(req.Email)
	if err != nil {
		_ = c.Error(apierr.BadRequestError("invalid email"))
		return
	}

	// Trigger OTP regeneration and re-transmission through the domain auth service
	ctx := c.Request.Context()
	err = h.authService.ResendChallenge(
		ctx,
		email,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Return confirmation message to client
	c.JSON(http.StatusOK, api.MessageResponse{
		Message: "the code sent successfully",
	})
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
	// Retrieve authenticated user UUID established in Gin context by authentication middleware
	userID := c.MustGet("userID").(uuid.UUID)

	// Fetch fresh user record from the database to reflect up-to-date role and status
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Map user role tier
	roleStr := string(user.Role)

	// Return sanitized user profile
	c.JSON(
		http.StatusOK,
		api.DataResponse{
			Data: MeResponse{
				ID:         user.ID,
				Identifier: user.Identifier,
				Status:     string(user.Status),
				Role:       roleStr,
				CreatedAt:  user.CreatedAt,
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
	// Extract authenticated user ID from context
	userID := c.MustGet("userID").(uuid.UUID)

	// Query Redis for all active sessions registered under the user's session set
	sessions, err := h.sessionService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Return session list with IP address, user-agent, creation, and last activity timestamps
	c.JSON(http.StatusOK, api.DataResponse{Data: sessions})
}

// DeleteSessionByID godoc
//
//	@Summary		Revoke session
//	@Description	Revokes and deletes a specific active session from Redis.
//	@Tags			Auth
//	@Produce		json
//	@Param			id	path		string	true	"Session ID to revoke"
//	@Success		204	"No Content"
//	@Failure		401	{object}	api.UnauthorizedResponse
//	@Failure		500	{object}	api.InternalServerErrorResponse
//	@Router			/auth/sessions/{id} [delete]
func (h *Handler) DeleteSessionByID(c *gin.Context) {
	// Extract authenticated user ID from context
	userID := c.MustGet("userID").(uuid.UUID)
	sessionID := c.Param("id")

	// Revoke the specified session record from Redis
	err := h.sessionService.DeleteSessionByID(c.Request.Context(), userID, sessionID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// If the revoked sessionID matches the caller's current active session, clear the session cookies from the client browser
	if activeSessionID, exists := c.Get("sessionID"); exists && activeSessionID == sessionID {
		h.clearAuthCookies(c)
	}

	// Return 204 No Content upon successful revocation
	c.Status(http.StatusNoContent)
}

// DeleteAllSessions godoc
//
//	@Summary		Revoke all active user sessions
//	@Description	Revokes and deletes all active sessions from Redis and clears auth cookies.
//	@Tags			Auth
//	@Produce		json
//	@Success		204	"No Content"
//	@Failure		401	{object}	api.UnauthorizedResponse
//	@Failure		500	{object}	api.InternalServerErrorResponse
//	@Router			/auth/sessions [delete]
func (h *Handler) DeleteAllSessions(c *gin.Context) {
	// Extract caller's user UUID from context
	userID := c.MustGet("userID").(uuid.UUID)

	// Delete all session keys and the user's tracking set from Redis
	err := h.sessionService.DeleteAllSessions(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	// Clear authentication cookies on the current client browser
	h.clearAuthCookies(c)
	c.Status(http.StatusNoContent)
}

// Logout godoc
//
//	@Summary		Logout
//	@Description	Revokes the current session and clears auth cookies.
//	@Tags			Auth
//	@Produce		json
//	@Success		204	"No Content"
//	@Failure		401	{object}	api.UnauthorizedResponse
//	@Failure		500	{object}	api.InternalServerErrorResponse
//	@Router			/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// Extract authenticated user ID from context
	userID := c.MustGet("userID").(uuid.UUID)

	// Extract active session ID injected by authentication middleware
	sessionID := ""
	if sid, exists := c.Get("sessionID"); exists && sid != "" {
		sessionID = sid.(string)
	}

	// Delete current session from Redis store. Revocation failure is surfaced
	// instead of swallowed: cookies are only cleared once the server-side
	// session is confirmed revoked, otherwise the session stays valid.
	if sessionID != "" {
		if err := h.sessionService.DeleteSessionByID(c.Request.Context(), userID, sessionID); err != nil {
			_ = c.Error(err)
			return
		}
	}

	// Clear auth cookies on client browser and respond with 204 No Content
	h.clearAuthCookies(c)
	c.Status(http.StatusNoContent)
}
