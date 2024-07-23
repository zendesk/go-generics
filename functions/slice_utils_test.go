package functions

import (
	"testing"

	"github.com/zendesk/go-generics/internal/test"
	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

func TestJoin(t *testing.T) {
	test.CheckEqual(Join([]string{"foo", "bar", "baz"}, ","), "Test1", "foo,bar,baz", t)
	test.CheckEqual(Join([]int{1, 2, 3, 4}, ","), "Test1", "1,2,3,4", t)
	test.CheckEqual(Join([]float64{1.11, 2.22, 3.33, 4.44}, ","), "Test1", "1.11,2.22,3.33,4.44", t)
}

func TestGeneralize(t *testing.T) {
	ints := []int{1, 2, 3, 4}
	generalized := Generalize(ints)
	var generalType []interface{}

	test.CheckOk(utils.TypesMatch(generalized, generalType), "Expected []interface{} but did not get that", t)
}

func TestShuffle(t *testing.T) {
	testSlice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	updated := Shuffle(testSlice)

	test.CheckComparableEqualIgnoreOrder(testSlice, "Shuffled slices do not have the same items", updated, t)

	if testSlice[0] == updated[0] &&
		testSlice[1] == updated[1] &&
		testSlice[5] == updated[5] &&
		testSlice[2] == updated[2] {
		t.Fatalf("This slice was not shuffled properly, or we're REALLY unlucky.")
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name   string
		sliceA []int
		sliceB []int
		want   []int
	}{
		{
			"it returns the intersection when a matching element is found",
			[]int{1, 2, 3, 4, 5, 6},
			[]int{1, 8, 9, 10, 11, 6},
			[]int{1, 6},
		},
		{
			"it returns empty when no matching element is found",
			[]int{1, 2, 3, 4, 5, 6},
			[]int{7, 8, 9, 10, 11},
			[]int{},
		},
		{
			"it works with identical slices",
			[]int{1, 2, 3, 4, 5, 6},
			[]int{1, 2, 3, 4, 5, 6},
			[]int{1, 2, 3, 4, 5, 6},
		},
		{
			"it works with inverted slices",
			[]int{1, 2, 3, 4, 5, 6},
			[]int{6, 5, 4, 3, 2, 1},
			[]int{1, 2, 3, 4, 5, 6},
		},
		{
			"it works with supersets slices in B",
			[]int{1, 2, 3, 4, 5, 6},
			[]int{8, 6, 5, 4, 3, 2, 1, 7},
			[]int{1, 2, 3, 4, 5, 6},
		},
		{
			"it works with supersets slices in A",
			[]int{7, 1, 2, 3, 9, 4, 5, 6, 8},
			[]int{6, 5, 4, 3, 2, 1},
			[]int{1, 2, 3, 4, 5, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.CheckComparableEqualIgnoreOrder(Intersection(tt.sliceA, tt.sliceB), tt.name, tt.want, t)
		})
	}
}
