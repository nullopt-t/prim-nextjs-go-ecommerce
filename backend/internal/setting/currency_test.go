package setting

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCurrencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{}

	r := gin.New()
	r.GET("/api/v1/settings/currencies", h.GetCurrencies)

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/v1/settings/currencies", nil)
	require.NoError(t, err)

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []CurrencyInfo `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.NotEmpty(t, resp.Data)
	codes := make([]string, 0, len(resp.Data))
	for _, c := range resp.Data {
		codes = append(codes, c.Code)
		assert.NotEmpty(t, c.Symbol)
		assert.Greater(t, c.RateToEGP, 0.0)
	}
	assert.Contains(t, codes, "EGP")
	assert.Contains(t, codes, "USD")
	assert.Contains(t, codes, "EUR")
}
