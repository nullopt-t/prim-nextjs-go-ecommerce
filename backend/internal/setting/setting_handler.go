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
	if s == nil {
		panic("setting: service is required")
	}
	return &Handler{
		service: s,
	}
}

// SettingsMap is the key/value representation of the settings payload.
type SettingsMap map[string]interface{}

// CurrencyInfo represents a supported currency and its display metadata and conversion rate relative to base.
type CurrencyInfo struct {
	Code      string  `json:"code" example:"USD"`
	Symbol    string  `json:"symbol" example:"$"`
	Decimals  int     `json:"decimals" example:"2"`
	RateToEGP float64 `json:"rateToEgp" example:"48.5"`
}

type UpdateSettingsRequest struct {
	// Key/value pairs to upsert (e.g. {"currency": "EGP"}).
	Items map[string]interface{} `json:"items" binding:"required" example:"{\"currency\":\"EGP\"}"`
}

// GetSettings godoc
//
//	@Summary		Get administrative settings
//	@Description	Returns the resolved admin settings (persisted overrides merged onto defaults).
//	@Tags			Admin Settings
//	@Accept			json
//	@Produce		json
//	@Failure		401	{object}	api.UnauthorizedErrorResponse	"Authentication required"
//	@Failure		403	{object}	api.ForbiddenErrorResponse		"Admin privileges required"
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=Config}	"Current settings"
//	@Router			/admin/settings [get]
func (h *Handler) GetSettings(c *gin.Context) {
	cfg, err := h.service.GetConfig(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: cfg})
}

// GetPublicSettings godoc
//
//	@Summary		Get public storefront settings
//	@Description	Returns the non-sensitive settings used by the storefront (identity, currency, announcement, shipping, maintenance state).
//	@Tags			Settings
//	@Accept			json
//	@Produce		json
//	@Failure		500	{object}	api.InternalServerErrorResponse	"Internal server error"
//	@Success		200	{object}	api.DataResponse{data=PublicConfig}	"Public settings"
//	@Router			/settings/public [get]
func (h *Handler) GetPublicSettings(c *gin.Context) {
	cfg, err := h.service.GetPublicConfig(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, api.DataResponse{Data: cfg})
}

// GetCurrencies godoc
//
//	@Summary		Get supported currencies and exchange rates
//	@Description	Returns supported currencies, their symbols, decimal places, and exchange conversion rates relative to the store base currency.
//	@Tags			Settings
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	api.DataResponse{data=[]CurrencyInfo}	"Supported currencies"
//	@Router			/settings/currencies [get]
func (h *Handler) GetCurrencies(c *gin.Context) {
	currencies := []CurrencyInfo{
		{Code: "EGP", Symbol: "E£", Decimals: 2, RateToEGP: 1.0},
		{Code: "USD", Symbol: "$", Decimals: 2, RateToEGP: 48.5},
		{Code: "EUR", Symbol: "€", Decimals: 2, RateToEGP: 52.0},
		{Code: "SAR", Symbol: "SR", Decimals: 2, RateToEGP: 12.9},
		{Code: "AED", Symbol: "AED", Decimals: 2, RateToEGP: 13.2},
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: currencies})
}

// UpdateSettings godoc
//
//	@Summary		Update administrative settings (Admin)
//	@Description	Upserts the provided key/value settings within a single transaction. Admin only.
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
