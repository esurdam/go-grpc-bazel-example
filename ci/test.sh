#!/bin/bash
# Run the lint gate, then execute all Bazel tests with the Go race detector.
#
# Race is scoped to test targets (via query) rather than //...: the OCI image
# targets transition to a cgo-disabled static platform, which is incompatible
# with race instrumentation.

set -eu

bash ci/lint.sh

set -x
bazel test --config=ci --config=race $(bazel query 'kind(".*_test rule", //...)')
