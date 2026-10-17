package setting

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
)

// Handler exposes administrative settings endpoints over HTTP.
type Handler struct {
	service *Service
}

// NewHandler constructs a settings HTTP handler.
func NewHandler(s *Service) *Handler {
	return &Handler{
		service: s,
	}
}

// SettingsMap is the key/value representation of the settings payload.
type SettingsMap map[string]interface{}

type UpdateSettingsRequest struct {
	// Key/value pairs to upsert (e.g. {"currency": "EGP"}).
	Items map[string]interface{} `json:"items" binding:"required" example:"{\"currency\":\"EGP\"}"`
}

// GetSettings godoc
//
//	@Summary		Get administrative settings
//	@Description	Returns all persisted admin settings as a key/value map.
//	@Tags			Admin Settings
//	@Accept			json
//	@Produce		json
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Authentication required"
//	@Failure		403	{object}	api.ForbiddenErrorResponse		"Admin privileges required"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=SettingsMap}	"Current settings"
//	@Router			/admin/settings [get]
func (h *Handler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: settings})
}

// UpdateSettings godoc
//
//	@Summary		Update administrative settings
//	@Description	Upserts the provided key/value settings within a single transaction.
//	@Tags			Admin Settings
//	@Accept			json
//	@Produce		json
//	@Param			data	body		UpdateSettingsRequest					true	"Settings key/value payload"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Invalid request body"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse			"Authentication required"
//	@Failure		403		{object}	api.ForbiddenErrorResponse				"Admin privileges required"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200		{object}	api.MessageResponse						"Update confirmation message"
//	@Router			/admin/settings [put]
func (h *Handler) UpdateSettings(c *gin.Context) {
	var body UpdateSettingsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		validation.ValidationError(c, err)
		return
	}

	if err := h.service.UpdateSettings(c.Request.Context(), body.Items); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.MessageResponse{Message: "Settings updated successfully"})
}
