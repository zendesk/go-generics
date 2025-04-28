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
	go test -v -timeout 20m ./cache
	go test -v -timeout 20m ./datastructures
	go test -v -timeout 20m ./encryption
	go test -v -timeout 20m ./functions
	go test -v -timeout 20m ./ratelimit
	go test -v -timeout 20m ./serialize

.PHONY: test-fuzz
test-fuzz:
	./scripts/run_fuzz_tests.sh 30

.PHONY: unit-with-coverage
test-unit-with-coverage:
	go test -v -timeout 45m ./... -coverprofile cover.out

#make test-one TEST=YourTestName
.PHONY: test-one
test-one:
	go clean -testcache
	go test -v -timeout 45m ./... -run ^$(TEST)$

