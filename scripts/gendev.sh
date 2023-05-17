#!/bin/bash

# Shortcut for running ictcc on fishi.md for when we build a new ictcc bin.
#
# Generates a dev binary and used for checking output without actually replacing
# the current frontend. All options passed to the ictcc binary located in the
# root of the repo; to get one, call scripts/build.sh.

# assumes we are in a dir called 'scripts' in the repo root:
cd "$(dirname "$0")/.."

./ictcc --clr \
	--ir 'github.com/dekarrin/ictiobus/fishi/syntax.AST' \
	--dest .testout \
	-l FISHI -v 1.0 \
	--hooks fishi/syntax \
	-d fishic \
	--dev \
	-f fishi/format \
	docs/fishi.md "$@"
