//go:build test
// +build test

package datastructures

import (
	"testing"

	"github.com/zendesk/go-generics/internal/test"
)

func TestStack(t *testing.T) {
	stack := Stack[int]{}
	items := []int{1, 2, 3, 4, 5, 6}
	for _, i := range items {
		stack.Push(i)
	}

	test.CheckEqual(stack.Size(), "Size", len(items), t)
	test.CheckEqual(stack.Peek(), "Peek", items[len(items)-1], t)
	test.CheckOk(stack.HasMore(), "HasMore", t)
	test.CheckEqual(stack.Copy(), "Copy", stack, t)

	for idx, _ := range items {
		test.CheckEqual(stack.Pop(), "Popped Items", items[len(items)-1-idx], t)
	}
}
