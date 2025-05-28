package concurrency

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis"
)

type RedisLockBackend struct {
	locker *redsync.Redsync
}

func NewRedisLockBackend(pools ...redis.Pool) *RedisLockBackend {
	locker := redsync.New(pools...)
	return &RedisLockBackend{
		locker: locker,
	}
}

type redisMutex struct {
	mutex *redsync.Mutex
}

func (r *RedisLockBackend) ObtainLock(ctx context.Context, name string, ttl time.Duration) (Lock, error) {
	mutex := r.locker.NewMutex(name, redsync.WithExpiry(ttl))
	if err := mutex.LockContext(ctx); err != nil {
		return nil, WrapError(ErrorLockNotAcquired, fmt.Sprintf("failed to obtain lock for name %s", name))
	}
	return &redisMutex{mutex: mutex}, nil
}

// AcquireLock attempts to acquire a lock with the given name and TTL. It returns a RedisMutex if successful.
func (m *redisMutex) Release(ctx context.Context) error {
	if _, err := m.mutex.UnlockContext(ctx); err != nil {
		return WrapError(ErrorLockNotReleased, "failed to release lock")
	}
	return nil
}

// RefreshLock attempts to extend the expiration of the lock. It extends it by the original TTL if it was set, or to the default expiration time if not specified.
func (m *redisMutex) Refresh(ctx context.Context) error {
	ok, err := m.mutex.ExtendContext(ctx)
	if err != nil {
		return WrapError(ErrorLockNotRefreshed, "failed to refresh lock")
	}
	if !ok {
		return WrapError(ErrorLockNotRefreshed, "lock not refreshed")
	}
	return nil
}

// assert that RedisLockBackend implements LockBackend
var _ LockBackend = (*RedisLockBackend)(nil)

// RedisMutex implements Mutex
var _ Lock = (*redisMutex)(nil)
