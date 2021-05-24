# go-grpc-bazel-example

This repo is an example monorepo which utilized grpc+bazel.

If proto implementation will be shared across services, implementation should reside in `pkg/`. Otherwise,
implementation can be included in `services/{{.packageName}}/pkg/`.

In this case, `helloworld` will be implemented in `pkg`.

(Running locally)[#running-locally]
(Deployment)[#deployment]

## Layout

```
ci          # contains ci/automation scripts
cmd         # command line tool entrypoints
pb          # contains all proto definitions and gen output
pkg         # contains proto implementations
services    # entrypoints for kubernetes defined microservices
tools       # tool versioning
```

## Generating BUILD files

Run `make gazelle` to generate/update BUILD files.

This also updates the WORKSPACE with required deps.

BUILD.bazel files located in pb directory will contain grpc rules.

## Generating proto files (development)

Run `*_go_proto_link` rule to generate `.pb.go` files and add them to the proto directory.

Generated files don't necessarily need to be checked in to repo.
In this example, generated files are checked in. They are only necessary for local development.
Otherwise, Bazel will handle generating the pb file during build.

```bash
bazel run //pb/helloworld:helloworld_go_proto_link
```

## Running service locally

```bash
bazel run //services/helloworld:helloworld
```

## Deployment

CI checks for formatting; ensure formatting with `make fmt`

```bash
make fmt
```