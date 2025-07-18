package concurrency

import (
	"errors"
	"fmt"
)

var (
	// ErrorLockNotAcquired is returned when a lock cannot be acquired. This means the lock is already held.
	ErrorLockNotAcquired = errors.New("go-generics: concurrency: lock not acquired")
	// ErrorLockNotRefreshed is returned when a lock cannot be refreshed. This means the lock is not held or has expired.
	ErrorLockNotRefreshed = errors.New("go-generics: concurrency: lock not refreshed")
	// ErrorLockNotReleased is returned when trying to release an inactive lock.
	ErrorLockNotReleased = errors.New("go-generics: concurrency: lock not held")
)

func WrapError(err error, message string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", message, err)
	}
	return nil
}
