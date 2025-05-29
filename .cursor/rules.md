# Cursor Rules for go-grpc-bazel-example

## General Structure

- **Monorepo**: All microservices, protobufs, and shared code live in a single repository.
- **Directory Layout**:
  - `pb/`: All protobuf definitions and generated code.
  - `pkg/`: Go implementations of proto services.
  - `services/`: Entrypoints for deployable microservices.
  - `cmd/`: Standalone command-line tools.
  - `ci/`: CI/CD scripts and Kubernetes manifests.
  - `tools/`: Tool versioning and helper scripts.

## Protobufs

- All proto files must be placed under `pb/`.
- Generated code should be tracked in the repo for reproducibility.
- Use `make link` to generate/update protobuf Go code for local development.
- Do not manually edit generated files.

## Bazel

- Use `make gazelle` to generate/update all `BUILD.bazel` files.
- All new services and packages must have corresponding Bazel build rules.
- Do not check in hand-written `BUILD.bazel` files unless necessary; prefer auto-generation.

## Services

- Each new service must have:
  - Proto definition in `pb/<service>/<service>.proto`
  - Implementation in `pkg/<service>/server/server.go`
  - Entrypoint in `services/<service>/main.go`
  - Kubernetes manifest in `ci/services/<service>.yaml`
  - Aggregated in root `BUILD` file

## Testing

- All code must have tests.
- Run `make test` before submitting changes.
- Individual package tests can be run with `bazel test ...`.

## Certificates

- Use self-signed certificates for local development.
- Store local certs in the `ssl/` directory (do not commit private keys).

## Docker & Deployment

- Use Bazel rules (`oci_push`, etc.) for building and pushing images.
- Do not use deprecated `rules_docker` for new code.
- Use `make push` and `make deploy` for CI/CD flows.

## Swagger/OpenAPI

- Expose Swagger JSON at `/swagger.json` for each service.
- Use `gateway_grpc_library` and `gateway_openapiv2_compile` Bazel rules for OpenAPI generation.
