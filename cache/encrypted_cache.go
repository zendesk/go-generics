package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/zendesk/go-generics/encryption"
)

// Item represents a single item in the cache, and if it is expired
type Item[K comparable, V any] interface {
	Value() V
	IsExpired() bool
}

type EncryptedCache[K comparable, V any] struct {
	ed      *encryption.EncryptorDecryptor[V]
	backend CacheBackendAdapter[K, []byte]
	cfg     cacheBackendCfg[K, V]
}

// NewEncryptedCache provides a new EncryptedCache with default configurations
// that balance performance and security.
//
// salt must be at least encryption.MinSaltLength (16) bytes and iterations
// must be at least encryption.MinIterations (100 000); otherwise construction
// fails. Legacy persisted data that predates these minimums can opt out by
// passing cache.WithEncryptionOptions(encryption.WithAllowLegacyParameters()).
func NewEncryptedCache[K comparable, V any](backend CacheBackendAdapter[K, []byte], salt []byte, iterations int, opts ...CacheBackendOption[K, V]) (*EncryptedCache[K, V], error) {
	return NewEncryptedCacheWithIterations[K, V](backend, salt, iterations, opts...)
}

// NewEncryptedCacheWithIterations allows the user to specify a number of
// iterations which influences work factor for the cache. iterations must be at
// least encryption.MinIterations (100 000) unless the caller opts out via
// cache.WithEncryptionOptions(encryption.WithAllowLegacyParameters()).
func NewEncryptedCacheWithIterations[K comparable, V any](backend CacheBackendAdapter[K, []byte], salt []byte, iterations int, opts ...CacheBackendOption[K, V]) (*EncryptedCache[K, V], error) {
	cfg := setBackendOpts(opts...)

	ed, err := encryption.New[V](salt, iterations, cfg.encryptionOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating new encryptor: %w", err)
	}
	e := &EncryptedCache[K, V]{
		ed:      ed,
		backend: backend,
		cfg:     cfg,
	}

	return e, nil
}

// NewEncryptedCacheWithPassword creates a new EncryptedCache with a caller-supplied password.
// A fresh random nonce is generated for each encryption and prepended to the ciphertext.
func NewEncryptedCacheWithPassword[K comparable, V any](backend CacheBackendAdapter[K, []byte], password, salt []byte, iterations int, opts ...CacheBackendOption[K, V]) (*EncryptedCache[K, V], error) {
	cfg := setBackendOpts(opts...)

	ed, err := encryption.NewWithPassword[V](password, salt, iterations, cfg.encryptionOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating new encryptor: %w", err)
	}

	e := &EncryptedCache[K, V]{
		ed:      ed,
		backend: backend,
		cfg:     cfg,
	}

	return e, nil
}

// Deprecated: NewEncryptedCacheWithPasswordNonce is replaced by NewEncryptedCacheWithPassword.
// Nonces are now generated per-encryption. This alias is kept for API compatibility.
func NewEncryptedCacheWithPasswordNonce[K comparable, V any](backend CacheBackendAdapter[K, []byte], password, salt []byte, iterations int, opts ...CacheBackendOption[K, V]) (*EncryptedCache[K, V], error) {
	return NewEncryptedCacheWithPassword[K, V](backend, password, salt, iterations, opts...)
}

func (c *EncryptedCache[K, V]) Get(key K, opts ...OperationOption) (v V, wasFound bool, err error) {
	encryptedBytes, _, err := c.backend.Get(key, opts...)
	if err != nil {
		return v, false, err
	}

	// If there are no encrypted bytes found, return nil without error. Nothing exists in cache for key.
	if encryptedBytes == nil {
		return v, false, nil
	}

	v, err = c.ed.Decrypt(encryptedBytes)
	if err != nil {
		return v, false, fmt.Errorf("error decrypting bytes returned from cache: %w", err)
	}

	return v, true, nil
}

func (c *EncryptedCache[K, V]) Delete(key K, opts ...OperationOption) error {
	return c.backend.Delete(key, opts...)
}

func (c *EncryptedCache[K, V]) Set(key K, value V, opts ...OperationOption) error {
	// Encrypt the binary value
	encryptedBytes, err := c.ed.Encrypt(value)
	if err != nil {
		return fmt.Errorf("error encrypting value: %w", err)
	}

	// Store the encrypted value in the cache
	err = c.backend.Set(key, encryptedBytes, opts...)
	if err == nil || c.cfg.ignoreCacheSetErrors {
		return nil
	}

	// Err on set, so return CacheSetErr
	err = CacheSetError{Message: err.Error()}
	return err
}

func (c *EncryptedCache[K, V]) Purge() error {
	return c.backend.Purge()
}

func (c *EncryptedCache[K, V]) GetOrSet(key K, orSet func() (V, error), opts ...OperationOption) (val V, wasFoundInCache bool, err error) {
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

func (c *EncryptedCache[K, V]) encodeToBinary(value V) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(value); err != nil {
		return nil, fmt.Errorf("error encoding value to binary: %w", err)
	}

	return buf.Bytes(), nil
}
