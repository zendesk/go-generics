package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/zendesk/lockbox-shared-lib/lockbox/utils"
	logger "github.com/zendesk/zendesk_logging_go"
)

type GateType bool

const (
	// RateWaitDuration is the default amount of time before the rate limiter will check the backend for more available rate.
	RateWaitDuration = time.Millisecond * 25
	// FailOpen will make errors from Redis result in automatic rate being allowed
	FailOpen GateType = true
	// FailClosed will make errors from Redis result in automatic rate being denied.
	FailClosed GateType = false
)

type RateLimitConfig struct {
	limiterPrefix    string
	rateWaitDuration time.Duration
}

func NewRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		rateWaitDuration: RateWaitDuration,
	}
}

type rateLimitOption func(cfg RateLimitConfig) RateLimitConfig

// WithPrefixOption - if provided, all keys used in the backend will be prefixed with the provided string `${prefix}-${clientID}`
// This is useful if you're using multiple rate limiters with the same backend and the same client set but wish to rate-limit them
// independently.
var WithPrefixOption = func(prefix string) rateLimitOption {
	return func(cfg RateLimitConfig) RateLimitConfig {
		cfg.limiterPrefix = prefix
		return cfg
	}
}

// WithRateWaitDuration - if provided, the rate limiter will wait for the provided duration before checking the backend for more available rate.
// Default is 25ms - you may want to increase this if your backend is a remote service to prevent high traffic throughput checking for available rate.
var WithRateWaitDuration = func(duration time.Duration) rateLimitOption {
	return func(cfg RateLimitConfig) RateLimitConfig {
		cfg.rateWaitDuration = duration
		return cfg
	}
}

type RateLimitBackend interface {
	GetRate(ctx context.Context, clientID string) (bool, error)
	SetThroughput(rate int, overTime time.Duration, burstCapacity int)
	GetThroughput() (int, time.Duration, int)
}

type RateLimiter struct {
	backend          RateLimitBackend
	gateType         GateType
	cfg              RateLimitConfig
	rateWaitDuration time.Duration
	defaultClient    string
}

// NewRateLimiter creates a new rate limiter with a redis backend
func NewRateLimiter(gateType GateType, backend RateLimitBackend, opts ...rateLimitOption) *RateLimiter {
	cfg := NewRateLimitConfig()
	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return &RateLimiter{
		backend:       backend,
		gateType:      gateType,
		cfg:           cfg,
		defaultClient: utils.GenerateRandomLetterString(10),
	}
}

func (rl *RateLimiter) getRate(ctx context.Context, clientID string) bool {
	hasRate, err := rl.getRateFromBackend(ctx, clientID)

	if err != nil {
		if rl.gateType == FailOpen {
			logger.Errorf(ctx, "Got error attempting to receive rate. Allowing rate b/c gate is FailOpen. Err: %+v", err)
			return true
		} else {
			logger.Errorf(ctx, "Got error attempting to receive rate. Denying rate b/c gate is FailClosed. Err: %+v", err)
			return false
		}
	}

	return hasRate
}

func (rl *RateLimiter) getRateWithError(ctx context.Context, clientID string) (bool, error) {
	return rl.getRateFromBackend(ctx, clientID)
}

// WaitForRate blocks until rate is available. When rate is available, true is returned.
func (rl *RateLimiter) WaitForRate(ctx context.Context) bool {
	return rl.WaitForRateForClient(ctx, rl.defaultClient)
}

// WaitForRateWithTimeout returns True if rate is available and false if timeout has expired with no rate available
func (rl *RateLimiter) WaitForRateWithTimeout(ctx context.Context, timeout time.Duration) bool {
	return rl.WaitForRateWithTimeoutForClient(ctx, rl.defaultClient, timeout)
}

// WaitForRateWithError blocks until rate is available or an error occurs. When rate is available, true is returned. If an error occurs,
// false is returned.
func (rl *RateLimiter) WaitForRateWithError(ctx context.Context) (bool, error) {
	return rl.WaitForRateWithErrorForClient(ctx, rl.defaultClient)
}

