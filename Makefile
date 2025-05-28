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
	go test -v -timeout 30m -tags=test ./cache
	go test -v -timeout 30m -tags=test ./concurrency
	go test -v -timeout 30m -tags=test ./datastructures
	go test -v -timeout 30m -tags=test ./encryption
	go test -v -timeout 30m -tags=test ./functions
	go test -v -timeout 30m -tags=test ./ratelimit
	go test -v -timeout 30m -tags=test ./serialize
	go test -v -timeout 1m -tags=test ./test

.PHONY: test-fuzz
test-fuzz:
	./scripts/run_fuzz_tests.sh 30

.PHONY: unit-with-coverage
test-coverage:
	make redis-up
	go clean -testcache
	go test -v -timeout 30m -tags=test ./cache -coverprofile cache.out
	go test -v -timeout 30m -tags=test ./datastructures -coverprofile datastructures.out
	go test -v -timeout 30m -tags=test ./encryption -coverprofile encryption.out
	go test -v -timeout 30m -tags=test ./functions -coverprofile functions.out
	go test -v -timeout 30m -tags=test ./ratelimit -coverprofile ratelimit.out
	go test -v -timeout 30m -tags=test ./serialize -coverprofile serialize.out
	go test -v -timeout 30m -tags=test,redis ./concurrency -coverprofile concurrency.out
	go test -v -timeout 1m -tags=test ./test -coverprofile test.out

# Detect container runtime (docker or podman)
CONTAINER_RUNTIME := $(shell command -v docker 2> /dev/null)
ifndef CONTAINER_RUNTIME
    CONTAINER_RUNTIME := $(shell command -v podman 2> /dev/null)
    ifeq ($(CONTAINER_RUNTIME),)
        $(error Neither docker nor podman found. Please install one of them.)
    endif
    # Use podman-compose if available, otherwise use podman compose
    COMPOSE_CMD := $(shell command -v podman-compose 2> /dev/null || echo "$(CONTAINER_RUNTIME) compose")
else
    COMPOSE_CMD := $(CONTAINER_RUNTIME) compose
endif

.PHONY: redis-up
redis-up:
	$(COMPOSE_CMD) up -d redis
	@echo "Waiting for Redis to be ready..."
	@timeout 30 sh -c 'until $(CONTAINER_RUNTIME) exec $$($(COMPOSE_CMD) ps -q redis) redis-cli ping | grep -q PONG; do sleep 1; done' || (echo "Redis failed to start" && exit 1)
	@echo "Redis is ready!"

.PHONY: redis-down
redis-down:
	$(COMPOSE_CMD) down

.PHONY: redis-logs
redis-logs:
	$(COMPOSE_CMD) logs -f redis

.PHONY: test-redis
test-redis: redis-up
	@echo "Running redis tests..."
	go test -v -timeout 30m -tags=redis ./concurrency
	@echo "Redis tests completed."

.PHONY: test-redis-cleanup
test-redis-cleanup: test-redis redis-down

.PHONY: tidy
tidy:
	cd ./cache && go mod tidy
	cd ./concurrency && go mod tidy
	cd ./datastructures && go mod tidy
	cd ./encryption && go mod tidy
	cd ./functions && go mod tidy
	cd ./ratelimit && go mod tidy
	cd ./serialize && go mod tidy
	cd ./test && go mod tidy



