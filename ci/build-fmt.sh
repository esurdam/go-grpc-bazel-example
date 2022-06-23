#!/bin/bash
# This script is used to format all go files and BAZEL build files
# It is required to successfully pass a Travis build on the master branch.

echo "Checking go formatting..."
gofiles=$(git ls-files | grep '.go$' | grep -v 'pb.go$' | grep -v 'pb.gw.go$')
gofmtfiles=$(echo $gofiles | xargs gofmt -e -l -s )
for i in ${gofmtfiles[@]}; do
    echo "formatting $i"
    gofmt -w -s "${i}"
done

echo "Checking goimports formatting..."
goimpfiles=$(echo $gofiles | xargs goimports -e -l )
for i in ${goimpfiles[@]}; do
    echo "formatting $i"
    goimports -w "${i}"
done

buildifier -mode=fix --lint=fix $(find . -type f \( -iname BUILD -or -iname BUILD.bazel -or -iname WORKSPACE -or -iname go.bzl \) | grep -v node_modules | grep -v vendor)
