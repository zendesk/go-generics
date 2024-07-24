package types

import (
	"fmt"
	"math"
	"testing"

	"github.com/zendesk/go-generics/internal/test"
)

type setTest[K comparable] struct {
	name      string
	ops       func(ISet[K])
	expected  []K
	validator func(ISet[K]) (result bool, msg string)
}

func generateSetTests[K comparable](seeds []K) []setTest[K] {
	var dedupedSeeds = dedupe(seeds)

	var kMinus1 []K
	if len(dedupedSeeds) > 0 {
		kMinus1 = make([]K, len(dedupedSeeds)-1)
		if len(dedupedSeeds) > 0 {
			copy(kMinus1, dedupedSeeds[:len(dedupedSeeds)-1])
		}
	}

	return []setTest[K]{
		{
			name: "Put",
			ops: func(s ISet[K]) {
				for _, seed := range seeds {
					s.Put(seed)
				}
			},
			expected: dedupedSeeds,
		},
		{
			name: "Remove",
			ops: func(s ISet[K]) {
				for _, seed := range seeds {
					s.Put(seed)
				}
				if len(dedupedSeeds) > 0 {
					s.Remove(dedupedSeeds[len(dedupedSeeds)-1])
				}
			},
			expected: kMinus1,
			validator: func(s ISet[K]) (bool, string) {
				if len(dedupedSeeds) > 0 {
					return !s.Has(dedupedSeeds[len(dedupedSeeds)-1]), "Last seed was not removed"
				}
				return true, ""
			},
		},
		{
			name: "Has",
			ops: func(s ISet[K]) {
				for _, seed := range seeds {
					s.Put(seed)
				}
			},
			expected: dedupedSeeds,
			validator: func(s ISet[K]) (bool, string) {
				if len(seeds) > 1 {
					return s.Has(seeds[0]), ""
				}
				if len(seeds) > 0 {
					return s.Has(seeds[0]), fmt.Sprintf("Seed missing: %+v", seeds[0])
				}
				return true, ""
			},
		},
		{
			name: "Size",
			ops: func(s ISet[K]) {
				for _, seed := range seeds {
					s.Put(seed)
				}
			},
			expected: dedupedSeeds,
			validator: func(s ISet[K]) (bool, string) {
				return s.Size() == len(dedupedSeeds), fmt.Sprintf("Size was not: %d", len(dedupedSeeds))
			},
		},
		{
			name: "Clear",
			ops: func(s ISet[K]) {
				for _, seed := range seeds {
					s.Put(seed)
				}
				s.Clear()
			},
			expected: []K{},
			validator: func(s ISet[K]) (bool, string) {
				return s.Size() == 0, "Size was not zero"
			},
		},
		{
			name: "New",
			ops: func(s ISet[K]) {
				for _, seed := range seeds {
					s.Put(seed)
				}
			},
			expected: dedupedSeeds,
			validator: func(s ISet[K]) (bool, string) {
				n := s.New()
				return n.Size() == 0, fmt.Sprintf("Size was not zero. Got: %+v", n)
			},
		},
		{
			name: "Copy",
			ops: func(s ISet[K]) {
				for _, seed := range seeds {
					s.Put(seed)
				}
			},
			expected: dedupedSeeds,
			validator: func(s ISet[K]) (bool, string) {
				n := s.Copy()
				s.Clear()
				if len(seeds) > 0 {
					return n.Has(seeds[0]) && !s.Has(seeds[0]), fmt.Sprintf("N: %+v, S: %+v", n, s)
				}
				return true, ""
			},
		},
	}
}

func Test_ConcreteSets(t *testing.T) {
	initSets := func() map[string]ISet[int] {
		return map[string]ISet[int]{
			"ComparableSet": newComparableSet[int](),
			"HashSet":       newHashSet[int](),
		}
	}

	input := []int{1, 2, 3, 4, 5, 4, 4, 4}
	for _, tc := range generateSetTests[int](input) {
		t.Run(tc.name, func(t *testing.T) {
			for setName, set := range initSets() {
				// Run operator
				tc.ops(set)
				actual := set.Values()
				test.CheckComparableEqualIgnoreOrder(actual, setName, tc.expected, t)
				if tc.validator != nil {
					if result, msg := tc.validator(set); !result {
						t.Errorf("%s: Validator failed for test: %s. Error: %s input: %+v.", setName, tc.name, msg, input)
					}
				}
			}
		})
	}
}

