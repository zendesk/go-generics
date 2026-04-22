package cache

import "github.com/zendesk/go-generics/encryption"

type cacheBackendCfg[K comparable, V any] struct {
	failThroughCache     CacheBackendAdapter[K, V]
	observer             CacheObserver[K]
	capacity             uint64
	ignoreCacheSetErrors bool
	encryptionOpts       []encryption.Option
}

type CacheBackendOption[K comparable, V any] func(cfg cacheBackendCfg[K, V]) cacheBackendCfg[K, V]

func setBackendOpts[K comparable, V any](opts ...CacheBackendOption[K, V]) cacheBackendCfg[K, V] {
	cfg := cacheBackendCfg[K, V]{
		capacity: ^uint64(0), // default capacity to maximum size,
	}

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return cfg
}

// WithFailThroughCache will set a fail-through cache. If the primary cache does not have the value, it will attempt to get it from the fail-through cache, and
// update the primary cache if the value is found.
func WithFailThroughCache[K comparable, V any](failThrough CacheBackendAdapter[K, V]) CacheBackendOption[K, V] {
	return func(cfg cacheBackendCfg[K, V]) cacheBackendCfg[K, V] {
		cfg.failThroughCache = failThrough
		return cfg
	}
}

// WithCapacity will ensure the cache cannot exceed the # of objects as defined here.
func WithCapacity[K comparable, V any](maxObjects uint64) CacheBackendOption[K, V] {
	return func(cfg cacheBackendCfg[K, V]) cacheBackendCfg[K, V] {
		cfg.capacity = maxObjects
		return cfg
	}
}

// IgnoreCacheSetErrors will ignore any errors that occur when setting a value in the cache and will not return an error to the client
func IgnoreCacheSetErrors[K comparable, V any]() CacheBackendOption[K, V] {
	return func(cfg cacheBackendCfg[K, V]) cacheBackendCfg[K, V] {
		cfg.ignoreCacheSetErrors = true
		return cfg
	}
}

// WithEncryptionOptions forwards the given encryption.Option values to the
// underlying EncryptorDecryptor used by NewEncryptedCache* constructors.
// It has no effect on non-encrypted caches.
func WithEncryptionOptions[K comparable, V any](opts ...encryption.Option) CacheBackendOption[K, V] {
	return func(cfg cacheBackendCfg[K, V]) cacheBackendCfg[K, V] {
		cfg.encryptionOpts = append(cfg.encryptionOpts, opts...)
		return cfg
	}
}

type cacheCfg[K comparable, V any] struct {
	observer CacheObserver[K]
}

type CacheOption[K comparable, V any] func(cfg cacheCfg[K, V]) cacheCfg[K, V]

func WithObserver[K comparable, V any](observer CacheObserver[K]) CacheOption[K, V] {
	return func(cfg cacheCfg[K, V]) cacheCfg[K, V] {
		cfg.observer = observer
		return cfg
	}
}
func setCacheOpts[K comparable, V any](opts ...CacheOption[K, V]) cacheCfg[K, V] {
	cfg := cacheCfg[K, V]{}

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return cfg
}

// OperationOption configures the behavior of individual cache operations (Get, Set, Delete, etc.).
type OperationOption func(*operationCfg)

type operationCfg struct {
	prefix string
}

// WithPrefix prepends the given prefix to the hashed cache key for this operation.
// Both Redis and in-memory backends hash keys via HashAny; the prefix is prepended to that hash.
func WithPrefix(prefix string) OperationOption {
	return func(cfg *operationCfg) {
		cfg.prefix = prefix
	}
}

func resolveOperationOpts(opts ...OperationOption) operationCfg {
	var cfg operationCfg
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
