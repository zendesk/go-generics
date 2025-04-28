package functions

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/zendesk/go-generics/functions/internal/test"
)

func TestMapWithErrs(t *testing.T) {
	errOnEvens := func(i int) (int, error) {
		if i%2 == 0 {
			return -1, fmt.Errorf("error: %d.", i)
		} else {
			return i, nil
		}
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expected := []int{1, 3, 5, 7, 9}
	results, foundErrs := MapWithErrs(items, errOnEvens)
	resultsNoNegative := Filter(results, func(i int) bool {
		return i != -1
	})

	foundMergedErrs := MergeErrors(foundErrs...)
	expectedErrsFound := strings.Contains(foundMergedErrs.Error(), "error: 2.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 4.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 6.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 8.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 10.")

	missingErrorsAreMissing := strings.Contains(foundMergedErrs.Error(), "error: 1.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 3.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 5.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 7.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 9.")

	test.CheckEqual(resultsNoNegative, "Expected Odds", expected, t)

	test.CheckOk(expectedErrsFound, "Expected errors do not exist", t)
	test.CheckOk(!missingErrorsAreMissing, "Errors have been found that shouldn't be here!", t)
}

func TestMapMergeErrs(t *testing.T) {
	errOnEvens := func(i int) (int, error) {
		if i%2 == 0 {
			return -1, fmt.Errorf("error: %d.", i)
		} else {
			return i, nil
		}
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expected := []int{1, 3, 5, 7, 9}
	results, foundErrs := MapMergeErrs(items, errOnEvens)
	resultsNoNegative := Filter(results, func(i int) bool {
		return i != -1
	})

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

	test.CheckEqual(resultsNoNegative, "Expected Odds", expected, t)

	t.Logf(foundErrs.Error())
	test.CheckOk(expectedErrsFound, "Expected errors do not exist", t)
	test.CheckOk(!missingErrorsAreMissing, "Errors have been found that shouldn't be here!", t)
}

func TestDiscardResultsIfErr(t *testing.T) {
	errOnEvens := func(i int) (int, error) {
		if i%2 == 0 {
			return -1, fmt.Errorf("error: %d.", i)
		} else {
			return i, nil
		}
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expected := []int{1, 3, 5, 7, 9}
	results, foundErrs := GoMapWithErrs(items, errOnEvens, DiscardResultIfErrOption(), RandomOrderOption())
	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})

	foundMergedErrs := MergeErrors(foundErrs...)
	expectedErrsFound := strings.Contains(foundMergedErrs.Error(), "error: 2.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 4.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 6.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 8.") &&
		strings.Contains(foundMergedErrs.Error(), "error: 10.")

	missingErrorsAreMissing := strings.Contains(foundMergedErrs.Error(), "error: 1.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 3.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 5.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 7.") ||
		strings.Contains(foundMergedErrs.Error(), "error: 9.")

	test.CheckEqual(results, "Expected Odds", expected, t)

	test.CheckOk(expectedErrsFound, "Expected errors do not exist", t)
	test.CheckOk(!missingErrorsAreMissing, "Errors have been found that shouldn't be here!", t)
}

func FuzzMap(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)

		// remove any ones with dupe order b/c that's how we're ordering below.
		foos = DedupeByHash(foos, hashByOrder)
		bars := Map(foos, toBar)

		if len(bars) != len(foos) {
			t.Fatal("Mapping failed, data was lost.")
		}

		// Validate data was mutated properly by provided function
		sort.Slice(foos, func(i, j int) bool {
			return foos[i].Order > foos[j].Order
		})
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		for i, foo := range foos {
			test.CheckEqual(bars[i].Bing, "Bing", toBing(foo), t)
			test.CheckEqual(bars[i].Order, "Order", foo.Order, t)
			test.CheckEqual(bars[i], "Bar", toBar(foo), t)
		}
	})
}

func FuzzGoMap(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)

		// remove any ones with duplicate order value b/c that's how we're ordering below.
		foos = DedupeByHash(foos, hashByOrder)

		bars := GoMap(foos, toBar, RandomOrderOption())

		if len(bars) != len(foos) {
			t.Fatalf("Mapping failed, data was lost. Got length: %d but expected %d", len(bars), len(foos))
		}

		// Validate data was mutated properly by provided function
		sort.Slice(foos, func(i, j int) bool {
			return foos[i].Order > foos[j].Order
		})
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		for i, foo := range foos {
			test.CheckEqual(bars[i].Bing, "Bing", toBing(foo), t)
			test.CheckEqual(bars[i].Order, "Order", foo.Order, t)
			test.CheckEqual(bars[i], "Bar", toBar(foo), t)
		}
	})
}

func FuzzGoMapWithErrs(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)

		// remove any ones with duplicate order value b/c that's how we're ordering below.
		foos = DedupeByHash(foos, hashByOrder)

		var bars []*test.Bar
		var errs []error

		if num%2 == 0 {
			bars, errs = GoMapWithErrs(foos, toBarWithErr, DiscardResultIfErrOption())
		} else {
			bars, errs = GoMapWithErrs(foos, toBarWithErr, ConcurrencyLimitOption(num/3+1), DiscardResultIfErrOption(), RandomOrderOption())
		}

		if len(bars)+len(errs) != len(foos) {
			t.Fatalf("Mapping failed, data was lost. Got total: %d but expected %d", len(bars)+len(errs), len(foos))
		}

		var expectedErrs []error
		var expectedBars []*test.Bar
		for _, foo := range foos {
			bar, err := toBarWithErr(foo)
			if err != nil {
				expectedErrs = append(expectedErrs, err)
			} else {
				expectedBars = append(expectedBars, bar)
			}
		}

		// Order slices before compare
		sort.Slice(expectedBars, func(i, j int) bool {
			return expectedBars[i].Order > expectedBars[j].Order
		})
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		sort.Slice(expectedErrs, func(i, j int) bool {
			return expectedErrs[i].Error() > expectedErrs[j].Error()
		})
		sort.Slice(errs, func(i, j int) bool {
			return errs[i].Error() > errs[j].Error()
		})

		expectedErrStrs := Map(expectedErrs, func(e error) string { return e.Error() })
		errsStrs := Map(errs, func(e error) string { return e.Error() })

		test.CheckEqual(bars, "Bars", expectedBars, t)
		test.CheckEqual(errsStrs, "Errs", expectedErrStrs, t)
	})
}

func FuzzGoMapWithErrsRateLimitTest(f *testing.F) {
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
		ratePerTime = Max(ratePerTime, -ratePerTime)
		duration := time.Duration(durationNanoseconds)
		foos := test.MakeFoos(num)

		// estimate expected execution time given rate limit
		concurrency := randomNumberBetween(1, 20) // Concurrency doesn't matter, rate is limited across goroutines

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
		_, _ = GoMapWithErrs(foos, toBarWithErr, ConcurrencyLimitOption(concurrency), RateLimitOption(ratePerTime, duration))
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
