//go:build test
// +build test

package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/zendesk/go-generics/cache"
	"github.com/zendesk/go-generics/internal/test"
)

func Test_WithKeyPrefix_Redis_Set_Get(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	prefix := cache.WithKeyPrefix("mcp:")

	err := redisCache.Set("mykey", "myval", prefix)
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "SET key should have prefix", "mcp:"+cache.HashAny("mykey"), t)

	_, _, err = redisCache.Get("mykey", prefix)
	test.CheckErr(err, "Failed to get", t)
	test.CheckEqual(mockRedis.getKey, "GET key should have prefix", "mcp:"+cache.HashAny("mykey"), t)
}

func Test_WithKeyPrefix_Redis_GetOrSet(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	prefix := cache.WithKeyPrefix("ns:")

	// First Set so the mock has data, then GetOrSet should find it
	err := redisCache.Set("key1", "existing", prefix)
	test.CheckErr(err, "Failed to seed", t)

	val, fromCache, err := redisCache.GetOrSet("key1", func() (string, error) {
		return "should-not-be-used", nil
	}, prefix)
	test.CheckErr(err, "Failed GetOrSet", t)
	test.CheckOk(fromCache, "Should be from cache", t)
	test.CheckEqual(val, "Value", "existing", t)
	test.CheckEqual(mockRedis.getKey, "GET key should have prefix", "ns:"+cache.HashAny("key1"), t)
}

func Test_WithKeyPrefix_Redis_NoPrefix(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	// Without prefix, key should be plain hash
	err := redisCache.Set("mykey", "myval")
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "SET key should be plain hash", cache.HashAny("mykey"), t)
}

func Test_WithKeyPrefix_Redis_DifferentPrefixes_Isolate(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	// Set with prefix "a:"
	err := redisCache.Set("key", "val-a", cache.WithKeyPrefix("a:"))
	test.CheckErr(err, "Failed to set with prefix a:", t)

	// Set with prefix "b:" (overwrites mock's stored value)
	err = redisCache.Set("key", "val-b", cache.WithKeyPrefix("b:"))
	test.CheckErr(err, "Failed to set with prefix b:", t)

	// Get with prefix "b:" should return val-b (last written to mock)
	got, _, err := redisCache.Get("key", cache.WithKeyPrefix("b:"))
	test.CheckErr(err, "Failed to get with prefix b:", t)
	test.CheckEqual(got, "Value with prefix b:", "val-b", t)
	test.CheckEqual(mockRedis.getKey, "GET key should have b: prefix", "b:"+cache.HashAny("key"), t)
}

func Test_WithKeyPrefix_InMemory_Set_Get(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	prefix := cache.WithKeyPrefix("mcp:")

	err := memCache.Set("mykey", "myval", prefix)
	test.CheckErr(err, "Failed to set", t)

	val, found, err := memCache.Get("mykey", prefix)
	test.CheckErr(err, "Failed to get", t)
	test.CheckOk(found, "Should be found", t)
	test.CheckEqual(val, "Value", "myval", t)
}

func Test_WithKeyPrefix_InMemory_Isolation(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	// Set with prefix "a:"
	err := memCache.Set("key", "val-a", cache.WithKeyPrefix("a:"))
	test.CheckErr(err, "Failed to set with prefix a:", t)

	// Set with prefix "b:"
	err = memCache.Set("key", "val-b", cache.WithKeyPrefix("b:"))
	test.CheckErr(err, "Failed to set with prefix b:", t)

	// Get with "a:" should return val-a
	val, found, err := memCache.Get("key", cache.WithKeyPrefix("a:"))
	test.CheckErr(err, "Failed to get with prefix a:", t)
	test.CheckOk(found, "Should be found with prefix a:", t)
	test.CheckEqual(val, "Value with prefix a:", "val-a", t)

	// Get with "b:" should return val-b
	val, found, err = memCache.Get("key", cache.WithKeyPrefix("b:"))
	test.CheckErr(err, "Failed to get with prefix b:", t)
	test.CheckOk(found, "Should be found with prefix b:", t)
	test.CheckEqual(val, "Value with prefix b:", "val-b", t)

	// Get without prefix should not find either
	_, found, err = memCache.Get("key")
	test.CheckErr(err, "Failed to get without prefix", t)
	test.CheckNotOk(found, "Should not be found without prefix", t)
}

func Test_WithKeyPrefix_InMemory_Delete(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	prefix := cache.WithKeyPrefix("mcp:")

	err := memCache.Set("key", "val", prefix)
	test.CheckErr(err, "Failed to set", t)

	err = memCache.Delete("key", prefix)
	test.CheckErr(err, "Failed to delete", t)

	_, found, err := memCache.Get("key", prefix)
	test.CheckErr(err, "Failed to get after delete", t)
	test.CheckNotOk(found, "Should not be found after delete", t)
}

