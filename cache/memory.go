package cache

import (
	"time"

	"github.com/jellydator/ttlcache/v3"
)

// InMemoryCache adapts ttlCache to integrate it with CacheBackendAdapter.
// Keys are hashed to strings internally (via HashAny), matching Redis behavior.
type InMemoryCache[K comparable, V any] struct {
	cache       *ttlcache.Cache[string, V]
	ttl         time.Duration
	failThrough CacheBackendAdapter[K, V]
	cfg         cacheBackendCfg[K, V]
}

// NewInMemoryCache provisions a new InMemoryCache
func NewInMemoryCache[K comparable, V any](ttl time.Duration, opts ...CacheBackendOption[K, V]) CacheBackendAdapter[K, V] {
	cfg := setBackendOpts(opts...)

	ttlOpts := []ttlcache.Option[string, V]{
		ttlcache.WithTTL[string, V](ttl),
		ttlcache.WithDisableTouchOnHit[string, V](),
	}

	if cfg.capacity > 0 {
		ttlOpts = append(ttlOpts, ttlcache.WithCapacity[string, V](cfg.capacity))
	}

	ttlCache := ttlcache.New[string, V](ttlOpts...)

	cache := InMemoryCache[K, V]{
		cache:       ttlCache,
		ttl:         ttl,
		failThrough: cfg.failThroughCache,
		cfg:         cfg,
	}

	go cache.cache.Start()

	return &cache
}

func (c *InMemoryCache[K, V]) Get(key K, opts ...OperationOption) (V, bool, error) {
	var v V
	if c.cacheIsDisabled() {
		return v, false, nil
	}

	k := buildKey(key, c.cfg.keyPrefix, opts...)
	item := c.cache.Get(k)
	if item != nil {
		return item.Value(), true, nil
	}

	// If fail-through is configured, try fail-through
	if c.failThrough != nil {
		value, wasFound, err := c.failThrough.Get(key, opts...)
		if err != nil {
			return v, wasFound, err
		}

		// Add item to this cache before returning it. Do **NOT** call public set (c.Set) as it will reset the value in the
		// fail-through cache, so it will never expire (if it's a TTL based cache)
		if wasFound {
			_ = c.cache.Set(k, value, c.ttl)
			return value, true, nil
		}
	}

	// If item does not exist, do not return error, just return nil. This matches prior functionality
	return v, false, nil
}

func (c *InMemoryCache[K, V]) Set(key K, val V, opts ...OperationOption) error {
	if c.cacheIsDisabled() {
		return nil
	}

	k := buildKey(key, c.cfg.keyPrefix, opts...)
	_ = c.cache.Set(k, val, c.ttl)

	// if fail-through is enabled, set in fail-thorugh
	if c.failThrough != nil {
		err := c.failThrough.Set(key, val, opts...)
		if err != nil && !c.cfg.ignoreCacheSetErrors {
			// Error on SET, so we wrap it in RedisCacheSetError
			err = CacheSetError{Message: err.Error()}
			return err
		}
	}

	return nil
}

func (c *InMemoryCache[K, V]) Delete(key K, opts ...OperationOption) error {
	k := buildKey(key, c.cfg.keyPrefix, opts...)
	c.cache.Delete(k)

	// Delete from fail through
	if c.failThrough != nil {
		return c.failThrough.Delete(key, opts...)
	}

	return nil
}

func (c *InMemoryCache[K, V]) Purge() error {
	c.cache.DeleteAll()

	// Purge fail-through too
	if c.failThrough != nil {
		return c.failThrough.Purge()
	}
	return nil
}

func (c *InMemoryCache[K, V]) GetOrSet(key K, orSet func() (V, error), opts ...OperationOption) (val V, wasFoundInCache bool, err error) {
	if c.cacheIsDisabled() {
		val, err = orSet()
		return val, false, err
	}

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

	// Error on SET, so we wrap it in CacheSetError
	err = CacheSetError{Message: err.Error()}
	return val, false, err
}

// TTL is zero, or capacity is 0, we should consider this cache disabled
func (c *InMemoryCache[K, V]) cacheIsDisabled() bool {
	return c.ttl == 0 || c.cfg.capacity <= 0
}
