package functions

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/exp/slices"
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

// Remove deletes an element with index i from the slice and shifts all necessary items to preserver order
func Remove[T any](slice []T, i int) []T {
	return append(slice[:i], slice[i+1:]...)
}

// Generalize converts a slice[T] a slice[]interface{}. Rarely useful except in particular cases.
func Generalize[T any](from []T) []interface{} {
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