func Test_WithKeyPrefix_InMemory_GetOrSet(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	prefix := cache.WithKeyPrefix("ns:")

	val, fromCache, err := memCache.GetOrSet("key1", func() (string, error) {
		return "computed", nil
	}, prefix)
	test.CheckErr(err, "Failed GetOrSet", t)
	test.CheckNotOk(fromCache, "Should not be from cache on first call", t)
	test.CheckEqual(val, "Value", "computed", t)

	// Second call should hit cache
	val, fromCache, err = memCache.GetOrSet("key1", func() (string, error) {
		return "should-not-be-used", nil
	}, prefix)
	test.CheckErr(err, "Failed GetOrSet (2)", t)
	test.CheckOk(fromCache, "Should be from cache on second call", t)
	test.CheckEqual(val, "Value", "computed", t)
}

// Tests for the cache-constructor option WithPrefix (CacheBackendOption).
// The static prefix is stored on the backend and prepended to every key on
// every operation — no per-call OperationOption required.

func Test_WithPrefix_Redis_AppliesToAllOperations(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second,
		cache.WithPrefix[string, string]("svc:"))

	err := redisCache.Set("mykey", "myval")
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "SET key should have constructor prefix", "svc:"+cache.HashAny("mykey"), t)

	_, _, err = redisCache.Get("mykey")
	test.CheckErr(err, "Failed to get", t)
	test.CheckEqual(mockRedis.getKey, "GET key should have constructor prefix", "svc:"+cache.HashAny("mykey"), t)

	err = redisCache.Delete("mykey")
	test.CheckErr(err, "Failed to delete", t)
	test.CheckEqual(mockRedis.delKey, "DEL key should have constructor prefix", "svc:"+cache.HashAny("mykey"), t)
}

func Test_WithPrefix_Redis_StacksWithOperationPrefix(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second,
		cache.WithPrefix[string, string]("svc:"))

	// Backend prefix "svc:" + operation prefix "user:" + hash(key).
	err := redisCache.Set("k", "v", cache.WithKeyPrefix("user:"))
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "SET key should stack both prefixes",
		"svc:user:"+cache.HashAny("k"), t)
}

func Test_WithPrefix_InMemory_AppliesToAllOperations(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10*time.Second,
		cache.WithPrefix[string, string]("svc:"))

	err := memCache.Set("k", "v")
	test.CheckErr(err, "Failed to set", t)

	val, found, err := memCache.Get("k")
	test.CheckErr(err, "Failed to get", t)
	test.CheckOk(found, "Should be found with same backend prefix", t)
	test.CheckEqual(val, "Value", "v", t)

	err = memCache.Delete("k")
	test.CheckErr(err, "Failed to delete", t)

	_, found, err = memCache.Get("k")
	test.CheckErr(err, "Failed to get after delete", t)
	test.CheckNotOk(found, "Should not be found after delete", t)
}

func Test_WithPrefix_InMemory_IsolatesCachesWithDifferentBackendPrefixes(t *testing.T) {
	// Two backends with different static prefixes must never collide even when
	// callers use identical keys with no OperationOption.
	cacheA := cache.NewInMemoryCache[string, string](10*time.Second,
		cache.WithPrefix[string, string]("a:"))
	cacheB := cache.NewInMemoryCache[string, string](10*time.Second,
		cache.WithPrefix[string, string]("b:"))

	err := cacheA.Set("key", "val-a")
	test.CheckErr(err, "Failed to set on cacheA", t)
	err = cacheB.Set("key", "val-b")
	test.CheckErr(err, "Failed to set on cacheB", t)

	valA, foundA, err := cacheA.Get("key")
	test.CheckErr(err, "Failed to get on cacheA", t)
	test.CheckOk(foundA, "cacheA hit", t)
	test.CheckEqual(valA, "cacheA value", "val-a", t)

	valB, foundB, err := cacheB.Get("key")
	test.CheckErr(err, "Failed to get on cacheB", t)
	test.CheckOk(foundB, "cacheB hit", t)
	test.CheckEqual(valB, "cacheB value", "val-b", t)
}

func Test_WithPrefix_EmptyPrefix_BehavesLikeUnset(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second,
		cache.WithPrefix[string, string](""))

	err := redisCache.Set("k", "v")
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "SET key with empty backend prefix should be plain hash",
		cache.HashAny("k"), t)
}

func Test_WithPrefix_EncryptedCache_AppliesToBackend(t *testing.T) {
	// When an EncryptedCache wraps a backend that has WithPrefix set, the
	// prefix must reach Redis unchanged — encryption operates on the value,
	// not the key.
	mockRedis := &mockClient{}
	redisBackend := cache.NewRedisCache[string, []byte](context.Background(), mockRedis, 10*time.Second,
		cache.WithPrefix[string, []byte]("svc:"))

	encrypted, err := cache.NewEncryptedCacheWithPassword[string, string](redisBackend,
		[]byte("password-at-least-12-bytes"),
		[]byte("saltsaltsaltsalt"), // 16 bytes, meets MinSaltLength
		100_000)
	test.CheckErr(err, "Failed to construct EncryptedCache", t)

	err = encrypted.Set("k", "v")
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "encrypted-cache SET should reach Redis with the backend prefix",
		"svc:"+cache.HashAny("k"), t)
}
