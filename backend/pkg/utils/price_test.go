package utils_test

import (
	"testing"

	"github.com/nullopt-t/prim-backend/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		cents    int64
		currency string
		expected string
	}{
		{2499, "USD", "$24.99"},
		{100, "USD", "$1.00"},
		{0, "USD", "$0.00"},
		{129999, "USD", "$1299.99"},
		{4990, "EUR", "€49.90"},
		{3500, "GBP", "£35.00"},
		{1500, "EGP", "E£15.00"},
		{1999, "", "$19.99"},
		{100, "JPY", "¥100"},
		{1500, "KWD", "$1.500"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			res := utils.FormatPrice(tt.cents, tt.currency)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestExtractPrice(t *testing.T) {
	tests := []struct {
		cents    int64
		expected float64
	}{
		{2499, 24.99},
		{100, 1.00},
		{0, 0.0},
		{129999, 1299.99},
		{50, 0.50},
	}

	for _, tt := range tests {
		t.Run(utils.FormatPrice(tt.cents, "USD"), func(t *testing.T) {
			res := utils.ExtractPrice(tt.cents)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestExtractPriceForCurrency(t *testing.T) {
	assert.Equal(t, 24.99, utils.ExtractPriceForCurrency(2499, "USD"))
	assert.Equal(t, 100.0, utils.ExtractPriceForCurrency(100, "JPY"))
	assert.Equal(t, 1.5, utils.ExtractPriceForCurrency(1500, "KWD"))
}
