#!/bin/bash
# This script is used to format all go files and BAZEL build files
# It is required to successfully pass a Travis build on the master branch.

echo "Checking go formatting..."
gofiles=$(git ls-files | grep '.go$' | grep -v 'pb.go$' | grep -v 'pb.gw.go$' | xargs gofmt -e -l -s )
for i in ${gofiles[@]}; do
    echo "formatting $i"
    gofmt -w -s "${i}"
done

echo "Checking goimports formatting..."
gofiles=$(git ls-files | grep '.go$' | grep -v 'pb.go$' | grep -v 'pb.gw.go$' | xargs goimports -e -l )
for i in ${gofiles[@]}; do
    echo "formatting $i"
    goimports -w "${i}"
done

buildifier -mode=fix $(find . -type f \( -iname BUILD -or -iname BUILD.bazel \) | grep -v node_modules | grep -v vendor)
buildifier -mode=fix ./WORKSPACE
