package datastructures

import (
	"crypto/sha256"
	"fmt"
	"iter"
)

// HashFn is a type for hashing function that converts a value of any type to a string.
type HashFn[T any] func(t T) string

// newHashSet creates a new hashSet with a default hashing function.
func newHashSet[V any]() ISet[V] {
	return &hashSet[V]{
		values: make(map[string]V),
		order:  make([]string, 0),
		hashFn: hashAny[V],
	}
}

// newHashSetWithHashFn creates a new hashSet with a given hashing function.
func newHashSetWithHashFn[V any](fn HashFn[V]) ISet[V] {
	return &hashSet[V]{
		values: make(map[string]V),
		order:  make([]string, 0),
		hashFn: fn,
	}
}

// hashSet is an internal concrete implementation of a set using a map for storage.
type hashSet[V any] struct {
	values map[string]V
	order  []string // maintains the order of insertion
	hashFn HashFn[V]
}

// Put adds 'v' to the hashSet if it is not already present.
func (s *hashSet[V]) Put(v V) {
	key := s.hashFn(v)
	if _, ok := s.values[key]; !ok {
		s.values[key] = v
		s.order = append(s.order, key)
	}
}

// Has returns true if 'v' is in the hashSet.
func (s *hashSet[V]) Has(v V) bool {
	_, ok := s.values[s.hashFn(v)]
	return ok
}

// Remove removes 'v' from the hashSet.
func (s *hashSet[V]) Remove(v V) {
	key := s.hashFn(v)
	if _, ok := s.values[key]; ok {
		delete(s.values, key)
		for i, k := range s.order {
			if k == key {
				s.order = append(s.order[:i], s.order[i+1:]...)
				break
			}
		}
	}
}

// Values returns a slice containing all elements in the hashSet.
func (s *hashSet[V]) Values() []V {
	out := make([]V, len(s.values))
	i := 0
	for _, key := range s.order {
		out[i] = s.values[key]
		i++
	}
	return out
}

// Clear removes all elements from the hashSet.
func (s *hashSet[V]) Clear() {
	s.values = make(map[string]V)
	s.order = make([]string, 0)
}

// Size returns the number of elements in the hashSet.
func (s *hashSet[V]) Size() int {
	return len(s.values)
}

// Copy returns a copy of the current hashSet.
func (s *hashSet[V]) Copy() ISet[V] {
	return &hashSet[V]{
		values: s.copyItems(),
		order:  append([]string{}, s.order...), // Copy the order slice
		hashFn: s.hashFn,
	}
}

// New returns an empty instance of the hashSet.
func (s *hashSet[V]) New() ISet[V] {
	return newHashSet[V]()
}

// copyItems creates a copy of the internal map of the hashSet.
func (s *hashSet[V]) copyItems() map[string]V {
	values := make(map[string]V)
	for k, v := range s.values {
		values[k] = v
	}
	return values
}

// hashAny is the default hash function that generates a SHA-256 hash of the string representation of the value.
func hashAny[V any](v V) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", v)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// All returns an iterator that yields all elements in the hashSet in the order they were inserted.
func (s *hashSet[V]) All() iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		for i, key := range s.order {
			if !yield(i, s.values[key]) {
				break
			}
		}
	}
}
