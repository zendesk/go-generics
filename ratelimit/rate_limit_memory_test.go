package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/zendesk/go-generics/test"
)

func Test_RateLimiter_InMemoryBackend(t *testing.T) {
	rateDuration := time.Millisecond * 50
	backend := NewMemoryRateLimiterBackend(1, rateDuration, 0)
	limiter := NewRateLimiter(FailClosed, backend)
	ctx := context.Background()

	// Test GetRate
	result := limiter.GetRate(ctx)
	test.CheckOk(result, "Expected to get rate (1)", t)
	result = limiter.GetRate(ctx)
	test.CheckNotOk(result, "Didn't expect to get rate (1)", t)
	time.Sleep(rateDuration)
	result = limiter.GetRate(ctx)
	test.CheckOk(result, "Expected to get rate (2)", t)
	result = limiter.GetRate(ctx)
	test.CheckNotOk(result, "Didn't expect to get rate (2)", t)

	// Reset
	time.Sleep(time.Millisecond * 455)

	// Test GetRateWithError
	result, err := limiter.GetRateWithError(ctx)
	test.CheckOk(result, "Expected to get rate (3)", t)
	test.CheckErr(err, "Unexpected error on rate receive", t)
	result, err = limiter.GetRateWithError(ctx)
	test.CheckErr(err, "Unexpected error on rate receive", t)
	test.CheckNotOk(result, "Didn't expect to get rate (3)", t)
	time.Sleep(rateDuration)
	result, err = limiter.GetRateWithError(ctx)
	test.CheckErr(err, "Unexpected error on rate receive", t)
	test.CheckOk(result, "Expected to get rate (4)", t)
	result, err = limiter.GetRateWithError(ctx)
	test.CheckErr(err, "Unexpected error on rate receive", t)
	test.CheckNotOk(result, "Didn't expect to get rate (4)", t)

	start1 := time.Now()
	// Test WaitForRate
	gotRate := limiter.WaitForRate(ctx)
	start2 := time.Now()
	end1 := time.Now()
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	diff := end1.Sub(start1)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration duration, but it wasn't.", t)

	// Test WaitForRate 2
	gotRate = limiter.WaitForRate(ctx)
	start3 := time.Now()
	end2 := time.Now()
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	diff = end2.Sub(start2)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration duration, but it wasn't. (2)", t)

	// Test WaitForRateWithTimeout Success
	gotRate = limiter.WaitForRateWithTimeout(ctx, time.Second)
	end3 := time.Now()
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	diff = end3.Sub(start3)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration duration, but it wasn't. (2)", t)

	// Test WaitForRateWithTimeout -- Failure -- got timeout
	gotRate = limiter.WaitForRateWithTimeout(ctx, time.Millisecond*10)
	test.CheckNotOk(gotRate, "Did not expect rate but received it", t)

	// Reset
	time.Sleep(time.Millisecond * 500)
	limiter.GetRate(ctx)

	start := time.Now()
	gotRate, err = limiter.WaitForRateWithErrorAndTimeout(ctx, time.Second)
	test.CheckErr(err, "Unexpected err (1)", t)
	end := time.Now()
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	diff = end.Sub(start)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration duration, but it wasn't. (3)", t)

	// Test WaitForRateWithErrorAndTimeoutForClient -- Failure -- got timeout
	gotRate, err = limiter.WaitForRateWithErrorAndTimeout(ctx, time.Millisecond*10)
	test.CheckErr(err, "Unexpected err (2)", t)
	test.CheckNotOk(gotRate, "Did not expect rate but received it", t)
}

