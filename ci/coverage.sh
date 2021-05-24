#!/bin/bash
# This script is used to create a coverage report

env GO111MODULE=on

set -e
set -u
set -x

>coverage.txt

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