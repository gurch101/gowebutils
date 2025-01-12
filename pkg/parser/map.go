package parser

import "errors"

var ErrInvalidMapKey = errors.New("invalid key")

// ParseJSONMapInt64 parses a map[string]any and returns the value as an int64.
// Since json.Unmarshal() converts all numbers to float64, we need to convert them back to int64.
// Caller should ensure that castint from float64 to int64 is safe.
func ParseJSONMapInt64(m map[string]any, key string) (int64, error) {
	value, ok := m[key].(float64)
	if !ok {
		return 0, ErrInvalidMapKey
	}

	return int64(value), nil
}

// ParseJSONMapString parses a map[string]any and returns the value as a string.
func ParseJSONMapString(m map[string]any, key string) (string, error) {
	value, ok := m[key].(string)
	if !ok {
		return "", ErrInvalidMapKey
	}

	return value, nil
}
