#!/bin/bash
# This script is used to deploy services to k8s

env GO111MODULE=on

set -e
set -u
set -x

# Queries for all push rules and deploys them
for i in $(bazel query 'kind(".push rule", //...)'); do
  IMAGE_TAG=$(bazel run --platforms=@rules_go//go/toolchain:linux_amd64 --cpu=k8 "$i")
  VERSION=$(cat "$(bazel cquery --output=files "${i/:push/:stamped}")")
  root="${i/\/\//}" # remove the leading //
  core="${root/:push/}" # remove the trailing :push
  template=$(cat "ci/${core}.yaml" | sed "s#{{IMAGE_TAG}}#$IMAGE_TAG#g" | sed "s/{{VERSION}}/$VERSION/g")
  echo "$template"
done
