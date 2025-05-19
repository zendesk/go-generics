module github.com/zendesk/go-generics/ratelimit

go 1.23.0

require (
	github.com/go-redis/redis_rate/v10 v10.0.1
	github.com/redis/go-redis/v9 v9.7.3

	// Versions of go-generics are dynamically updated at release to reference the current version. This means all
	// dependencies across modules in go-generics depend on the same version
	github.com/zendesk/go-generics/test v1.5.1

)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/go-cmp v0.7.0 // indirect
)
