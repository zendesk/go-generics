#PACKAGE := "github.com/zendesk/go-generics"
#export GOPROXY ?= https://$(ARTIFACTORY_USERNAME):$(ARTIFACTORY_API_KEY)@zdrepo.jfrog.io/zdrepo/api/go/zen-go
#export GOSUMDB := off
AWS_PROFILE := sandbox1

default: build

.PHONY: build
build:
	go build $(PACKAGE) ./...

.PHONY: fmt
fmt:
	gofmt -w `find . -name '*.go'`

# make test TEST=MyTestName
.PHONY: test
test: test-unit

.PHONY: test-unit
test-unit:
	go clean -testcache
	go test -v -timeout 30m -tags=test ./...

.PHONY: test-fuzz
test-fuzz:
	./scripts/run_fuzz_tests.sh 30

.PHONY: unit-with-coverage
test-coverage:
	make redis-up
	go clean -testcache
	go test -v -timeout 30m -tags=test ./... -coverprofile cover.out

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
	go mod tidy



