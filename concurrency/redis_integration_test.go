//go:build redis
// +build redis

package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
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

func TestRedisLockBackend_Integration(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	ctx := context.Background()
	name := "test-redis-lock"
	ttl := 100 * time.Millisecond

	// Test obtaining a lock
	lock, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain Redis lock: %v", err)
	}

	// Test releasing the lock
	if err := lock.Release(ctx); err != nil {
		t.Fatalf("failed to release Redis lock: %v", err)
	}

	// Test refreshing the lock
	lock, err = backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain Redis lock for refresh test: %v", err)
	}
	defer lock.Release(ctx)

	if err := lock.Refresh(ctx); err != nil {
		t.Fatalf("failed to refresh Redis lock: %v", err)
	}
}

func TestRedisLockBackend_LockConflicts(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	ctx := context.Background()
	name := "conflict-redis-lock"
	ttl := 10 * time.Second

	// Obtain first lock
	lock1, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain first Redis lock: %v", err)
	}

	// Try to obtain the same lock - should fail
	_, err = backend.ObtainLock(ctx, name, ttl)
	if err == nil {
		t.Fatalf("expected error when obtaining conflicting Redis lock")
	}
	if !errors.Is(err, ErrorLockNotAcquired) {
		t.Fatalf("expected ErrorLockNotAcquired, got: %v", err)
	}

	// Release the first lock
	if err := lock1.Release(ctx); err != nil {
		t.Fatalf("failed to release Redis lock: %v", err)
	}

	// Now we should be able to obtain the lock again
	lock2, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain Redis lock after release: %v", err)
	}
	defer lock2.Release(ctx)
}

func TestRedisLockBackend_TTLExpiration(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	ctx := context.Background()
	name := "expire-redis-lock"
	shortTTL := 100 * time.Millisecond

	// Obtain lock with short TTL
	_, err := backend.ObtainLock(ctx, name, shortTTL)
	if err != nil {
		t.Fatalf("failed to obtain Redis lock: %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(shortTTL + 50*time.Millisecond)

	// Should be able to obtain lock again after expiration
	lock2, err := backend.ObtainLock(ctx, name, shortTTL)
	if err != nil {
		t.Fatalf("failed to obtain Redis lock after TTL expiration: %v", err)
	}
	defer lock2.Release(ctx)

	// Verify we got a new lock after TTL expiration
	if lock2 == nil {
		t.Fatalf("expected to get a new lock after TTL expiration")
	}
}

func TestRedisLockBackend_RefreshOperations(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	ctx := context.Background()
	name := "refresh-redis-lock"
	ttl := 200 * time.Millisecond

	// Obtain lock
	lock, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain Redis lock: %v", err)
	}
	defer lock.Release(ctx)

	// Wait a bit, then refresh
	time.Sleep(50 * time.Millisecond)
	if err := lock.Refresh(ctx); err != nil {
		t.Fatalf("failed to refresh Redis lock: %v", err)
	}

	// Multiple refreshes should work
	for i := 0; i < 3; i++ {
		time.Sleep(30 * time.Millisecond)
		if err := lock.Refresh(ctx); err != nil {
			t.Fatalf("failed to refresh Redis lock on iteration %d: %v", i, err)
		}
	}
}

func TestRedisLockBackend_ConcurrentAccess(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	ctx := context.Background()
	name := "concurrent-redis-lock"
	ttl := 20 * time.Second

	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex
	var successfulLock Lock

	// Launch multiple goroutines trying to acquire the same lock
	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			lock, err := backend.ObtainLock(ctx, name, ttl)
			if err == nil {
				mu.Lock()
				successCount++
				if successfulLock == nil {
					successfulLock = lock
				}
				mu.Unlock()

				// Hold the lock briefly
				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Clean up the successful lock
	if successfulLock != nil {
		successfulLock.Release(ctx)
	}

	// Only one goroutine should have successfully acquired the lock
	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful Redis lock acquisition, got: %d", successCount)
	}
}

func TestRedisLockBackend_MultipleLocks(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	ctx := context.Background()
	ttl := 200 * time.Millisecond

	// Obtain multiple locks with different names
	locks := make([]Lock, 5)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("multi-redis-lock-%d", i)
		lock, err := backend.ObtainLock(ctx, name, ttl)
		if err != nil {
			t.Fatalf("failed to obtain Redis lock %d: %v", i, err)
		}
		locks[i] = lock
	}

	// All locks should be independent and refreshable
	for i, lock := range locks {
		if err := lock.Refresh(ctx); err != nil {
			t.Fatalf("failed to refresh Redis lock %d: %v", i, err)
		}
	}

	// Release all locks
	for i, lock := range locks {
		if err := lock.Release(ctx); err != nil {
			t.Fatalf("failed to release Redis lock %d: %v", i, err)
		}
	}
}

