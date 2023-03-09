#!/bin/bash

set -e -o pipefail

# https://github.com/koalaman/shellcheck/wiki/SC2039#redirect-both-stdout-and-stderr
if ! command -v golangci-lint 2>&1 /dev/null; then
    echo "golangci-lint is not installed"
    exit 1
fi

exec golangci-lint --build-tags=race run "$@"
