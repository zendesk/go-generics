##@ Build

IMAGE_REPOSITORY = 713408432298.dkr.ecr.us-west-2.amazonaws.com/prod/zendesk/go-generics

.DEFAULT_GOAL := help

.PHONY: build-image push-image update-digest retrieve-digest
build-image: ## Build docker image
	docker build -t $(IMAGE_REPOSITORY):$(TAG) .

push-image: ## Push docker image
	docker push $(IMAGE_REPOSITORY):$(TAG)

##@ Help

.PHONY: help
help:  ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