func Test_RateLimiter_InMemoryBackend_MultipleClients(t *testing.T) {
	clientID1 := "client1"
	clientID2 := "client2"
	rateDuration := time.Millisecond * 50
	buffer := time.Millisecond * 4 // gives async process a little buffer to add rate back to the buckte
	backend := NewMemoryRateLimiterBackend(1, rateDuration, 0)
	limiter := NewRateLimiter(FailClosed, backend)
	ctx := context.Background()
	time.Sleep(time.Second) // wait for init

	// Test GetRate
	result := limiter.GetRateForClient(ctx, clientID1)
	test.CheckOk(result, "Expected to get rate (1)", t)
	result = limiter.GetRateForClient(ctx, clientID1)
	test.CheckNotOk(result, "Didn't expect to get rate (1)", t)
	time.Sleep(rateDuration + buffer)
	result = limiter.GetRateForClient(ctx, clientID1)
	test.CheckOk(result, "Expected to get rate (2)", t)
	result = limiter.GetRateForClient(ctx, clientID1)
	test.CheckNotOk(result, "Didn't expect to get rate (2)", t)

	start1 := time.Now()
	// Test WaitForRate
	gotRate := limiter.WaitForRateForClient(ctx, clientID1)
	start2 := time.Now() // start timer for next check
	end1 := time.Now()
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	diff := end1.Sub(start1)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration duration, but it wasn't.", t)

	// Test WaitForRate 2
	gotRate = limiter.WaitForRateForClient(ctx, clientID1)
	start3 := time.Now() // start timer for next check
	end2 := time.Now()
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	diff = end2.Sub(start2)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration duration, but it wasn't. (2)", t)

	// Test WaitForRateWithTimeout Success
	gotRate = limiter.WaitForRateWithTimeoutForClient(ctx, clientID1, time.Second)
	start4 := time.Now()
	end3 := time.Now()
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	diff = end3.Sub(start3)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration duration, but it wasn't. (3)", t)

	// Test WaitForRateWithTimeout -- Failure -- got timeout
	gotRate = limiter.WaitForRateWithTimeoutForClient(ctx, clientID1, time.Millisecond*10)
	test.CheckNotOk(gotRate, "Did not expect rate but received it", t)

	var err error

	gotRate, err = limiter.WaitForRateWithErrorAndTimeoutForClient(ctx, clientID1, time.Second)
	end4 := time.Now()
	diff = end4.Sub(start4)
	test.CheckErr(err, "Unexpected err (1)", t)
	test.CheckOk(gotRate, "Expected rate but didn't get rate", t)
	test.CheckOk(diff >= rateDuration, "Expected diff to be at least rateDuration (1) duration, but it wasn't. (4)", t)

	// Test WaitForRateWithErrorAndTimeoutForClient -- Failure -- got timeout
	gotRate, err = limiter.WaitForRateWithErrorAndTimeoutForClient(ctx, clientID1, time.Millisecond*10)
	test.CheckErr(err, "Unexpected err (2)", t)
	test.CheckNotOk(gotRate, "Did not expect rate but received it", t)

	// reset
	time.Sleep(rateDuration + buffer)

	// Interleave, verify client 2 doesn't affect client 1
	gotRate = limiter.GetRateForClient(ctx, clientID2)
	test.CheckOk(gotRate, "expected to get rate for client 2", t)
	gotRate = limiter.GetRateForClient(ctx, clientID2)
	test.CheckNotOk(gotRate, "didn't expect to get rate for client 2", t)
	gotRate = limiter.GetRateForClient(ctx, clientID1)
	test.CheckOk(gotRate, "expected to get rate for client 1", t)
	gotRate = limiter.GetRateForClient(ctx, clientID1)
	test.CheckNotOk(gotRate, "didn't expect to get rate for client 1", t)

	// reset
	time.Sleep(rateDuration + buffer)
	// Interleave, verify client 2 doesn't affect client 1
	gotRate, err = limiter.GetRateWithErrorForClient(ctx, clientID2)
	test.CheckErr(err, "Unexpected error on rate get", t)
	test.CheckOk(gotRate, "expected to get rate for client 2", t)
	gotRate, err = limiter.GetRateWithErrorForClient(ctx, clientID2)
	test.CheckErr(err, "Unexpected error on rate get", t)
	test.CheckNotOk(gotRate, "didn't expect to get rate for client 2", t)
	gotRate, err = limiter.GetRateWithErrorForClient(ctx, clientID1)
	test.CheckErr(err, "Unexpected error on rate get", t)
	test.CheckOk(gotRate, "expected to get rate for client 1a", t)
	gotRate, err = limiter.GetRateWithErrorForClient(ctx, clientID1)
	test.CheckErr(err, "Unexpected error on rate get", t)
	test.CheckNotOk(gotRate, "didn't expect to get rate for client 1", t)
}

func Test_RateLimiter_InMemoryBackend_BurstCapacity(t *testing.T) {
	rateDuration := time.Millisecond * 10
	backend := NewMemoryRateLimiterBackend(1, rateDuration, 3)
	limiter := NewRateLimiter(FailClosed, backend)
	ctx := context.Background()

	// Test GetRate
	result := limiter.GetRate(ctx)
	test.CheckOk(result, "Expected to get rate (1)", t)
	result = limiter.GetRate(ctx)
	test.CheckNotOk(result, "Didn't expect to get rate (1)", t)
	time.Sleep(rateDuration)
	result = limiter.GetRate(ctx)
	test.CheckOk(result, "Expected to get rate (2)", t)
	result = limiter.GetRate(ctx)
	test.CheckNotOk(result, "Didn't expect to get rate but got rate (2)", t)

	time.Sleep(rateDuration * 3)
	result = limiter.GetRate(ctx)
	test.CheckOk(result, "Expected to get rate (3)", t)
	result = limiter.GetRate(ctx)
	test.CheckOk(result, "Expected to get rate (3a)", t)
	result = limiter.GetRate(ctx)
	test.CheckOk(result, "Expected to get rate (3b)", t)

}
