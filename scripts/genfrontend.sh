#!/bin/bash

# LIVE FRONTEND GENERATION; use this script to generate own fishi frontend.

# If this fails with weird 'undefined' errors with certain types that have been
# created or updated since the latest release of ictiobus, you may need to run
# this with the flag `--dev` to use local ictiobus sources instead of the latest
# release.

# assumes we are in a dir called 'scripts' in the repo root:
cd "$(dirname "$0")/.."

if ! [ "$1" = "--yes" ]
then
  read -r -p "Completely replace current fishi frontend by generating from fishi.md? " replace_it
  replace_it="$(echo "$replace_it" | tr '[:upper:]' '[:lower:]')"
  if ! [ "$replace_it" = 'y' -o "$replace_it" = 'ye' -o "$replace_it" = 'yes' ]
  then
    echo "'y'/'yes' not typed; abort" >&2
    exit 1
  fi
else
  shift
fi

echo "Building current ictcc bin..."
scripts/build.sh || exit 2

echo "Generating new frontend in fishi/fe..."
./ictcc --lalr \
	--ir 'github.com/dekarrin/ictiobus/fishi/syntax.AST' \
	--dest fishi/fe \
	-l FISHI -v 1.0.0 \
	--hooks fishi/syntax \
	docs/fishi.md "$@"

