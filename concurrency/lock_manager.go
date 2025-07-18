package concurrency

import (
	"context"
	"errors"
	"time"
)

type LockManager struct {
	backend LockBackend
}

// NewLockManager creates a new LockManager with the provided backend
func NewLockManager(backend LockBackend) *LockManager {
	return &LockManager{
		backend: backend,
	}
}

func (lm *LockManager) Acquire(ctx context.Context, key string, ttl time.Duration) (Lock, bool, error) {
	lock, err := lm.backend.ObtainLock(ctx, key, ttl)
	if err != nil {
		if errors.Is(err, ErrorLockNotAcquired) {
			return nil, false, nil // Not acquired, but not an error
		}
		return lock, false, err // Other errors are actual errors
	}
	return lock, true, nil
}

// ExecuteWithLock attempts to acquire a lock for the given key and executes the provided function.
// If the lock cannot be acquired immediately, it will retry with exponential backoff until the timeout is reached.
// If the lock is acquired, it will ensure that the lock is released after the function execution.
// If the function returns an error, it will be propagated back to the caller.
func (lm *LockManager) ExecuteWithLock(ctx context.Context, key string, lockTTL, timeout time.Duration, fn func() error) error {
	lock, err := lm.acquireWithRetry(ctx, key, lockTTL, timeout)
	if err != nil {
		return err
	}

	return lm.executeWithLockHeld(ctx, lock, fn)
}

// acquireWithRetry attempts to acquire a lock with retry logic and exponential backoff
func (lm *LockManager) acquireWithRetry(ctx context.Context, key string, lockTTL, timeout time.Duration) (Lock, error) {
	// Try to acquire the lock immediately
	lock, acquired, err := lm.Acquire(ctx, key, lockTTL)
	if err != nil {
		return nil, err
	}
	if acquired {
		return lock, nil
	}

	// If not acquired immediately, retry with backoff
	return lm.retryAcquireWithBackoff(ctx, key, lockTTL, timeout)
}

// retryAcquireWithBackoff implements the retry logic with exponential backoff
func (lm *LockManager) retryAcquireWithBackoff(ctx context.Context, key string, lockTTL, timeout time.Duration) (Lock, error) {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	backoff := 50 * time.Millisecond
	maxBackoff := time.Second

	for {
		select {
		case <-waitCtx.Done():
			return nil, LockTimeoutError{Key: key}
		case <-time.After(backoff):
			lock, acquired, err := lm.Acquire(ctx, key, lockTTL)
			if err != nil {
				return nil, err
			}
			if acquired {
				return lock, nil
			}

			backoff = lm.calculateNextBackoff(backoff, maxBackoff)
		}
	}
}

// calculateNextBackoff calculates the next backoff duration with exponential growth and jitter
func (lm *LockManager) calculateNextBackoff(current, max time.Duration) time.Duration {
	// Exponential backoff with jitter
	next := time.Duration(float64(current)*1.5) + randomDurationBetween(time.Millisecond, time.Millisecond*50)
	if next > max {
		return max
	}
	return next
}

// executeWithLockHeld executes the function while ensuring the lock is released
func (lm *LockManager) executeWithLockHeld(ctx context.Context, lock Lock, fn func() error) error {
	defer func() {
		_ = lock.Release(ctx)
	}()

	return fn()
}
