//go:build redis
// +build redis

package cache

import (
	"context"
	"testing"
	"time"

	goredislib "github.com/redis/go-redis/v9"
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

	testKey := "test-expiration-key"
	testValue := "test-value-that-should-expire"

	// Step 1: Set an item in the cache
	err := cache.Set(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set item in cache: %v", err)
	}

	// Step 2: Wait a bit (but less than TTL), then get the item to verify it exists
	time.Sleep(50 * time.Millisecond)

	retrievedValue, found, err := cache.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get item from cache: %v", err)
	}
	if !found {
		t.Fatal("Expected item to be found in cache but it wasn't")
	}
	if retrievedValue != testValue {
		t.Fatalf("Expected value %s, got %s", testValue, retrievedValue)
	}

	// Step 3: Wait until after the original TTL should have expired
	// The item should be expired now since Redis cache doesn't refresh TTL on Get
	// Total wait time: 50 + 190 = 240ms > 200ms TTL but less than  250 (get @ 50 + 200 (ttl). if it was refreshed it would still be in the cache
	time.Sleep(190 * time.Millisecond)

	// Step 4: Verify that the item no longer exists in the cache
	_, found, err = cache.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get item from cache after expiration: %v", err)
	}
	if found {
		t.Fatal("Expected item to be expired and not found in cache, but it was found")
	}
}

func TestRedisCache_MultipleItemsExpiration(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	ctx := context.Background()

	// Create cache with short TTL for testing
	cacheTTL := 150 * time.Millisecond
	cache := NewRedisCache[string, int](ctx, client, cacheTTL)

	// Set multiple items
	testData := map[string]int{
		"item1": 100,
		"item2": 200,
		"item3": 300,
	}

	// Set all items
	for key, value := range testData {
		err := cache.Set(key, value)
		if err != nil {
			t.Fatalf("Failed to set item %s in cache: %v", key, err)
		}
	}

	// Wait a bit and verify all items exist
	time.Sleep(50 * time.Millisecond)

	for key, expectedValue := range testData {
		retrievedValue, found, err := cache.Get(key)
		if err != nil {
			t.Fatalf("Failed to get item %s from cache: %v", key, err)
		}
		if !found {
			t.Fatalf("Expected item %s to be found in cache but it wasn't", key)
		}
		if retrievedValue != expectedValue {
			t.Fatalf("Expected value %d for key %s, got %d", expectedValue, key, retrievedValue)
		}
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond) // Total: 200ms > 150ms TTL

	// Verify all items are expired
	for key := range testData {
		_, found, err := cache.Get(key)
		if err != nil {
			t.Fatalf("Failed to get item %s from cache after expiration: %v", key, err)
		}
		if found {
			t.Fatalf("Expected item %s to be expired and not found in cache, but it was found", key)
		}
	}
}
