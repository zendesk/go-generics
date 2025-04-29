//go:build test
// +build test

package datastructures

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/zendesk/go-generics/test"
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
	nilSet := Set[int]{}
	nilSet2 := Set[int]{}

	// NewFunc, Func name (for test errs)
	var newFuncs = []func(values ...int) (Set[int], string){
		func(values ...int) (Set[int], string) {
			return NewSet(values...), "NewSet"
		},
		func(values ...int) (Set[int], string) {
			return NewHashSet[int](values...), "NewHashSet"
		},
		func(values ...int) (Set[int], string) {
			return NewHashSetWithHashFn[int](func(i int) string {
				return hashAny[int](i)
			}, values...), "NewHashSetWithHashFn"
		},
	}

	for _, newFunc := range newFuncs {
		aSet, name := newFunc(a...)
		bSet, _ := newFunc(b...)
		cSet, _ := newFunc(c...)
		dSet, _ := newFunc(d...)
		dSetCopy, _ := newFunc(d...)

		// Difference
		test.CheckComparableEqualIgnoreOrder(aSet.Difference(bSet).Values(), fmt.Sprintf("%s: %s", name, "A - B"), []int{1, 2}, t)
		test.CheckComparableEqualIgnoreOrder(bSet.Difference(aSet).Values(), fmt.Sprintf("%s: %s", name, "B - A"), []int{4, 5}, t)

		// Diff with nil
		test.CheckComparableEqualIgnoreOrder(aSet.Difference(nilSet).Values(), fmt.Sprintf("%s: %s", name, "A - nil"), []int{1, 2, 3}, t)
		test.CheckComparableEqualIgnoreOrder(nilSet.Difference(nilSet2).Values(), fmt.Sprintf("%s: %s", name, "nil - nil"), []int{}, t)

		// Intersection
		test.CheckComparableEqualIgnoreOrder(aSet.Intersection(bSet).Values(), fmt.Sprintf("%s: %s", name, "A ∩ B"), []int{3}, t)
		test.CheckComparableEqualIgnoreOrder(bSet.Intersection(aSet).Values(), fmt.Sprintf("%s: %s", name, "B ∩ A"), []int{3}, t)

		// Intersect with nil
		test.CheckComparableEqualIgnoreOrder(aSet.Intersection(nilSet).Values(), fmt.Sprintf("%s: %s", name, "A ∩ nil"), []int{}, t)

		// Union
		test.CheckComparableEqualIgnoreOrder(aSet.Union(bSet).Values(), fmt.Sprintf("%s: %s", name, "A U B"), []int{1, 2, 3, 4, 5}, t)
		test.CheckComparableEqualIgnoreOrder(bSet.Union(aSet).Values(), fmt.Sprintf("%s: %s", name, "B U A"), []int{1, 2, 3, 4, 5}, t)

		// Union with nil
		test.CheckComparableEqualIgnoreOrder(aSet.Union(nilSet).Values(), fmt.Sprintf("%s: %s", name, "A U nil"), []int{1, 2, 3}, t)

		// IsDisjoint
		test.CheckNotOk(aSet.IsDisjoint(bSet), fmt.Sprintf("%s: %s", name, "Not Disjoint"), t)
		test.CheckOk(bSet.IsDisjoint(cSet), fmt.Sprintf("%s: %s", name, "Is Disjoint"), t)

		// IsSubset
		test.CheckOk(dSet.IsSubset(aSet), fmt.Sprintf("%s: %s", name, "Is Subset"), t)
		test.CheckNotOk(bSet.IsSubset(aSet), fmt.Sprintf("%s: %s", name, "Not Subset"), t)

		// Superset
		test.CheckOk(aSet.IsSuperset(dSet), fmt.Sprintf("%s: %s", name, "Is SuperSet"), t)
		test.CheckNotOk(bSet.IsSuperset(aSet), fmt.Sprintf("%s: %s", name, "Not Superset"), t)
		test.CheckNotOk(dSet.IsSuperset(aSet), fmt.Sprintf("%s: %s", name, "Not SuperSet (2)"), t)

		// Proper Subset
		test.CheckOk(dSet.IsProperSubset(aSet), fmt.Sprintf("%s: %s", name, "Proper subset"), t)
		test.CheckNotOk(dSet.IsProperSubset(bSet), fmt.Sprintf("%s: %s", name, "Not Proper subset"), t)
		test.CheckNotOk(dSet.IsProperSubset(dSetCopy), fmt.Sprintf("%s: %s", name, "Not Proper subset (equal)"), t)

		// Proper SuperSet
		test.CheckOk(aSet.IsProperSuperset(dSet), fmt.Sprintf("%s: %s", name, "Proper superset"), t)
		test.CheckNotOk(aSet.IsProperSuperset(bSet), fmt.Sprintf("%s: %s", name, "Not Proper superset"), t)
		test.CheckNotOk(dSet.IsProperSuperset(dSetCopy), fmt.Sprintf("%s: %s", name, "Not Proper superset (equal)"), t)

		// Equal
		test.CheckOk(dSet.Equal(dSetCopy), fmt.Sprintf("%s: %s", name, "Is Equal"), t)

		// SymmetricDifference
		test.CheckComparableEqualIgnoreOrder(aSet.SymmetricDifference(bSet).Keys(), fmt.Sprintf("%s: %s", name, "Symmetric Diff"), []int{1, 2, 4, 5}, t)

		// In Place
		aClone := aSet.Clone()
		test.CheckComparableEqualIgnoreOrder(aClone.InPlaceIntersection(bSet).Values(), fmt.Sprintf("%s: %s", name, "InPlaceIntersection"), []int{3}, t)
		aClone = aSet.Clone()
		test.CheckComparableEqualIgnoreOrder(aClone.InPlaceIntersection(nilSet).Values(), fmt.Sprintf("%s: %s", name, "InPlaceIntersection nil"), []int{}, t)

		aClone = aSet.Clone()
		test.CheckComparableEqualIgnoreOrder(aClone.InPlaceUnion(bSet).Values(), fmt.Sprintf("%s: %s", name, "InPlaceUnion"), []int{1, 2, 3, 4, 5}, t)
		aClone = aSet.Clone()
		test.CheckComparableEqualIgnoreOrder(aClone.InPlaceUnion(nilSet).Values(), fmt.Sprintf("%s: %s", name, "InPlaceUnion nil"), []int{1, 2, 3}, t)

		aClone = aSet.Clone()
		test.CheckComparableEqualIgnoreOrder(aClone.InPlaceDifference(bSet).Values(), fmt.Sprintf("%s: %s", name, "InPlaceDifference"), []int{1, 2}, t)
		aClone = aSet.Clone()
		test.CheckComparableEqualIgnoreOrder(aClone.InPlaceDifference(nilSet).Values(), fmt.Sprintf("%s: %s", name, "InPlaceDifference nil"), []int{1, 2, 3}, t)

		// EachWithErrs
		errs := aSet.EachWithErrs(func(i int) error {
			if i%2 == 0 {
				return errors.New("fake error")
			}
			return nil
		})
		numExpectedErrs := aSet.Size() / 2
		test.CheckEqual(len(errs), fmt.Sprintf("%s: %s", name, "EachWithErrs"), numExpectedErrs, t)
	}
}

