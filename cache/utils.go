package cache

import (
	"crypto/sha256"
	"fmt"
)

// Redis keys must be strings so always use the hash
func hashAny(obj any) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", obj)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
