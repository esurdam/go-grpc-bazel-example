#!/bin/bash
# This script is used to create a coverage report

env GO111MODULE=on

set -e
set -u
set -x

# Queries for all push rules and runs them
for i in $(bazel query 'kind(".push rule", //...)'); do
  bazel run \
    --platforms=@rules_go//go/toolchain:linux_amd64 \
    --cpu=k8 \
    "$i"
done