func Test_Set(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{3, 4, 5}
	c := []int{99, 100, 101}
	d := []int{1, 2}

	aSet := NewSet(a...)
	bSet := NewSet(b...)
	cSet := NewSet(c...)
	dSet := NewSet(d...)
	dSetCopy := NewSet(d...)

	// Difference
	test.CheckComparableEqualIgnoreOrder(aSet.Difference(bSet).Values(), "A - B", []int{1, 2}, t)
	test.CheckComparableEqualIgnoreOrder(bSet.Difference(aSet).Values(), "B - A", []int{4, 5}, t)

	// Intersection
	test.CheckComparableEqualIgnoreOrder(aSet.Intersection(bSet).Values(), "A ∩ B", []int{3}, t)
	test.CheckComparableEqualIgnoreOrder(bSet.Intersection(aSet).Values(), "B ∩ A", []int{3}, t)

	// Union
	test.CheckComparableEqualIgnoreOrder(aSet.Union(bSet).Values(), "A U B", []int{1, 2, 3, 4, 5}, t)
	test.CheckComparableEqualIgnoreOrder(bSet.Union(aSet).Values(), "B U A", []int{1, 2, 3, 4, 5}, t)

	// IsDisjoint
	test.CheckNotOk(aSet.IsDisjoint(bSet), "Not Disjoint", t)
	test.CheckOk(bSet.IsDisjoint(cSet), "Is Disjoint", t)

	// IsSubset
	test.CheckOk(dSet.IsSubset(aSet), "Is Subset", t)
	test.CheckNotOk(bSet.IsSubset(aSet), "Not Subset", t)

	// Superset
	test.CheckOk(aSet.IsSuperset(dSet), "Is SuperSet", t)
	test.CheckNotOk(bSet.IsSuperset(aSet), "Not Superset", t)
	test.CheckNotOk(dSet.IsSuperset(aSet), "Not SuperSet (2)", t)

	// Proper Subset
	test.CheckOk(dSet.IsProperSubset(aSet), "Proper subset", t)
	test.CheckNotOk(dSet.IsProperSubset(bSet), "Not Proper subset", t)
	test.CheckNotOk(dSet.IsProperSubset(dSetCopy), "Not Proper subset (equal)", t)

	// Proper SuperSet
	test.CheckOk(aSet.IsProperSuperset(dSet), "Proper superset", t)
	test.CheckNotOk(aSet.IsProperSuperset(bSet), "Not Proper superset", t)
	test.CheckNotOk(dSet.IsProperSuperset(dSetCopy), "Not Proper superset (equal)", t)

	// Equal
	test.CheckOk(dSet.Equal(dSetCopy), "Is Equal", t)

	// SymmetricDifference
	test.CheckComparableEqualIgnoreOrder(aSet.SymmetricDifference(bSet).Keys(), "Symmetric Diff", []int{1, 2, 4, 5}, t)

	// In Place
	aClone := aSet.Clone()
	test.CheckComparableEqualIgnoreOrder(aClone.InPlaceIntersection(bSet).Values(), "InPlaceIntersection", []int{3}, t)
	aClone = aSet.Clone()
	test.CheckComparableEqualIgnoreOrder(aClone.InPlaceUnion(bSet).Values(), "InPlaceUnion", []int{1, 2, 3, 4, 5}, t)
	aClone = aSet.Clone()
	test.CheckComparableEqualIgnoreOrder(aClone.InPlaceDifference(bSet).Values(), "InPlaceDifference", []int{1, 2}, t)
}

// Fuzz_Set is a fuzz test for the Set operations, byte slice is the only compatible dynamic length input for fuzzing
func Fuzz_Set(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 22, 44})
	f.Add([]byte{2, 2, 2})

	f.Fuzz(func(t *testing.T, bytes []byte) {
		initSets := func() map[string]ISet[byte] {
			return map[string]ISet[byte]{
				"ComparableSet": newComparableSet[byte](),
				"HashSet":       newHashSet[byte](),
			}
		}
		for _, tc := range generateSetTests[byte](bytes) {
			for setName, s := range initSets() {
				// Run operator
				tc.ops(s)
				actual := s.Values()
				test.CheckComparableEqualIgnoreOrder(actual, setName, tc.expected, t)
				if tc.validator != nil {
					if result, msg := tc.validator(s); !result {
						t.Errorf("%s: Validator failed for test: %s. Error: %s input: %+v.", setName, tc.name, msg, bytes)
					}
				}
			}
		}
	})
}

// Fuzz_Set_StringsRandomLength tests random sets of strings of random length with a max size of length 65535
func Fuzz_Set_Strings_Random_Length(f *testing.F) {
	f.Add(uint16(0))
	f.Add(uint16(math.MaxUint16))

	f.Fuzz(func(t *testing.T, size uint16) {
		items := makeItems(size)
		initSets := func() map[string]ISet[string] {
			return map[string]ISet[string]{
				"ComparableSet": newComparableSet[string](),
				"HashSet":       newHashSet[string](),
			}
		}
		for _, tc := range generateSetTests[string](items) {
			for setName, s := range initSets() {
				// Run operator
				tc.ops(s)
				actual := s.Values()
				test.CheckComparableEqualIgnoreOrder(actual, setName, tc.expected, t)
				if tc.validator != nil {
					if result, msg := tc.validator(s); !result {
						t.Errorf("%s: Validator failed for test: %s. Error: %s.", setName, tc.name, msg)
					}
				}
			}
		}
	})
}

func makeItems(length uint16) []string {
	items := make([]string, length)
	for i := range items {
		items[i] = fmt.Sprintf("item-%s", test.GenerateRandomLetterString(500))
	}
	return items
}

func dedupe[T comparable](items []T) []T {
	itemMap := make(map[T]bool)
	for _, item := range items {
		itemMap[item] = true
	}

	var deduped []T
	for k := range itemMap {
		deduped = append(deduped, k)
	}

	return deduped
}
