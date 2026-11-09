package setting

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nullopt-t/prim-backend/internal/shared/validation"
	"github.com/nullopt-t/prim-backend/pkg/api"
	"github.com/nullopt-t/prim-backend/pkg/api/apierr"
)

// Handler exposes administrative settings endpoints over HTTP.
type Handler struct {
	service         *Service
	currencyService *CurrencyService
}

// NewHandler constructs a settings HTTP handler.
func NewHandler(s *Service, cs ...*CurrencyService) *Handler {
	if s == nil {
		panic("setting: service is required")
	}
	var currencySvc *CurrencyService
	if len(cs) > 0 && cs[0] != nil {
		currencySvc = cs[0]
	} else {
		currencySvc = NewCurrencyService(s.dr, s.repo)
	}
	return &Handler{
		service:         s,
		currencyService: currencySvc,
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
	if h.currencyService != nil {
		currencies, err := h.currencyService.GetCurrencies(c.Request.Context())
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(http.StatusOK, api.DataResponse{Data: currencies})
		return
	}
	c.JSON(http.StatusOK, api.DataResponse{Data: DefaultCurrencies})
}

// UpdateCurrencies godoc
//
//	@Summary		Update currency exchange rates (Admin)
//	@Description	Updates custom exchange rates relative to the store base currency (EGP). Admin only.
//	@Tags			Admin Settings
//	@Accept			json
//	@Produce		json
//	@Param			data	body		UpdateSettingsRequest					true	"Exchange rates map"
//	@Failure		400		{object}	api.BadRequestErrorResponse				"Invalid exchange rates payload"
//	@Failure		401		{object}	api.UnauthorizedErrorResponse			"Authentication required"
//	@Failure		403		{object}	api.ForbiddenErrorResponse				"Admin privileges required"
//	@Failure		500		{object}	api.InternalServerErrorResponse			"Internal server error"
//	@Success		200		{object}	api.DataResponse{data=[]CurrencyInfo}	"Updated currencies"
//	@Router			/admin/settings/currencies [put]
func (h *Handler) UpdateCurrencies(c *gin.Context) {
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		validation.ValidationError(c, err)
		return
	}

	rates := make(map[string]float64)
	if r, ok := raw["rates"].(map[string]interface{}); ok {
		for k, v := range r {
			if num, ok := v.(float64); ok {
				rates[k] = num
			}
		}
	} else {
		for k, v := range raw {
			if num, ok := v.(float64); ok {
				rates[k] = num
			}
		}
	}

	if len(rates) == 0 {
		_ = c.Error(apierr.BadRequestError("Rates payload must contain at least one valid currency exchange rate"))
		return
	}

	if h.currencyService == nil {
		_ = c.Error(apierr.InternalServerError("Currency service is unavailable"))
		return
	}

	if err := h.currencyService.UpdateRates(c.Request.Context(), rates); err != nil {
		_ = c.Error(err)
		return
	}

	currencies, err := h.currencyService.GetCurrencies(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
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
