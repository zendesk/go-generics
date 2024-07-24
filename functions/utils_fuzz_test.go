package functions

import (
	"math/rand"
	"testing"
	"time"
)

const (
	fuzzSeedIterations = 5000
)

func Fuzz_randomNumberBetweenTest(f *testing.F) {
	maxInt := 2147483647
	for i := 0; i < fuzzSeedIterations; i++ {
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Nanosecond)
		low := rand.Intn(maxInt - 1)
		high := rand.Intn(maxInt - low)
		f.Add(low, low+high)
	}

	f.Fuzz(func(t *testing.T, minNumber, maxNumber int) {
		num := randomNumberBetween(minNumber, maxNumber)
		if num < minNumber || num > maxNumber {
			t.Fatalf("Number returned: %d is not valid! It must be between %d and %d", num, minNumber, maxNumber)
		}
	})
}

func Fuzz_randomDurationBetweenTest(f *testing.F) {
	var maxInt64 int64 = 9223372036854775807
	for i := 0; i < fuzzSeedIterations; i++ {
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Nanosecond)
		low := rand.Int63n(maxInt64)
		high := rand.Int63n(maxInt64 - low)
		f.Add(low, low+high)
	}

	f.Fuzz(func(t *testing.T, minNumber, maxNumber int64) {
		min := time.Duration(minNumber)
		max := time.Duration(maxNumber)
		num := randomDurationBetween(min, max)
		if num.Nanoseconds() < min.Nanoseconds() || num.Nanoseconds() > max.Nanoseconds() {
			t.Fatalf("Duration returned: %d is not valid! It must be between %d and %d", num, minNumber, maxNumber)
		}
	})
}
