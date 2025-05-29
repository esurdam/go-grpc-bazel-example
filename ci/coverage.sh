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
  bazel coverage --combined_report=lcov //...
  genhtml --branch-coverage --output genhtml "$(bazel info output_path)/_coverage/_coverage_report.dat"

  echo "Coverage completed."
  echo "Open genhtml/index.html to view result."
}

arg="${1:-}"
if [ -n "$arg" ]; then
  coverWithGo
else
  coverWithBazel
fi