//go:build test
// +build test

package functions

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/zendesk/go-generics/test"
)

func FuzzGoMapToMany(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)
		barsPerFoo := randomNumberBetween(1, maxMutationExpansion)

		//remove any ones with duplicate order value b/c that's how we're ordering below.
		foos = DedupeByHash(foos, hashByOrder)

		//execute
		var bars []*test.Bar
		if num%2 == 0 {
			bars = GoMapToMany(foos, toManyBars(barsPerFoo))
		} else {
			bars = GoMapToMany(foos, toManyBars(barsPerFoo), ConcurrencyLimitOption(num/3+1))
		}

		// Validate data was mutated properly by provided function
		var expectedBars []*test.Bar
		for _, foo := range foos {
			expectedBars = append(expectedBars, toManyBars(barsPerFoo)(foo)...)
		}

		sort.Slice(expectedBars, func(i, j int) bool {
			return expectedBars[i].Order > expectedBars[j].Order
		})

		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		test.CheckEqual(len(bars), "Bar Length is not equal expected bar length", len(expectedBars), t)

		test.CheckEqual(bars, "Bars", expectedBars, t)
	})
}

func FuzzGoMapToManyWithErrs(f *testing.F) {
	for i := 0; i < seedIterations; i++ {
		f.Add(randomNumber(maxSliceSizeLength))
	}

	f.Fuzz(func(t *testing.T, num int) {
		foos := test.MakeFoos(num)
		barsPerFoo := randomNumberBetween(1, maxMutationExpansion)

		// remove any ones with duplicate order value b/c that's how we're ordering below.
		foos = DedupeByHash(foos, hashByOrder)

		// execute
		var bars []*test.Bar
		var errs []error

		if num%2 == 0 {
			bars, errs = GoMapToManyWithErrs(foos, toManyBarsWithErr(barsPerFoo), DiscardResultIfErrOption())
		} else {
			bars, errs = GoMapToManyWithErrs(foos, toManyBarsWithErr(barsPerFoo), ConcurrencyLimitOption(num/3+1), DiscardResultIfErrOption())
		}

		// Validate data was mutated properly by provided function
		var expectedBars []*test.Bar
		var expectedErrs []error
		for _, foo := range foos {
			newBars, newErr := toManyBarsWithErr(barsPerFoo)(foo)
			if newErr != nil {
				expectedErrs = append(expectedErrs, newErr)
			} else {
				expectedBars = append(expectedBars, newBars...)
			}
		}

		sort.Slice(expectedBars, func(i, j int) bool {
			return expectedBars[i].Order > expectedBars[j].Order
		})

		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Order > bars[j].Order
		})

		sort.Slice(errs, func(i, j int) bool {
			return errs[i].Error() > errs[j].Error()
		})

		sort.Slice(expectedErrs, func(i, j int) bool {
			return expectedErrs[i].Error() > expectedErrs[j].Error()
		})

		test.CheckEqual(len(bars), "Bar Length is not equal expected bar length", len(expectedBars), t)
		test.CheckEqual(bars, "Bars", expectedBars, t)

		expectedStrs := Map(expectedErrs, func(e error) string { return e.Error() })
		errStrs := Map(errs, func(e error) string { return e.Error() })
		test.CheckEqual(expectedStrs, "Errs", errStrs, t)
	})
}

func FuzzGoMapToManyWithErrsRateLimitTest(f *testing.F) {
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
	}

	f.Fuzz(func(t *testing.T, num int, ratePerTime int, durationNanoseconds int64) {
		ratePerTime = Max(ratePerTime)
		duration := time.Duration(durationNanoseconds)
		foos := test.MakeFoos(num)
		barsPerFoo := randomNumberBetween(1, maxMutationExpansion)

		// estimate expected execution time given rate limit
		concurrency := randomNumberBetween(1, 20) // Concurrency doesn't matter, rate is limited across goroutines
		t.Logf("Len: %d, rate: %d, duration: %d, - converted len %f, rate %f duration %f", len(foos), ratePerTime, duration.Milliseconds(), float64(len(foos)), float64(ratePerTime), float64(duration.Milliseconds()))

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
		_, _ = GoMapToManyWithErrs(foos, toManyBarsWithErr(barsPerFoo), ConcurrencyLimitOption(concurrency), RateLimitOption(ratePerTime, duration))
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

func Test_GoMapToMany_DiscardResultsIfErr(t *testing.T) {
	errOnEvens := func(i int) ([]int, error) {
		if i%2 == 0 {
			return []int{i, i}, fmt.Errorf("error: %d.", i)
		} else {
			return []int{i, i}, nil
		}
	}

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	expected := []int{1, 1, 3, 3, 5, 5, 7, 7, 9, 9}
	results, foundErrs := GoMapToManyWithErrs(items, errOnEvens, DiscardResultIfErrOption(), RandomOrderOption())
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
	t.Logf("results: %+v", results)

	test.CheckOk(expectedErrsFound, "Expected errors do not exist", t)
	test.CheckOk(!missingErrorsAreMissing, "Errors have been found that shouldn't be here!", t)
}
