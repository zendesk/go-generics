//go:build test
// +build test

package cache_test

import (
	"testing"
	"time"

	"github.com/zendesk/go-generics/cache"
	"github.com/zendesk/go-generics/cache/internal/test"
)

const (
	salt string = "salt"
)

func TestNew(t *testing.T) {
	ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)

	c, err := cache.NewEncryptedCache[string, int](ttlCache, []byte(salt), 25)
	test.CheckErr(err, "Unexpected err", t)
	test.ExpectNotNil("cache", c, t)
}

func TestSetGetDelete(t *testing.T) {
	ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)

	c, err := cache.NewEncryptedCache[string, string](ttlCache, []byte(salt), 25)
	test.CheckErr(err, "Unexpected err", t)

	// Set a value in the cache
	err = c.Set("key", "value")
	test.CheckErr(err, "Unexpected err", t)

	// Get the value from the cache
	value, _, err := c.Get("key")
	test.CheckErr(err, "Unexpected err", t)
	test.CheckEqual(value, "value", "value", t)

	err = c.Delete("key")
	test.CheckErr(err, "Unexpected err", t)
	value, _, err = c.Get("key")
	test.CheckEqual("", "value", value, t)
	test.CheckErr(err, "Unexpected err", t)
}

func TestGetError(t *testing.T) {
	ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)

	// Create a new EncryptedCache
	c, err := cache.NewEncryptedCache[string, string](ttlCache, []byte(salt), 25)
	test.CheckErr(err, "Unexpected err", t)

	// Try to get a value that hasn't been set
	v, _, err := c.Get("key")
	test.CheckErr(err, "Unexpeted err", t)
	test.CheckEqual(v, "value", "", t)
}

func TestSetGetComplexStruct(t *testing.T) {

	type ComplexStruct struct {
		Name    string
		Numbers []int
		Nested  struct {
			Field     string
			SubNested struct {
				SubField int
			}
		}
		Map map[string]struct {
			Value int
		}
	}

	// Create a new EncryptedCache
	ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)
	c, err := cache.NewEncryptedCache[string, ComplexStruct](ttlCache, []byte(salt), 25)
	test.CheckErr(err, "Unexpected err", t)

	// Create a complex struct
	value := ComplexStruct{
		Name:    "test",
		Numbers: []int{1, 2, 3},
		Nested: struct {
			Field     string
			SubNested struct {
				SubField int
			}
		}{
			Field: "nested",
			SubNested: struct {
				SubField int
			}{
				SubField: 10,
			},
		},
		Map: map[string]struct {
			Value int
		}{
			"one": {Value: 1},
			"two": {Value: 2},
		},
	}

	// Set the complex struct in the cache
	err = c.Set("key", value)
	test.CheckErr(err, "Unexpected err", t)

	// Get the complex struct from the cache
	retrievedValue, _, err := c.Get("key")
	test.CheckErr(err, "Unexpected err", t)
	test.CheckEqual(retrievedValue, "retrievedValue", value, t)
}

func TestSetGetComplexKey(t *testing.T) {

	type ComplexKey struct {
		Field1 string
		Field2 int
	}

	// Create a new EncryptedCache
	ttlCache := cache.NewInMemoryCache[ComplexKey, []byte](10 * time.Second)
	c, err := cache.NewEncryptedCache[ComplexKey, string](ttlCache, []byte(salt), 25)
	test.CheckErr(err, "Unexpected err", t)

	// Create a complex key
	key := ComplexKey{
		Field1: "field1",
		Field2: 2,
	}

	// Set a value with the complex key
	err = c.Set(key, "value")
	test.CheckErr(err, "Unexpected err", t)

	// Get the value with the complex key
	value, _, err := c.Get(key)
	test.CheckErr(err, "Unexpected err", t)
	test.CheckEqual(value, "value", "value", t)
}

func TestSetInvalidValue(t *testing.T) {

	// Create a new EncryptedCache
	ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)
	c, err := cache.NewEncryptedCache[string, chan int](ttlCache, []byte(salt), 25)
	test.CheckErr(err, "Unexpected err", t)

	// Try to set an invalid value
	err = c.Set("key", make(chan int))
	test.ExpectErr(err, "Expected err", t)
}

func BenchmarkEncrypt(b *testing.B) {

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Include key generation in benchmark test since that's the part that takes the longest
		ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)
		c, err := cache.NewEncryptedCache[string, string](ttlCache, []byte(salt), 25)
		if err != nil {
			b.Fatal(err)
		}
		if err := c.Set("key", "value"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	// Check that the benchmark took less than 75ms
	if avg := b.Elapsed() / time.Duration(b.N); avg > 75*time.Millisecond {
		b.Fatalf("encrypt took too long: %v", avg)
	}
}

func BenchmarkDecrypt(b *testing.B) {

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Include key generation in benchmark test since that's the part that takes the longest
		ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)
		c, err := cache.NewEncryptedCache[string, string](ttlCache, []byte(salt), 25)
		if err != nil {
			b.Fatal(err)
		}

		// Encrypt the value
		if err := c.Set("key", "value"); err != nil {
			b.Fatal(err)
		}
		if _, _, err := c.Get("key"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()

	// Check that the benchmark took less than 75ms
	if avg := b.Elapsed() / time.Duration(b.N); avg > 75*time.Millisecond {
		b.Fatalf("decrypt took too long: %v", avg)
	}
}

func TestPurge(t *testing.T) {
	ttlCache := cache.NewInMemoryCache[string, []byte](10 * time.Second)

	c, err := cache.NewEncryptedCache[string, string](ttlCache, []byte(salt), 25)
	test.CheckErr(err, "Unexpected err", t)

	// Set a value in the cache
	err = c.Set("key", "value")
	test.CheckErr(err, "Unexpected err", t)

	// Set another value in the cache
	err = c.Set("key2", "value2")
	test.CheckErr(err, "Unexpected err", t)

	// Get the value from the cache
	value, _, err := c.Get("key")
	test.CheckErr(err, "Unexpected err", t)
	test.CheckEqual(value, "value", "value", t)

	err = c.Purge()
	test.CheckErr(err, "Unexpected err", t)
	value, _, err = c.Get("key")
	test.CheckEqual(value, "value", "", t)
	test.CheckErr(err, "Unexpected err", t)
	value, _, err = c.Get("key2")
	test.CheckEqual(value, "value", "", t)
	test.CheckErr(err, "Unexpected err", t)
}