func TestRedisLockBackend_LockAfterRelease(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	ctx := context.Background()
	name := "reuse-redis-lock"
	ttl := 100 * time.Millisecond

	// Obtain and release lock multiple times
	for i := 0; i < 3; i++ {
		lock, err := backend.ObtainLock(ctx, name, ttl)
		if err != nil {
			t.Fatalf("failed to obtain Redis lock on iteration %d: %v", i, err)
		}

		// Refresh to ensure it works
		if err := lock.Refresh(ctx); err != nil {
			t.Fatalf("failed to refresh Redis lock on iteration %d: %v", i, err)
		}

		// Release the lock
		if err := lock.Release(ctx); err != nil {
			t.Fatalf("failed to release Redis lock on iteration %d: %v", i, err)
		}
	}
}

func TestRedisLockBackend_WithLockManager(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	manager := NewLockManager(backend)
	ctx := context.Background()
	key := "test-redis-manager-key"

	// Test acquiring a lock through the manager
	lock, acquired, err := manager.Acquire(ctx, key, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to acquire Redis lock through manager: %v", err)
	}
	if !acquired {
		t.Fatalf("Redis lock not acquired through manager")
	}
	defer lock.Release(ctx)

	// Test ExecuteWithLock
	executed := false
	err = manager.ExecuteWithLock(ctx, "another-redis-key", 100*time.Millisecond, 500*time.Millisecond, func() error {
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("failed to execute with Redis lock: %v", err)
	}
	if !executed {
		t.Fatalf("function was not executed with Redis lock")
	}
}

func TestRedisLockBackend_ContextCancellation(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool})
	name := "context-cancel-lock"
	ttl := 100 * time.Millisecond

	// Test context cancellation during lock acquisition
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := backend.ObtainLock(ctx, name, ttl)
	if err == nil {
		t.Fatalf("expected error when acquiring lock with canceled context")
	}

	// Test normal acquisition then context cancellation during operations
	normalCtx := context.Background()
	lock, err := backend.ObtainLock(normalCtx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain Redis lock: %v", err)
	}

	// Use canceled context for refresh - behavior may vary by implementation
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Try to refresh with canceled context
	err = lock.Refresh(canceledCtx)
	// Note: redsync might handle this differently, so we just ensure it doesn't panic

	// Clean up with normal context
	if err := lock.Release(normalCtx); err != nil {
		t.Fatalf("failed to release Redis lock: %v", err)
	}
}

func TestRedisLockBackend_WithPrefix(t *testing.T) {
	client := setupRedisClient(t)
	defer client.Close()

	pool := goredis.NewPool(client)
	prefix := "svc:"
	backend := NewRedisLockBackend([]redsyncredis.Pool{pool}, WithPrefix(prefix))
	ctx := context.Background()
	name := "prefix-redis-lock"
	ttl := 10 * time.Second

	lock, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain prefixed Redis lock: %v", err)
	}
	defer lock.Release(ctx)

	// The prefixed key should exist; the unprefixed key should not.
	if exists, err := client.Exists(ctx, prefix+name).Result(); err != nil {
		t.Fatalf("failed to check prefixed key existence: %v", err)
	} else if exists != 1 {
		t.Fatalf("expected prefixed key %q to exist, got exists=%d", prefix+name, exists)
	}
	if exists, err := client.Exists(ctx, name).Result(); err != nil {
		t.Fatalf("failed to check unprefixed key existence: %v", err)
	} else if exists != 0 {
		t.Fatalf("expected unprefixed key %q to not exist, got exists=%d", name, exists)
	}

	// A second backend with the same prefix should conflict on the same name.
	backend2 := NewRedisLockBackend([]redsyncredis.Pool{pool}, WithPrefix(prefix))
	if _, err := backend2.ObtainLock(ctx, name, ttl); !errors.Is(err, ErrorLockNotAcquired) {
		t.Fatalf("expected ErrorLockNotAcquired from conflicting prefixed lock, got: %v", err)
	}

	// A backend with a different prefix should not conflict.
	backend3 := NewRedisLockBackend([]redsyncredis.Pool{pool}, WithPrefix("other:"))
	lock3, err := backend3.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("expected independent lock under different prefix, got: %v", err)
	}
	_ = lock3.Release(ctx)
}
