package cache

import (
	"crypto/sha256"
	"fmt"
)

// Redis keys must be strings so always use the hash
func HashAny(obj any) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

// buildKey hashes the key and prepends an optional prefix from OperationOptions.
// Both Redis and in-memory backends use this to produce the internal storage key.
func buildKey(key any, opts ...OperationOption) string {
	cfg := resolveOperationOpts(opts...)
	return cfg.prefix + HashAny(key)
}
