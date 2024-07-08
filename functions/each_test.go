package functions_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/zendesk/go-generics/functions"
	"github.com/zendesk/go-generics/internal/test"
	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

func TestEachMergeErrs(t *testing.T) {
	errOnEvens := func(i int) error {
		if i%2 == 0 {
			return fmt.Errorf("error: %d.", i)
		} else {
			return nil
		}
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	foundErrs := functions.EachMergeErrs(items, errOnEvens)

	expectedErrsFound := strings.Contains(foundErrs.Error(), "error: 2.") &&
		strings.Contains(foundErrs.Error(), "error: 4.") &&
		strings.Contains(foundErrs.Error(), "error: 6.") &&
		strings.Contains(foundErrs.Error(), "error: 8.") &&
		strings.Contains(foundErrs.Error(), "error: 10.")

	missingErrorsAreMissing := strings.Contains(foundErrs.Error(), "error: 1.") ||
		strings.Contains(foundErrs.Error(), "error: 3.") ||
		strings.Contains(foundErrs.Error(), "error: 5.") ||
		strings.Contains(foundErrs.Error(), "error: 7.") ||
		strings.Contains(foundErrs.Error(), "error: 9.")

	t.Logf(foundErrs.Error())
	test.CheckOk(expectedErrsFound, "Expected errors do not exist", t)
	test.CheckOk(!missingErrorsAreMissing, "Errors have been found that shouldn't be here!", t)
}

func FuzzEach(f *testing.F) {
	for i := 0; i < 1; i++ {
		f.Add(utils.RandomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)

		functions.Each(foos, mutateFoo)
		var expectedErrs []error
		for _, foo := range foos {
			err := mutateFooWithErr(foo)
			if err != nil {
				expectedErrs = append(expectedErrs, err)
			}
		}

		for _, foo := range foos {
			test.CheckEqual(foo.Bar, fmt.Sprintf("Foo: %d", foo.Order), foo.Baz+fmt.Sprintf("%d", foo.Order), t)
		}
	})
}

func FuzzGoEach(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(utils.RandomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)

		if num%2 == 0 {
			functions.GoEach(foos, mutateFoo)
		} else {
			functions.GoEach(foos, mutateFoo, functions.ConcurrencyLimitOption(num/3+1))
		}

		var expectedErrs []error
		for _, foo := range foos {
			err := mutateFooWithErr(foo)
			if err != nil {
				expectedErrs = append(expectedErrs, err)
			}
		}

		for _, foo := range foos {
			test.CheckEqual(foo.Bar, fmt.Sprintf("Foo: %d", foo.Order), foo.Baz+fmt.Sprintf("%d", foo.Order), t)
		}
	})
}

func FuzzGoEachWithErrs(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(utils.RandomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)

		var errs []error

		if num%2 == 0 {
			errs = functions.GoEachWithErrs(foos, mutateFooWithErr)
		} else {
			errs = functions.GoEachWithErrs(foos, mutateFooWithErr, functions.ConcurrencyLimitOption(num/3+1))
		}

		var expectedErrs []error
		for _, foo := range foos {
			err := mutateFooWithErr(foo)
			if err != nil {
				expectedErrs = append(expectedErrs, err)
			}
		}

		// Order slices before compare
		sort.Slice(expectedErrs, func(i, j int) bool {
			return expectedErrs[i].Error() > expectedErrs[j].Error()
		})
		sort.Slice(errs, func(i, j int) bool {
			return errs[i].Error() > errs[j].Error()
		})

		test.CheckEqual(errs, "Errs", expectedErrs, t)
		for _, foo := range foos {
			test.CheckEqual(foo.Bar, fmt.Sprintf("Foo: %d", foo.Order), foo.Baz+fmt.Sprintf("%d", foo.Order), t)
		}
	})
}

func FuzzGoEachWithErrsRateLimitTest(f *testing.F) {
	for i := 0; i < seedRateLimitIterations; i++ {
		sliceSize := utils.RandomNumber(maxSliceSizeLengthRateLimit)
		// we want to ensure rate < sliceSize otherwise no throttling will occur and we cannot estimate expectedDuration. Also rate cannot be 0
		rate := utils.RandomNumberBetween(minRatePerInterval, (sliceSize+1)/5+1)
		duration := utils.RandomDurationBetween(time.Millisecond, time.Second).Nanoseconds()

		// If rate is very low, reset rate to ensure we don't run TOO long (max 50 seconds with this change)
		if sliceSize != 0 && sliceSize/rate > 50 {
			rate = sliceSize / 10
		}

		f.Add(sliceSize, rate, duration)
		f.Logf("Adding: %d, %d, %d", sliceSize, rate, duration)
	}

	f.Fuzz(func(t *testing.T, num int, ratePerTime int, durationNanoseconds int64) {
		duration := time.Duration(durationNanoseconds)
		foos := test.MakeFoos(num)

		// estimate expected execution time given rate limit
		concurrency := utils.RandomNumberBetween(1, 20) // Concurrency doesn't matter, rate is limited across goroutines

		var expectedDurationMillis float64
		// excluding the first batch, we can assume rate-limiting for all subsequent batches at the per-time interval. First batch starts immediately
		if ratePerTime > 0 {
			expectedDurationMillis = float64(len(foos)-ratePerTime)/(float64(ratePerTime)/float64(duration.Milliseconds())) - 1
		} else {
			// if no rate limit is specified, expect this to be very fast
			expectedDurationMillis = 0
		}

		// execute
		start := time.Now().UnixMilli()
		_ = functions.GoEachWithErrs(foos, mutateFooWithErr, functions.ConcurrencyLimitOption(concurrency), functions.RateLimitOption(ratePerTime, duration))
		finish := time.Now().UnixMilli()

		totalTime := float64(finish - start)

		// If no rate limiting was happening, the actual process time would be nanoseconds long
		// and far below minProcessTime. Min process time is best-case scenario
		// We cannot reasonably estimate max process time because we're on a system that is loaded by concurrent Fuzz tests
		// and CPU wait is a real thing. What we _do_ know, is that the test should not finish before minProcessTime elapses
		minProcessTime := expectedDurationMillis

		if minProcessTime <= totalTime {
			t.Logf("SUCCESS: Process took %f millis. Expected at least %f. Inputs(%d, %d, %d)", totalTime, minProcessTime, num, ratePerTime, durationNanoseconds)
		} else {
			t.Fatalf("FAILURE: Process took %f millis. Expected at least %f. Inputs(%d, %d, %d)", totalTime, minProcessTime, num, ratePerTime, durationNanoseconds)
		}
	})

}
