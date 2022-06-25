#!/bin/bash
# This script is used to test the repo

env GO111MODULE=on

set -eux

gocount=$(git ls-files | grep '.go$' | grep -v 'bindata_assetfs.go$' | grep -v 'bindatafs.go$' | grep -v 'pb.go$' | grep -v 'bindata.go$' | grep -v 'pb.gw.go$' | xargs gofmt -e -l -s | wc -l)
if [ "$gocount" -gt 0 ]; then
  echo "Some Go files are not formatted. Check your formatting!"
  exit 1
fi

buildcount=$(buildifier -mode=check $(find . -type f \( -iname BUILD -or -iname BUILD.bazel \) | grep -v node_modules | grep -v vendor) | wc -l)
if [ "$buildcount" -gt 0 ]; then
    echo "Some BUILD files are not formatted. Run make fmt"
    exit 1
fi

bazeltests=$(bazel query 'kind(".*_test rule", //...)')
for i in "${bazeltests[@]}"; do
    echo "testing $i"
    bazel test --features race \
  --verbose_failures \
  --test_output=errors \
  --action_env=CI=true "${i}"
done