package test

import (
	"hash/fnv"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomNumber(max int) int {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	return rand.Intn(max)
}

func GenerateRandomLetterString(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// RandomDurationBetween returns a random duration between min/max duration.
// If min > max, then max is returned.
func RandomDurationBetween(min time.Duration, max time.Duration) time.Duration {
	if min > max {
		return min
	}

	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	num := rand.Int63n(max.Nanoseconds() - min.Nanoseconds())

	return time.Duration(num + min.Nanoseconds())
}

// RandomNumberBetween returns a random number between min and max. If Min > Max, then max is returned.
func RandomNumberBetween(min int, max int) int {
	if min >= max {
		return max
	}
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	num := rand.Intn(max - min)
	return num + min
}

// Hash64 quickly converts a string to an int64
func Hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
