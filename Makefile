.PHONY: build clean gazelle link fmt

.DEFAULT_GOAL = build

build:
	make clean

fmt:
	bash ci/build-fmt.sh

gazelle:
	go mod tidy
	bazel run //:gazelle update
	bazel run //:gazelle -- update-repos -from_file=go.mod -prune=true

link:
	bazel run //pb/helloworld:helloworld_go_proto_link

