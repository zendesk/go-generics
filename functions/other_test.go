package functions

import (
	"fmt"
	"testing"

	"github.com/zendesk/go-generics/functions/internal/test"
)

func TestMin(t *testing.T) {
	test.CheckEqual(Min(5, 1, 2, 4), "Test1", 1, t)
	test.CheckEqual(Min(1, 1), "Test2", 1, t)
	test.CheckEqual(Min(1), "Test3", 1, t)
	test.CheckEqual(Min(99, 9, 9, 9), "Test4", 9, t)
	test.CheckEqual(Min(99, -14, 9, 9), "Test5", -14, t)
	var empty []int
	test.CheckEqual(Min(empty...), "Test6", 0, t)
}

func TestMax(t *testing.T) {
	test.CheckEqual(Max(5, 1, 2, 4), "Test1", 5, t)
	test.CheckEqual(Max(1, 1, 4), "Test2", 4, t)
	test.CheckEqual(Max(1, 1, 1, 0, -10), "Test3", 1, t)
	test.CheckEqual(Max(-9999, 99999, 0), "Test4", 99999, t)
	var empty []int
	test.CheckEqual(Max(empty...), "Test6", 0, t)
}

func TestCopy(t *testing.T) {
	testMap1 := map[int]string{
		1:    "asdf",
		2:    "bcd",
		9999: "abc",
	}

	copyMap := Copy(testMap1)

	test.CheckEqual(copyMap, "Test1", testMap1, t)
	test.CheckNotOk(&testMap1 == &copyMap, "Test2", t)
}

func TestConvert(t *testing.T) {
	toString := func(t int) string {
		return fmt.Sprintf("%d", t)
	}

	from := 1
	expected := "1"

	result := Convert(from, toString)
	test.CheckEqual(result, "Test1", expected, t)
}
