# go-grpc-bazel-example

[![test](https://github.com/AdGreetz/go-grpc-bazel-example/actions/workflows/go.yml/badge.svg)](https://github.com/AdGreetz/go-grpc-bazel-example/actions/workflows/go.yml)

## Project Overview

This repository demonstrates a modern monorepo setup for building scalable gRPC microservices in Go, using Bazel for reproducible builds and dependency management. It provides a reference architecture for teams looking to:
- Share protobuf definitions and implementations across services
- Use Bazel for fast, hermetic builds and CI/CD
- Integrate gRPC, REST (via grpc-gateway), and OpenAPI/Swagger documentation
- Deploy services to Kubernetes and container registries

The example service, `helloworld`, showcases the recommended project structure, build rules, and development workflow.

## Features

- **Monorepo structure** for multiple Go microservices
- **gRPC** and **grpc-gateway** for both gRPC and RESTful APIs
- **Bazel** for fast, reproducible builds and dependency management
- **Protobuf** definitions and code generation
- **OpenAPI/Swagger** documentation generation
- **Docker/OCI** image builds and publishing
- **Kubernetes** deployment manifests and scripts
- **CI/CD** with GitHub Actions

## Quickstart

Follow these steps to get the example service running locally:

1. **Install Prerequisites**
   - [Go](https://golang.org/doc/install) (>= 1.23)
   - [Bazel](https://docs.bazel.build/versions/master/install.html) (tested with 8.2.1)
   - [Docker](https://www.docker.com/) (optional, for container builds)

2. **Clone the Repository**
   ```bash
   git clone https://github.com/AdGreetz/go-grpc-bazel-example.git
   cd go-grpc-bazel-example
   ```

3. **Generate Protobuf and Build Files**
   ```bash
   make link      # Generates protobuf Go code for local development (optional)
   ```

4. **Run Tests**
   ```bash
   make test
   ```

5. **Run the Example Service Locally**
   - Generate self-signed TLS certificates:
     ```bash
     mkdir -p ssl
     go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 2048 --host 127.0.0.1,::1,localhost,localhost:4443 --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h -certfile ssl/cert.pem -keyfile ssl/key.pem
     ```
   - Start the service:
     ```bash
     bazel run //services/helloworld:helloworld -- -http-port 4443 -cert $(pwd)/ssl/cert.pem -key $(pwd)/ssl/key.pem
     ```
   - Test with cURL:
     ```bash
     curl -X POST -k https://localhost:4443/v1/greeter -d '{"name": "TestName"}'
     ```

6. **View Swagger/OpenAPI Docs**
   - Open [https://localhost:4443/swagger.json](https://localhost:4443/swagger.json) in your browser.

_For more details, see the sections below._

## Project Layout

```
.github/     # github Action CI configs
ci/          # contains ci/automation scripts
cmd/         # command line tool entrypoints
pb/          # contains all proto definitions and gen output
pkg/         # contains proto implementations
services/    # entrypoints for kubernetes defined microservices
tools/       # tool versioning
BUILD        # root bazel BUILD definitions; aggregates services
WORKSPACE    # bazel workspace rules; external code 
```

## Requirements

- Go (>= 1.23)
- Bazel (tested with 8.2.1)

## Development Workflow

### Creating a New Service

1. Create proto file and define types/service:
   ```bash
   touch pb/helloworld/helloworld.proto # Add definitions to this file
   ```
2. Generate `BUILD.bazel` files and add generated files locally:
   ```bash
   make gazelle
   make link
   ```
3. Implement proto service in `pkg`:
   ```bash
   mkdir -p pkg/helloworld/server
   touch pkg/helloworld/server/server.go
   ```
4. Create service entrypoint in `services`:
   ```bash
   mkdir services/helloworld
   touch services/helloworld/main.go
   ```
5. Define kubernetes service in `ci/services`:
   ```bash
   touch ci/services/helloworld.yaml
   ```
6. Add the service definition to aggregate rule in `BUILD`.

Example service scaffold:
```
project   
│
└───ci
│   └───services
│       │   helloworld.yaml
│
└───pb
│   └───helloworld
│       │   helloworld.proto
│   
└───pkg
│   └───helloworld
│       └───server
│           │   server.go
|
└───services
│   └───helloworld
│       │   main.go
|
```

### Generating BUILD Files

Run `make gazelle` to generate/update BUILD files (which include test and binaries). This also updates the WORKSPACE with required deps. BUILD.bazel files located in pb directory will contain grpc rules.

### Generating Proto Files

Generated files don't necessarily need to be checked in to repo. In this example, generated files are checked in. They are only necessary for local development. Otherwise, Bazel will handle generating the pb file during build.

It is generally a good idea to track changes to pb in repo.

```bash
make link
```

### Testing

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
  //pkg/helloworld/server:server_test
```

Tests can also be aggregated into test groups to be tested at once.

### Running Locally

Generate a self-signed certificate (cert.pem & key.pem) to run services locally. Required for multiplexing grpc/http2 over single port. 

```bash
mkdir -p ssl && \
  (cd ssl && \
    go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 2048 --host 127.0.0.1,::1,localhost,localhost:443,localhost:4443 --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h)
```

Use bazel to run the service
```bash
bazel run //services/helloworld:helloworld -- -http-port 4443 -cert $(pwd)/ssl/cert.pem -key $(pwd)/ssl/key.pem
```

Then we use cURL to send HTTP requests

```bash
curl -X POST -k https://localhost:4443/v1/greeter -d '{"name": "TestName"}'
```

You can view the swagger at [https://localhost:4443/swagger.json](https://localhost:4443/swagger.json)

Or use the client:
With the server running, you can test command line tools from `cmd`.
```bash
$ bazel run //cmd/helloworld-client -- \
    --name "Beutiful" \
    --server-addr localhost:4443 \
    --ca-cert $(pwd)/ssl/cert.pem

INFO: Analyzed target //cmd/helloworld-client:helloworld-client (0 packages loaded, 0 targets configured).
INFO: Found 1 target...
Target //cmd/helloworld-client:helloworld-client up-to-date:
  bazel-bin/cmd/helloworld-client/helloworld-client_/helloworld-client
INFO: Elapsed time: 0.365s, Critical Path: 0.00s
INFO: 1 process: 1 internal.
INFO: Build completed successfully, 1 total action
INFO: Running command line: bazel-bin/cmd/helloworld-client/helloworld-client_/helloworld-client --name 'Beutiful' --server-addr localhost:4443 --cert ...
INFO: Build completed successfully, 1 total action

2022/10/22 18:03:59 message:"Hello Beutiful!"
```

### Running with Docker

```
oci_load(
    name = "load",
    # Use the image built for the target platform
    image = ":transitioned_image",
    repo_tags = ["ghcr.io/adgreetz/go-grpc-bazel-example/cmd/helloworld-client:latest"],
)
```

For example, to load tarball with current architecture:
```bash
bazel run //services/helloworld:load

# Run the loaded image
docker run --rm -v $(pwd)/ssl:/ssl -p 4443:4443 ghcr.io/adgreetz/go-grpc-bazel-example/services/helloworld:latest --http-port 4443 --cert /ssl/cert.pem --key /ssl/key.pem
```

arch example:
```bash
 bazel run \
  --platforms=@rules_go//go/toolchain:linux_amd64 \
  --cpu=k8 \
  //services/helloworld:load
```

## Swagger + JSON Gateway

Services may utilize `grpc-gateway` for a JSON-to-GRPC proxy. This is autogenerated with the `gateway_grpc_library` rule. Swagger json is autogenerated via the `gateway_openapiv2_compile` rule.

```
load("@rules_proto_grpc//grpc-gateway:defs.bzl", "gateway_grpc_compile", "gateway_grpc_library", "gateway_openapiv2_compile")

gateway_grpc_library(
    name = "helloworld_gateway_lib_proto",
    importpath = "github.com/AdGreetz/go-grpc-bazel-example/pb/helloworld",
    protos = [":helloworld_proto"],
    visibility = ["//visibility:public"],
)

gateway_openapiv2_compile(
    name = "helloworld_gateway_grpc",
    protos = [":helloworld_proto"],
    visibility = ["//visibility:public"],
)
```

Services can then use the `embed` rule to embed the `swagger.json` compiled output.

View the `genrule` in [BUILD.bazel](services/helloworld/BUILD.bazel)

```
//go:embed helloworld_openapi_swagger.json
var Data []byte
```

Services can then expose the `swagger.json` file directly.

## Deployment

### Pushing to Container Registry

Used to deploy a service to the container registry.

Each service should contain a `oci_push` rule, which defines the container registry and stamped image. The `:transitioned_image` contains multi-arch capabilites. 

```
oci_push(
    name = "push",
    image = ":transitioned_image",
    remote_tags = ":stamped",
    repository = "ghcr.io/adgreetz/go-grpc-bazel-example/cmd/helloworld-client",
    visibility = ["//visibility:public"],
)
```

To push an individual service (where version is the container label):

Iterates through all `:push` targets and pushes to the container registry.
```bash
make push
```

e.g.
```bash
bazel run --platforms=@rules_go//go/toolchain:linux_amd64 \
  --cpu=k8 \
  //services/helloworld:push
```

See [ci/push-service.sh](ci/push-service.sh)

### Kubernetes Deployment

Since [rules_docker](https://github.com/bazelbuild/rules_docker?) has been deprecated, we can no longer use the `k8s_deploy` rule to deploy to k8s. Instead, we can use the `oci_push` rule to push the image to the container registry, and then use `kubectl` to apply the deployment.

Iterates through all `:push` targets, uses the stamp to update the image tag and applies the k8s deployment.
```bash
make deploy
```

See [ci/deploy.sh](ci/deploy.sh)

## Useful Links

**GRPC**

- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)

**Bazelbuild rules**

- [rules_oci](https://github.com/bazel-contrib/rules_oci/tree/main)
- [rules_docker](https://github.com/bazelbuild/rules_docker)
- [rules_go](https://github.com/bazelbuild/rules_go)
- [rules_k8s](https://github.com/bazelbuild/rules_k8s)
- [rules_proto](https://github.com/bazelbuild/rules_proto)
