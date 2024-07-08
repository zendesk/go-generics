package functions

import (
	"cmp"
)

// Min returns the least of the provided values
func Min[T cmp.Ordered](values ...T) T {
	var zero T
	if len(values) == 0 {
		return zero
	}

	min := values[0]
	if len(values) > 1 {
		for _, val := range values {
			if val < min {
				min = val
			}
		}
	}

	return min
}

// Max returns the largest of the provided values
func Max[T cmp.Ordered](values ...T) T {
	var zero T
	if len(values) == 0 {
		return zero
	}

	max := values[0]
	if len(values) > 1 {
		for _, val := range values {
			if val > max {
				max = val
			}
		}
	}
	return max
}

// Copy deep copies a map to a new map of the same type. This does not mutate the original
func Copy[K comparable, V any](items map[K]V) map[K]V {
	newMap := make(map[K]V)
	for k, v := range items {
		newMap[k] = v
	}

	return newMap
}

// Convert converts T to Y via a provided converter function
func Convert[T any, Y any](from T, converter func(T) Y) Y {
	return converter(from)
}
