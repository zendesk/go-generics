package concurrency

import (
	"context"
	"sync"
	"time"
)

type MemoryLockBackend struct {
	locks sync.Map
}

type memoryLock struct {
	name      string
	expiresAt time.Time
	ttl       time.Duration
	backend   *MemoryLockBackend
}

func NewMemoryLockBackend() *MemoryLockBackend {
	return &MemoryLockBackend{}
}

func (mlb *MemoryLockBackend) ObtainLock(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
	lock := &memoryLock{
		name:      name,
		expiresAt: time.Now().Add(ttl),
		ttl:       ttl,
		backend:   mlb,
	}

	actual, loaded := mlb.locks.LoadOrStore(name, lock)
	if loaded {
		existingLock := actual.(*memoryLock)
		if time.Now().Before(existingLock.expiresAt) {
			return nil, ErrorLockNotAcquired
		}
		mlb.locks.Store(name, lock)
	}

	return lock, nil
}

func (ml *memoryLock) Release(ctx context.Context) error {
	actual, loaded := ml.backend.locks.Load(ml.name)
	if !loaded || actual != ml {
		return ErrorLockNotReleased
	}

	ml.backend.locks.Delete(ml.name)
	return nil
}

func (ml *memoryLock) Refresh(ctx context.Context) error {
	actual, loaded := ml.backend.locks.Load(ml.name)
	if !loaded || actual != ml {
		return ErrorLockNotRefreshed
	}

	ml.expiresAt = time.Now().Add(ml.ttl)
	return nil
}

// assert that MemoryLockBackend implements LockBackend
var _ LockBackend = (*MemoryLockBackend)(nil)

// MemoryLock implements Lock
var _ Lock = (*memoryLock)(nil)
