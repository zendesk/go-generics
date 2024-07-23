#PACKAGE := "github.com/zendesk/lockbox-shared-lib"
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
	go test -v -timeout 45m ./...

.PHONY: test-fuzz
test-fuzz:
	./scripts/run_fuzz_tests.sh 30

.PHONY: unit-with-coverage
test-unit-with-coverage:
	LOCAL_DEV=true go test -v -timeout 45m ./... -coverprofile cover.out

#make test-one TEST=YourTestName
.PHONY: test-one
test-one:
	go clean -testcache
	go test -v -timeout 45m ./... -run ^$(TEST)$
