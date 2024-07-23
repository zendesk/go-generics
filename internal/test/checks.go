package test

import (
	"fmt"
	"reflect"
	"testing"

	deep "github.com/go-test/deep"
)

// CheckErr will fail a test if an error is detected
func CheckErr(err error, message string, t *testing.T) {
	if err != nil {
		t.Fatal(fmt.Sprintf("%s - error: %v", message, err))
	}
}

// CheckErrs will fail a test if an error is detected
func CheckErrs(errs []error, message string, t *testing.T) {
	if len(errs) > 0 {
		var err error = fmt.Errorf("")
		for _, e := range errs {
			err = fmt.Errorf("%s;%s", e.Error(), err.Error())
		}
		t.Fatalf("%s %+v", message, err)
	}
}

// ExpectNil will fail a test if a value is not nil
func ExpectNil(value interface{}, t *testing.T) {
	if value != nil && !reflect.ValueOf(value).IsNil() {
		t.Fatal(fmt.Sprintf("expected nil but got: %+v", value))
	}
}

// ExpectNotEmpty will fail a test if a value is empty
func ExpectNotEmpty(value interface{}, valueName string, t *testing.T) {
	if value == "" {
		t.Fatal(fmt.Sprintf("expected not empty for value: %s", valueName))
	}
}

// ExpectNotNil will fail a test if a value is nil
func ExpectNotNil(varName string, value interface{}, t *testing.T) {
	if value == nil {
		t.Fatal(fmt.Sprintf("expected not nil for variable %s but got nil instead", varName))
	}
}

// ExpectErr will fail a test if an error is _not_ detected.
func ExpectErr(err error, message string, t *testing.T) {
	if err == nil {
		t.Fatal(fmt.Sprintf("%s - error: %v", message, err))
	}
}

func CheckEqual(result interface{}, resultName string, expectedResult interface{}, t *testing.T) {
	if diff := deep.Equal(result, expectedResult); diff != nil {
		t.Error(diff)
		t.Fatalf("Unexpected results for result: %s", resultName)
	}
}

func CheckComparableEqualIgnoreOrder[T comparable](result []T, resultName string, expected []T, t *testing.T) {
	areEqual := equalIgnoreOrder(result, expected)
	if !areEqual {
		t.Fatalf("%s: Provided slices: result: [%v] expected: [%v] are not equal.", resultName, result, expected)
	}
}

func CheckContainsDeepEqual[T any](list []T, item T) bool {
	for _, listItem := range list {
		if diff := deep.Equal(item, listItem); diff == nil {
			return true
		}
	}
	return false
}

func CheckContains[T any](list []T, item T, t *testing.T) {
	for _, listItem := range list {
		if diff := deep.Equal(item, listItem); diff == nil {
			return
		}
	}

	t.Fatalf("Item: %+v not found in list", item)
}

func CheckOk(ok bool, message string, t *testing.T) {
	if !ok {
		t.Fatal(message)
	}
}

func CheckNotOk(ok bool, message string, t *testing.T) {
	if ok {
		t.Fatal(message)
	}
}

func equalIgnoreOrder[T comparable](slices ...[]T) bool {
	if len(slices) <= 1 {
		return true
	}

	sliceMap := make([]map[T]int, len(slices))
	firstSliceLen := len(slices[0])
	// Add all items across all slices to map, if any items are not identical, we'll end up with a map longer than the provided slices.
	for i := 0; i < len(slices); i++ {
		sliceMap[i] = make(map[T]int)
		// Short circuit false if slices aren't all the same length
		if firstSliceLen != len(slices[i]) {
			return false
		}

		// Increment item in map
		for _, item := range slices[i] {
			sliceMap[i][item]++
		}
	}

	isEqual := true
	// iterate over each map (except the last), and compare it to the map after it
	for i := 0; i < len(sliceMap)-1; i++ {

		for k, v := range sliceMap[i] {
			isEqual = isEqual && sliceMap[i+1][k] == v
		}

		if !isEqual {
			return false
		}
	}

	return isEqual
}
