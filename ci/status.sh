#!/bin/bash
# This script is used in as an argument to bazel --workspace_status_command
# it populates BUILD files with the below variables

echo STABLE_GIT_COMMIT "$(git rev-parse --short HEAD)"
echo BUILD_TIME "$(date -u '+%Y-%m-%d_%H:%M:%S')"
echo VERSION "$(openssl rand -base64 8 |md5 |head -c8)"
#echo BUILD_EMBED_LABEL "$(openssl rand -base64 8 |md5 |head -c8)"
echo ENV "$(echo $env)"
echo PUSH_REPO "$(echo $PUSH_REPO)"