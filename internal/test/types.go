package test

type Foo struct {
	Bar   string
	Baz   string
	Order int
}

type Bar struct {
	Bing  string
	Order int
}

func MakeFoos(num int) []*Foo {
	var foos []*Foo
	for i := 0; i < num; i++ {
		foos = append(foos, &Foo{
			Bar:   GenerateRandomLetterString(RandomNumber(12)),
			Baz:   GenerateRandomLetterString(RandomNumber(12)),
			Order: RandomNumber(999999),
		})
	}
	return foos
}

func MakeFoosOrderly(num int) []*Foo {
	var foos = make([]*Foo, 0)
	for i := 0; i < num; i++ {
		foos = append(foos, &Foo{
			Bar:   GenerateRandomLetterString(RandomNumber(12)),
			Baz:   GenerateRandomLetterString(RandomNumber(12)),
			Order: i,
		})
	}
	return foos
}

func MakeFooMaps(num int) map[int]*Foo {
	fooMap := make(map[int]*Foo)
	foos := MakeFoos(num)
	for i := 0; i < num; i++ {
		fooMap[i] = foos[i]
	}
	return fooMap
}
