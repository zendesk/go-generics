package types

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

type Set[V any] struct {
	set ISet[V]
}

func NewSet[V comparable]() Set[V] {
	return Set[V]{set: newComparableSet[V]()}
}

func NewCustomSet[V any](set ISet[V]) Set[V] {
	return Set[V]{set: set}
}

func (s Set[V]) Intersection(others ...Set[V]) Set[V] {
	return s.Clone().InPlaceIntersection(others...)
}
func (s Set[V]) Difference(others ...Set[V]) Set[V] {
	return s.Clone().InPlaceDifference(others...)
}
func (s Set[V]) Union(others ...Set[V]) Set[V] {
	return s.Clone().InPlaceUnion(others...)
}

// Each calls 'fn' on every item in the set in no particular order.
func (s Set[V]) Each(fn func(v V)) {
	for _, val := range s.Values() {
		fn(val)
	}
}

// EachWithErrs calls 'fn' on every item in the set in no particular order. All errors are aggregated and returned
func (s Set[V]) EachWithErrs(fn func(key V) error) []error {
	var errs []error
	for _, key := range s.Values() {
		err := fn(key)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (s Set[V]) Has(k V) bool {
	return s.set.Has(k)
}

func (s Set[V]) Remove(k V) {
	s.set.Remove(k)
}

// Put adds 'val' to the set.
func (s Set[V]) Put(k V) {
	s.set.Put(k)
}

func (s Set[V]) Clear() {
	s.set.Clear()
}

func (s Set[V]) Size() int {
	return s.Size()
}

func (s Set[V]) Clone() Set[V] {
	new := Set[V]{set: s.set.New()}
	new.set = s.set.Copy()
	return new
}

func (s Set[V]) SymmetricDifference(others ...Set[V]) Set[V] {
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

func (s Set[V]) InPlaceIntersection(others ...Set[V]) Set[V] {
	for _, other := range others {
		s.Each(func(key V) {
			if !other.Has(key) {
				s.Remove(key)
			}
		})
	}
	return s
}

func (s Set[V]) InPlaceDifference(others ...Set[V]) Set[V] {
	for _, other := range others {
		other.Each(func(key V) {
			s.Remove(key)
		})
	}
	return s
}

func (s Set[V]) InPlaceUnion(others ...Set[V]) Set[V] {
	for _, other := range others {
		other.Each(func(key V) {
			s.Put(key)
		})
	}
	return s
}

func (s Set[V]) Values() []V {
	return s.set.Values()
}

func (s Set[V]) Keys() []V {
	return s.Values()
}

func (s Set[V]) IsDisjoint(other Set[V]) bool {
	return s.Intersection(other).Size() > 0
}

func (s Set[V]) IsSubset(of Set[V]) bool {
	subset := true
	s.Each(func(key V) {
		if !of.Has(key) {
			subset = false
		}
	})
	return subset
}

func (s Set[V]) IsSuperset(of Set[V]) bool {
	superset := true
	of.Each(func(key V) {
		if !s.Has(key) {
			superset = false
		}
	})
	return superset
}

func (s Set[V]) Equal(to Set[V]) bool {
	if s.Size() != to.Size() {
		return false
	}
	return s.Union(to).Size() == s.Size()
}

func (s Set[V]) IsProperSubset(to Set[V]) bool {
	if s.Equal(to) {
		return false
	}
	return s.IsSubset(to)
}

func (s Set[V]) IsProperSuperset(to Set[V]) bool {
	if s.Equal(to) {
		return false
	}
	return s.IsSuperset(to)
}
