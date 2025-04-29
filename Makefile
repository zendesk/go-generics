#PACKAGE := "github.com/zendesk/go-generics"
#export GOPROXY ?= https://$(ARTIFACTORY_USERNAME):$(ARTIFACTORY_API_KEY)@zdrepo.jfrog.io/zdrepo/api/go/zen-go
#export GOSUMDB := off
AWS_PROFILE := sandbox1

default: build

.PHONY: build
build:
	go build $(PACKAGE) ./...

.PHONY: ensure_deps
ensure_deps:
	go mod vendor
	go mod tidy

.PHONY: fmt
	gofmt -w `find . -name '*.go'`

# make test TEST=MyTestName
.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit:
	go clean -testcache
	go test -v -timeout 20m -tags=test ./cache
	go test -v -timeout 20m -tags=test ./datastructures
	go test -v -timeout 20m -tags=test ./encryption
	go test -v -timeout 20m -tags=test ./functions
	go test -v -timeout 20m -tags=test ./ratelimit
	go test -v -timeout 20m -tags=test ./serialize
	go test -v -timeout 1m -tags=test ./test

.PHONY: test-fuzz
test-fuzz:
	./scripts/run_fuzz_tests.sh 30

.PHONY: unit-with-coverage
test-unit-with-coverage:
	go clean -testcache
	go test -v -timeout 20m -tags=test ./cache -coverprofile cache.out
	go test -v -timeout 20m -tags=test ./datastructures -coverprofile datastructures.out
	go test -v -timeout 20m -tags=test ./encryption -coverprofile encryption.out
	go test -v -timeout 20m -tags=test ./functions -coverprofile functions.out
	go test -v -timeout 20m -tags=test ./ratelimit -coverprofile ratelimit.out
	go test -v -timeout 20m -tags=test ./serialize -coverprofile serialize.out
