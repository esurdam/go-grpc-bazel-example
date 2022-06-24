.PHONY: build clean gazelle link fmt test coverage
.DEFAULT_GOAL = help

build: ## Build services
	bash ci/build-service.sh

coverage: ## Generate coverage report
	bash ci/coverage.sh

fmt: ## Run build-fmt
	bash ci/build-fmt.sh

gazelle: ## Run go mod and gazelle
	go mod tidy
	@bazel run //:gazelle fix
	@bazel run //:gazelle -- update-repos -from_file=go.mod -prune=true -build_file_proto_mode=disable -to_macro go.bzl%go_deps

link: ## Link bazel build proto to local
	bash ci/link.sh

test: ## Run test
	bash ci/test.sh

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
