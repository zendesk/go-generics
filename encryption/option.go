package encryption

// Option configures an EncryptorDecryptor constructor.
type Option func(*options)

type options struct {
	allowLegacyParameters bool
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
