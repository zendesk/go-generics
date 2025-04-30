package test

import (
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// CheckErr will fail a test if an error is detected
func CheckErr(err error, message string, t TestT) {
	if err != nil {
		t.Fatal(fmt.Sprintf("%s - error: %v", message, err))
	}
}

// CheckErrs will fail a test if an error is detected
func CheckErrs(errs []error, message string, t TestT) {
	if len(errs) > 0 {
		var err error = fmt.Errorf("")
		for _, e := range errs {
			err = fmt.Errorf("%s;%s", e.Error(), err.Error())
		}
		t.Fatalf("%s %+v", message, err)
	}
}

// ExpectNil will fail a test if a value is not nil
func ExpectNil(value interface{}, t TestT) {
	if !isNil(value) {
		t.Fatal(fmt.Sprintf("expected nil but got: %+v", value))
	}
}

// ExpectNotEmpty will fail a test if a value is empty
func ExpectNotEmpty(value interface{}, valueName string, t TestT) {
	if value == "" {
		t.Fatal(fmt.Sprintf("expected not empty for value: %s", valueName))
	}
}

// ExpectNotNil will fail a test if a value is nil
func ExpectNotNil(varName string, value interface{}, t TestT) {
	if isNil(value) {
		t.Fatal(fmt.Sprintf("expected not nil for variable %s but got nil instead", varName))
	}
}

// ExpectErr will fail a test if an error is _not_ detected.
func ExpectErr(err error, message string, t TestT) {
	if err == nil {
		t.Fatal(fmt.Sprintf("%s - error: %v", message, err))
	}
}

func CheckEqual(result interface{}, resultName string, expectedResult interface{}, t TestT) {
	if equal := cmp.Equal(result, expectedResult, cmp.Exporter(func(v reflect.Type) bool {
		return true
	})); !equal {
		t.Fatalf("Unexpected results for result: %s. Expected: %+v, Got: %+v", resultName, expectedResult, result)
	}
}

func CheckContains[T any](list []T, item T, t TestT) {
	for _, listItem := range list {
		if matches := cmp.Equal(item, listItem); matches {
			return
		}
	}

	t.Fatalf("Item: %+v not found in list", item)
}

// CheckEqualEquateEmptyIgnoreOrder will treat nil and empty values as equal for maps and slices
func CheckEqualEquateEmptyIgnoreOrder(result interface{}, resultName string, expectedResult interface{}, t TestT) {
	if isEqual := cmp.Equal(result, expectedResult, cmpopts.EquateEmpty(), cmpopts.SortSlices(func(a, b any) bool {
		return fmt.Sprintf("%+v", a) < fmt.Sprintf("%+v", b)
	})); !isEqual {
		t.Errorf("Error for: %s, Result: %+v and Expected: %+v differ.", resultName, result, expectedResult)
		t.Fatalf("Unexpected results for result: %s", resultName)
	}
}

// CheckEqualEquateEmpty will treat nil and empty values as equal for maps and slices
func CheckEqualEquateEmpty(result interface{}, resultName string, expectedResult interface{}, t TestT) {
	if isEqual := cmp.Equal(result, expectedResult, cmpopts.EquateEmpty()); !isEqual {
		t.Errorf("Error for: %s, Result: %+v and Expected: %+v differ.", resultName, result, expectedResult)
		t.Fatalf("Unexpected results for result: %s", resultName)
	}
}

func CheckEqualErrs(result interface{}, resultName string, expectedResult interface{}, t TestT) {
	if equal := cmp.Equal(result, expectedResult, cmpopts.EquateErrors()); !equal {
		t.Fatalf("Unexpected results for result: %s. Expected: %+v, Got: %+v", resultName, expectedResult, result)
	}
}

func CheckComparableEqualIgnoreOrder[T comparable](result []T, resultName string, expected []T, t TestT) {
	if !equalIgnoreOrder(result, expected) {
		t.Fatalf("%s: Provided slices: result: [%v] expected: [%v] are not equal.", resultName, result, expected)
	}
}

func CheckOk(ok bool, message string, t TestT) {
	if !ok {
		t.Fatal(message)
	}
}

func CheckNotOk(ok bool, message string, t TestT) {
	if ok {
		t.Fatal(message)
	}
}

// equalIgnoreOrder compares N slices for equal values ignoring the order of the items in the slices. Items must be comparable.
func equalIgnoreOrder[T comparable](slices ...[]T) bool {
	if len(slices) <= 1 {
		return true
	}

	sliceMap := make([]map[T]int, len(slices))
	firstSliceLen := len(slices[0])
	// Add all items across all slices to map, if any items are not identical, we'll end up with a map longer than the provided slices.
	for i := 0; i < len(slices); i++ {
		// Short circuit false if slices aren't all the same length
		if firstSliceLen != len(slices[i]) {
			return false
		}

		sliceMap[i] = make(map[T]int, firstSliceLen)

		// Increment item in map
		for _, item := range slices[i] {
			sliceMap[i][item]++
		}
	}

	isEqual := true
	// iterate over each item in the first map
	for k, v := range sliceMap[0] {
		// iterate over each map and compare it to the first maps item
		for i := 1; i < len(sliceMap); i++ {
			isEqual = sliceMap[i][k] == v

			if !isEqual {
				return false
			}
		}
	}

	return isEqual
}

func isNil(i any) bool {
	if i == nil {
		return true
	}

	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
