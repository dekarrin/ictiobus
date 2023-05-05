#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[1/1] Build with ictcc:"
./ictcc --ll \
	-l SimpleMath -v 1.0.0 \
	-d "$script_path/testdiag" \
    --hooks "$script_path/hooks" \
	--ir 'int' \
	-n \
	"$script_path/simplemath.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"
