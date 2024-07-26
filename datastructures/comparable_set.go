package datastructures

func newComparableSet[V comparable]() ISet[V] {
	return &comparableSet[V]{values: make(map[V]struct{})}
}

// InternalSet concrete implementation.
type comparableSet[V comparable] struct {
	values map[V]struct{}
}

// Put adds 'val' to the comparableSet.
func (s *comparableSet[V]) Put(k V) {
	s.values[k] = struct{}{}
}

// Has returns true only if 'val' is in the comparableSet.
func (s *comparableSet[V]) Has(k V) bool {
	_, ok := s.values[k]
	return ok
}

// Remove removes 'val' from the comparableSet.
func (s *comparableSet[V]) Remove(k V) {
	delete(s.values, k)
}

// Values returns all elements in the comparableSet.
func (s *comparableSet[V]) Values() []V {
	out := make([]V, len(s.values))
	i := 0
	for val := range s.values {
		out[i] = val
		i++
	}

	return out
}

// Clear removes all elements from the comparableSet.
func (s *comparableSet[V]) Clear() {
	s.values = make(map[V]struct{})
}

// Size returns the number of elements in the comparableSet.
func (s *comparableSet[V]) Size() int {
	return len(s.values)
}

// Copy returns a copy of this comparableSet.
func (s *comparableSet[V]) Copy() ISet[V] {
	return &comparableSet[V]{
		values: s.copyItems(),
	}
}

// New returns an empty SetOf[V]
func (s *comparableSet[V]) New() ISet[V] {
	return newComparableSet[V]()
}

func (s *comparableSet[V]) copyItems() map[V]struct{} {
	values := make(map[V]struct{})
	for k, v := range s.values {
		values[k] = v
	}
	return values
}
