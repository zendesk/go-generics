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

// buildKey hashes the key and prepends the backend-level and operation-level
// prefixes. The backend-level prefix (from the CacheBackendOption WithPrefix)
// comes first and is fixed at cache construction time; the operation-level
// prefix (from the OperationOption WithKeyPrefix) comes next and may vary per
// call.
//
// Final layout: <backendPrefix><operationPrefix><hash(key)>
//
// Both Redis and in-memory backends use this to produce the internal storage key.
func buildKey(key any, backendPrefix string, opts ...OperationOption) string {
	cfg := resolveOperationOpts(opts...)
	return backendPrefix + cfg.prefix + HashAny(key)
}
