package utils

import (
	"os"
	"strconv"
)

func GetEnvAsInt(name string, defaultVal int) int {
	if value, exists := os.LookupEnv(name); exists {
		if val, err := strconv.Atoi(value); err == nil {
			return val
		}
	}
	return defaultVal
}

func GetEnvAsBool(name string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(name); exists {
		if val, err := strconv.ParseBool(value); err == nil {
			return val
		}
	}
	return defaultVal
}

func GetEnvOrDefault(name string, defaultVal string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}
	return defaultVal
}
