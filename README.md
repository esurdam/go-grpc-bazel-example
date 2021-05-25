# go-grpc-bazel-example

[![test](https://github.com/AdGreetz/go-grpc-bazel-example/actions/workflows/go.yml/badge.svg)](https://github.com/AdGreetz/go-grpc-bazel-example/actions/workflows/go.yml)

This repo is an example monorepo which utilizes grpc+bazel.

If proto implementation will be shared across services, implementation should reside in `pkg/`. Otherwise,
implementation can be included in `services/{{.packageName}}/pkg/`.

In this case, `helloworld` will be implemented in `pkg`.

- [Create new service](#create-a-new-service)
- [Generating BUILD.bazel files](#generating-build-files)
- [Generating proto files](#generating-proto-files-(development))
- [Test locally](#test-repo)
- [Running locally](#running-service-locally)
- [Deployment](#deployment)
- [Useful Links/Resources](#useful-links)

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

## Create a new service

Create proto file and define types/service.
```bash
touch pb/helloworld/helloworld.proto # Add definitions to this file
```

Generate `BUILD.bazel` files which contain proto and library definitions, then run `make link` to add
the generated files locally.
```bash
make gazelle
make link
```

Implement proto service in `pkg`
```bash
mkdir pkg/helloworld/server
touch pkg/helloworld/server/server.go
```

Create service entrypoint in `services`
```bash
mkdir services/helloworld
touch services/helloworld/main.go
```

Define kubernetes service in `ci/services`
```bash
touch ci/services/helloworld.yml
```

Lastly, add the service definition to aggregate rule in `BUILD`.

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

Then we use cURL to send HTTP requests:

```bash
curl -X POST -k http://localhost:8090/v1/greeter -d '{"name": "TestName"}'
```
```json
{"message":"Hello TestName!"}
```

You can view the swagger at [http://localhost:8090/swagger.json](http://localhost:8090/swagger.json)

### Running with Docker locally

Each service should contain a `docker` rule, which builds the binary in a docker image:

```
go_image(
    name = "docker",
    embed = [
        ":helloworld_lib",
    ],
)
```
To run the binary in docker:

```bash
bazel run \
  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
  --cpu=k8 \
  //services/helloworld:docker
```

## Deployment

CI checks for formatting; ensure formatting with `make fmt`

```bash
make fmt
```

### Kubernetes deployment

Each service should contain a `k8s_deploy` rule, which defines the cluster deployment.

This rule builds the binary in Docker, pushes the image to the container registry, 
and deploys the service to the defined Kubernetes cluster. 

```
k8s_deploy(
    name = "k8s",
    images = {
        "services/helloworld:latest": "//services/helloworld:docker",
    },
    template = "//ci/services:helloworld.yaml",
)
```

Each service is expected to have an exported yaml file for configuration exposed in the `ci/services` directory.

### Pushing service to container registry

Used to deploy a service to the container registry.

Each service should contain a `docker_push` rule, which defines the container registry and path.

```
docker_push(
    name = "push",
    image = ":docker",
    registry = "ghcr.io",
    repository = "adgreetz/go-grpc-bazel-example/services/helloworld",
    tag = "$(version)",
)
```

To push an individual service (where version is the container label):

```bash
bazel run \
  --define version="$(openssl rand -base64 8 |md5 |head -c8)" \ 
  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
  --cpu=k8 \
  //services/helloworld:push
```

## Useful Links

GRPC
- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)

Bazelbuild rules
- [golink](https://github.com/nikunjy/golink)
- [rules_docker](https://github.com/bazelbuild/rules_docker)
- [rules_go](https://github.com/bazelbuild/rules_go)
- [rules_k8s](https://github.com/bazelbuild/rules_k8s)
- [rules_proto](https://github.com/bazelbuild/rules_proto)