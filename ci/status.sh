#!/bin/bash
# This script is used in as an argument to bazel --workspace_status_command
# it populates BUILD files with the below variables

echo STABLE_GIT_COMMIT "$(git rev-parse --short HEAD)"
echo BUILD_TIME "$(date -u '+%Y-%m-%d_%H:%M:%S')"
echo VERSION "$(openssl rand -base64 8 | sha256sum | head -c8)"
echo PUSH_REPO "$(echo $PUSH_REPO)"
# dynamically pass env variables to the build files
echo ENV "$(echo $env)"