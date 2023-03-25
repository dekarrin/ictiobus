#!/bin/bash

#!/bin/bash

set -eo pipefail

if ! which go >/dev/null 2>&1
then
  echo "The Go language compiler does not appear to be installed" >&2
  echo "Install Go from https://go.dev/doc/install and try again" >&2
  exit 1
fi

cd "$(dirname "$0")"

ext=
for_windows=
[ "$(go env GOOS)" = "windows" ] && for_windows=1 && ext=".exe"

env CGO_ENABLED=0 go build -o ictcc$ext cmd/ictcc/*.go
