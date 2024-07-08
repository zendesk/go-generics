package functions

import (
	"context"

	"github.com/zendesk/go-generics/types"
)

// MapToSlice converts a map[K]V to a slice of type Z via provided converter function.
func MapToSlice[K comparable, V any, Z any](from map[K]V, converter func(k K, v V) Z) []Z {
	var results []Z

	for key, value := range from {
		result := converter(key, value)
		results = append(results, result)
	}
	return results
}

// GoMapToSlice - Fast Map of a Map[T]Y to []Z -- runs in a goroutine with safe aggregation
// This function is particularly useful if your converter function has a form of IO Wait (i.e, queries a DB, makes a HTTP call, or reads from disk)
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
func GoMapToSlice[K comparable, V any, Z any](items map[K]V, converter func(K, V) Z, opts ...Option) []Z {
	zs, _ := GoMapToSliceWithErrs(items, func(k K, v V) (Z, error) {
		return converter(k, v), nil
	}, opts...)

	return zs
}

// GoMapToSliceWithErrs - Fast Map of a Map[K]V to []Z -- runs in a goroutine with safe aggregation of results and errors
// This function is particularly useful if your converter function has a form of IO Wait (i.e, queries a DB, makes a HTTP call, or reads from disk)
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
//   - RetryOption
func GoMapToSliceWithErrs[K comparable, V any, Z any](items map[K]V, converter func(K, V) (Z, error), opts ...Option) (results []Z, errs []error) {
	cfg := setOpts(opts)

	// If our limit > # items, set limit to the # items. This prevents over-spawning of goroutines
	if len(items) < cfg.ConcurrencyLimit {
		cfg.ConcurrencyLimit = Max(len(items), 1)
	}

	// Do not allow user to provision excessive concurrency
	cfg.ConcurrencyLimit = Min(cfg.ConcurrencyLimit, AbsoluteMaxConcurrency)

	// Channel type is a map that contains a batch of values to map
	jobChan := make(chan types.Tuple[K, V], len(items))
	resultChan := make(chan types.Tuple[Z, error], len(items))
	defer close(resultChan)

	// Run Consumer jobs
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		go func() {
			for item := range jobChan {
				cfg.limiter.WaitForRate(context.Background())
				converted, err := RunMapToSliceWithRetries(converter, item.A, item.B, cfg.RetryCount, cfg.RetryBackoffInterval)
				resultChan <- types.Tuple[Z, error]{A: converted, B: err}
			}
		}()
	}

	for k, v := range items {
		jobChan <- types.Tuple[K, V]{A: k, B: v}
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
