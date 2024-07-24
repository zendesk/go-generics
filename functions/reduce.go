package functions

// Reduce converts []T to a single Y via a reducer function. `to` is provided to each iteration of
// reducer and will be returned at the end.
func Reduce[T any, Y any](from []T, to Y, reducer func(T, Y) Y) Y {
	for _, t := range from {
		to = reducer(t, to)
	}
	return to
}
