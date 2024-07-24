package functions

import (
	"context"

	"github.com/zendesk/go-generics/types"
)

// Map converts a slice of T to a slice of Y via a converter function
func Map[T any, Y any](from []T, converter func(T) Y) []Y {
	var ys []Y

	for _, t := range from {
		converted := converter(t)
		ys = append(ys, converted)
	}

	return ys
}

// MapWithErrs converts a slice of T to a slice of Y. Errors are aggregated and returned.
func MapWithErrs[T any, Y any](from []T, converter func(T) (Y, error)) ([]Y, []error) {
	var ys []Y
	var errs []error

	for _, t := range from {
		converted, err := converter(t)
		ys = append(ys, converted)
		errs = append(errs, err)
	}

	return ys, errs
}

// MapMergeErrs converts a slice of T to a slice of Y. Errors are merged, then returned.
func MapMergeErrs[T any, Y any](from []T, converter func(T) (Y, error)) ([]Y, error) {
	var ys []Y
	var errs []error

	for _, t := range from {
		converted, err := converter(t)
		ys = append(ys, converted)
		errs = append(errs, err)
	}

	return ys, MergeErrors(errs...)
}

// GoMap - Fast Map inside of a goroutine with safe aggregation
// This function is particularly useful if your converter function has a form of IO Wait (i.e, queries a DB, makes a HTTP call, or reads from disk)
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
//   - RandomOrderOption
func GoMap[T any, Y any](items []T, converter func(T) Y, opts ...Option) []Y {
	ys, _ := GoMapWithErrs(items, func(t T) (Y, error) {
		return converter(t), nil
	}, opts...)

	return ys
}

// GoMapWithErrs - Fast Map inside of a goroutine with safe aggregation of results & errors
// This function is particularly useful if your converter function has a form of IO Wait (i.e, queries a DB, makes a HTTP call, or reads from disk)
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
//   - RandomOrderOption
//   - RetryOption
func GoMapWithErrs[T any, Y any](items []T, f func(T) (Y, error), opts ...Option) (results []Y, errs []error) {
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
	resultChan := make(chan types.Tuple[Y, error], len(items))
	defer close(resultChan)

	// Run Consumer jobs
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		go func() {
			for item := range jobChan {
				cfg.limiter.WaitForRate(context.Background())
				converted, err := runMapWithRetries(f, item, cfg.RetryCount, cfg.RetryBackoffInterval)
				resultChan <- types.Tuple[Y, error]{A: converted, B: err}
			}
		}()
	}

	// Add items to channel
	for _, i := range items {
		jobChan <- i
	}
	close(jobChan)

	for i := 0; i < len(items); i++ {
		result := <-resultChan
		if result.B != nil {
			errs = append(errs, result.B)
			if cfg.DiscardResultsIfErr {
				continue
			}
		}

		results = append(results, result.A)
	}

	return
}
