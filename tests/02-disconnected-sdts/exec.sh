#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[PRE] Build diag binary:"
./ictcc --clr \
	--ir 'int' \
	-l SimpleMath -v 1.0.0 \
	-d "$script_path/testdiag" \
	--hooks "$script_path/.hooks" \
	--dev \
	-S all \
	-nq \
	"$script_path/simplemath.md" > /dev/null || { echo "FAIL" >&2 ; exit 1 ; }

echo "(done)"
# above should produce warnings but no actual issues. we can verify specific
# suppressions with manual simulation
echo "[1/3] Warns expected during simulation:"

"$script_path/testdiag" --sim -q
echo "(done)"

echo "[2/3] Warns should cause failure when set with -F:"
"$script_path/testdiag" --sim -q -F validation
echo "(done)"

echo "[3/3] Warns should be suppressed when set with -S:"
"$script_path/testdiag" --sim -q -S validation
echo "(done)"