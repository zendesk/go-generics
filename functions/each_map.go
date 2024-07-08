package functions

import (
	"context"

	"github.com/zendesk/go-generics/types"
)

// GoEachMapWithErrs runs a function across each item of a map concurrently. Errors are aggregated and returned.
// Options Support:
//   - ConcurrencyLimitOption
//   - RateLimitOption
func GoEachMapWithErrs[K comparable, V any](items map[K]V, fn func(K, V) error, opts ...Option) (errs []error) {
	cfg := setOpts(opts)

	// If our limit > # items, set limit to the # items. This prevents over-spawning of goroutines
	if len(items) < cfg.ConcurrencyLimit {
		cfg.ConcurrencyLimit = Max(len(items), 1)
	}

	// Do not allow user to provision excessive concurrency
	cfg.ConcurrencyLimit = Min(cfg.ConcurrencyLimit, AbsoluteMaxConcurrency)

	jobChan := make(chan types.Tuple[K, V], len(items))
	resultChan := make(chan error, len(items))
	defer close(resultChan)

	// Run Consumer jobs
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		go func() {
			for item := range jobChan {
				cfg.limiter.WaitForRate(context.Background())
				err := fn(item.A, item.B)
				resultChan <- err
			}
		}()
	}

	for k, v := range items {
		jobChan <- types.Tuple[K, V]{A: k, B: v}
	}
	close(jobChan)

	for i := 0; i < len(items); i++ {
		err := <-resultChan
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
