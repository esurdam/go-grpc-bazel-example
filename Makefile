.PHONY: build clean gazelle link fmt test coverage
.DEFAULT_GOAL = build

build:
	bash ci/build-service.sh

coverage:
	bash ci/coverage.sh

fmt:
	bash ci/build-fmt.sh

gazelle:
	go mod tidy
	bazel run //:gazelle fix
	bazel run //:gazelle -- update-repos -from_file=go.mod -prune=true

link:
	bash ci/link.sh

test:
	bash ci/test.sh

