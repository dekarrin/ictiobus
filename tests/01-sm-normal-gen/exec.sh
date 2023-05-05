#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[PRE] Build diag binary:"
./ictcc --clr \
	--ir 'int' \
	-l SimpleMath -v 1.0.0 \
	-d "$script_path/testdiag" \
    --hooks "$script_path/hooks" \
	--dev \
	-n \
	"$script_path/simplemath.md" >/dev/null || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[1/3] Evaluate 2+3:"
"$script_path"/testdiag -C "2+3" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[2/3] Evaluate 2:"
"$script_path"/testdiag -C "2"   || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[3/3] Evaluate 2*3:"
"$script_path"/testdiag -C "2*3" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"
