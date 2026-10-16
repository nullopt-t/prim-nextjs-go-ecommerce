// Package utils provides utility functions for pricing, currency representation, and formatting.
package utils

import (
	"fmt"
	"strings"
)

// currencySymbols maps ISO 4217 uppercase currency codes to their standard symbol representations.
var currencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"CAD": "CA$",
	"AUD": "AU$",
	"JPY": "¥",
	"EGP": "E£",
}

// CurrencyDecimals returns the number of decimal digits (exponent) for a given ISO 4217 currency.
// Most currencies have 2 decimal digits. Zero-decimal currencies (JPY, KRW, VND, CLP, etc.) have 0.
// Three-decimal currencies (BHD, KWD, OMR, TND, etc.) have 3.
func CurrencyDecimals(currency string) int {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "JPY", "KRW", "VND", "CLP", "BIF", "DJF", "GNF", "ISK", "KMF", "PYG", "RWF", "UGX":
		return 0
	case "BHD", "KWD", "OMR", "TND", "IQD", "JOD", "LYD":
		return 3
	default:
		return 2
	}
}

// currencyDivisor returns the divisor corresponding to the currency decimal places.
func currencyDivisor(currency string) float64 {
	dec := CurrencyDecimals(currency)
	switch dec {
	case 0:
		return 1.0
	case 1:
		return 10.0
	case 3:
		return 1000.0
	case 4:
		return 10000.0
	default:
		return 100.0
	}
}

// FormatPrice formats an integer price in minor units into a display string with currency symbol.
//
// Parameters:
//   - cents: The price amount represented in minor units (e.g. 2499 for $24.99, or 1000 for ¥1000).
//   - currency: The ISO currency code (e.g. "USD", "EUR", "JPY"). If empty, defaults to "USD".
//
// Returns:
//   - string: Formatted price string with currency symbol.
func FormatPrice(cents int64, currency string) string {
	// Normalize currency string by trimming whitespace and capitalizing.
	currUpper := strings.ToUpper(strings.TrimSpace(currency))
	// Default to EGP if no currency string is provided.
	if currUpper == "" {
		currUpper = "EGP"
	}
	// Look up currency symbol in predefined symbol table.
	symbol, ok := currencySymbols[currUpper]
	// Fallback to "E£" if currency is unknown.
	if !ok {
		symbol = "E£"
	}

	divisor := currencyDivisor(currUpper)
	val := float64(cents) / divisor
	decimals := CurrencyDecimals(currUpper)

	return fmt.Sprintf("%s%.*f", symbol, decimals, val)
}

// ExtractPrice converts an integer price in minor units (cents) to a float64 major currency value (e.g. 2499 -> 24.99).
// Assumes standard 2-decimal scale for backward compatibility.
//
// Parameters:
//   - cents: The price amount represented in minor units.
//
// Returns:
//   - float64: The amount in major currency units.
func ExtractPrice(cents int64) float64 {
	// Divide cents by 100.0 to produce major units.
	return float64(cents) / 100.0
}

// ExtractPriceForCurrency converts an integer price in minor units to a float64 major currency value
// taking the currency's decimal places into account (e.g. JPY 100 -> 100.0, KWD 1000 -> 1.0, USD 2499 -> 24.99).
func ExtractPriceForCurrency(cents int64, currency string) float64 {
	divisor := currencyDivisor(currency)
	return float64(cents) / divisor
}

