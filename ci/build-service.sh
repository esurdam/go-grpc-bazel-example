#!/bin/bash
# This script is used to build/deploy services.

env GO111MODULE=on

set -e
set -u
set -x

services=$(bazel query 'kind(".*_binary rule", //services/...)')
for i in "${services[@]}"; do
    echo "building $i"
    bazel build "${i}"
done