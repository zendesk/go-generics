## Contributing to go-generics

## Repo structure

Go-generics is a multi-module repository that exports the following go modules:

- github.com/zendesk/go-generics/cache
- github.com/zendesk/go-generics/datastructures
- github.com/zendesk/go-generics/encryption
- github.com/zendesk/go-generics/functions
- github.com/zendesk/go-generics/ratelimit
- github.com/zendesk/go-generics/serialize
- github.com/zendesk/go-generics/test

Each new version of go-generics publishes a new version for each of these modules, and dependencies between modules
are within a single larger go-generics version. 

**This means that github.com/zendesk/go-generics/cache@v1.0.1 will depend on github.com/zendesk/go-generics/serialize@1.0.1**

This provides an advantage that locally you may develop new changes and reference changes across-module without having to release
one module to import its changes into another. This is achieved by the included `go.work` file. Changes may be made across modules,
tested, and published together.


## Why go through this trouble?

Publishing these modules independently keeps transitive dependencies minimal for consumers. For instance, the cache module has a dependency on Redis; however, 
a user of the functions package shouldn't have to import this dependency as an indirect dependency by adopting the functions module. This could be 
achieved (to some extent) by leveraging custom interfaces (which we do in most cases), but there are some attached usability tradeoffs as a result. Publishing individual
modules provides the best overall user experience and minimizes unnecessary indirect imports (which must be security patched) for customers.

## Build + Release

During the build process the `internal/build.go` file is executed. This script updates all module's go.mod files to reference the new version, then 
generates valid checksums for each module and writes them into dependent go.sum files. This process is complex, but enables all new versions of go-generics to be published simultaneously 
while referencing their sibling version. 


## Local Development

The `go.work` file is committed here to simplify local development. You may modify any individual module, and test it individually, or with a group of modules. To run
the whole test suite, run `make test`. To tidy all modules, `make tidy`. 