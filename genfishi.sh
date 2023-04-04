#!/bin/bash

# shortcut for running ictcc on fishi.md for when we build a new ictcc bin

./ictcc --clr \
	--ir '[]github.com/dekarrin/ictiobus/fishi/syntax.Block' \
	--dest .testout \
	-l fishi --lang-ver 1.0.0 \
	--hooks fishi/syntax \
	fishi.md "$@"
