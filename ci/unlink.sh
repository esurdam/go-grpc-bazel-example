#!/bin/bash
# This script is used to unlink bazel compiled grpc files from the local workspace.

shopt -s globstar

# delete local pb files
for i in $(ls $(bazel info bazel-genfiles)/pb/**/*.go | grep -v '_rpg') ; do
    OUTPUT="pb/${i##*/pb/}"
    echo "deleting $OUTPUT" ;
    [[ -f "$OUTPUT" ]] && rm -f "$OUTPUT"
done