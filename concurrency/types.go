package concurrency

import (
	"context"
	"fmt"
	"time"
)

// LockBackend defines the interface for lock storage backends
type LockBackend interface {
	ObtainLock(ctx context.Context, name string, ttl time.Duration) (Lock, error)
}

// Lock defines the interface for a lock.
type Lock interface {
	Release(ctx context.Context) error
	Refresh(ctx context.Context) error
}

// LockError represents an error that occurred during locking operations
type LockError struct {
	Message string
}

func (le LockError) Error() string {
	return le.Message
}

// LockTimeoutError represents an error when a lock operation times out
type LockTimeoutError struct {
	Key string
}

func (lte LockTimeoutError) Error() string {
	return fmt.Sprintf("timed out waiting for lock: %s", lte.Key)
}
