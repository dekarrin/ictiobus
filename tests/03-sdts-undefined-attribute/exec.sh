#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[1/1] Building diag binary should fail during SDTS validation:"
./ictcc --clr \
	--ir 'int' \
	-l SimpleMath -v 1.0.0 \
	-d "$script_path/testdiag" \
    --hooks "$script_path/hooks" \
	--dev \
	-S all \
	-nq \
	"$script_path/simplemath.md" > /dev/null || { echo "FAIL" >&2 ; exit 1 ; }

echo "(done)"
