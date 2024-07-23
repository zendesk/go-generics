package functions

import (
	"testing"
	"time"

	"github.com/zendesk/lockbox-shared-lib/lockbox/test"
	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

type Foo struct {
	Foo string
}

type Bar struct {
	bar string
}

func TestTypesMatch(t *testing.T) {
	foo1 := Foo{}
	foo2 := Foo{
		Foo: "123",
	}

	bar1 := Bar{}

	test.CheckOk(utils.TypesMatch(foo1, foo2), "Types do not match but should", t)
	test.CheckOk(!utils.TypesMatch(foo1, bar1), "Types match but should not", t)
	test.CheckOk(!utils.TypesMatch(foo1, nil), "Types match but should not", t)
	test.CheckOk(!utils.TypesMatch(nil, bar1), "Types match but should not", t)
}

func TestRandomDurationBetweenMinGTMax(t *testing.T) {
	result := utils.RandomDurationBetween(time.Second, time.Microsecond)
	test.CheckEqual(result, "Min duration was not returned but was expected.", time.Second, t)
}

func TestFirstNotEmpty(t *testing.T) {
	items := []string{"", "", "", "abc", "def"}
	first, ok := utils.FirstNotEmpty(items...)
	test.CheckEqual(first, "first item", "abc", t)
	test.CheckOk(ok, "Found item", t)

	items2 := []string{"", "", ""}
	first1, ok1 := utils.FirstNotEmpty(items2...)
	test.CheckEqual(first1, "first item", "", t)
	test.CheckNotOk(ok1, "Item not found", t)
}
