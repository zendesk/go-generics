package functions_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/zendesk/go-generics/functions"
	"github.com/zendesk/go-generics/internal/test"
	"github.com/zendesk/lockbox-shared-lib/lockbox/generics"
	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

func TestGoEachMapWithErrs(t *testing.T) {
	fooMap := test.MakeFooMaps(5000)

	errs := functions.GoEachMapWithErrs(fooMap, func(k int, v *test.Foo) error {
		if v.Order%2 == 0 {
			return fmt.Errorf("ERROR EVEN NUMBER")
		}
		return nil
	})

	numErrs := 0
	for _, v := range fooMap {
		if v.Order%2 == 0 {
			numErrs++
		}
	}

	test.CheckEqual(numErrs, "Num Errors", len(errs), t)
}

func FuzzGoEachMapWithErrsRateLimitTest(f *testing.F) {
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
		foos := test.MakeFooMaps(num)

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

		errorOnEvens := func(i int, f *test.Foo) error {
			if f.Order%2 == 0 {
				return fmt.Errorf("ERROR on even!")
			}

			return nil
		}

		// execute
		start := time.Now().UnixMilli()
		_ = generics.GoEachMapWithErrs(foos, errorOnEvens, generics.ConcurrencyLimitOption(concurrency), generics.RateLimitOption(ratePerTime, duration))
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
