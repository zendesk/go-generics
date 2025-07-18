//go:build test
// +build test

package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zendesk/go-generics/test"
)

func Test_RateLimiter_MultipleLimitersWithDifferentPrefixOnSameBackend(t *testing.T) {
	rateDuration := time.Millisecond * 50
	backend := NewMemoryRateLimiterBackend(1, rateDuration, 0)
	limiter1 := NewRateLimiter(FailClosed, backend, WithPrefixOption("limiter1"))
	limiter2 := NewRateLimiter(FailClosed, backend, WithPrefixOption("limiter2"))
	limiter3 := NewRateLimiter(FailClosed, backend, WithPrefixOption("limiter3"))

	ctx := context.Background()
	ok := limiter1.GetRateForClient(ctx, "client1")
	test.CheckOk(ok, "Rate expected but was not available for client1 on limiter 1", t)
	ok = limiter2.GetRateForClient(ctx, "client1")
	test.CheckOk(ok, "Rate expected but was not available for client1 on limiter 2", t)
	ok = limiter3.GetRateForClient(ctx, "client1")
	test.CheckOk(ok, "Rate expected but was not available for client1 on limiter 3", t)
}

func Test_RateLimiter_RateWaitDuration(t *testing.T) {
	rateDuration := time.Millisecond * 50
	rateWaitDuration := time.Second * 2
	backend := NewMemoryRateLimiterBackend(1, rateDuration, 0)
	limiter1 := NewRateLimiter(FailClosed, backend, WithRateWaitDuration(rateWaitDuration))

	ctx := context.Background()
	ok := limiter1.GetRateForClient(ctx, "client1")
	test.CheckOk(ok, "Rate expected but was not available for client1 on limiter 1", t)

	// No rate is available, so this should wait ~5 seconds before returning
	start := time.Now()
	ok = limiter1.WaitForRateForClient(ctx, "client1")
	t.Logf("Since: %v", time.Since(start))
	test.CheckOk(time.Since(start) > rateWaitDuration, "Expected to wait at least rateWaitDuration but did not.", t)
}

func Test_RateLimiter_AdjustThroughput(t *testing.T) {
	rateDuration := time.Millisecond * 50
	backend := NewMemoryRateLimiterBackend(1, rateDuration, 0)
	limiter1 := NewRateLimiter(FailClosed, backend)

	ctx := context.Background()
	ok := limiter1.GetRateForClient(ctx, "client1")
	test.CheckOk(ok, "Rate expected but was not available for client1", t)

	ok = limiter1.GetRateForClient(ctx, "client2")
	test.CheckOk(ok, "Rate expected but was not available for client2", t)

	ok = limiter1.GetRateForClient(ctx, "client1")
	test.CheckNotOk(ok, "Rate not expected but as available for client1", t)

	// Wait for new rate to be available
	time.Sleep(rateDuration)

	// Change rate to 10/s
	limiter1.SetThroughput(10, rateDuration, 0)

	for i := 0; i < 10; i++ {
		ok = limiter1.GetRateForClient(ctx, "client1")
		test.CheckOk(ok, fmt.Sprintf("Rate expected but was not available for client1: %d", i), t)
	}

	// Client should be throttled with no rate left
	ok = limiter1.GetRateForClient(ctx, "client1")
	test.CheckNotOk(ok, "Rate not expected but as available for client1", t)
}

func Test_RateLimiter_AdjustThroughput_WithThroughputProvider(t *testing.T) {
	var throughput = 1
	rateDuration := time.Millisecond * 50
	provider := func() (rate int, overTime time.Duration, burstCapacity int) {
		return throughput, rateDuration, 0
	}

	backend := NewMemoryRateLimiterBackend(1, rateDuration, 0)
	limiter1 := NewRateLimiter(FailClosed, backend, WithThroughputProvider(provider, time.Millisecond))

	ctx := context.Background()
	ok := limiter1.GetRateForClient(ctx, "client1")
	test.CheckOk(ok, "Rate expected but was not available for client1", t)

	ok = limiter1.GetRateForClient(ctx, "client2")
	test.CheckOk(ok, "Rate expected but was not available for client2", t)

	ok = limiter1.GetRateForClient(ctx, "client1")
	test.CheckNotOk(ok, "Rate not expected but as available for client1", t)

	// Wait for new rate to be available
	time.Sleep(rateDuration)

	// Change rate to 10/s
	throughput = 10
	time.Sleep(time.Millisecond * 2)

	for i := 0; i < 10; i++ {
		ok = limiter1.GetRateForClient(ctx, "client1")
		test.CheckOk(ok, fmt.Sprintf("Rate expected but was not available for client1: %d", i), t)
	}

	// Client should be throttled with no rate left
	ok = limiter1.GetRateForClient(ctx, "client1")
	test.CheckNotOk(ok, "Rate not expected but as available for client1", t)
}
