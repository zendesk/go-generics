package datastructures

import (
	"crypto/sha256"
	"fmt"
)

type HashFn[T any] func(t T) string

func newHashSet[V any]() ISet[V] {
	return &hashSet[V]{
		values: make(map[string]V),
		hashFn: hashAny[V],
	}
}

func newHashSetWithHashFn[V any](fn HashFn[V]) ISet[V] {
	return &hashSet[V]{
		values: make(map[string]V),
		hashFn: fn,
	}
}

// InternalSet concrete implementation.
type hashSet[V any] struct {
	values map[string]V
	hashFn HashFn[V]
}

// Put adds 'val' to the hashSet.
func (s *hashSet[V]) Put(v V) {
	s.values[s.hashFn(v)] = v
}

// Has returns true only if 'val' is in the hashSet.
func (s *hashSet[V]) Has(v V) bool {
	_, ok := s.values[s.hashFn(v)]
	return ok
}

// Remove removes 'val' from the hashSet.
func (s *hashSet[V]) Remove(v V) {
	delete(s.values, s.hashFn(v))
}

func (s *hashSet[V]) Values() []V {
	out := make([]V, len(s.values))
	i := 0
	for _, val := range s.values {
		out[i] = val
		i++
	}
	return out
}

// Clear removes all elements from the hashSet.
func (s *hashSet[V]) Clear() {
	s.values = make(map[string]V)
}

// Size returns the number of elements in the hashSet.
func (s *hashSet[V]) Size() int {
	return len(s.values)
}

// Copy returns a copy of this hashSet.
func (s *hashSet[V]) Copy() ISet[V] {
	return &hashSet[V]{
		values: s.copyItems(),
		hashFn: s.hashFn,
	}
}

// New returns an empty SetOf[V]
func (s *hashSet[V]) New() ISet[V] {
	return newHashSet[V]()
}

func (s *hashSet[V]) copyItems() map[string]V {
	values := make(map[string]V)
	for k, v := range s.values {
		values[k] = v
	}
	return values
}

func hashAny[V any](v V) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", v)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
