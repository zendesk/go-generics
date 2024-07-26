package functions

import (
	"context"

	"github.com/zendesk/go-generics/datastructures"
)

// GoMapToMany - Fast Map inside of a goroutine where each function run may return N elements with safe aggregation
// This function is particularly useful if your converter function has a form of IO Wait (i.e, queries a DB, makes a HTTP call, or reads from disk)
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
func GoMapToMany[T any, Y any](items []T, converter func(T) []Y, opts ...Option) (results []Y) {
	results, _ = GoMapToManyWithErrs(items, func(t T) ([]Y, error) {
		return converter(t), nil
	}, opts...)

	return results
}

// GoMapToManyWithErrs - Fast Map inside a goroutine where each function run may return N elements or an error with safe aggregation
// This function is particularly useful if your converter function has a form of IO Wait (i.e, queries a DB, makes a HTTP call, or reads from disk)
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
//   - DiscardResultIfErrOption
//   - RandomOrderOption
//   - RetryOption
func GoMapToManyWithErrs[T any, Y any](items []T, converter func(T) ([]Y, error), opts ...Option) (results []Y, errs []error) {
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
	resultChan := make(chan datastructures.Tuple[[]Y, error], len(items))

	// Run Consumer jobs
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		go func() {
			for item := range jobChan {
				cfg.limiter.WaitForRate(context.Background())
				converted, err := converter(item)
				resultChan <- datastructures.Tuple[[]Y, error]{A: converted, B: err}
			}
		}()
	}

	// Add items to channel
	for _, i := range items {
		jobChan <- i
	}
	close(jobChan)

	// Aggregate results
	for i := 0; i < len(items); i++ {
		result := <-resultChan
		if result.B != nil {
			errs = append(errs, result.B)
			if cfg.DiscardResultsIfErr {
				continue
			}
		}

		results = append(results, result.A...)
	}

	return
}
