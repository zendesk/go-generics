package functions

import (
	"reflect"

	"github.com/zendesk/generic/set"
)

// EqualIgnoreOrder compares N slices for equal values ignoring the order of the items in the slices. Items must be comparable.
func EqualIgnoreOrder[T comparable](slices ...[]T) bool {
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

// Contains returns true/false on whether item is contained in list.
func Contains[T comparable](list []T, item T) bool {
	for _, listItem := range list {
		if listItem == item {
			return true
		}
	}
	return false
}

// ContainsAny compares slices A and B and returns true if at least one element of A exists in B
func ContainsAny[T comparable](A []T, B []T) bool {
	set := make(map[T]bool)
	for _, a := range A {
		set[a] = true
	}

	for _, b := range B {
		if set[b] {
			return true
		}
	}
	return false
}

// ContainsDeepEqual operates similar to Contains except performs comparisons by reflect.DeepEqual. This is capable of comparing
// types that do not implement 'comparable'
func ContainsDeepEqual[T any](list []T, item T) bool {
	for _, listItem := range list {
		if reflect.DeepEqual(item, listItem) {
			return true
		}
	}
	return false
}

// Dedupe the items in a comparable slice.
func Dedupe[T comparable](items []T) []T {
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

// DedupeByHash dedupes the items between 2 slices using a hashFn that should return equal values for two items that are considered equal
func DedupeByHash[T comparable](items []T, hashFn func(t T) uint64) []T {
	deduped := set.NewHashset(0, func(a, b T) bool {
		return hashFn(a) == hashFn(b)
	}, hashFn)

	for _, item := range items {
		deduped.Put(item)
	}

	return deduped.Keys()
}

// RemoveNils removes any nil values from a slice of T -- this is safe for finding boxed and unboxed nils.
func RemoveNils[T any](from []T) []T {
	var noNils = []T{}
	for _, t := range from {
		v := reflect.ValueOf(t)
		kind := v.Kind()
		switch kind {
		case reflect.Invalid:
			continue
		case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
			if !v.IsNil() {
				noNils = append(noNils, t)
			}
		default:
			// all other types are not nillable
			noNils = append(noNils, t)
		}
	}
	return noNils
}
