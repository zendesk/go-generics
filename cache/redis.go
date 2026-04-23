package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zendesk/go-generics/serialize"
)

type Redis interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	FlushDB(ctx context.Context) *redis.StatusCmd
}

// Todo: Implement WithFailThroughCache && Capacity options. THESE ARE NOT YET IMPLEMENTED
type RedisCache[K comparable, V any] struct {
	client Redis
	ctx    context.Context
	ttl    time.Duration
	cfg    cacheBackendCfg[K, V]
}

// NewRedisCache provisions a new RedisCache
func NewRedisCache[K comparable, V any](ctx context.Context, client Redis, ttl time.Duration, opts ...CacheBackendOption[K, V]) CacheBackendAdapter[K, V] {
	cfg := setBackendOpts(opts...)

	cache := RedisCache[K, V]{
		ctx:    ctx,
		ttl:    ttl,
		client: client,
		cfg:    cfg,
	}

	return &cache
}

func (c *RedisCache[K, V]) Get(key K, opts ...OperationOption) (V, bool, error) {
	var v V
	bytes, err := c.client.Get(c.ctx, buildKey(key, c.cfg.keyPrefix, opts...)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return v, false, nil
		}
		return v, false, err
	}

	result, err := serialize.NewSerializer[V]().FromBytes(bytes).ToDynamicType(serialize.Reflect, v)
	if err != nil {
		return v, false, err
	}

	v, ok := result.(V)
	if !ok {
		return v, false, fmt.Errorf("failed to convert result to type. result: %+v, converted: %+v", result, v)
	}
	return v, true, nil
}

func (c *RedisCache[K, V]) Set(key K, val V, opts ...OperationOption) error {
	bytes, err := serialize.NewSerializer[V]().FromDynamicType(val).ToBytes()
	if err != nil {
		return err
	}
	err = c.client.Set(c.ctx, buildKey(key, c.cfg.keyPrefix, opts...), bytes, c.ttl).Err()

	if err == nil || c.cfg.ignoreCacheSetErrors {
		return nil
	}

	// Error on SET, so we wrap it in CacheSetError
	err = CacheSetError{Message: err.Error()}
	return err
}

func (c *RedisCache[K, V]) Delete(key K, opts ...OperationOption) error {
	return c.client.Del(c.ctx, buildKey(key, c.cfg.keyPrefix, opts...)).Err()
}

func (c *RedisCache[K, V]) Purge() error {
	return c.client.FlushDB(c.ctx).Err()
}

// GetOrSet gets the value from cache for key K, or sets it value and return it via func orSet.
func (c *RedisCache[K, V]) GetOrSet(key K, orSet func() (V, error), opts ...OperationOption) (val V, wasFoundInCache bool, err error) {
	item, wasFound, err := c.Get(key, opts...)
	if wasFound {
		return item, wasFound, err
	}

	val, err = orSet()
	if err != nil {
		return val, false, err
	}

	err = c.Set(key, val, opts...)
	if err == nil || c.cfg.ignoreCacheSetErrors {
		return val, false, nil
	}

	return val, false, err
}
