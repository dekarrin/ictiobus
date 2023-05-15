#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[PRE] Build with ictcc:"
./ictcc --ll \
	-l SimpleMath -v 1.0.0 \
	-d "$script_path/testdiag" \
	--hooks "$script_path/.hooks" \
	--ir 'int' \
	--dev \
	-nq \
	"$script_path/simplemath-ll.md" || { echo "FAIL" >&2 ; exit 1 ; }

echo "(done)"

echo "[1/4] Evaluate 2+3:"
"$script_path"/testdiag -C "2+3" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[2/4] Evaluate 2:"
"$script_path"/testdiag -C "2"   || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[3/4] Evaluate 2*3:"
"$script_path"/testdiag -C "2*3" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[4/4] Evaluate (3+4) * 5:"
"$script_path"/testdiag -C "(3+4) * 5" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"
