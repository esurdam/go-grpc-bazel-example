# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build, Test & Lint Commands
- **Build all services**: `make build` or `bazel build :build_all`
- **Run all tests**: `make test` or `bazel test //...`
- **Run single test**: `bazel test --features race --verbose_failures --test_output=errors //pkg/path:target_test`
- **Format code**: `make fmt`
- **Generate protos**: `make link`
- **Update BUILD files**: `make gazelle`
- **Deploy to k8s**: `make deploy`

## Code Style Guidelines
- Follow Go standard formatting with `gofmt -s`
- Organize imports with `goimports`
- Format Bazel files with `buildifier`
- Use table-driven tests with descriptive names
- Error handling: Return simple error objects (`errors.New()`) for validation
- Package naming: Use domain-oriented packages under `pkg/` 
- Proto implementation: Shared services in `pkg/`, service-specific in `services/{name}/pkg/`
- For new services, follow the scaffold pattern in README.md
- Write comprehensive unit tests for all implementations