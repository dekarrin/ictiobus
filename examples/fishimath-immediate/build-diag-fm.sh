#!/bin/bash

set -e

# This script builds an Ictiobus diagnostics binary called `diag-fmi` for the
# specified language.

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

# is there a 'test-expect.txt' file present? if so, expect us to be in a repo
# checked out for local dev and include relevant options
local_repo=
[ -f "$script_path/test-expect.txt" ] && local_repo=1

cd "$script_path"

echo "Building FM diagnostic binary..." >&2

echo "" >&2

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
    --ir '[]github.com/dekarrin/ictfishimath_eval/fmhooks.FMValue' \
    -l FISHIMath -v 1.0 \
    -d "$script_path/diag-fm" \
    --hooks "fmhooks" \
	--dev \
    --no-gen \
    "$script_path/fm-eval.md" || { echo "" >&2 ; echo "ICTCC FAILED" >&2 ; exit 1 ; }
else
  # not in a local repo, so we must use the version of ictcc available to the
  # system.

  # first, in a distribution, the binary should still be available two levels
  # up.
  ictcc_path="$(cd "$script_path/../.." >/dev/null ; pwd -P)/ictcc"

  if [ ! -x "$ictcc_path" ]
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

    ictcc_path="$(command -v ictcc)"
  else
    echo "Using ictcc executable included with distribution" >&2
  fi

  echo "" >&2

  "$ictcc_path" --slr \
    --ir '[]github.com/dekarrin/ictfishimath_eval/fmhooks.FMValue' \
    -l FISHIMath -v 1.0 \
    -d "$script_path/diag-fm" \
    --hooks "fmhooks" \
    --no-gen \
    "$script_path/fm-eval.md" || { echo "" >&2 ; echo "ICTCC FAILED" >&2 ; exit 1 ; }
fi

echo "" >&2
echo "Binary diag-fm successfully built" >&2
