#!/bin/bash
# This script is used to deploy services to k8s
# It will:
# 1. Query all push rules
# 2. For each push rule, it will:
#    a. Check if the yaml file exists
#    b. Run the push rule to get the image tag 
#    c. Get the version from the stamped file
#    d. Replace the image tag and version in the yaml file
#    e. Apply the yaml file to k8s

env GO111MODULE=on

set -e
set -u
set -x

TOOLCHAIN="@rules_go//go/toolchain:linux_amd64" # TODO: Make this configurable
CPU="k8"  # k8 is the default cpu for linux

process_service() {
  local i=$1
  local root="${i/\/\//}" # remove the leading //
  local core="${root/:push/}" # remove the trailing :push
  # check if yaml exists
  if [ ! -f "ci/${core}.yaml" ]; then
    echo "yaml file not found: ci/${core}.yaml"
    return 0
  fi
  local IMAGE_TAG=$(bazel run --platforms=$TOOLCHAIN --cpu=$CPU "$i")
  local VERSION=$(cat "$(bazel cquery --output=files "${i/:push/:stamped}")")
  # replace the image tag and version in the yaml file
  local template=$(cat "ci/${core}.yaml" | sed "s#{{IMAGE_TAG}}#$IMAGE_TAG#g" | sed "s/{{VERSION}}/$VERSION/g")
  echo "$template"
}

# Queries for all push rules and deploys them
for i in $(bazel query 'kind(".push rule", //...)'); do
  process_service "$i"
done
