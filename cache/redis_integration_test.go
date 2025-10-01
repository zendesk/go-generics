//go:build redis
// +build redis

package cache

import (
	"context"
	"testing"
	"time"

	goredislib "github.com/redis/go-redis/v9"
	"github.com/zendesk/go-generics/internal/test"
)

const (
	redisAddr = "localhost:6379"
	redisDB   = 0
)

func setupRedisClient(t *testing.T) *goredislib.Client {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: redisAddr,
		DB:   redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at %s: %v", redisAddr, err)
	}

	// Clear any existing test keys
	client.FlushDB(ctx)

	return client
}

func TestRedisCache_ExpirationWithoutRefresh(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	ctx := context.Background()

	// Create cache with short TTL for testing
	cacheTTL := 200 * time.Millisecond
	cache := NewRedisCache[string, string](ctx, client, cacheTTL)

	testKey := test.GenerateRandomLetterString(10)
	testValue := "test-value-that-should-expire"

	// Step 1: Set an item in the cache
	err := cache.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set item in cache: %v", err)
	}

	// Step 2: Wait a bit (but less than TTL), then get the item to verify it exists
	time.Sleep(50 * time.Millisecond)

	retrievedValue, found, err := cache.Get(testKey)
	test.CheckErr(err, "Failed to get item from cache", t)
	test.CheckOk(found, "Expected item to be found in cache but it was not", t)
	test.CheckEqual(retrievedValue, "Value", testValue, t)

	// Step 3: Wait until after the original TTL should have expired
	// The item should be expired now since Redis cache doesn't refresh TTL on Get
	// Total wait time: 50 + 190 = 240ms > 200ms TTL but less than  250 (get @ 50 + 200 (ttl). if it was refreshed it would still be in the cache
	time.Sleep(190 * time.Millisecond)

	// Step 4: Verify that the item no longer exists in the cache
	_, found, err = cache.Get(testKey)
	test.CheckErr(err, "Failed to get item from cache, got unexpected error", t)
	test.CheckNotOk(found, "Expected item to NOT bein cache but it was found", t)
}
