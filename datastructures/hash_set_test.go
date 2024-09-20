package datastructures

import (
	"reflect"
	"sort"
	"testing"

	"github.com/zendesk/go-generics/internal/test"
)

// Define a struct to use as input
type Person struct {
	Name string
	Age  int
}

func Test_HashSet_Struct(t *testing.T) {
	alice := Person{"Alice", 30}
	bob := Person{"Bob", 25}
	steve := Person{"Steve", 14}

	tests := []struct {
		name      string
		input     []Person
		ops       func(set ISet[Person])
		expected  []Person
		validator func(ISet[Person], *testing.T)
	}{
		{
			name: "Put",
			input: []Person{
				{"Alice", 30},
				{"Bob", 25},
			},
			ops: func(set ISet[Person]) {
				for _, p := range set.Values() {
					set.Put(p)
				}
			},
			expected: []Person{
				{"Alice", 30},
				{"Bob", 25},
			},
		},
		{
			name: "Remove",
			input: []Person{
				{"Alice", 30},
				{"Bob", 25},
			},
			ops: func(set ISet[Person]) {
				set.Remove(Person{"Alice", 30})
			},
			expected: []Person{
				{"Bob", 25},
			},
		},
		{
			name: "Has",
			input: []Person{
				{"Alice", 30},
				{"Bob", 25},
			},
			ops: func(set ISet[Person]) {
				set.Put(alice)
				set.Remove(alice)
			},
			expected: []Person{
				{"Bob", 25},
			},
			validator: func(s ISet[Person], t *testing.T) {
				test.CheckOk(s.Has(bob), "Bob should be in the set", t)
				test.CheckNotOk(s.Has(alice), "Alice should not be in the set", t)
				test.CheckNotOk(s.Has(steve), "Steve should not be in the set", t)
			},
		},
		{
			name: "Clear",
			input: []Person{
				{"Alice", 30},
				{"Bob", 25},
			},
			ops: func(set ISet[Person]) {
				set.Clear()
			},
			expected: []Person{},
			validator: func(s ISet[Person], t *testing.T) {
				test.CheckOk(s.Size() == 0, "Size should be 0", t)
			},
		},
		{
			name: "Size",
			input: []Person{
				{"Alice", 30},
				{"Bob", 25},
			},
			ops: func(set ISet[Person]) {
			},
			expected: []Person{
				alice, bob,
			},
			validator: func(s ISet[Person], t *testing.T) {
				test.CheckOk(s.Size() == 2, "Size should be 0", t)
			},
		},
		{
			name: "Copy",
			input: []Person{
				{"Alice", 30},
				{"Bob", 25},
			},
			ops: func(set ISet[Person]) {
			},
			expected: []Person{
				alice, bob,
			},
			validator: func(s ISet[Person], t *testing.T) {
				newSet := s.Copy()
				newValues := newSet.Values()
				oldValues := s.Values()
				// Sort before compare
				sort.Slice(newValues, func(i, j int) bool { return newValues[i].Name < newValues[j].Name })
				sort.Slice(oldValues, func(i, j int) bool { return oldValues[i].Name < oldValues[j].Name })
				test.CheckEqual(newValues, "Deep equal", oldValues, t)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set := newHashSet[Person]()
			for _, p := range tc.input {
				set.Put(p)
			}
			tc.ops(set)
			actual := set.Values()

			// sort slices before compare
			sort.Slice(actual, func(i, j int) bool { return actual[i].Name < actual[j].Name })
			sort.Slice(tc.expected, func(i, j int) bool { return actual[i].Name < actual[j].Name })
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Test %s failed: expected %v, got %v", tc.name, tc.expected, actual)
			}

			if tc.validator != nil {
				tc.validator(set, t)
			}
		})
	}
}

func Test_HashSet_All(t *testing.T) {
	alice := Person{"Alice", 30}
	bob := Person{"Bob", 25}
	steve := Person{"Steve", 14}

	tests := []struct {
		name     string
		input    []Person
		expected []Person
	}{
		{
			name:     "Empty Set",
			input:    []Person{},
			expected: []Person{},
		},
		{
			name: "Single Element Set",
			input: []Person{
				alice,
			},
			expected: []Person{
				alice,
			},
		},
		{
			name: "Multiple Elements Set",
			input: []Person{
				alice, bob, steve,
			},
			expected: []Person{
				alice, bob, steve,
			},
		},
		{
			name: "Elements with the same values",
			input: []Person{
				alice, alice, bob,
			},
			expected: []Person{
				alice, bob,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set := newHashSet[Person]()
			for _, p := range tc.input {
				set.Put(p)
			}

			var iterated []Person
			for _, v := range set.All() {
				iterated = append(iterated, v)
			}

			test.CheckEqualEquateEmpty(iterated, "equal", tc.expected, t)
		})
	}
}

func Test_HashSet_OrderMaintained(t *testing.T) {
	alice := Person{"Alice", 30}
	bob := Person{"Bob", 25}
	steve := Person{"Steve", 14}
	emma := Person{"Emma", 40}

	tests := []struct {
		name     string
		ops      func(set ISet[Person])
		expected []Person
	}{
		{
			name: "Add and Remove in Different Orders",
			ops: func(set ISet[Person]) {
				set.Put(alice)
				set.Put(bob)
				set.Remove(alice)
				set.Put(steve)
				set.Put(alice) // Alice added again
			},
			expected: []Person{bob, steve, alice},
		},
		{
			name: "Add Remove and Re-Add the Same Item",
			ops: func(set ISet[Person]) {
				set.Put(alice)
				set.Remove(alice)
				set.Put(alice) // Alice re-added
				set.Put(bob)
				set.Put(steve)
			},
			expected: []Person{alice, bob, steve},
		},
		{
			name: "Removing Non-Existing Item",
			ops: func(set ISet[Person]) {
				set.Put(alice)
				set.Remove(bob) // Bob is not in the set
				set.Put(bob)
				set.Put(steve)
			},
			expected: []Person{alice, bob, steve},
		},
		{
			name: "Adding, Removing, and Re-Adding Items in Complex Order",
			ops: func(set ISet[Person]) {
				set.Put(alice)
				set.Put(bob)
				set.Put(steve)
				set.Remove(bob)
				set.Put(emma)
				set.Remove(alice)
				set.Put(bob)   // Bob re-added
				set.Put(alice) // Alice re-added
			},
			expected: []Person{steve, emma, bob, alice},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			set := newHashSet[Person]()
			tc.ops(set)

			var iterated []Person
			for _, v := range set.All() {
				iterated = append(iterated, v)
			}

			if !reflect.DeepEqual(iterated, tc.expected) {
				t.Errorf("Test %s failed: expected %v, got %v", tc.name, tc.expected, iterated)
			}
		})
	}
}
