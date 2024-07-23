package functions

import (
	"testing"

	"github.com/zendesk/go-generics/internal/test"
)

func TestReduce(t *testing.T) {
	start := []int{1, 2, 3, 3, 9, 1, 1}
	sum := 0

	sum = Reduce(start, sum, func(from int, to int) int {
		return from + to
	})

	test.CheckEqual(sum, "TestIntReduce", 20, t)

	startFloat := []float64{1.9, 2.1, 3.3, 3.09099099, 9, 1, 1.10}
	sumFloat := 0.0

	sumFloat = Reduce(startFloat, sumFloat, func(from float64, to float64) float64 {
		return from + to
	})

	test.CheckEqual(sumFloat, "TestFloatReduce", 21.49099099, t)
}
