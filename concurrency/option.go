package concurrency

// LockBackendOption configures a lock backend at construction time.
type LockBackendOption func(*lockBackendCfg)

type lockBackendCfg struct {
	prefix string
}

// WithPrefix prepends the given prefix to every lock name obtained through the backend.
// This is useful when the underlying store (e.g. AWS ElastiCache) enforces ACLs on key
// patterns — every lock key will start with the configured prefix so it satisfies a
// pattern like `service-name:*`.
func WithPrefix(prefix string) LockBackendOption {
	return func(cfg *lockBackendCfg) {
		cfg.prefix = prefix
	}
}

func resolveLockBackendOpts(opts ...LockBackendOption) lockBackendCfg {
	var cfg lockBackendCfg
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
