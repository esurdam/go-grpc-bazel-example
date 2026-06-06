# AGENTS.md

Guidance for AI coding agents (Claude Code, Cursor, etc.) working in this repository.

## Repository Structure

Monorepo — all microservices, protobufs, and shared code live in a single repository.

- `pb/` — All protobuf definitions and generated code.
- `pkg/` — Go implementations of proto services.
- `services/` — Entrypoints for deployable microservices.
- `cmd/` — Standalone command-line tools.
- `ci/` — CI/CD scripts and Kubernetes manifests.
- `tools/` — Tool versioning and helper scripts.
- `ssl/` — Local self-signed certificates (never commit private keys).

## Build, Test & Lint Commands

- **Build all services**: `make build` or `bazel build :build_all`
- **Run all tests**: `make test` or `bazel test //...`
- **Run single test**: `bazel test --features race --verbose_failures --test_output=errors //pkg/path:target_test`
- **Format code**: `make fmt`
- **Generate protos**: `make link`
- **Update BUILD files**: `make gazelle`
- **Push images**: `make push`
- **Deploy to k8s**: `make deploy`

## Protobufs

- Place all proto files under `pb/`.
- Generated code is tracked in the repo for reproducibility — do not manually edit generated files.
- Use `make link` to generate/update protobuf Go code for local development.

## Bazel

- Use `make gazelle` to generate/update all `BUILD.bazel` files.
- Every new service and package must have corresponding Bazel build rules.
- Prefer auto-generation; do not check in hand-written `BUILD.bazel` files unless necessary.

## Services

Each new service must have:

- Proto definition in `pb/<service>/<service>.proto`
- Implementation in `pkg/<service>/server/server.go`
- Entrypoint in `services/<service>/main.go`
- Kubernetes manifest in `ci/services/<service>.yaml`
- Aggregation in the root `BUILD` file

See the scaffold pattern in `README.md` for full details.

## Testing

- All code must have tests. Use table-driven tests with descriptive names.
- Write comprehensive unit tests for all implementations.
- Run `make test` before submitting changes.
- Run an individual package test with `bazel test --features race --verbose_failures --test_output=errors //pkg/path:target_test`.

## Code Style

- Follow Go standard formatting with `gofmt -s`.
- Organize imports with `goimports`.
- Format Bazel files with `buildifier`.
- Error handling: return simple error objects (`errors.New()`) for validation.
- Package naming: use domain-oriented packages under `pkg/`.
- Proto implementation: shared services in `pkg/`, service-specific in `services/{name}/pkg/`.

## Certificates

- Use self-signed certificates for local development.
- Store local certs in `ssl/` — do not commit private keys.

## Docker & Deployment

- Use Bazel rules (`oci_push`, etc.) for building and pushing images.
- Do not use the deprecated `rules_docker` for new code.
- Use `make push` and `make deploy` for CI/CD flows.

## Swagger / OpenAPI

- Expose Swagger JSON at `/swagger.json` for each service.
- Use the `gateway_grpc_library` and `gateway_openapiv2_compile` Bazel rules for OpenAPI generation.
