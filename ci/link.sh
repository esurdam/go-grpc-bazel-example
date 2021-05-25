#!/bin/bash
# This script is used to link bazel compiled grpc files to the local workspace.

shopt -s globstar

for i in $( bazel query 'kind(".*_library rule", //pb/...)' | grep -E 'gateway') ; do \
    bazel build "${i}_pb" ; \
done

for i in $(ls $(bazel info bazel-genfiles)/pb/**/*.go | grep -v '_rpg') ; do
    OUTPUT="pb/${i##*/pb/}"
    echo "copying $(basename $i) to $OUTPUT" ;
    [[ -f "$OUTPUT" ]] && rm -f "$OUTPUT"
    cp "${i}" "$OUTPUT"
done