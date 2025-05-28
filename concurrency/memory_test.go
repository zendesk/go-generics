package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMemoryLockBackend(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	name := "test-lock"
	ttl := 100 * time.Millisecond

	// Test obtaining a lock
	lock, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain lock: %v", err)
	}

	// Test releasing the lock
	if err := lock.Release(ctx); err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// Test refreshing the lock
	lock, err = backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain lock: %v", err)
	}
	if err := lock.Refresh(ctx); err != nil {
		t.Fatalf("failed to refresh lock: %v", err)
	}
}

func TestMemoryLockBackend_LockConflicts(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	name := "conflict-lock"
	ttl := 200 * time.Millisecond

	// Obtain first lock
	lock1, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain first lock: %v", err)
	}

	// Try to obtain the same lock - should fail
	_, err = backend.ObtainLock(ctx, name, ttl)
	if err == nil {
		t.Fatalf("expected error when obtaining conflicting lock")
	}
	if !errors.Is(err, ErrorLockNotAcquired) {
		t.Fatalf("expected ErrorLockNotAcquired, got: %v", err)
	}

	// Release the first lock
	if err := lock1.Release(ctx); err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// Now we should be able to obtain the lock again
	lock2, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain lock after release: %v", err)
	}
	defer lock2.Release(ctx)
}

func TestMemoryLockBackend_TTLExpiration(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	name := "expire-lock"
	shortTTL := 50 * time.Millisecond

	// Obtain lock with short TTL
	lock1, err := backend.ObtainLock(ctx, name, shortTTL)
	if err != nil {
		t.Fatalf("failed to obtain lock: %v", err)
	}

	// Wait for TTL to expire
	time.Sleep(shortTTL + 10*time.Millisecond)

	// Should be able to obtain lock again after expiration
	lock2, err := backend.ObtainLock(ctx, name, shortTTL)
	if err != nil {
		t.Fatalf("failed to obtain lock after TTL expiration: %v", err)
	}
	defer lock2.Release(ctx)

	// First lock should still exist but operations on it should fail
	err = lock1.Refresh(ctx)
	if err == nil {
		t.Fatalf("expected error when refreshing expired lock")
	}
	if !errors.Is(err, ErrorLockNotRefreshed) {
		t.Fatalf("expected ErrorLockNotRefreshed, got: %v", err)
	}
}

func TestMemoryLockBackend_RefreshOperations(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	name := "refresh-lock"
	ttl := 100 * time.Millisecond

	// Obtain lock
	lock, err := backend.ObtainLock(ctx, name, ttl)
	if err != nil {
		t.Fatalf("failed to obtain lock: %v", err)
	}
	defer lock.Release(ctx)

	// Get initial expiry time
	memLock := lock.(*memoryLock)
	initialExpiry := memLock.expiresAt

	// Wait a bit, then refresh
	time.Sleep(30 * time.Millisecond)
	if err := lock.Refresh(ctx); err != nil {
		t.Fatalf("failed to refresh lock: %v", err)
	}

	// Expiry should be extended
	newExpiry := memLock.expiresAt
	if !newExpiry.After(initialExpiry) {
		t.Fatalf("lock expiry was not extended after refresh")
	}

	// Multiple refreshes should work
	for i := 0; i < 3; i++ {
		time.Sleep(20 * time.Millisecond)
		if err := lock.Refresh(ctx); err != nil {
			t.Fatalf("failed to refresh lock on iteration %d: %v", i, err)
		}
	}
}

func TestMemoryLockBackend_ConcurrentAccess(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	name := "concurrent-lock"
	ttl := 200 * time.Millisecond

	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex

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
				mu.Unlock()

				// Hold the lock briefly
				time.Sleep(50 * time.Millisecond)
				lock.Release(ctx)
			}
		}(i)
	}

	wg.Wait()

	// Only one goroutine should have successfully acquired the lock
	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful lock acquisition, got: %d", successCount)
	}
}

func TestMemoryLockBackend_MultipleLocks(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	ttl := 100 * time.Millisecond

	// Obtain multiple locks with different names
	locks := make([]Lock, 5)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("multi-lock-%d", i)
		lock, err := backend.ObtainLock(ctx, name, ttl)
		if err != nil {
			t.Fatalf("failed to obtain lock %d: %v", i, err)
		}
		locks[i] = lock
	}

	// All locks should be independent and refreshable
	for i, lock := range locks {
		if err := lock.Refresh(ctx); err != nil {
			t.Fatalf("failed to refresh lock %d: %v", i, err)
		}
	}

	// Release all locks
	for i, lock := range locks {
		if err := lock.Release(ctx); err != nil {
			t.Fatalf("failed to release lock %d: %v", i, err)
		}
	}
}

func TestMemoryLockBackend_LockAfterRelease(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	name := "reuse-lock"
	ttl := 100 * time.Millisecond

	// Obtain and release lock multiple times
	for i := 0; i < 3; i++ {
		lock, err := backend.ObtainLock(ctx, name, ttl)
		if err != nil {
			t.Fatalf("failed to obtain lock on iteration %d: %v", i, err)
		}

		// Refresh to ensure it works
		if err := lock.Refresh(ctx); err != nil {
			t.Fatalf("failed to refresh lock on iteration %d: %v", i, err)
		}

		// Release the lock
		if err := lock.Release(ctx); err != nil {
			t.Fatalf("failed to release lock on iteration %d: %v", i, err)
		}
	}
}

func TestMemoryLockBackend_ZeroTTL(t *testing.T) {
	backend := NewMemoryLockBackend()
	ctx := context.Background()
	name := "zero-ttl-lock"

	// Obtain lock with zero TTL (should expire immediately)
	lock, err := backend.ObtainLock(ctx, name, 0)
	if err != nil {
		t.Fatalf("failed to obtain lock with zero TTL: %v", err)
	}

	// Lock should be immediately available for re-acquisition
	lock2, err := backend.ObtainLock(ctx, name, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to obtain lock after zero TTL: %v", err)
	}
	defer lock2.Release(ctx)

	// First lock operations should fail
	if err := lock.Refresh(ctx); err == nil {
		t.Fatalf("expected error when refreshing expired zero-TTL lock")
	}
}

func TestLockManager(t *testing.T) {
	backend := NewMemoryLockBackend()
	manager := NewLockManager(backend)
	ctx := context.Background()
	key := "test-key"

	// Test acquiring a lock
	firstLock, acquired, err := manager.Acquire(ctx, key, 100*time.Second) // intentionally long TTL
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	if !acquired {
		t.Fatalf("lock not acquired")
	}

	// Test ExecuteWithLock
	err = manager.ExecuteWithLock(ctx, key, 100*time.Millisecond, 500*time.Millisecond, func() error {
		return nil
	})
	// expect error since TTL of first lock has not yet completed
	if err == nil {
		t.Fatalf("expected error due to lock TTL not completed")
	}
	// Release the first lock
	if err := firstLock.Release(ctx); err != nil {
		t.Fatalf("failed to release first lock: %v", err)
	}
	// Try to acquire the lock again
	// try to ExecuteWithLock again
	err = manager.ExecuteWithLock(ctx, key, 100*time.Millisecond, 500*time.Millisecond, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("failed to execute with lock: %v", err)
	}
}
