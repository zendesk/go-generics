//go:build test

package functions

import (
	"reflect"
	"testing"
)

// Table driven test for Batch function
func TestBatch(t *testing.T) {
	tests := []struct {
		name      string
		items     []int
		batchSize int
		expected  [][]int
	}{
		{
			name:      "empty slice",
			items:     []int{},
			batchSize: 2,
			expected:  nil,
		},
		{
			name:      "single item",
			items:     []int{1},
			batchSize: 2,
			expected:  [][]int{{1}},
		},
		{
			name:      "multiple items",
			items:     []int{1, 2, 3, 4, 5},
			batchSize: 2,
			expected:  [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name:      "batch size larger than slice",
			items:     []int{1, 2, 3},
			batchSize: 10,
			expected:  [][]int{{1, 2, 3}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Batch(test.items, test.batchSize)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func FuzzBatch(f *testing.F) {
	// Add a seed corpus
	f.Add(uint16(29), 2)
	f.Add(uint16(18), 1)
	f.Add(uint16(1), 0)
	f.Add(uint16(0), 3)
	f.Add(uint16(5000), 10)

	f.Fuzz(func(t *testing.T, sliceSize uint16, batchSize int) {
		if batchSize <= 0 {
			batchSize = 1
		}

		// Make slice of items
		items := make([]int, sliceSize)
		for i := uint16(0); i < sliceSize; i++ {
			items[i] = int(i)
		}

		result := Batch(items, batchSize)
		if len(result) == 0 && len(items) > 0 {
			t.Errorf("expected non-empty result for non-empty input")
		}

		remainder := len(items) % batchSize
		expectedBatches := int(len(items) / batchSize)
		if remainder > 0 {
			expectedBatches++
		}
		if len(result) != expectedBatches {
			t.Errorf("expected %d batches, got %d witn # items: %d and batch size: %d", expectedBatches, len(result), len(items), batchSize)
		}
	})
}
