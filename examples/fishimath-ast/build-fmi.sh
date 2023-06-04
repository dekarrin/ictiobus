#!/bin/bash

set -e

# This script builds the example program fmi. It first uses ictcc to generate
# all sources from the FISHI markdown file, then builds the `fmi` binary in this
# directory.
#
# It also creates an Ictiobus diagnostics binary called `diag-fmi`.

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

# is there a 'test-expect.txt' file present? if so, expect us to be in a repo
# checked out for local dev and include relevant options
local_repo=
[ -f "$script_path/test-expect.txt" ] && local_repo=1

cd "$script_path"

echo "Ictiobus frontend sources for FM will be placed in ./fmfront" >&2
echo "Ictiobus diagnostic binary for FM will be placed in ./diag-fm" >&2

echo "Building FM frontend sources and diagnostic binary..." >&2

if [ -n "$local_repo" -a "$1" != "--non-local" ]
then
  # if we're in a local repo then we expect current ictcc to live two levels up
  # in the repo root.
  repo_root="$(cd "$script_path/../.." >/dev/null ; pwd -P)"

  export ICTIOBUS_SOURCE="${ICTIOBUS_SOURCE:-$repo_root}"

  if [ ! -x "$repo_root"/ictcc ]
  then
    echo "You appear to be running this script in a cloned repo of ictiobus, but the ictcc" >&2
    echo "binary is not present in the repo root. Execute scripts/build.sh from the repo" >&2
    echo "root and try again, or do --non-local to explicitly select use of the installed" >&2
    echo "version of ictcc and the latest release sources as opposed to what is available" >&2
    echo "in the local repo clone." >&2
    echo "" >&2
    echo "ICTCC FAILED" >&2
    exit 1
  else
    echo "Detected execution from within cloned Ictiobus repo; using ictcc binary located" >&2
    echo "in repo root and sources located at $ICTIOBUS_SOURCE." >&2
  fi

  echo "" >&2

  "$repo_root/ictcc" --slr \
    --ir 'github.com/dekarrin/ictfishimath_ast/fmhooks.AST' \
    -l FISHIMath -v 1.0 \
    -d "$script_path/diag-fm" \
    --hooks "fmhooks" \
    --dest "$script_path/fmfront" --pkg fmfront \
    --dev \
    "$script_path/fm-ast.md" || { echo "" >&2 ; echo "ICTCC FAILED" >&2 ; exit 1 ; }
else
  # not in a local repo, so we must use the version of ictcc available to the
  # system.

  # first, in a distribution, the binary should still be available two levels
  # up.
  ictcc_path="../../ictcc"

  if [ ! -x "$bin_path" ]
  then
    # looks like it's not there. that's okay, as long as it's installed
    # somewhere on the system.
    if ! command -v ictcc &> /dev/null
    then
      echo "No ictcc executable found in PATH or in distribution." >&2
      echo "Cannot proceed." >&2
      echo ""
      echo "ICTCC FAILED" >&2
      exit 1
    fi

    echo "Using ictcc executable available in \$PATH" >&2

    bin_path="$(command -v ictcc)"
  else
    echo "Using ictcc executable included with distribution" >&2
  fi

  echo "" >&2

  "$bin_path" --slr \
    --ir 'github.com/dekarrin/ictfishimath_ast/fmhooks.AST' \
    -l FISHIMath -v 1.0 \
    -d "$script_path/diag-fm" \
    --hooks "fmhooks" \
    --dest "$script_path/fmfront" --pkg fmfront \
    "$script_path/fm-ast.md" || { echo "" >&2 ; echo "ICTCC FAILED" >&2 ; exit 1 ; }
fi

echo "" >&2
echo "Frontend generation completed successfully; building fmi..." >&2

fmi_bin="fmi"
[ "$(go env GOOS)" = "windows" ] && fmi_bin="$fmi_bin.exe"

go clean
env CGO_ENABLED=0 go build -o "$fmi_bin" "./cmd/fmi" || { echo "" >&2 && echo "BUILD FAILED" >&2 ; exit 2 ; }

echo "" >&2
echo "Binary $fmi_bin successfully built" >&2
