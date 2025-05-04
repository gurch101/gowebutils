package collectionutils

// ContainsAll checks if all elements in the collection satisfy the predicate.
func ContainsAll[T any](collection []T, predicate func(T) bool) bool {
	for _, item := range collection {
		if !predicate(item) {
			return false
		}
	}

	return true
}
