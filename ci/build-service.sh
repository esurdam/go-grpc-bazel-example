#!/bin/bash
# This script is used to build/deploy services.

env GO111MODULE=on

set -e
set -u
set -x

bazel build \
  --define version="$(openssl rand -base64 8 |md5 |head -c8)" \
  --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
  --cpu=k8 //:bundle