package functions

import (
	"reflect"
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

// ContainsAll checks if all elements of list1 are present in list2.
// It does not require the elements to be contiguous.
func ContainsAll[T comparable](list1, list2 []T) bool {
	// Create a set (map) from list2 for efficient lookup
	list2Set := make(map[T]bool)
	for _, item := range list2 {
		list2Set[item] = true
	}

	// Check each item from list1 exists in the set
	for _, item := range list1 {
		if !list2Set[item] {
			return false
		}
	}

	return true
}
