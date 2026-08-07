#!/bin/bash
# This script is used to create a coverage report

env GO111MODULE=on

set -e
set -u
set -x

>coverage.txt

coverWithGo() {
  TEST_PKGS=$(go list ./pkg/...)
  for d in ${TEST_PKGS}; do
      go test -race -coverprofile=profile.out -covermode=atomic "${d}"
      if [ -f profile.out ]; then
        cat profile.out >>coverage.txt
        rm profile.out
      fi
  done

  go tool cover --html=coverage.txt -o coverage.html

  echo "Coverage completed."
  echo "Open coverage.html to view result."
}


coverWithBazel() {
  # This pass doubles as the CI test gate: it runs every test with the Go race
  # detector (--config=race) while collecting coverage, so there is no separate
  # race-test pass. Race is scoped to test targets (via query) because the OCI
  # image targets transition to a cgo-disabled platform that race rejects.
  bazel coverage --config=ci --config=race --combined_report=lcov \
    $(bazel query 'kind(".*_test rule", //...)')

  local report
  report="$(bazel info output_path)/_coverage/_coverage_report.dat"
  # Stable path for Codecov / other coverage services.
  cp "${report}" coverage.lcov
  genhtml --branch-coverage --output genhtml "${report}"

  echo "Coverage completed."
  echo "Open genhtml/index.html to view result."
}

arg="${1:-}"
if [ -n "$arg" ]; then
  coverWithGo
else
  coverWithBazel
fi