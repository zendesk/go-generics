package functions

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/zendesk/generic/set"
)

// Intersection returns the intersection of two slices
func Intersection[T comparable](a, b []T) []T {
	set := make(map[T]bool)
	var intersection []T
	for _, item := range a {
		set[item] = true
	}

	for _, item := range b {
		if set[item] {
			intersection = append(intersection, item)
		}
	}

	return intersection
}

// Join merges a slice[T] to a string separated by the provided separator. Items are rendered via fmt %v syntax.
func Join[T any](items []T, separator string) string {
	var stringSlice []string

	for _, item := range items {
		stringSlice = append(stringSlice, fmt.Sprintf("%v", item))
	}

	return strings.Join(stringSlice, separator)
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

// Remove deletes an element with index i from the slice and shifts all necessary items to preserver order
func Remove[T any](slice []T, i int) []T {
	return append(slice[:i], slice[i+1:]...)
}

// Generalize converts a slice[T] a slice[]any. Rarely useful except in particular cases.
func Generalize[T any](from []T) []any {
	var generalized []interface{}

	for _, t := range from {
		generalized = append(generalized, t)
	}

	return generalized
}

// Shuffle returns a COPY of the provided slice with shuffled elements
func Shuffle[T any](items []T) []T {
	copySlice := slices.Clone(items)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(copySlice), func(i, j int) { copySlice[i], copySlice[j] = copySlice[j], copySlice[i] })
	return copySlice
}
