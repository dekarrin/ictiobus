#!/bin/bash

# Runs all integration tests, exiting if there is any error.

# assumes we are in a dir called 'scripts' in the repo root:
cd "$(dirname "$0")/.."

go test ./... -count 1 -timeout 30s || { echo "Unit tests failed; not running int tests" >&2 ; exit 1 ; }
echo "" >&2
echo "Unit tests passed; running int tests..." >&2
scripts/int-tests.sh || { echo "Integration tests failed; not running example tests" >&2 ; exit 2 ; }
echo "" >&2
echo "Integration tests passed; running example tests..." >&2
scripts/example-tests.sh
