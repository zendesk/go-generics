package datastructures

// A simple set implementation with comparable keys. Use this instead of manually managing your own map[V]struct{}
// Inspired by https://github.com/zyedidia/generic
// For more complex uses (hashset, mapset), try: https://github.com/zyedidia/generic

// ISet is the internal representation of a set, you may implement your own if desired.
type ISet[V any] interface {
	Has(k V) bool
	Remove(k V)
	Put(k V)
	Clear()
	Copy() ISet[V]
	Size() int
	Values() []V
	New() ISet[V]
}

// Set is the public interface for a set
type Set[V any] interface {
	Values() []V
	Keys() []V
	Equal(to Set[V]) bool
	Has(k V) bool
	Remove(k V)
	Put(k V)
	Clear()
	Size() int
	Clone() Set[V]
	Intersection(others ...Set[V]) Set[V]
	Difference(others ...Set[V]) Set[V]
	Union(others ...Set[V]) Set[V]
	Each(fn func(v V))
	EachWithErrs(fn func(key V) error) []error
	SymmetricDifference(others ...Set[V]) Set[V]
	InPlaceIntersection(others ...Set[V]) Set[V]
	InPlaceDifference(others ...Set[V]) Set[V]
	InPlaceUnion(others ...Set[V]) Set[V]
	IsDisjoint(other Set[V]) bool
	IsSubset(of Set[V]) bool
	IsSuperset(of Set[V]) bool
	IsProperSubset(to Set[V]) bool
	IsProperSuperset(to Set[V]) bool
}

type setOf[V any] struct {
	set ISet[V]
}

// NewSet creates a new set with the provided values. Supports only comparable types.
func NewSet[V comparable](values ...V) Set[V] {
	s := setOf[V]{set: newComparableSet[V]()}

	for _, val := range values {
		s.Put(val)
	}

	return s
}

// NewHashSet creates a new set with the provided values. Supports any type.
func NewHashSet[V any](values ...V) Set[V] {
	s := setOf[V]{set: newHashSet[V]()}

	for _, val := range values {
		s.Put(val)
	}

	return s
}

func NewHashSetWithHashFn[V any](hashFn HashFn[V], values ...V) Set[V] {
	s := setOf[V]{set: newHashSetWithHashFn[V](hashFn)}

	for _, val := range values {
		s.Put(val)
	}

	return s
}

// NewCustomSet creates a new set with the provided ISet implementation
func NewCustomSet[V any](set ISet[V]) Set[V] {
	return setOf[V]{set: set}
}

// Intersection returns a new set with the intersection of the current set and others
func (s setOf[V]) Intersection(others ...Set[V]) Set[V] {
	return s.Clone().InPlaceIntersection(others...)
}

// Difference returns a new set with the difference of the current set and others
func (s setOf[V]) Difference(others ...Set[V]) Set[V] {
	return s.Clone().InPlaceDifference(others...)
}

// Union returns a new set with the union of the current set and others
func (s setOf[V]) Union(others ...Set[V]) Set[V] {
	return s.Clone().InPlaceUnion(others...)
}

// Each calls 'fn' on every item in the set in no particular order.
func (s setOf[V]) Each(fn func(v V)) {
	for _, val := range s.Values() {
		fn(val)
	}
}

// EachWithErrs calls 'fn' on every item in the set in no particular order. All errors are aggregated and returned
func (s setOf[V]) EachWithErrs(fn func(key V) error) []error {
	var errs []error
	for _, key := range s.Values() {
		err := fn(key)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Has returns true only if 'val' is in the set.
func (s setOf[V]) Has(k V) bool {
	return s.set.Has(k)
}

// Remove removes 'val' from the set.
func (s setOf[V]) Remove(k V) {
	s.set.Remove(k)
}

// Put adds 'val' to the set.
func (s setOf[V]) Put(k V) {
	s.set.Put(k)
}

// Clear removes all elements from the set.
func (s setOf[V]) Clear() {
	s.set.Clear()
}

// Size returns the number of elements in the set.
func (s setOf[V]) Size() int {
	return s.set.Size()
}

// Clone returns a copy of this set.
func (s setOf[V]) Clone() Set[V] {
	new := setOf[V]{set: s.set.New()}
	new.set = s.set.Copy()
	return new
}

// SymmetricDifference returns a new set with the symmetric difference of the current set and others
func (s setOf[V]) SymmetricDifference(others ...Set[V]) Set[V] {
	new := s.Clone()
	seen := new.Clone()
	for _, other := range others {
		other.Each(func(key V) {
			if seen.Has(key) {
				new.Remove(key)
				return
			}
			new.Put(key)
			seen.Put(key)
		})
	}
	return new
}

// InPlaceIntersection will modify the current set in place to become the intersection with others
func (s setOf[V]) InPlaceIntersection(others ...Set[V]) Set[V] {
	for _, other := range others {
		s.Each(func(key V) {
			if !other.Has(key) {
				s.Remove(key)
			}
		})
	}
	return s
}

// InPlaceDifference will modify the current set in place to become the difference with others
func (s setOf[V]) InPlaceDifference(others ...Set[V]) Set[V] {
	for _, other := range others {
		other.Each(func(key V) {
			s.Remove(key)
		})
	}
	return s
}

// InPlaceUnion will modify the current set in place to become the union with others
func (s setOf[V]) InPlaceUnion(others ...Set[V]) Set[V] {
	for _, other := range others {
		other.Each(func(key V) {
			s.Put(key)
		})
	}
	return s
}

func (s setOf[V]) Values() []V {
	return s.set.Values()
}

func (s setOf[V]) Keys() []V {
	return s.Values()
}

// IsDisjoint returns true if the current set has no elements in common with others
func (s setOf[V]) IsDisjoint(other Set[V]) bool {
	return s.Intersection(other).Size() == 0
}

// IsSubset returns true if the current set is a subset of others
func (s setOf[V]) IsSubset(of Set[V]) bool {
	subset := true
	s.Each(func(key V) {
		if !of.Has(key) {
			subset = false
		}
	})
	return subset
}

// IsSuperset returns true if the current set is a superset of others
func (s setOf[V]) IsSuperset(of Set[V]) bool {
	superset := true
	of.Each(func(key V) {
		if !s.Has(key) {
			superset = false
		}
	})
	return superset
}

// Equal returns true if the current set is equal to others
func (s setOf[V]) Equal(to Set[V]) bool {
	if s.Size() != to.Size() {
		return false
	}
	return s.Union(to).Size() == s.Size()
}

// IsProperSubset returns true if the current set is a proper subset of others
func (s setOf[V]) IsProperSubset(to Set[V]) bool {
	if s.Equal(to) {
		return false
	}
	return s.IsSubset(to)
}

// IsProperSuperset returns true if the current set is a proper superset of others
func (s setOf[V]) IsProperSuperset(to Set[V]) bool {
	if s.Equal(to) {
		return false
	}
	return s.IsSuperset(to)
}
