#!/bin/bash
# Lint gate: verify Go source and Bazel BUILD files are correctly formatted.
# Extracted from ci/test.sh so CI can run the fast format checks as their own
# step, independent of the (heavier) Bazel test/coverage pass.

env GO111MODULE=on
BUILDIFIER_VERSION="5.4.0"

which buildifier >/dev/null
if [ $? -ne 0 ]; then
  go install github.com/bazelbuild/buildtools/buildifier@$BUILDIFIER_VERSION
fi

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
