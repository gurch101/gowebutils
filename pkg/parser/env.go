package parser

import (
	"fmt"
	"os"
	"strconv"
)

func ParseEnvString(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	return val
}

func ParseEnvStringPanic(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("missing required env var: " + key)
	}

	return val
}

func ParseEnvInt(key string, defaultValue int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue, nil
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to parse env var %s as int: %w", key, err)
	}

	return intVal, nil
}

func ParseEnvIntPanic(key string) int {
	val := os.Getenv(key)
	if val == "" {
		panic("missing required env var: " + key)
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Errorf("failed to parse env var %s as int: %w", key, err))
	}

	return intVal
}

func ParseEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	return val == "true"
}

func ParseEnvFloat64(key string, defaultValue float64) (float64, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue, nil
	}

	floatVal, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse env var %s as float64: %w", key, err)
	}

	return floatVal, nil
}
