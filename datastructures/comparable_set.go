package datastructures

func newComparableSet[V comparable]() ISet[V] {
	return &set[V]{values: make(map[V]struct{})}
}

// InternalSet concrete implementation.
type set[V comparable] struct {
	values map[V]struct{}
}

// Put adds 'val' to the set.
func (s *set[V]) Put(k V) {
	s.values[k] = struct{}{}
}

// Has returns true only if 'val' is in the set.
func (s *set[V]) Has(k V) bool {
	_, ok := s.values[k]
	return ok
}

// Remove removes 'val' from the set.
func (s *set[V]) Remove(k V) {
	delete(s.values, k)
}

// Values returns all elements in the set.
func (s *set[V]) Values() []V {
	out := make([]V, len(s.values))
	i := 0
	for val := range s.values {
		out[i] = val
		i++
	}

	return out
}

// Clear removes all elements from the set.
func (s *set[V]) Clear() {
	s.values = make(map[V]struct{})
}

// Size returns the number of elements in the set.
func (s *set[V]) Size() int {
	return len(s.values)
}

// Copy returns a copy of this set.
func (s *set[V]) Copy() ISet[V] {
	return &set[V]{
		values: s.copyItems(),
	}
}

// New returns an empty SetOf[V]
func (s *set[V]) New() ISet[V] {
	return newComparableSet[V]()
}

func (s *set[V]) copyItems() map[V]struct{} {
	values := make(map[V]struct{})
	for k, v := range s.values {
		values[k] = v
	}
	return values
}
