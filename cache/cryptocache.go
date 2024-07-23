package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/zendesk/lockbox-shared-lib/cryptoutil"
)

// Item represents a single item in the cache, and if it is expired
type Item[K comparable, V any] interface {
	Value() V
	IsExpired() bool
}

type CryptoCacher[K comparable, V any] struct {
	ed      *cryptoutil.EncryptorDecryptor
	backend CacheBackendAdapter[K, []byte]
	cfg     CacheCfg[K, V]
}

// New provides a new CryptoCacher with default configurations that provides a balance of performance and security. Selecting a # of iterations that is difficult to guess
// and high enough to make encrypt / decrypt durations sufficiently long to deter brute force attack is recommended. Minimum 2048 iterations is recommended
func New[K comparable, V any](backend CacheBackendAdapter[K, []byte], iterations int, opts ...CacheOption[K, V]) (*CryptoCacher[K, V], error) {
	return NewWithIterations[K, V](backend, iterations, opts...)
}

// NewWithIterations allows the user to specify a number of iterations which influences work factor for the cache.
func NewWithIterations[K comparable, V any](backend CacheBackendAdapter[K, []byte], iterations int, opts ...CacheOption[K, V]) (*CryptoCacher[K, V], error) {
	cfg := SetOpts(opts...)

	ed, err := cryptoutil.NewWithIterations(iterations)
	if err != nil {
		return nil, fmt.Errorf("error creating new encryptor: %w", err)
	}
	e := &CryptoCacher[K, V]{
		ed:      ed,
		backend: backend,
		cfg:     cfg,
	}

	return e, nil
}

func NewWithPasswordNonceIterations[K comparable, V any](backend CacheBackendAdapter[K, []byte], password []byte, nonce []byte, iterations int, opts ...CacheOption[K, V]) (*CryptoCacher[K, V], error) {
	cfg := SetOpts(opts...)

	ed, err := cryptoutil.NewWithPasswordNonceIterations(password, nonce, iterations)
	if err != nil {
		return nil, fmt.Errorf("error creating new encryptor: %w", err)
	}

	e := &CryptoCacher[K, V]{
		ed:      ed,
		backend: backend,
		cfg:     cfg,
	}

	return e, nil
}

func (c *CryptoCacher[K, V]) Get(key K) (v V, wasFound bool, err error) {
	encryptedBytes, _, err := c.backend.Get(key)
	if err != nil {
		return v, false, err
	}

	// If there are no encrypted bytes found, return nil without error. Nothing exists in cache for key.
	if encryptedBytes == nil {
		return v, false, nil
	}

	decryptedBytes, err := c.decrypt(encryptedBytes)
	if err != nil {
		return v, false, fmt.Errorf("error decrypting bytes returned from cache: %w", err)
	}

	v, err = c.decodeBinary(decryptedBytes)
	if err != nil {
		return v, false, fmt.Errorf("error decoding binary: %w", err)
	}
	return v, true, nil
}

func (c *CryptoCacher[K, V]) Delete(key K) error {
	return c.backend.Delete(key)
}

func (c *CryptoCacher[K, V]) Set(key K, value V) error {
	// Convert the value to binary
	binaryValue, err := c.encodeToBinary(value)
	if err != nil {
		return fmt.Errorf("error encoding value to binary: %w", err)
	}

	// Encrypt the binary value
	encryptedBytes, err := c.encrypt(binaryValue)
	if err != nil {
		return fmt.Errorf("error encrypting value: %w", err)
	}

	// Store the encrypted value in the cache
	err = c.backend.Set(key, encryptedBytes)
	if err == nil || c.cfg.ignoreCacheSetErrors {
		return nil
	}

	// Err on set, so return CacheSetErr
	err = CacheSetError{Message: err.Error()}
	return err
}

func (c *CryptoCacher[K, V]) Purge() error {
	return c.backend.Purge()
}

func (c *CryptoCacher[K, V]) GetOrSet(key K, orSet func() (V, error)) (val V, wasFoundInCache bool, err error) {
	item, wasFound, err := c.Get(key)
	if wasFound {
		return item, wasFound, err
	}

	val, err = orSet()
	if err != nil {
		return val, false, err
	}

	err = c.Set(key, val)
	if err == nil || c.cfg.ignoreCacheSetErrors {
		return val, false, nil
	}

	return val, false, err
}

func (c *CryptoCacher[K, V]) encrypt(value []byte) ([]byte, error) {
	return c.ed.Encrypt(value)
}

func (c *CryptoCacher[K, V]) decrypt(ciphertext []byte) ([]byte, error) {
	return c.ed.Decrypt(ciphertext)
}

func (c *CryptoCacher[K, V]) encodeToBinary(value V) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(value); err != nil {
		return nil, fmt.Errorf("error encoding value to binary: %w", err)
	}

	return buf.Bytes(), nil
}

func (c *CryptoCacher[K, V]) decodeBinary(b []byte) (V, error) {
	r := bytes.NewReader(b)
	enc := gob.NewDecoder(r)

	var v V
	if err := enc.Decode(&v); err != nil {
		return v, fmt.Errorf("error decoding binary: %w", err)
	}
	return v, nil
}
