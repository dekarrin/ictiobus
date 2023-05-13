#!/bin/bash

# Runs all integration tests, exiting if there is any error.

# assumes we are in a dir called 'scripts' in the repo root:
cd "$(dirname "$0")/.."

go test ./... -count 1 -timeout 30s || { echo "Unit tests failed; not running int tests" >&2 ; exit 1 ; }
echo "" >&2
echo "Unit tests passed; running int tests..." >&2
scripts/run-int-tests.sh
