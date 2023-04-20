#!/bin/bash

# LIVE FRONTEND GENERATION; use this script to generate own fishi frontend.

cd "$(dirname "$0")"

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
./build.sh || exit 2

echo "Generating new frontend in fishi/fe..."
./ictcc --clr \
	--ir 'github.com/dekarrin/ictiobus/fishi/syntax.AST' \
	--dest fishi/fe \
	-l FISHI -v 1.0.0 \
	--hooks fishi/syntax \
	fishi.md "$@"

