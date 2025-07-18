package concurrency

import (
	"math/rand"
	"time"
)

// randomDurationBetween returns a random duration between min and max durations.
// If min > max, then min is returned.
func randomDurationBetween(min, max time.Duration) time.Duration {
	if min > max {
		return min
	}

	// Calculate the random duration between min and max
	// no need to seed the random number generator in Go 1.20 and later, as it is already seeded by the runtime.
	delta := max.Nanoseconds() - min.Nanoseconds()
	randomNanos := rand.Int63n(delta + 1) // +1 to include max value

	return min + time.Duration(randomNanos)
}
