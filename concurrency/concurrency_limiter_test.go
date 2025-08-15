//go:build test
// +build test

package concurrency

import (
	"testing"
	"time"

	"github.com/zendesk/go-generics/test"
)

func Test_ConcurrencyLimiter(t *testing.T) {
	testFn := func() {
		time.Sleep(time.Second * 2)
	}

	callbackChan := make(chan bool)
	callback := func() {
		callbackChan <- true
	}

	limiter := NewConcurrencyLimiter(3)

	start := time.Now()

	// 3 should start instantly
	limiter.Run(testFn)
	limiter.Run(testFn)
	limiter.Run(testFn)

	if time.Since(start) > time.Millisecond*50 {
		t.Fatal("Unexpected amount of time to start goroutines. They should start and return instantly")
	}

	// 4th should wait
	limiter.Run(testFn)

	if time.Since(start) < time.Second*2 {
		t.Fatalf("Unexpected amount of time to start goroutines. more time should have elapsed before the last goroutine started. Got: %f seconds", time.Since(start).Seconds())
	}

	// Test callback
	limiter.Run(testFn, WithOnCompleteCallback(callback))
	result := <-callbackChan
	test.CheckOk(result, "Expected callback to be called", t)
}
