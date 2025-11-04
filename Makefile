.PHONY: build clean gazelle help link fmt test coverage upgrade _godeps push deploy tidy unlink _tidy
.DEFAULT_GOAL = help
VERSION ?= $(shell openssl rand -base64 8 |md5 |head -c8)

build: ## Build services
	bazel build :build_all

coverage: ## Generate coverage report
	bash ci/coverage.sh

deploy: ## Deploy services to k8s
	bash ci/deploy.sh

fmt: ## Run build-fmt
	bash ci/build-fmt.sh

gazelle: ## Run gazelle update
	bazel run //:gazelle -- update

link: ## Link bazel build proto to local
	bash ci/link.sh

push: ## Push all 'push' to registry
	bash ci/push-service.sh

_tidy:
	bazel run @rules_go//go -- mod tidy

tidy: link _tidy unlink ## Run go mod tidy
	
test: ## Run test
	bash ci/test.sh

unlink: ## Unlink bazel compiled grpc files from local workspace
	bash ci/unlink.sh

_godeps:
	go get -u ./...

upgrade: _godeps gazelle fmt ## Upgrade all deps

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
