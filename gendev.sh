#!/bin/bash

# shortcut for running ictcc on fishi.md for when we build a new ictcc bin.
# generates a dev binary and used for checking output without actually replacing the current frontend

cd "$(dirname "$0")"

./ictcc --clr \
	--ir 'github.com/dekarrin/ictiobus/fishi/syntax.AST' \
	--dest .testout \
	-l FISHI -v 1.0.0 \
	--hooks fishi/syntax \
	-d fishic \
	--dev \
	-f fishi/format \
	fishi.md "$@"
