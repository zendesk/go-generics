package functions

import (
	"hash/fnv"
	"math/rand"
	"reflect"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func typesMatch(a interface{}, b interface{}) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}

func randomNumber(max int) int {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	return rand.Intn(max)
}

func generateRandomLetterString(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// randomDurationBetween returns a random duration between min/max duration.
// If min > max, then max is returned.
func randomDurationBetween(min time.Duration, max time.Duration) time.Duration {
	if min > max {
		return min
	}

	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	num := rand.Int63n(max.Nanoseconds() - min.Nanoseconds())

	return time.Duration(num + min.Nanoseconds())
}

// randomNumberBetween returns a random number between min and max. If Min > Max, then max is returned.
func randomNumberBetween(min int, max int) int {
	if min >= max {
		return max
	}
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Nanosecond * 1)
	num := rand.Intn(max - min)
	return num + min
}

// hash64 quickly converts a string to an int64
func hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// firstNotEmpty returns the first non-empty string of the provided set. Strings are compared in order provided.
// Returns first non-empty item and T/F of whether or not an item was found.
func firstNotEmpty(items ...string) (string, bool) {
	for _, i := range items {
		if i != "" {
			return i, true
		}
	}
	return "", false
}
