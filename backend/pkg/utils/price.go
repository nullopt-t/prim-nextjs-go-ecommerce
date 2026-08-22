package utils

import (
	"fmt"
	"strings"
)

var currencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"CAD": "CA$",
	"AUD": "AU$",
	"JPY": "¥",
	"EGP": "E£",
}

// FormatPrice formats an integer price in minor units (cents) into a display string with currency symbol (e.g. "$24.99").
func FormatPrice(cents int64, currency string) string {
	currUpper := strings.ToUpper(strings.TrimSpace(currency))
	if currUpper == "" {
		currUpper = "USD"
	}
	symbol, ok := currencySymbols[currUpper]
	if !ok {
		symbol = "$"
	}
	val := float64(cents) / 100.0
	return fmt.Sprintf("%s%.2f", symbol, val)
}

// ExtractPrice converts an integer price in minor units (cents) to a float64 major currency value (e.g. 2499 -> 24.99).
func ExtractPrice(cents int64) float64 {
	return float64(cents) / 100.0
}
