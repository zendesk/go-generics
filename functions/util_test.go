//go:build test
// +build test

package functions

import (
	"testing"
	"time"

	"github.com/zendesk/go-generics/internal/test"
)

type Foo struct {
	Foo string
}

type Bar struct {
	bar string
}

func Test_typesMatch(t *testing.T) {
	foo1 := Foo{}
	foo2 := Foo{
		Foo: "123",
	}

	bar1 := Bar{}

	test.CheckOk(typesMatch(foo1, foo2), "Types do not match but should", t)
	test.CheckOk(!typesMatch(foo1, bar1), "Types match but should not", t)
	test.CheckOk(!typesMatch(foo1, nil), "Types match but should not", t)
	test.CheckOk(!typesMatch(nil, bar1), "Types match but should not", t)
}

func Test_randomDurationBetweenMinGTMax(t *testing.T) {
	result := randomDurationBetween(time.Second, time.Microsecond)
	test.CheckEqual(result, "Min duration was not returned but was expected.", time.Second, t)
}

func Test_firstNotEmpty(t *testing.T) {
	items := []string{"", "", "", "abc", "def"}
	first, ok := firstNotEmpty(items...)
	test.CheckEqual(first, "first item", "abc", t)
	test.CheckOk(ok, "Found item", t)

	items2 := []string{"", "", ""}
	first1, ok1 := firstNotEmpty(items2...)
	test.CheckEqual(first1, "first item", "", t)
	test.CheckNotOk(ok1, "Item not found", t)
}
