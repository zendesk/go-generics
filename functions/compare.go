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

// ContainsAll checks if all elements of the second list exist in the first list, ignoring order.
func ContainsAll[T comparable](list1, list2 []T) bool {

	if len(list2) == 0 {
		return false
	}

	// Create a map to count occurrences of elements in list1
	countMap := make(map[T]int)

	// Populate the map with counts of each element in list1
	for _, item := range list1 {
		countMap[item]++
	}

	// Check if all elements in list2 exist in the countMap
	for _, item := range list2 {
		if countMap[item] == 0 {
			return false // If an element in list2 is not found in list1, return false
		}
		countMap[item]-- // Decrease the count for the found element
	}

	// All elements in list2 are found in list1
	return true
}
