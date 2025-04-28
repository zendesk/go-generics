package cache_test

import (
	"testing"
	"time"

	"github.com/zendesk/go-generics/cache"
	"github.com/zendesk/go-generics/cache/internal/test"
)

type cacheObs[K comparable] struct {
	count map[string]int
}

func (c *cacheObs[K]) Hit(k K) {
	c.count["hit"]++
}
func (c *cacheObs[K]) Miss(k K) {
	c.count["miss"]++
}
func (c *cacheObs[K]) Get(k K) {
	c.count["get"]++
}
func (c *cacheObs[K]) Set(k K) {
	c.count["set"]++
}
func (c *cacheObs[K]) Delete(k K) {
	c.count["delete"]++
}
func (c *cacheObs[K]) Purge() {
	c.count["purge"]++
}

func Test_Cache_Observer(t *testing.T) {
	memBackend := cache.NewInMemoryCache[string, string](time.Second*20,
		cache.WithCapacity[string, string](uint64(1000)))

	observer := cacheObs[string]{count: make(map[string]int)}

	cash := cache.NewCache(memBackend, cache.WithObserver[string, string](&observer))
	_, _, _ = cash.Get("missing")

	// Test observer get
	test.CheckEqual(observer.count["get"], "Get: Get count should be 1", 1, t)
	test.CheckEqual(observer.count["hit"], "Get: Hit count should be 0", 0, t)
	test.CheckEqual(observer.count["miss"], "Get: Miss count should be 1", 1, t)

	// Reset
	_ = cash.Purge()
	observer.count = make(map[string]int)

	// Test observer set
	_ = cash.Set("foo", "bar")
	test.CheckEqual(observer.count["set"], "Set: Set count should be 1", 1, t)
	test.CheckEqual(observer.count["hit"], "Set: Hit count should be 1", 0, t)
	test.CheckEqual(observer.count["mis"], "Set: Miss count should be 1", 0, t)

	// Reset
	_ = cash.Purge()
	observer.count = make(map[string]int)

	// Test observer delete
	_ = cash.Set("foo", "bar")
	_ = cash.Delete("foo")
	test.CheckEqual(observer.count["delete"], "Delete: Delete count should be 1", 1, t)
	test.CheckEqual(observer.count["hit"], "Delete: Hit count should be 0", 0, t)
	test.CheckEqual(observer.count["miss"], "Delete: Miss count should be 0", 0, t)

	// Reset
	_ = cash.Purge()
	observer.count = make(map[string]int)

	// Test observer purge
	_ = cash.Purge()
	test.CheckEqual(observer.count["purge"], "Purge: Purge count should be 1", 1, t)
	test.CheckEqual(observer.count["hit"], "Purge: Hit count should be 0", 0, t)
	test.CheckEqual(observer.count["miss"], "Purge: Miss count should be 0", 0, t)

	// Reset
	_ = cash.Purge()
	observer.count = make(map[string]int)

	// Test cash GetOrSet

	// First GetorSet should miss but set value
	_, _, _ = cash.GetOrSet("foo", func() (string, error) {
		return "bar", nil
	})
	test.CheckEqual(observer.count["get"], "GetOrSet: Get count should be 1", 1, t)
	test.CheckEqual(observer.count["set"], "GetOrSet: Set count should be 1", 1, t)
	test.CheckEqual(observer.count["miss"], "GetOrSet: Miss count should be 1", 1, t)
	test.CheckEqual(observer.count["hit"], "GetOrSet: Hit count should be 0", 0, t)

	// Run again
	_, _, _ = cash.GetOrSet("foo", func() (string, error) {
		return "bar", nil
	})

	test.CheckEqual(observer.count["get"], "GetOrSet: Get count should be 2", 2, t)
	test.CheckEqual(observer.count["set"], "GetOrSet: Set count should be 1", 1, t)
	test.CheckEqual(observer.count["miss"], "GetOrSet: Miss count should be 1", 1, t)
	test.CheckEqual(observer.count["hit"], "GetOrSet: Hit count should be 1", 1, t)
}
