package collectionutils

// Equals checks if two slices are equal.
//
//nolint:varnamelen
func Equals[T comparable](a, b []T) bool {
	setA := make(map[T]struct{})
	setB := make(map[T]struct{})

	for _, v := range a {
		setA[v] = struct{}{}
	}

	for _, v := range b {
		setB[v] = struct{}{}
	}

	if len(setA) != len(setB) {
		return false
	}

	for k := range setA {
		if _, ok := setB[k]; !ok {
			return false
		}
	}

	return true
}
