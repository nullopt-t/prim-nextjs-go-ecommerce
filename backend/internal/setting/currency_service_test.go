package setting

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyService_GetCurrencies_Defaults(t *testing.T) {
	runner := &fakeRunner{}
	repo := NewRepository()
	svc := NewCurrencyService(runner, repo)

	currencies, err := svc.GetCurrencies(context.Background())
	require.NoError(t, err)
	require.Len(t, currencies, 5)

	var egp CurrencyInfo
	for _, c := range currencies {
		if c.Code == "EGP" {
			egp = c
		}
	}
	assert.Equal(t, "EGP", egp.Code)
	assert.Equal(t, 1.0, egp.RateToBase)
	assert.True(t, egp.IsBase)
}

func TestCurrencyService_GetCurrencies_WithCustomRates(t *testing.T) {
	customRates := map[string]float64{
		"USD": 50.0,
		"EUR": 55.0,
	}

	runner := &fakeRunner{
		items: []Item{
			{
				Key:       SettingKeyCurrencyRates,
				Value:     customRates,
				UpdatedAt: time.Now(),
			},
		},
	}
	repo := NewRepository()
	svc := NewCurrencyService(runner, repo)

	currencies, err := svc.GetCurrencies(context.Background())
	require.NoError(t, err)

	for _, c := range currencies {
		if c.Code == "USD" {
			assert.Equal(t, 50.0, c.RateToEGP)
			assert.Equal(t, 50.0, c.RateToBase)
		} else if c.Code == "EUR" {
			assert.Equal(t, 55.0, c.RateToEGP)
			assert.Equal(t, 55.0, c.RateToBase)
		} else if c.Code == "SAR" {
			assert.Equal(t, 12.9, c.RateToBase) // default preserved
		}
	}

	// Verify caching works
	cached, ok := svc.getCached()
	assert.True(t, ok)
	assert.Len(t, cached, 5)
}

func TestCurrencyService_Convert(t *testing.T) {
	runner := &fakeRunner{}
	repo := NewRepository()
	svc := NewCurrencyService(runner, repo)
	ctx := context.Background()

	// Same currency
	val, err := svc.Convert(ctx, 100, "USD", "USD")
	require.NoError(t, err)
	assert.Equal(t, 100.0, val)

	// USD to EGP: 10 USD * 48.5 = 485 EGP
	val, err = svc.Convert(ctx, 10, "USD", "EGP")
	require.NoError(t, err)
	assert.InDelta(t, 485.0, val, 0.001)

	// EGP to USD: 485 EGP / 48.5 = 10 USD
	val, err = svc.Convert(ctx, 485, "EGP", "USD")
	require.NoError(t, err)
	assert.InDelta(t, 10.0, val, 0.001)

	// Cross currency: 100 USD (4850 EGP) / 52.0 EUR
	val, err = svc.Convert(ctx, 100, "USD", "EUR")
	require.NoError(t, err)
	expected := (100.0 * 48.5) / 52.0
	assert.InDelta(t, expected, val, 0.001)

	// Case insensitivity
	valLower, err := svc.Convert(ctx, 10, "usd", "egp")
	require.NoError(t, err)
	assert.InDelta(t, 485.0, valLower, 0.001)

	// Unsupported source currency
	_, err = svc.Convert(ctx, 10, "XYZ", "USD")
	assert.Error(t, err)

	// Unsupported target currency
	_, err = svc.Convert(ctx, 10, "USD", "XYZ")
	assert.Error(t, err)
}

func TestCurrencyService_UpdateRates_Validation(t *testing.T) {
	runner := &fakeRunner{}
	repo := NewRepository()
	svc := NewCurrencyService(runner, repo)
	ctx := context.Background()

	// Empty payload
	err := svc.UpdateRates(ctx, map[string]float64{})
	assert.Error(t, err)

	// Zero rate
	err = svc.UpdateRates(ctx, map[string]float64{"USD": 0})
	assert.Error(t, err)

	// Negative rate
	err = svc.UpdateRates(ctx, map[string]float64{"USD": -5.0})
	assert.Error(t, err)

	// Valid rate
	err = svc.UpdateRates(ctx, map[string]float64{"USD": 49.5})
	assert.NoError(t, err)
}

func TestHandler_UpdateCurrencies_Endpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	runner := &fakeRunner{}
	repo := NewRepository()
	currencySvc := NewCurrencyService(runner, repo)
	settingSvc := NewService(runner, repo)
	h := NewHandler(settingSvc, currencySvc)

	r := gin.New()
	r.PUT("/api/v1/admin/settings/currencies", h.UpdateCurrencies)

	payload := map[string]interface{}{
		"rates": map[string]float64{
			"USD": 51.5,
		},
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/settings/currencies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
