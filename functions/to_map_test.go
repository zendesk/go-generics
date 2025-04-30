//go:build test
// +build test

package functions

import (
	"fmt"
	"testing"

	"github.com/zendesk/go-generics/test"
)

func TestToMap(t *testing.T) {
	ints := []int{1, -20, 0, 11}

	squaresMap := ToMap(ints, func(i int) (int, int) {
		return i, i * i
	})

	expected := map[int]int{
		1:   1,
		-20: 400,
		0:   0,
		11:  121,
	}

	test.CheckEqual(squaresMap, "Squares", expected, t)
}

func FuzzGoToMap(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoosOrderly(num)
		fooMap := GoToMap(foos, toKeyValue)

		if len(fooMap) != len(foos) {
			t.Fatalf("Mapping failed, data was lost. Got length: %d but expected %d", len(fooMap), len(foos))
		}

		for k, v := range fooMap {
			test.CheckEqual(k, "Order Key", foos[k].Order, t)
			test.CheckEqual(v, "Bar", foos[k].Bar, t)
		}
	})
}

func FuzzGoToMapWithErrs(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		num = Max(1, num)
		foos := test.MakeFoosOrderly(num)

		fooMap, errs := GoToMapWithErrs(foos, toKeyValueWithErrs)

		expectedErrCount := (num-1)/3 + 1

		test.CheckEqual(len(errs), "Err count", expectedErrCount, t)
		test.CheckEqual(len(fooMap), "Foo map length", num, t)

		for k, v := range fooMap {
			test.CheckEqual(k, "Order Key", foos[k].Order, t)
			test.CheckEqual(v, "Bar", foos[k].Bar, t)
		}
	})
}

func FuzzGoToMapWithErrsDropErrorRecords(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		num = Max(1, num)
		foos := test.MakeFoosOrderly(num)

		fooMap, errs := GoToMapWithErrs(foos, toKeyValueWithErrs, DiscardResultIfErrOption())

		expectedErrCount := (num-1)/3 + 1

		test.CheckEqual(len(errs), fmt.Sprintf("Err count from num: %d", num), expectedErrCount, t)
		test.CheckEqual(len(fooMap), "Foo map length", num-expectedErrCount, t)

		for k, v := range fooMap {
			test.CheckEqual(k, "Order Key", foos[k].Order, t)
			test.CheckEqual(v, "Bar", foos[k].Bar, t)
			test.CheckOk(k%3 != 0, "Unexpected record returned. This should have been an error", t)
			test.CheckOk(foos[k].Order%3 != 0, "Unexpected record returned. This should have been an error", t)
		}
	})
}
