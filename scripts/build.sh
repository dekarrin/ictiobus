#!/bin/bash

# Builds the ictcc binary and places it in the root of the repo.

set -eo pipefail

# assumes we are in a dir called 'scripts' in the repo root:
cd "$(dirname "$0")/.."

if ! which go >/dev/null 2>&1
then
  echo "The Go language compiler does not appear to be installed" >&2
  echo "Install Go from https://go.dev/doc/install and try again" >&2
  exit 1
fi

ext=
for_windows=
[ "$(go env GOOS)" = "windows" ] && for_windows=1 && ext=".exe"

env CGO_ENABLED=0 go build -o ictcc$ext cmd/ictcc/*.go
