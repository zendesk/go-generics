package functions

import (
	"time"

	"github.com/zendesk/go-generics/ratelimit"
	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
)

const (
	AbsoluteMaxConcurrency = 100
	DefaultMaxConcurrency  = 20
)

type OptionConfig struct {
	limiter              *ratelimit.RateLimiter
	ConcurrencyLimit     int
	RandomOrder          bool
	DiscardResultsIfErr  bool
	RetryCount           int
	RetryBackoffInterval time.Duration
}

type Option func(option *OptionConfig) *OptionConfig

// ConcurrencyLimitOption limits the concurrency of a concurrent mapping function to protect against open file limits, connection limits, etc.
var ConcurrencyLimitOption = func(limit int) Option {
	return func(opt *OptionConfig) *OptionConfig {
		opt.ConcurrencyLimit = Max(limit, 1) // User may not set < 0 concurrency
		return opt
	}
}

// RateLimitOption limits maximum iterations that may be executed over a specified timeframe
// For instance, if rate is 5 and time.duration is 1 millisecond. The max throughput of the
// optioned function is 5 executions per millisecond.
var RateLimitOption = func(rate int, rateTimeframe time.Duration) Option {
	return func(oc *OptionConfig) *OptionConfig {
		// if negative number is provided for rate, convert to its absolute value
		rate = Max(rate, -rate)
		memoryBackend := ratelimit.NewMemoryRateLimiterBackend(rate, rateTimeframe, rate)
		oc.limiter = ratelimit.NewRateLimiter(ratelimit.FailClosed, memoryBackend)
		return oc
	}
}

// DiscardResultIfErrOption will make mapping functions not map results where errors are detected. For instance,
// if the provided fn() (result, error) function returns an error the result will not be added to the mapped result.
var DiscardResultIfErrOption = func() Option {
	return func(oc *OptionConfig) *OptionConfig {
		oc.DiscardResultsIfErr = true
		return oc
	}
}

// RandomOrderOption will make the targeted function randomly order it's execution rather than iterating over elements
// in order.
var RandomOrderOption = func() Option {
	return func(oc *OptionConfig) *OptionConfig {
		oc.RandomOrder = true
		return oc
	}
}

// RetryOption will make the targeted function attempt to retry any executed functions numRetries times.
// For instance, if a list has 5 items, and numRetries is 2, the provided mapping function could be executed a total of 15 times,
// 3 per item, due to 2 retries. With each retry, backoffInterval grows by backoffInterval. If backoffInterval is 1s, the first
// retry will be 1s after the first attempt, and the second will wait 2s before trying again, and so on.
var RetryOption = func(numRetries int, backoffInterval time.Duration) Option {
	return func(oc *OptionConfig) *OptionConfig {
		oc.RetryCount = numRetries
		oc.RetryBackoffInterval = backoffInterval
		return oc
	}
}

// Used internally to chain options to build a complete OptionConfig
func setOpts(options []Option) *OptionConfig {
	cfg := &OptionConfig{
		ConcurrencyLimit:    DefaultMaxConcurrency,
		DiscardResultsIfErr: false,
		RetryCount:          0,
	}

	for _, opt := range options {
		cfg = opt(cfg)
	}

	if cfg.limiter == nil {
		memoryBackend := ratelimit.NewMemoryRateLimiterBackend(ratelimit.UnlimitedRate, time.Second, utils.UnlimitedRate)
		cfg.limiter = ratelimit.NewRateLimiter(ratelimit.FailClosed, memoryBackend)
	}

	return cfg
}
