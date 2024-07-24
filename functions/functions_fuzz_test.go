package functions

import (
	"crypto/sha256"
	"fmt"

	"github.com/zendesk/go-generics/internal/test"
)

const (
	seedIterations              = 50   // Each fuzz test will run this many times to generate seeds
	seedRateLimitIterations     = 10   // rate limit tests will run this many times to generate seeds
	maxSliceSizeLength          = 5000 // Slices provided by fuzz tests to generic functions will never exceed this length
	maxSliceSizeLengthRateLimit = 1500 // cannot be too large due to time constraints
	minRatePerInterval          = 100  // minimum throughput for a rate limit test for each interval (which has a max of 1 second)
	maxMutationExpansion        = 4    // during some tests, 1 foo will be transformed to this maximum # of bars -- recommend low number to prevent memory problems.
)

var toBing = func(foo *test.Foo) string {
	return fmt.Sprintf("%s-%s", foo.Bar, foo.Baz)
}

var toBar = func(f *test.Foo) *test.Bar {
	return &test.Bar{
		Order: f.Order,
		Bing:  toBing(f),
	}
}

var toKeyValue = func(f *test.Foo) (int, string) {
	return f.Order, f.Bar
}

var toKeyValueWithErrs = func(f *test.Foo) (int, string, error) {
	if f.Order%3 == 0 {
		return f.Order, f.Bar, fmt.Errorf("ERROR")
	}
	return f.Order, f.Bar, nil
}

var toManyBars = func(num int) func(*test.Foo) (bars []*test.Bar) {
	return func(f *test.Foo) (bars []*test.Bar) {
		for i := 0; i < num; i++ {
			bars = append(bars, &test.Bar{
				// Since a single foo is being mapped to many bars, space out bars ordering to prevent collisions
				Order: 5*f.Order + i,
				Bing:  fmt.Sprintf("%s-%d", toBing(f), i),
			})
		}
		return bars
	}
}

var toManyBarsWithErr = func(num int) func(*test.Foo) ([]*test.Bar, error) {
	return func(f *test.Foo) ([]*test.Bar, error) {
		if f.Order%5 == 0 {
			return nil, fmt.Errorf("error %d", f.Order)
		}
		return toManyBars(num)(f), nil
	}
}

var toBarWithErr = func(t *test.Foo) (*test.Bar, error) {
	if t.Order%5 == 0 {
		return nil, fmt.Errorf("error %d", t.Order)
	}

	return &test.Bar{
		Order: t.Order,
		Bing:  toBing(t),
	}, nil
}

var mutateFoo = func(f *test.Foo) {
	_ = mutateFooWithErr(f)
}

var mutateFooWithErr = func(t *test.Foo) error {
	t.Bar = t.Baz + fmt.Sprintf("%d", t.Order)

	if t.Order%5 == 0 {
		return fmt.Errorf("error %d", t.Order)
	}
	return nil
}

var hashByOrder = func(f *test.Foo) string {
	return hashAny(f.Order)
}

var mapMap = func(key int, val *test.Foo) *test.Bar {
	return &test.Bar{
		Bing:  fmt.Sprintf("%s-%s", val.Bar, val.Baz),
		Order: key,
	}
}

var mapMapWithErr = func(key int, val *test.Foo) (*test.Bar, error) {
	bar := &test.Bar{
		Bing:  fmt.Sprintf("%s-%s", val.Bar, val.Baz),
		Order: key,
	}

	if val.Order%5 == 0 {
		return bar, fmt.Errorf("error %d", val.Order)
	} else {
		return bar, nil
	}
}

func hashAny(obj any) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
