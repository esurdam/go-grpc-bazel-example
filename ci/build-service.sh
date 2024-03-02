#!/bin/bash
# This script is used to build/deploy services.

env GO111MODULE=on

set -e
set -u
set -x

BUILD_PATHS=(
  "//cmd/helloworld-client:helloworld-client"
  "//services/helloworld:helloworld"
)

for path in "${BUILD_PATHS[@]}"; do
  bazel build \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --cpu=k8 "$path"
done

