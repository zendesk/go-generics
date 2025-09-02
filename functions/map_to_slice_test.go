//go:build test
// +build test

package functions

import (
	"sort"
	"testing"
	"time"

	"github.com/zendesk/go-generics/internal/test"
)

func FuzzMapToSlice(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		fooMaps := test.MakeFooMaps(num)

		var bars []*test.Bar
		bars = MapToSlice(fooMaps, mapMap)

		var expectedBars []*test.Bar
		for k, v := range fooMaps {
			bar := mapMap(k, v)
			expectedBars = append(expectedBars, bar)
		}

		// Order slices before compare
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		sort.Slice(expectedBars, func(i, j int) bool {
			return expectedBars[i].Order > expectedBars[j].Order
		})

		test.CheckEqual(bars, "Bars", expectedBars, t)
	})
}

func FuzzGoMapToSlice(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		fooMaps := test.MakeFooMaps(num)

		var bars []*test.Bar
		if num%2 == 0 {
			bars = GoMapToSlice(fooMaps, mapMap)
		} else {
			bars = GoMapToSlice(fooMaps, mapMap, ConcurrencyLimitOption(num/3+1))
		}

		var expectedBars []*test.Bar
		for k, v := range fooMaps {
			bar := mapMap(k, v)
			expectedBars = append(expectedBars, bar)
		}

		// Order slices before compare
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		sort.Slice(expectedBars, func(i, j int) bool {
			return expectedBars[i].Order > expectedBars[j].Order
		})

		test.CheckEqual(bars, "Bars", expectedBars, t)
	})
}

func FuzzGoMapToSliceWithErrs(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		fooMaps := test.MakeFooMaps(num)

		var errs []error
		var bars []*test.Bar
		if num%2 == 0 {
			bars, errs = GoMapToSliceWithErrs(fooMaps, mapMapWithErr, DiscardResultIfErrOption())
		} else {
			bars, errs = GoMapToSliceWithErrs(fooMaps, mapMapWithErr, ConcurrencyLimitOption(num/3+1), DiscardResultIfErrOption())
		}

		var expectedErrs []error
		var expectedBars []*test.Bar
		for k, v := range fooMaps {
			bar, err := mapMapWithErr(k, v)
			if err != nil {
				expectedErrs = append(expectedErrs, err)
			} else {
				expectedBars = append(expectedBars, bar)
			}
		}

		// Order slices before compare
		sort.Slice(expectedErrs, func(i, j int) bool {
			return expectedErrs[i].Error() > expectedErrs[j].Error()
		})
		sort.Slice(errs, func(i, j int) bool {
			return errs[i].Error() > errs[j].Error()
		})

		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		sort.Slice(expectedBars, func(i, j int) bool {
			return expectedBars[i].Order > expectedBars[j].Order
		})

		test.CheckEqual(bars, "Bars", expectedBars, t)

		expectedStrs := Map(expectedErrs, func(e error) string { return e.Error() })
		errStrs := Map(errs, func(e error) string { return e.Error() })
		test.CheckEqual(expectedStrs, "Errs", errStrs, t)
	})
}

func FuzzGoMapToSliceWithErrsRateLimitTest(f *testing.F) {
	for i := 0; i < seedRateLimitIterations; i++ {
		sliceSize := randomNumber(maxSliceSizeLengthRateLimit)
		// we want to ensure rate < sliceSize otherwise no throttling will occur and we cannot estimate expectedDuration. Also rate cannot be 0
		rate := randomNumberBetween(minRatePerInterval, (sliceSize+1)/5+1)
		duration := randomDurationBetween(time.Millisecond, time.Second).Nanoseconds()

		// If rate is very low, reset rate to ensure we don't run TOO long (max 50 seconds with this change)
		if sliceSize != 0 && sliceSize/rate > 50 {
			rate = sliceSize / 10
		}

		f.Add(sliceSize, rate, duration)
		f.Logf("Adding: %d, %d, %d", sliceSize, rate, duration)
	}

	f.Fuzz(func(t *testing.T, num int, ratePerTime int, durationNanoseconds int64) {
		duration := time.Duration(durationNanoseconds)
		fooMaps := test.MakeFooMaps(num)

		// estimate expected execution time given rate limit
		concurrency := randomNumberBetween(1, 20) // Concurrency doesn't matter, rate is limited across goroutines

		var expectedDurationMillis float64
		// excluding the first batch, we can assume rate-limiting for all subsequent batches at the per-time interval. First batch starts immediately
		if ratePerTime > 0 {
			expectedDurationMillis = float64(len(fooMaps)-ratePerTime)/(float64(ratePerTime)/float64(duration.Milliseconds())) - 1
		} else {
			// if no rate limit is specified, expect this to be very fast
			expectedDurationMillis = 0
		}

		// execute
		start := time.Now().UnixMilli()
		_, _ = GoMapToSliceWithErrs(fooMaps, mapMapWithErr, ConcurrencyLimitOption(concurrency), RateLimitOption(ratePerTime, duration))
		finish := time.Now().UnixMilli()

		totalTime := float64(finish - start)

		// If no rate limiting was happening, the actual process time would be nanoseconds long
		// and far below minProcessTime. Min process time is best-case scenario
		// We cannot reasonably estimate max process time because we're on a system that is loaded by concurrent Fuzz tests
		// and CPU wait is a real thing. What we _do_ know, is that the test should not finish before minProcessTime elapses
		minProcessTime := expectedDurationMillis

		if minProcessTime <= totalTime {
			t.Logf("SUCCESS: Process took %f millis. Expected at least %f", totalTime, minProcessTime)
		} else {
			t.Fatalf("FAILURE: Process took %f millis. Expected at least %f", totalTime, minProcessTime)
		}
	})
}
