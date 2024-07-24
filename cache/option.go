package cache

type CacheCfg[K comparable, V any] struct {
	failThroughCache     CacheBackendAdapter[K, V]
	capacity             uint64
	ignoreCacheSetErrors bool
}

type CacheOption[K comparable, V any] func(cfg CacheCfg[K, V]) CacheCfg[K, V]

func SetOpts[K comparable, V any](opts ...CacheOption[K, V]) CacheCfg[K, V] {
	cfg := CacheCfg[K, V]{}
	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return cfg
}

// WithFailThroughCache will set a fail-through cache. If the primary cache does not have the value, it will attempt to get it from the fail-through cache, and
// update the primary cache if the value is found.
func WithFailThroughCache[K comparable, V any](failThrough CacheBackendAdapter[K, V]) CacheOption[K, V] {
	return func(cfg CacheCfg[K, V]) CacheCfg[K, V] {
		cfg.failThroughCache = failThrough
		return cfg
	}
}

// WithCapacity will ensure the cache cannot exceed the # of objects as defined here.
func WithCapacity[K comparable, V any](maxObjects uint64) CacheOption[K, V] {
	return func(cfg CacheCfg[K, V]) CacheCfg[K, V] {
		cfg.capacity = maxObjects
		return cfg
	}
}

// IgnoreCacheSetErrors will ignore any errors that occur when setting a value in the cache and will not return an error to the client
func IgnoreCacheSetErrors[K comparable, V any]() CacheOption[K, V] {
	return func(cfg CacheCfg[K, V]) CacheCfg[K, V] {
		cfg.ignoreCacheSetErrors = true
		return cfg
	}
}
