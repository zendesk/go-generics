//go:build test
// +build test

package functions

import (
	"testing"

	"github.com/zendesk/go-generics/functions/internal/test"
)

func TestFind(t *testing.T) {
	item, wasFound := Find([]string{"foo", "bar", "baz"}, func(f string) bool {
		return f == "foo"
	})
	test.CheckEqual(item, "Test1", "foo", t)
	test.CheckOk(wasFound, "Test1WasFound", t)

	item2, wasFound2 := Find([]int{2, 22, 999}, func(i int) bool {
		return i == 999
	})
	test.CheckEqual(item2, "Test2", 999, t)
	test.CheckOk(wasFound2, "Test2WasFound", t)

	item3, wasFound3 := Find([]float64{2.99, 22, 99.999}, func(i float64) bool {
		return i == 22
	})
	test.CheckEqual(item3, "Test3", 22.0, t)
	test.CheckOk(wasFound3, "Test3WasFound", t)

	item4, notFound1 := Find([]float64{2.99, 22, 99.999}, func(i float64) bool {
		return i == 22.000001
	})
	test.CheckEqual(item4, "Test4", 0.0, t)
	test.CheckOk(!notFound1, "Test4WasFound", t)

	item5, notFound5 := Find([]string{"foo", "bar", "baz"}, func(f string) bool {
		return f == "missingfoo"
	})
	test.CheckEqual(item5, "Test5", "", t)
	test.CheckOk(!notFound5, "Test5WasFound", t)

	item6, notFound6 := Find([]int{2, 22, 999}, func(i int) bool {
		return i == 9991
	})
	test.CheckEqual(item6, "Test6", 0, t)
	test.CheckOk(!notFound6, "Test6WasFound", t)
}

func TestFilter(t *testing.T) {
	ints := []int{1, 2, 3, 3, 9, 1, 1, 20, 9999, -15, -20}

	negativeNumbers := Filter(ints, func(i int) bool {
		return i < 0
	})

	gt9 := Filter(ints, func(i int) bool {
		return i > 9
	})

	evens := Filter(ints, func(i int) bool {
		return i%2 == 0
	})

	test.CheckEqual(negativeNumbers, "negativNumbers", []int{-15, -20}, t)
	test.CheckEqual(gt9, "greaterThan9", []int{20, 9999}, t)
	test.CheckEqual(evens, "evens", []int{2, 20, -20}, t)
}

func TestFilterMap(t *testing.T) {
	testMap := map[int]int{
		1: 1,
		2: 2,
		3: 3,
		4: 4,
		5: 5,
	}

	newMap := FilterMap(testMap, func(k, v int) bool {
		return k%2 == 0
	})

	expected := map[int]int{
		2: 2,
		4: 4,
	}

	test.CheckEqual(newMap, "Map filter", expected, t)

	newMap = FilterMap(testMap, func(k, v int) bool {
		return true
	})
	test.CheckEqual(newMap, "All items", testMap, t)

	newMap = FilterMap(testMap, func(k, v int) bool {
		return false
	})

	test.CheckEqual(newMap, "No items filter", map[int]int{}, t)
}
