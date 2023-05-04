#!/bin/bash

# Runs ictcc to generate a diagnostics binary 'smc' for simplemath.md. Execute
# this script from the root of the project after building ictcc.

script_path="$(dirname "$0")"

./ictcc --clr \
	--ir 'int' \
	-l SimpleMath -v 1.0.0 \
	--hooks "$script_path/hooks" \
	-d "$script_path/testdiag" \
	--dev \
	-n \
	"$script_path/simplemath.md" "$@"
