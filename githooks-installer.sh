#!/bin/sh

set -e

cp ./.githooks/pre-commit .git/hooks/pre-commit

echo "DONE"