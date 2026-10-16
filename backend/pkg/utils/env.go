package utils

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// GetEnvAsInt retrieves an environment variable by name and parses it as an integer.
// If the environment variable does not exist, is empty, or fails integer parsing,
// a warning is logged and defaultVal is returned.
//
// Parameters:
//   - name: The key of the environment variable.
//   - defaultVal: The integer value to return if lookup or parsing fails.
//
// Returns:
//   - int: Parsed integer or defaultVal.
func GetEnvAsInt(name string, defaultVal int) int {
	if value, exists := os.LookupEnv(name); exists {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			slog.Warn("environment variable is empty, using default value", slog.String("name", name), slog.Int("default", defaultVal))
			return defaultVal
		}

		val, err := strconv.ParseInt(trimmed, 10, 0)
		if err != nil {
			slog.Warn("invalid integer for environment variable, using default value",
				slog.String("name", name),
				slog.String("value", value),
				slog.Int("default", defaultVal),
				slog.Any("error", err),
			)
			return defaultVal
		}
		return int(val)
	}
	return defaultVal
}

// GetEnvAsBool retrieves an environment variable by name and parses it as a boolean.
// Accepts standard strconv.ParseBool representations (1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False)
// as well as common boolean flags (yes, no, on, off, enable, enabled, disable, disabled).
// If the environment variable does not exist, is empty, or fails boolean parsing,
// a warning is logged and defaultVal is returned.
//
// Parameters:
//   - name: The key of the environment variable.
//   - defaultVal: The boolean value to return if lookup or parsing fails.
//
// Returns:
//   - bool: Parsed boolean or defaultVal.
func GetEnvAsBool(name string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(name); exists {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" {
			slog.Warn("environment variable is empty, using default value", slog.String("name", name), slog.Bool("default", defaultVal))
			return defaultVal
		}

		switch trimmed {
		case "1", "t", "true", "yes", "y", "on", "enable", "enabled":
			return true
		case "0", "f", "false", "no", "n", "off", "disable", "disabled":
			return false
		default:
			slog.Warn("invalid boolean syntax for environment variable, using default value",
				slog.String("name", name),
				slog.String("value", value),
				slog.Bool("default", defaultVal),
			)
			return defaultVal
		}
	}
	return defaultVal
}

// GetEnvOrDefault retrieves an environment variable by name as a string, returning defaultVal
// if the environment variable is not defined or is an empty/whitespace string.
//
// Parameters:
//   - name: The key of the environment variable.
//   - defaultVal: The string value to return if the environment variable is not set or empty.
//
// Returns:
//   - string: The environment variable value or defaultVal.
func GetEnvOrDefault(name string, defaultVal string) string {
	if value, exists := os.LookupEnv(name); exists && strings.TrimSpace(value) != "" {
		return value
	}
	return defaultVal
}

// GetEnvNonEmptyOrDefault is an alias for GetEnvOrDefault ensuring non-empty configuration values.
func GetEnvNonEmptyOrDefault(name string, defaultVal string) string {
	return GetEnvOrDefault(name, defaultVal)
}
