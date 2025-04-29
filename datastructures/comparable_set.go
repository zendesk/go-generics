package datastructures

import (
	"fmt"
	"iter"
)

func newComparableSet[V comparable]() ISet[V] {
	return &comparableSet[V]{
		values: make(map[V]struct{}),
		order:  make([]V, 0),
	}
}

// comparableSet is an internal concrete implementation of a set using a map for storage.
type comparableSet[V comparable] struct {
	values map[V]struct{}
	order  []V // maintains the order of insertion
}

// Put adds 'val' to the comparableSet if it is not already present.
func (s *comparableSet[V]) Put(k V) {
	if _, ok := s.values[k]; !ok {
		s.values[k] = struct{}{}
		s.order = append(s.order, k)
	}
}

// Has returns true only if 'val' is in the comparableSet.
func (s *comparableSet[V]) Has(k V) bool {
	_, ok := s.values[k]
	return ok
}

// Remove removes 'val' from the comparableSet.
func (s *comparableSet[V]) Remove(k V) {
	if _, ok := s.values[k]; ok {
		delete(s.values, k)
		for i, v := range s.order {
			if v == k {
				s.order = append(s.order[:i], s.order[i+1:]...)
				break
			}
		}
	}
}

// Values returns all elements in the comparableSet.
func (s *comparableSet[V]) Values() []V {
	//out := make([]V, len(s.order))
	//copy(out, s.order)
	return s.order
}

// Clear removes all elements from the comparableSet.
func (s *comparableSet[V]) Clear() {
	s.values = make(map[V]struct{})
	s.order = make([]V, 0)
}

// Size returns the number of elements in the comparableSet.
func (s *comparableSet[V]) Size() int {
	return len(s.values)
}

// Copy returns a copy of this comparableSet.
func (s *comparableSet[V]) Copy() ISet[V] {
	return &comparableSet[V]{
		values: s.copyItems(),
		order:  append([]V{}, s.order...),
	}
}

// New returns an empty SetOf[V].
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

// All returns an iterator that yields all elements in the comparableSet in the order they were inserted.
func (s *comparableSet[V]) All() iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		for i, val := range s.order {
			if !yield(i, val) {
				break
			}
		}
	}
}

// String returns a string representation of the comparableSet.
func (s *comparableSet[V]) String() string {
	str := ""
	for i, val := range s.order {
		if i > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%v", val)
	}
	return str
}

// ensure at compile time this satisfies the ISet[V] interface
var _ ISet[int] = &comparableSet[int]{}
