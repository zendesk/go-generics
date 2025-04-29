module github.com/zendesk/go-generics/cache

go 1.23.0

require (
	github.com/jellydator/ttlcache/v3 v3.2.0
	github.com/redis/go-redis/v9 v9.7.3

	// Versions of go-generics are dynamically updated at release to reference the current version. This means all
	// dependencies across modules in go-generics depend on the same version
	github.com/zendesk/go-generics/encryption v1.4.23
	github.com/zendesk/go-generics/serialize v1.4.23
	github.com/zendesk/go-generics/test v1.4.23
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
)
