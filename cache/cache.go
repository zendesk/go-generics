package cache

type Cache[K comparable, V any] interface {
	Get(key K) (v V, wasFound bool, err error)
	Delete(key K) error
	Set(key K, value V) error
	Purge() error
	// Get the value from cache for key K, or setItsValue and return it via func orSet
	GetOrSet(key K, orSet func() (V, error)) (val V, wasFoundInCache bool, err error)
}

// CacheBackendAdapter manges reading/writing/deleting to and from a cache backend. E.G. Memory backend, Redis Backend, etc.
type CacheBackendAdapter[K comparable, V any] interface {
	Get(key K) (value V, wasFound bool, err error)
	Set(key K, value V) error
	Delete(key K) error
	Purge() error // Purges the cache
	// Get the value from cache for key K, or setItsValue and return it via func orSet
	GetOrSet(key K, orSet func() (V, error)) (val V, wasFoundInCache bool, err error)
}

type CacheSetError struct {
	Message string
}

func (cse CacheSetError) Error() string {
	return cse.Message
}

// CacheObserver may be provided to track changes within the cache. These may be be used to push metrics, initiate cache purges, invalidations, etc.
type CacheObserver[K comparable] interface {
	Hit(k K)
	Miss(k K)
	Get(k K)
	Set(k K)
	Delete(k K)
	Purge()
}

func NewCache[K comparable, V any](backend CacheBackendAdapter[K, V], opts ...CacheOption[K, V]) Cache[K, V] {
	cfg := setCacheOpts(opts...)

	cash := &cache[K, V]{
		cfg:     cfg,
		backend: backend,
	}

	if cfg.observer != nil {
		cash.observer = cfg.observer
	}

	return cash
}

type cache[K comparable, V any] struct {
	observer CacheObserver[K]
	backend  CacheBackendAdapter[K, V]
	cfg      cacheCfg[K, V]
}

func (c *cache[K, V]) Get(key K) (V, bool, error) {

	val, fromCache, err := c.backend.Get(key)
	if c.observer != nil {
		c.observer.Get(key)
		if fromCache {
			c.observer.Hit(key)
		} else {
			c.observer.Miss(key)
		}
	}

	return val, fromCache, err
}

func (c *cache[K, V]) Set(key K, val V) error {
	if c.observer != nil {
		c.observer.Set(key)
	}

	return c.backend.Set(key, val)
}

func (c *cache[K, V]) Delete(key K) error {
	if c.observer != nil {
		c.observer.Delete(key)
	}

	return c.backend.Delete(key)
}

func (c *cache[K, V]) Purge() error {
	if c.observer != nil {
		c.observer.Purge()
	}

	return c.backend.Purge()
}

func (c *cache[K, V]) GetOrSet(key K, orSet func() (V, error)) (val V, wasFoundInCache bool, err error) {
	v, fromCache, err := c.backend.GetOrSet(key, orSet)

	if c.observer != nil {
		c.observer.Get(key)
		if fromCache {
			c.observer.Hit(key)
		} else {
			c.observer.Miss(key)
			c.observer.Set(key)
		}
	}

	return v, fromCache, err
}
