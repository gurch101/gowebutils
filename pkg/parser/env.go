package parser

import (
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
	return strconv.Atoi(val)
}

func ParseEnvIntPanic(key string) int {
	val := os.Getenv(key)
	if val == "" {
		panic("missing required env var: " + key)
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	return intVal
}
