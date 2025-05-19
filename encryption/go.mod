module github.com/zendesk/go-generics/encryption

go 1.23.0

require (
	// Versions of go-generics are dynamically updated at release to reference the current version. This means all
	// dependencies across modules in go-generics depend on the same version
	github.com/zendesk/go-generics/serialize v1.5.1
	golang.org/x/crypto v0.35.0
)
