#!/bin/bash

# Runs all integration tests, exiting if there is any error.

# assumes we are in a dir called 'scripts' in the repo root:
cd "$(dirname "$0")/.."

examples/run-all-tests.sh "$@"
