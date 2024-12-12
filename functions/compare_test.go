package functions

import (
	"math/rand"
	"testing"
	"time"

	"github.com/zendesk/go-generics/internal/test"
)

func TestDedupe(t *testing.T) {
	test.CheckComparableEqualIgnoreOrder(Dedupe([]int{1, 1, 5, 5, 7, 9, 999}), "Test1", []int{1, 5, 7, 9, 999}, t)
	test.CheckComparableEqualIgnoreOrder(Dedupe([]string{"abc", "bc", "a", "b", "b", "g", "abc"}), "Test2", []string{"abc", "bc", "a", "b", "g"}, t)
	test.CheckComparableEqualIgnoreOrder(Dedupe([]int{-99, -9999, 9999, 99, -99, 99}), "Test3", []int{-99, -9999, 9999, 99}, t)
	test.CheckComparableEqualIgnoreOrder(Dedupe([]byte{0x1, 0x2, 0x9, 0x10, 0x1}), "Test4", []byte{0x1, 0x2, 0x9, 0x10}, t)
	test.CheckComparableEqualIgnoreOrder(Dedupe([]float64{1.11, 1.99, 99.99, 1.91, 99.99, 1.11, 1.44}), "Test3", []float64{1.11, 1.99, 99.99, 1.91, 1.44}, t)
}

