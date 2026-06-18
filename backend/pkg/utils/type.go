package utils

import (
	"strconv"

	"github.com/google/uuid"
)

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func StringToInt(s string, defaultVal int) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return num
}

func StringToFloat(s string, defaultVal float64) float64 {
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return num
}

func StringToBool(s string, defaultVal bool) bool {
	val, err := strconv.ParseBool(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func IntToString(i int, defaultVal string) string {
	return strconv.Itoa(i)
}
