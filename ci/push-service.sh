#!/bin/bash
# This script is used to create a coverage report

env GO111MODULE=on

set -e
set -u
set -x

IMAGE_TAG=`bazel run --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 --cpu=k8 //services/helloworld:push`
VERSION=`cat "$(bazel cquery --output=files //services/helloworld:stamped)"`
template=`cat "ci/services/helloworld.yaml" | sed "s#{{IMAGE_TAG}}#$IMAGE_TAG#g" | sed "s/{{VERSION}}/$VERSION/g"`
echo "$template" | kubectl apply --namespace ops -f -