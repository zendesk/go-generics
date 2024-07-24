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