// WaitForRateWithErrorAndTimeout returns True if rate is available and false if timeout has expired with no rate available
func (rl *RateLimiter) WaitForRateWithErrorAndTimeout(ctx context.Context, timeout time.Duration) (bool, error) {
	return rl.WaitForRateWithErrorAndTimeoutForClient(ctx, rl.defaultClient, timeout)
}

// GetRate Returns T/F if there is rate available for execution, decrements if return is true. If an error occurs, the
// rate will be made available, or rejected, depending on GateType configuration.
func (rl *RateLimiter) GetRate(ctx context.Context) bool {
	return rl.getRate(ctx, rl.defaultClient)
}

// GetRateWithError Returns T/F if there is rate available for execution, or error if rate cannot be determined
func (rl *RateLimiter) GetRateWithError(ctx context.Context) (bool, error) {
	return rl.getRateFromBackend(ctx, rl.defaultClient)
}

// GetRateWithErrorForClient Returns T/F if there is rate available for execution, or error if rate cannot be determined
func (rl *RateLimiter) GetRateWithErrorForClient(ctx context.Context, clientID string) (bool, error) {
	return rl.getRateFromBackend(ctx, clientID)
}

// GetRateForClient Returns T/F if there is rate available for execution for the specific clientID, decrements if return is true
func (rl *RateLimiter) GetRateForClient(ctx context.Context, clientID string) bool {
	return rl.getRate(ctx, clientID)
}

// WaitForRateForClient blocks until rate is available for the specified clientID. When rate is available, true is returned.
func (rl *RateLimiter) WaitForRateForClient(ctx context.Context, clientID string) bool {
	for {
		if rl.getRate(ctx, clientID) {
			return true
		} else {
			time.Sleep(rl.cfg.rateWaitDuration)
		}
	}
}

func (rl *RateLimiter) WaitForRateWithErrorForClient(ctx context.Context, clientID string) (bool, error) {
	for {
		hasRate, err := rl.getRateWithError(ctx, clientID)
		if err != nil {
			return hasRate, err
		}

		if hasRate {
			return true, nil
		} else {
			time.Sleep(rl.cfg.rateWaitDuration)
		}
	}
}

// WaitForRateWithTimeoutForClient returns True if rate is available for the specified clientID and false if timeout has expired with no rate available
func (rl *RateLimiter) WaitForRateWithTimeoutForClient(ctx context.Context, clientID string, timeout time.Duration) bool {
	start := time.Now()
	for {
		if rl.getRate(ctx, clientID) {
			return true
		} else {
			if time.Since(start) > timeout {
				return false
			}
			time.Sleep(rl.cfg.rateWaitDuration)
		}
	}
}

// WaitForRateWithErrorAndTimeoutForClient returns True if rate is available and false if timeout has expired with no rate available, or an error if an error is encountered while waiting
func (rl *RateLimiter) WaitForRateWithErrorAndTimeoutForClient(ctx context.Context, clientID string, timeout time.Duration) (bool, error) {
	start := time.Now()
	for {
		hasRate, err := rl.getRateWithError(ctx, clientID)
		if err != nil {
			return hasRate, err
		}

		if hasRate {
			return true, nil
		} else {
			if time.Since(start) > timeout {
				return false, nil
			}
			time.Sleep(rl.cfg.rateWaitDuration)
		}
	}
}

func (rl *RateLimiter) SetThroughput(rate int, overTime time.Duration, burstCapacity int) {
	rl.backend.SetThroughput(rate, overTime, burstCapacity)
}

func (rl *RateLimiter) GetThroughput() (rate int, overTime time.Duration, burstCapacity int) {
	return rl.backend.GetThroughput()
}

func (rl *RateLimiter) getRateFromBackend(ctx context.Context, clientID string) (bool, error) {
	clientID = rl.initClientId(clientID)
	return rl.backend.GetRate(ctx, clientID)
}

// Returns the client id to be used for rate limiting with any per-rate-limiter prefix attached
func (rl *RateLimiter) initClientId(clientID string) string {
	if rl.cfg.limiterPrefix == "" {
		return clientID
	}

	return fmt.Sprintf("%s-%s", rl.cfg.limiterPrefix, clientID)
}
