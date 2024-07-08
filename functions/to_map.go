package functions

import (
	"context"

	"github.com/zendesk/go-generics/types"
)

// ToMap converts a slice[] to a map[K]V via a provided converter function.
func ToMap[T any, K comparable, V any](from []T, converter func(T) (K, V)) map[K]V {
	var results = make(map[K]V)

	for _, t := range from {
		k, v := converter(t)
		results[k] = v
	}

	return results
}

// GoToMap converts a slice[] to a map[K]V via a provided converter function.
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
//   - RandomOrderOption
func GoToMap[T any, K comparable, V any](items []T, f func(T) (K, V), opts ...Option) map[K]V {
	results, _ := GoToMapWithErrs(items, func(t T) (K, V, error) {
		k, v := f(t)
		return k, v, nil
	}, opts...)
	return results
}

// GoToMapWithErrs - ToMap converts a slice[] to a map[K]V via a provided converter function.
// This function is particularly useful if your converter function has a form of IO Wait (i.e, queries a DB, makes a HTTP call, or reads from disk)
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
//   - RandomOrderOption
//   - RetryOption
func GoToMapWithErrs[T any, K comparable, V any](items []T, f func(T) (K, V, error), opts ...Option) (results map[K]V, errs []error) {
	cfg := setOpts(opts)
	results = make(map[K]V, len(items))

	if cfg.RandomOrder {
		items = Shuffle(items)
	}

	// If our limit > # items, set limit to the # items, or our max concurrency. This prevents over-spawning of goroutines
	if len(items) < cfg.ConcurrencyLimit {
		cfg.ConcurrencyLimit = Max(len(items), 1)
	}

	// Do not allow user to provision excessive concurrency
	cfg.ConcurrencyLimit = Min(cfg.ConcurrencyLimit, AbsoluteMaxConcurrency)

	jobChan := make(chan T, len(items))
	resultChan := make(chan types.TupleWithErr[K, V], len(items))
	defer close(resultChan)

	// Run Consumer jobs
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		go func() {
			for item := range jobChan {
				cfg.limiter.WaitForRate(context.Background())
				k, v, err := RunToMapWithRetries(f, item, cfg.RetryCount, cfg.RetryBackoffInterval)
				resultChan <- types.TupleWithErr[K, V]{A: k, B: v, E: err}
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
		result := <-resultChan
		if result.HasError() {
			errs = append(errs, result.E)
			if cfg.DiscardResultsIfErr {
				continue
			}
		}

		results[result.A] = result.B
	}

	return
}
