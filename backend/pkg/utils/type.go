// Package utils provides common type parsing, conversion, and validation functions.
package utils

import (
	"errors"
	"math"
	"strconv"

	"github.com/google/uuid"
)

// IsValidUUID checks if the given string is a valid non-nil RFC 4122 UUID format (UUIDv1-v7).
//
// Parameters:
//   - s: The string to test.
//
// Returns:
//   - bool: true if s is a valid, non-nil UUID, false otherwise.
func IsValidUUID(s string) bool {
	parsed, err := uuid.Parse(s)
	return err == nil && parsed != uuid.Nil
}

// IsValidNonNilUUID checks if the given string is a valid non-nil RFC 4122 UUID.
func IsValidNonNilUUID(s string) bool {
	return IsValidUUID(s)
}

// IsValidAnyUUID checks if the given string is any valid RFC 4122 UUID, including the nil UUID.
func IsValidAnyUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// StringToInt converts a string to an integer, returning defaultVal if parsing fails.
//
// Parameters:
//   - s: The string to parse.
//   - defaultVal: The integer fallback value.
//
// Returns:
//   - int: Parsed integer or defaultVal.
func StringToInt(s string, defaultVal int) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return num
}

// ParseInt parses a string into an integer, returning an error if parsing fails.
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// StringToFloat converts a string to a 64-bit floating point number, returning defaultVal on failure.
// NaN and infinite values are rejected and fall back to defaultVal.
//
// Parameters:
//   - s: The string to parse.
//   - defaultVal: The fallback float64 value.
//
// Returns:
//   - float64: Parsed float64 or defaultVal.
func StringToFloat(s string, defaultVal float64) float64 {
	num, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(num) || math.IsInf(num, 0) {
		return defaultVal
	}
	return num
}

// ParseFloat parses a string into a float64, returning an error on failure or if the value is NaN or Inf.
func ParseFloat(s string) (float64, error) {
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(num) || math.IsInf(num, 0) {
		return 0, errors.New("value cannot be NaN or Inf")
	}
	return num, nil
}

// StringToBool converts a string representation of a boolean to a bool, returning defaultVal on failure.
//
// Parameters:
//   - s: The string to parse (e.g., "true", "false", "1", "0").
//   - defaultVal: The fallback bool value.
//
// Returns:
//   - bool: Parsed boolean value or defaultVal.
func StringToBool(s string, defaultVal bool) bool {
	val, err := strconv.ParseBool(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// ParseBool parses a string representation of a boolean to a bool, returning an error if parsing fails.
func ParseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}

// IntToString converts an integer to its decimal string representation.
//
// Parameters:
//   - i: The integer value to convert.
//
// Returns:
//   - string: Decimal string representation of i.
func IntToString(i int) string {
	return strconv.Itoa(i)
}

