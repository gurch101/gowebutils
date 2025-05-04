package collectionutils

// FindFirst returns the first element in the slice for which the predicate returns true.
// If no element satisfies the predicate, it returns the zero value of the type and false.
func FindFirst[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}

	var zero T

	return zero, false
}
