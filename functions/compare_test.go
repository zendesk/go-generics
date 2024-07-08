package functions_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/zendesk/go-generics/functions"
	"github.com/zendesk/go-generics/internal/test"
	"github.com/zendesk/lockbox-shared-lib/lockbox/generics"
	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

func TestDedupe(t *testing.T) {
	test.CheckEqualIgnoreOrder(functions.Dedupe([]int{1, 1, 5, 5, 7, 9, 999}), "Test1", []int{1, 5, 7, 9, 999}, t)
	test.CheckEqualIgnoreOrder(functions.Dedupe([]string{"abc", "bc", "a", "b", "b", "g", "abc"}), "Test2", []string{"abc", "bc", "a", "b", "g"}, t)
	test.CheckEqualIgnoreOrder(functions.Dedupe([]int{-99, -9999, 9999, 99, -99, 99}), "Test3", []int{-99, -9999, 9999, 99}, t)
	test.CheckEqualIgnoreOrder(functions.Dedupe([]byte{0x1, 0x2, 0x9, 0x10, 0x1}), "Test4", []byte{0x1, 0x2, 0x9, 0x10}, t)
	test.CheckEqualIgnoreOrder(functions.Dedupe([]float64{1.11, 1.99, 99.99, 1.91, 99.99, 1.11, 1.44}), "Test3", []float64{1.11, 1.99, 99.99, 1.91, 1.44}, t)
}

func TestContains(t *testing.T) {
	test.CheckEqual(functions.Contains([]string{"foo", "bar", "baz"}, "foo"), "Test1", true, t)
	test.CheckEqual(functions.Contains([]int{1, 2, 3, 4}, 4), "Test1", true, t)
	test.CheckEqual(functions.Contains([]float64{1.11, 2.22, 3.33, 4.44}, 4.44), "Test1", true, t)
	test.CheckEqual(functions.Contains([]string{"foo", "bar", "baz"}, "fo"), "Test1", false, t)
	test.CheckEqual(functions.Contains([]int{1, 2, 3, 4}, 0), "Test1", false, t)
	test.CheckEqual(functions.Contains([]float64{1.11, 2.22, 3.33, 4.44}, 4.444), "Test1", false, t)
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

	test.CheckEqual(functions.RemoveNils([]*string{&foo, &bar, nil, &baz}), "Test1", []*string{&foo, &bar, &baz}, t)
	test.CheckEqual(functions.RemoveNils([]*string{&foo, &bar, nil, &baz, &thing, thing2}), "Test2", []*string{&foo, &bar, &baz, &thing}, t)
	test.CheckEqual(functions.RemoveNils([]IsAThing3{thing3, nil}), "Test3", []IsAThing3{}, t)
	test.CheckEqual(functions.RemoveNils([]IsAThing3{thing5, nil, nil}), "Test4", []IsAThing3{}, t)
	test.CheckEqual(functions.RemoveNils([]IsAThing3{thing5, nil, realThing3}), "Test4", []IsAThing3{realThing3}, t)
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

	notFound := functions.ContainsDeepEqual(missing, toFind)
	test.CheckOk(!notFound, "Item was found but should be missing!", t)

	// Add the duplicate, but don't add toFind b/c we want to ensure deepEqual is not comparing memory addresses.
	notMissing := append(missing, toAdd)

	shouldBeFound := functions.ContainsDeepEqual(notMissing, toFind)
	test.CheckOk(shouldBeFound, "Item was not found but should be!", t)

}

func TestEqualIgnoreOrder(t *testing.T) {
	first := []int{3, 2, 1, 7, 7}
	second := []int{2, 1, 7, 7, 3}
	third := []int{1, 2, 3, 7, 7}
	different := []int{9, 9, 9, 9}

	test.CheckOk(functions.EqualIgnoreOrder(first, second, third), "Test 1", t)
	test.CheckOk(functions.EqualIgnoreOrder(first, third), "Test 2", t)
	test.CheckNotOk(functions.EqualIgnoreOrder(first, second, third, different), "Test 3", t)
	test.CheckNotOk(functions.EqualIgnoreOrder(different, first, second), "Test 4", t)
	test.CheckOk(functions.EqualIgnoreOrder(first, second), "Test 5", t)
	test.CheckOk(functions.EqualIgnoreOrder(first, third), "Test 6", t)
	test.CheckOk(functions.EqualIgnoreOrder(different, different), "Test 7", t)
	test.CheckOk(functions.EqualIgnoreOrder(first), "Test 8", t)
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
			test.CheckEqual(functions.SliceContainsAny(tt.sliceA, tt.sliceB), tt.name, tt.want, t)
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
			test.CheckEqual(functions.SliceContainsAny(tt.sliceA, tt.sliceB), tt.name, tt.want, t)
		})
	}
}

//
//func TestDedupeByHash(t *testing.T) {
//	foos := []test.Foo{
//		{
//			Bar:   "bar1",
//			Baz:   "baz1",
//			Order: 0,
//		},
//		{
//			Bar:   "bar2",
//			Baz:   "baz2",
//			Order: 1,
//		},
//		{
//			Bar:   "bar3",
//			Baz:   "baz3",
//			Order: 2,
//		}, // Below are dupes depending on fn provided
//	}
//
//	dupe1 := test.Foo{
//		Bar:   "bar1",
//		Baz:   "nadfdsafads",
//		Order: 9999,
//	}
//
//	dupe2 := test.Foo{
//		Bar:   "adfasdf",
//		Baz:   "baz2",
//		Order: 9998,
//	}
//
//	dupe3 := test.Foo{
//		Bar:   "adfasdf",
//		Baz:   "nadfdsafads",
//		Order: 1,
//	}
//
//	dupe1Foos := append(foos, dupe1)
//	dupe2Foos := append(foos, dupe2)
//	dupe3Foos := append(foos, dupe3, dupe1, dupe2)
//
//	// test.Bar: "bar1" should be removed.
//	dedupe1 := functions.DedupeByHash(dupe1Foos, func(i test.Foo) uint64 {
//		return utils.Hash64(i.Bar)
//	})
//
//	test.CheckEqualIgnoreOrder(dedupe1, "dedupe1", foos, t)
//
//	// Baz: "baz2" should be removed
//	dedupe2 := functions.DedupeByHash(dupe2Foos, func(i test.Foo) uint64 {
//		return utils.Hash64(i.Baz)
//	})
//	test.CheckEqualIgnoreOrder(dedupe2, "dedupe2", foos, t)
//
//	// Order: 1 should be removed.
//	dedupe3 := functions.DedupeByHash(dupe3Foos, func(i test.Foo) uint64 {
//		return uint64(i.Order)
//	})
//	test.CheckEqualIgnoreOrder(dedupe3, "dedupe3", append(foos, dupe1, dupe2), t)
//}

func FuzzTestEqualIgnoreOrder(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		sliceSize := uint(utils.RandomNumber(maxSliceSizeLength))
		numSlices := uint(utils.RandomNumber(1000))
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
			number := utils.RandomNumber(99999)
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
			sliceToBreak := utils.RandomNumberBetween(0, int(numSlices)-1)
			itemToBreak := utils.RandomNumberBetween(0, int(sliceLength)-1)
			comparables[sliceToBreak][itemToBreak] = comparables[sliceToBreak][itemToBreak] + 1
			test.CheckNotOk(generics.EqualIgnoreOrder(comparables...), "Comparable slices were equal but shouldn't be!", t)
		} else {
			test.CheckOk(generics.EqualIgnoreOrder(comparables...), "Comparable were not equal but should be!", t)
		}
	})
}
