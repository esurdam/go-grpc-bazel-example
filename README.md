# go-grpc-bazel-example

This repo is an example monorepo which utilizes grpc+bazel.

If proto implementation will be shared across services, implementation should reside in `pkg/`. Otherwise,
implementation can be included in `services/{{.packageName}}/pkg/`.

In this case, `helloworld` will be implemented in `pkg`.

- [Generating BUILD.bazel files](#generating-build-files)
- [Generating proto files](#generating-proto-files-(development))
- [Test locally](#test-repo)
- [Running locally](#running-service-locally)
- [Deployment](#deployment)

## Layout

```
ci          # contains ci/automation scripts
cmd         # command line tool entrypoints
pb          # contains all proto definitions and gen output
pkg         # contains proto implementations
services    # entrypoints for kubernetes defined microservices
tools       # tool versioning
```

## Requirements

`Go` and `Bazel` are the only two requirements.

[Install Bazel](https://docs.bazel.build/versions/master/install.html)

## Generating BUILD files

Run `make gazelle` to generate/update BUILD files (which include test and binaries).

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

## Test repo

To run all tests:
```bash
make test
```

Test individual package:
```bash
bazel test --features race \
  --verbose_failures \
  --test_output=errors \
  --action_env=CI=true \
  //pkg/helloworld/server:go_default_test
```

Tests can also be aggregated into test groups to be tested at once.

## Running service locally

```bash
bazel run //services/helloworld:helloworld
```

## Deployment

CI checks for formatting; ensure formatting with `make fmt`

```bash
make fmt
```