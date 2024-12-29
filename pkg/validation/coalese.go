package validation

// Coalesce returns the first non-nil value from the given values.
func Coalesce[T any](ptr *T, defaultVal T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultVal
}
