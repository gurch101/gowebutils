package collectionutils

// Contains checks if any element in the collection satisfies the predicate.
func Contains[T any](collection []T, predicate func(T) bool) bool {
	for _, item := range collection {
		if predicate(item) {
			return true
		}
	}

	return false
}
