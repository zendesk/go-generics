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

func Test_WithPrefix_Redis_Set_Get(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	prefix := cache.WithPrefix("mcp:")

	err := redisCache.Set("mykey", "myval", prefix)
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "SET key should have prefix", "mcp:"+cache.HashAny("mykey"), t)

	_, _, err = redisCache.Get("mykey", prefix)
	test.CheckErr(err, "Failed to get", t)
	test.CheckEqual(mockRedis.getKey, "GET key should have prefix", "mcp:"+cache.HashAny("mykey"), t)
}

func Test_WithPrefix_Redis_GetOrSet(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	prefix := cache.WithPrefix("ns:")

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

func Test_WithPrefix_Redis_NoPrefix(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	// Without prefix, key should be plain hash
	err := redisCache.Set("mykey", "myval")
	test.CheckErr(err, "Failed to set", t)
	test.CheckEqual(mockRedis.setKey, "SET key should be plain hash", cache.HashAny("mykey"), t)
}

func Test_WithPrefix_Redis_DifferentPrefixes_Isolate(t *testing.T) {
	mockRedis := &mockClient{}
	redisCache := cache.NewRedisCache[string, string](context.Background(), mockRedis, 10*time.Second)

	// Set with prefix "a:"
	err := redisCache.Set("key", "val-a", cache.WithPrefix("a:"))
	test.CheckErr(err, "Failed to set with prefix a:", t)

	// Set with prefix "b:" (overwrites mock's stored value)
	err = redisCache.Set("key", "val-b", cache.WithPrefix("b:"))
	test.CheckErr(err, "Failed to set with prefix b:", t)

	// Get with prefix "b:" should return val-b (last written to mock)
	got, _, err := redisCache.Get("key", cache.WithPrefix("b:"))
	test.CheckErr(err, "Failed to get with prefix b:", t)
	test.CheckEqual(got, "Value with prefix b:", "val-b", t)
	test.CheckEqual(mockRedis.getKey, "GET key should have b: prefix", "b:"+cache.HashAny("key"), t)
}

func Test_WithPrefix_InMemory_Set_Get(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	prefix := cache.WithPrefix("mcp:")

	err := memCache.Set("mykey", "myval", prefix)
	test.CheckErr(err, "Failed to set", t)

	val, found, err := memCache.Get("mykey", prefix)
	test.CheckErr(err, "Failed to get", t)
	test.CheckOk(found, "Should be found", t)
	test.CheckEqual(val, "Value", "myval", t)
}

func Test_WithPrefix_InMemory_Isolation(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	// Set with prefix "a:"
	err := memCache.Set("key", "val-a", cache.WithPrefix("a:"))
	test.CheckErr(err, "Failed to set with prefix a:", t)

	// Set with prefix "b:"
	err = memCache.Set("key", "val-b", cache.WithPrefix("b:"))
	test.CheckErr(err, "Failed to set with prefix b:", t)

	// Get with "a:" should return val-a
	val, found, err := memCache.Get("key", cache.WithPrefix("a:"))
	test.CheckErr(err, "Failed to get with prefix a:", t)
	test.CheckOk(found, "Should be found with prefix a:", t)
	test.CheckEqual(val, "Value with prefix a:", "val-a", t)

	// Get with "b:" should return val-b
	val, found, err = memCache.Get("key", cache.WithPrefix("b:"))
	test.CheckErr(err, "Failed to get with prefix b:", t)
	test.CheckOk(found, "Should be found with prefix b:", t)
	test.CheckEqual(val, "Value with prefix b:", "val-b", t)

	// Get without prefix should not find either
	_, found, err = memCache.Get("key")
	test.CheckErr(err, "Failed to get without prefix", t)
	test.CheckNotOk(found, "Should not be found without prefix", t)
}

func Test_WithPrefix_InMemory_Delete(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	prefix := cache.WithPrefix("mcp:")

	err := memCache.Set("key", "val", prefix)
	test.CheckErr(err, "Failed to set", t)

	err = memCache.Delete("key", prefix)
	test.CheckErr(err, "Failed to delete", t)

	_, found, err := memCache.Get("key", prefix)
	test.CheckErr(err, "Failed to get after delete", t)
	test.CheckNotOk(found, "Should not be found after delete", t)
}

func Test_WithPrefix_InMemory_GetOrSet(t *testing.T) {
	memCache := cache.NewInMemoryCache[string, string](10 * time.Second)

	prefix := cache.WithPrefix("ns:")

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