// Test_Size_Should_Not_Panic tests that calling Size() on a struct with a Set field should not panic
func Test_Size_Should_Not_Panic(t *testing.T) {
	type SomeStruct struct {
		MySet Set[int]
	}
	s := SomeStruct{}
	s.MySet.Size()
}

// Fuzz_Set is a fuzz test for the SetOf operations, byte slice is the only compatible dynamic length input for fuzzing
func Fuzz_Set(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 22, 44})
	f.Add([]byte{2, 2, 2})

	f.Fuzz(func(t *testing.T, bytes []byte) {
		initSets := func() map[string]ISet[byte] {
			return map[string]ISet[byte]{
				"ComparableSet": newComparableSet[byte](),
				"HashSet":       newHashSet[byte](),
				"HashSetWithCustomHasFn": newHashSetWithHashFn[byte](func(t byte) string {
					return hashAny[byte](t)
				}),
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

// Fuzz_Set_StringsRandomLength tests random sets of strings of random length with a max size of length 10k
func Fuzz_Set_Strings_Random_Length(f *testing.F) {
	f.Add(uint16(0))
	f.Add(uint16(15000))

	f.Fuzz(func(t *testing.T, size uint16) {
		items := makeItems(size)
		initSets := func() map[string]ISet[string] {
			return map[string]ISet[string]{
				"ComparableSet": newComparableSet[string](),
				"HashSet":       newHashSet[string](),
				"HashSetWithCustomHasFn": newHashSetWithHashFn[string](func(t string) string {
					return hashAny[string](t)
				}),
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

func Test_HashSet_CustomType(t *testing.T) {
	// Hash set automatically dedupes non-comparable types
	type foo struct {
		Name string
		Age  int
	}

	foos := []foo{{"James", 30}, {"Bob", 44}, {"James", 30}, {"James", 31}}
	myHashSet := NewHashSet(foos...)
	// [{"James", 30}, {"Bob", 44}, {"James", 31}]

	test.CheckEqual(myHashSet.Size(), "Size", 3, t)
	test.CheckContains(myHashSet.Values(), foo{"James", 30}, t)
	test.CheckContains(myHashSet.Values(), foo{"James", 31}, t)
	test.CheckContains(myHashSet.Values(), foo{"Bob", 44}, t)
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

func Test_JSON_Support(t *testing.T) {
	// Test if set is uniitialized
	set := Set[int]{}
	jsonBytes, err := json.Marshal(set)
	test.CheckErr(err, "error marshalling json", t)
	test.CheckEqual("[]", "nil set", string(jsonBytes), t)

	set = NewSet(1, 2, 3, 4, 5)
	jsonBytes, err = json.Marshal(set)
	test.CheckErr(err, "error marshalling json", t)
	test.CheckEqual("[1,2,3,4,5]", "integers", string(jsonBytes), t)

	set2 := NewSet("1", "2", "3")
	jsonBytes, err = json.Marshal(set2)
	test.CheckErr(err, "error marshalling json", t)
	test.CheckEqual(`["1","2","3"]`, "strings", string(jsonBytes), t)
}