func TestContains(t *testing.T) {
	test.CheckEqual(Contains([]string{"foo", "bar", "baz"}, "foo"), "Test1", true, t)
	test.CheckEqual(Contains([]int{1, 2, 3, 4}, 4), "Test1", true, t)
	test.CheckEqual(Contains([]float64{1.11, 2.22, 3.33, 4.44}, 4.44), "Test1", true, t)
	test.CheckEqual(Contains([]string{"foo", "bar", "baz"}, "fo"), "Test1", false, t)
	test.CheckEqual(Contains([]int{1, 2, 3, 4}, 0), "Test1", false, t)
	test.CheckEqual(Contains([]float64{1.11, 2.22, 3.33, 4.44}, 4.444), "Test1", false, t)
}
func TestContainsSublist(t *testing.T) {
	tests := []struct {
		name     string
		list1    []int
		list2    []int
		expected bool
	}{
		{
			name:     "Sublist exists in the middle",
			list1:    []int{2, 3},
			list2:    []int{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			name:     "Sublist exists at the start",
			list1:    []int{1, 2},
			list2:    []int{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			name:     "Sublist exists at the end",
			list1:    []int{4, 5},
			list2:    []int{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			name:     "Sublist does not need to be contiguous",
			list1:    []int{3, 5}, // Not contiguous
			list2:    []int{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			name:     "Empty list 1",
			list1:    []int{},
			list2:    []int{1, 2, 3, 4, 5},
			expected: true,
		},
		{
			name:     "Sublist larger than main list",
			list1:    []int{1, 2, 3},
			list2:    []int{1, 2},
			expected: false,
		},
		{
			name:     "Identical lists",
			list1:    []int{1, 2, 3},
			list2:    []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "list 2 is longer, but still contains all elements from list 1",
			list1:    []int{1, 2, 3},
			list2:    []int{1, 2, 3, 4, 5},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsAll(tt.list1, tt.list2)
			test.CheckEqual(result, tt.name, tt.expected, t)
		})
	}
}

type thing1 struct{}
type IsAThing = string
type IsAThing2 = *string
type IsAThing3 interface{}
type Thing5 struct {
	Thing3 IsAThing3
}

func MakeThing(t IsAThing3) IsAThing3 {
	return t
}

func TestRemoveNils(t *testing.T) {
	foo := "foo"
	bar := "bar"
	baz := "baz"

	var thing IsAThing
	var thing2 IsAThing2 = nil
	var thing3 IsAThing3
	// This is a boxed nil
	var thing4 *thing1 = nil
	var thing5 IsAThing3 = MakeThing(thing4)
	var realThing3 IsAThing3 = Thing5{
		Thing3: "asdf123",
	}

	t.Logf("thing5: %+v", thing5)         // thing 5 is nil but thing5 == nil is false. It's a BOXED nil.
	t.Logf("IS NIL5: %+v", thing5 == nil) // this is FALSE because this is a BOXED nil

	test.CheckEqual(RemoveNils([]*string{&foo, &bar, nil, &baz}), "Test1", []*string{&foo, &bar, &baz}, t)
	test.CheckEqual(RemoveNils([]*string{&foo, &bar, nil, &baz, &thing, thing2}), "Test2", []*string{&foo, &bar, &baz, &thing}, t)
	test.CheckEqual(RemoveNils([]IsAThing3{thing3, nil}), "Test3", []IsAThing3{}, t)
	test.CheckEqual(RemoveNils([]IsAThing3{thing5, nil, nil}), "Test4", []IsAThing3{}, t)
	test.CheckEqual(RemoveNils([]IsAThing3{thing5, nil, realThing3}), "Test4", []IsAThing3{realThing3}, t)
}

type ContainingStruct struct {
	Deep DeepStruct
	Foo  test.Foo
}
type DeepStruct struct {
	Bar test.Bar
}

func TestContainsDeepEqual(t *testing.T) {
	toFind := ContainingStruct{
		Foo: test.Foo{
			Bar:   "1234",
			Baz:   "12123",
			Order: 1234,
		},
		Deep: DeepStruct{
			Bar: test.Bar{
				Bing:  "Bong",
				Order: 111,
			},
		},
	}

	// Must have same values as toFind
	toAdd := ContainingStruct{
		Foo: test.Foo{
			Bar:   "1234",
			Baz:   "12123",
			Order: 1234,
		},
		Deep: DeepStruct{
			Bar: test.Bar{
				Bing:  "Bong",
				Order: 111,
			},
		},
	}

	missing := []ContainingStruct{
		{
			Foo: test.Foo{
				Bar:   "bar",
				Baz:   "baz",
				Order: 1234,
			},
			Deep: DeepStruct{
				Bar: test.Bar{},
			},
		}, {
			Foo: test.Foo{
				Bar:   "asdf1234123",
				Baz:   "adfasdf",
				Order: 222222,
			},
			Deep: DeepStruct{
				Bar: test.Bar{
					Bing:  "24141414",
					Order: 201231,
				},
			},
		},
	}

	notFound := ContainsDeepEqual(missing, toFind)
	test.CheckOk(!notFound, "Item was found but should be missing!", t)

	// Add the duplicate, but don't add toFind b/c we want to ensure deepEqual is not comparing memory addresses.
	notMissing := append(missing, toAdd)

	shouldBeFound := ContainsDeepEqual(notMissing, toFind)
	test.CheckOk(shouldBeFound, "Item was not found but should be!", t)

}

func TestEqualIgnoreOrder(t *testing.T) {
	first := []int{3, 2, 1, 7, 7}
	second := []int{2, 1, 7, 7, 3}
	third := []int{1, 2, 3, 7, 7}
	different := []int{9, 9, 9, 9}

	test.CheckOk(EqualIgnoreOrder(first, second, third), "Test 1", t)
	test.CheckOk(EqualIgnoreOrder(first, third), "Test 2", t)
	test.CheckNotOk(EqualIgnoreOrder(first, second, third, different), "Test 3", t)
	test.CheckNotOk(EqualIgnoreOrder(different, first, second), "Test 4", t)
	test.CheckOk(EqualIgnoreOrder(first, second), "Test 5", t)
	test.CheckOk(EqualIgnoreOrder(first, third), "Test 6", t)
	test.CheckOk(EqualIgnoreOrder(different, different), "Test 7", t)
	test.CheckOk(EqualIgnoreOrder(first), "Test 8", t)
}

func TestSliceContainsAny(t *testing.T) {
	tests := []struct {
		name   string
		sliceA []int
		sliceB []int
		want   bool
	}{
		{
			"it returns true when a matching element is found",
			[]int{1, 2, 3, 4, 5, 6},
			[]int{1, 8, 9, 10, 11},
			true,
		},
		{
			"it returns false when no matching element is found",
			[]int{1, 2, 3, 4, 5, 6},
			[]int{7, 8, 9, 10, 11},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.CheckEqual(ContainsAny(tt.sliceA, tt.sliceB), tt.name, tt.want, t)
		})
	}
}

func TestSliceContainsAnyBools(t *testing.T) {
	tests := []struct {
		name   string
		sliceA []bool
		sliceB []bool
		want   bool
	}{
		{
			"it returns true when a matching element is found",
			[]bool{false},
			[]bool{false},
			true,
		},
		{
			"it returns false when no matching element is found",
			[]bool{true},
			[]bool{false},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.CheckEqual(ContainsAny(tt.sliceA, tt.sliceB), tt.name, tt.want, t)
		})
	}
}

func FuzzTestEqualIgnoreOrder(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		sliceSize := uint(randomNumber(maxSliceSizeLength))
		numSlices := uint(randomNumber(1000))
		f.Add(numSlices, sliceSize)
	}

	f.Fuzz(func(t *testing.T, numSlices, sliceLength uint) {
		// expect failure when even number of slices is provided
		expectFailure := numSlices%2 == 0 && numSlices > 1 && sliceLength > 1

		var comparables = make([][]int, numSlices)

		// Init slices
		for i := 0; i < int(numSlices); i++ {
			comparables[i] = make([]int, sliceLength)
		}

		// Make identical slices
		for j := 0; j <= int(sliceLength)-1; j++ {
			number := randomNumber(99999)
			for i := 0; i <= int(numSlices)-1; i++ {
				comparables[i][j] = number
			}
		}

		// Shuffle slices
		for i := 0; i <= int(numSlices)-1; i++ {
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(comparables[i]), func(j, k int) {
				comparables[i][j], comparables[i][k] = comparables[i][k], comparables[i][j]
			})
		}

		if expectFailure {
			sliceToBreak := randomNumberBetween(0, int(numSlices)-1)
			itemToBreak := randomNumberBetween(0, int(sliceLength)-1)
			comparables[sliceToBreak][itemToBreak] = comparables[sliceToBreak][itemToBreak] + 1
			test.CheckNotOk(EqualIgnoreOrder(comparables...), "Comparable slices were equal but shouldn't be!", t)
		} else {
			test.CheckOk(EqualIgnoreOrder(comparables...), "Comparable were not equal but should be!", t)
		}
	})
}
