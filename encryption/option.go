package encryption

// Option configures an EncryptorDecryptor constructor.
type Option func(*options)

type options struct {
	allowLegacyParameters bool
	legacyNonce           []byte
	legacyDecryptNonce    []byte
}

func resolveOptions(opts ...Option) options {
	var cfg options
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// WithAllowLegacyParameters disables validation of the minimum salt length
// (MinSaltLength) and minimum iteration count (MinIterations).
//
// Use only when an existing production system relies on an already-persisted
// salt or iteration count that pre-dates the current minimums, and rotating
// them requires a coordinated data re-encryption that cannot ship with the
// upgrade. New deployments must not use this option.
func WithAllowLegacyParameters() Option {
	return func(o *options) {
		o.allowLegacyParameters = true
	}
}

// WithLegacyNonce selects the pre-2026-03 ciphertext wire format: the nonce is
// fixed at construction time (the one provided here) and is NOT prepended to
// ciphertext. Encrypt uses that nonce; Decrypt can read both the legacy and
// per-message-nonce formats to support a staged rollout to safe writes.
//
// NOT RECOMMENDED. This exists purely for backwards compatibility with data
// that was encrypted by an older go-generics release (pre-#18), which stored
// ciphertext as [ciphertext|tag] without a nonce prefix. New callers, and any
// caller that can migrate its persisted data, must not use this option — AES-GCM
// is catastrophically broken under nonce reuse (confidentiality AND authenticity
// fail), so a fixed-nonce encryptor should be treated as read-only whenever
// possible. Encrypting many messages under a single key+nonce with this option
// enabled is a security bug; prefer decrypting legacy rows with this option and
// re-encrypting them using a default (per-call random nonce) encryptor.
//
// The nonce must be exactly NonceLength (12) bytes. Passing a different length
// causes the constructor to return ErrInvalidNonce.
func WithLegacyNonce(nonce []byte) Option {
	return func(o *options) {
		o.legacyNonce = nonce
	}
}

// WithLegacyDecryptNonce reads ciphertext written by pre-#18 go-generics while
// keeping Encrypt on the safe, per-message-random-nonce format. Use this for
// migrating persisted ciphertext without continuing to write fixed-nonce data.
// Both formats are authenticated before plaintext is returned.
//
// All readers must support the new format before enabling new writes; older
// versions of go-generics cannot decrypt ciphertext written in this mode.
func WithLegacyDecryptNonce(nonce []byte) Option {
	return func(o *options) {
		o.legacyDecryptNonce = nonce
	}
}
