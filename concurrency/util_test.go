//go:build test
// +build test

package concurrency

import (
	"testing"
	"time"
)

func TestRandomDurationBetween(t *testing.T) {
	min := 10 * time.Millisecond
	max := 100 * time.Millisecond

	duration := randomDurationBetween(min, max)
	if duration < min || duration > max {
		t.Fatalf("duration out of range: %v", duration)
	}

	// Test when min > max
	duration = randomDurationBetween(max, min)
	if duration != max {
		t.Fatalf("expected max when min > max, got: %v", duration)
	}
}
