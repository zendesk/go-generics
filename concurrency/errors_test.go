package concurrency

import (
	"testing"
)

func TestWrapError(t *testing.T) {
	baseErr := ErrorLockNotAcquired
	wrappedErr := WrapError(baseErr, "additional context")

	if wrappedErr.Error() != "additional context: go-generics: concurrency: lock not acquired" {
		t.Fatalf("unexpected error message: %v", wrappedErr.Error())
	}
}
