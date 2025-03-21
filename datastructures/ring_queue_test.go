package datastructures

import (
	"fmt"
	"testing"

	"github.com/zendesk/go-generics/internal/test"
)

func Test_Iter2(t *testing.T) {
	capacity := 5
	numItems := 23
	queue := NewRingQueue[int](capacity, false)
	for i := 0; i < numItems; i++ {
		queue.Push(i)
	}

	for i, val := range queue.Iter2() {
		test.CheckEqual(val, fmt.Sprintf("Item: %d", i), numItems-capacity+i, t)
	}
}

func Test_Iter(t *testing.T) {
	capacity := 5
	numItems := 23
	queue := NewRingQueue[int](capacity, false)
	for i := 0; i < numItems; i++ {
		queue.Push(i)
	}

	i := 0
	for val := range queue.Iter() {
		test.CheckEqual(val, fmt.Sprintf("Item: %d", i), numItems-capacity+i, t)
		i++
	}
}

func Test_Push(t *testing.T) {
	capacity := 5
	overlap := 14
	queue := NewRingQueue[int](capacity, false)
	for i := 0; i < capacity+overlap; i++ {
		queue.Push(i)
	}

	var expected []int
	for i := overlap; i < capacity+overlap; i++ {
		expected = append(expected, i)
	}

	test.CheckEqual(queue.Items(), "Expected mismatch", expected, t)
}

func Test_Pop(t *testing.T) {
	capacity := 7
	queue := NewRingQueue[int](capacity, false)
	for i := 0; i < capacity; i++ {
		queue.Push(i)
	}

	t.Log(queue.Items())
	for i := 0; i < capacity; i++ {
		val := queue.Pop()
		test.CheckEqual(val, fmt.Sprintf("I: %d Val: %d", i, val), i, t)
	}

	test.CheckEqual(queue.Size(), "size", 0, t)
}

func Test_Size(t *testing.T) {
	capacity := 7
	size := 5
	queue := NewRingQueue[int](capacity, false)
	for i := 0; i < size; i++ {
		queue.Push(i)
	}

	test.CheckEqual(queue.Size(), "size", size, t)
}

func Test_Items(t *testing.T) {
	capacity := 5
	overlap := 1
	queue := NewRingQueue[int](capacity, false)
	for i := 0; i < capacity+overlap; i++ {
		queue.Push(i)
	}

	expected := []int{1, 2, 3, 4, 5}
	got := queue.Items()
	test.CheckEqual(got, "Items", expected, t)
}
