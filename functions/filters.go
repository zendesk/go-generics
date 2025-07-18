package functions

// Find retrieves the first element in a slice matching the filter. Nil of type T is returned if nothing matches
func Find[T interface{}](from []T, filter func(T) bool) (item T, wasFound bool) {
	wasFound = false
	for _, t := range from {
		if filter(t) {
			return t, true
		}
	}

	return
}

// Filter removes any items from []T that _do not_ match the provided filter function.
func Filter[T any](from []T, filter func(T) bool) []T {
	var remaining []T
	for _, t := range from {
		if filter(t) {
			remaining = append(remaining, t)
		}
	}

	return remaining
}

// FilterMap removes any items from map[K]V that _do not_ match the provided filter function.
func FilterMap[K comparable, V any](from map[K]V, filter func(k K, v V) bool) map[K]V {
	result := make(map[K]V)

	for k, v := range from {
		if filter(k, v) {
			result[k] = v
		}
	}

	return result
}
