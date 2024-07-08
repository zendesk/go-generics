package functions

import (
	"context"

	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

// Each runs fn() over every provided item.
func Each[T interface{}](items []T, fn func(T)) {
	for _, i := range items {
		fn(i)
	}
}

// EachMergeErrs runs the provided fn() over every item  and aggregates all errors into a single error.
// If no error is found, nil is returned.
func EachMergeErrs[T interface{}](items []T, fn func(T) error) error {
	var errs []error
	for _, i := range items {
		err := fn(i)
		errs = append(errs, err)
	}

	return utils.MergeErrors(errs...)
}

// GoEach runs a function across each item of a slice concurrently.
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - RandomOrderOption
func GoEach[T interface{}](items []T, fn func(T), opts ...Option) {
	_ = GoEachWithErrs(items, func(t T) error {
		fn(t)
		return nil
	}, opts...)
	return
}

// GoEachWithErrs runs a function across each item of a slice concurrently. Errors are aggregated and returned.
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - RandomOrderOption
//   - RetryOption
func GoEachWithErrs[T any](items []T, fn func(T) error, opts ...Option) (errs []error) {
	cfg := setOpts(opts)

	if cfg.RandomOrder {
		items = Shuffle(items)
	}

	// If our limit > # items, set limit to the # items. This prevents over-spawning of goroutines
	if len(items) < cfg.ConcurrencyLimit {
		cfg.ConcurrencyLimit = Max(len(items), 1)
	}

	// Do not allow user to provision excessive concurrency
	cfg.ConcurrencyLimit = Min(cfg.ConcurrencyLimit, AbsoluteMaxConcurrency)

	jobChan := make(chan T, len(items))
	resultChan := make(chan error, len(items))
	defer close(resultChan)

	// Run Consumer jobs
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		go func() {
			for item := range jobChan {
				cfg.limiter.WaitForRate(context.Background())
				err := RunWithRetries(fn, item, cfg.RetryCount, cfg.RetryBackoffInterval)
				resultChan <- err
			}
		}()
	}

	// Add items to channel
	for _, i := range items {
		jobChan <- i
	}
	close(jobChan)

	// Process results
	for i := 0; i < len(items); i++ {
		err := <-resultChan
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
