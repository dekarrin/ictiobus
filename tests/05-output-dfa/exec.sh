#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[1/1] Output CLR(1) DFA:"
./ictcc --clr \
	-l SimpleMath -v 1.0.0 \
	-nq \
	--dfa \
	"$script_path/simplemath.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[2/1] Output SLR(1) DFA:"
./ictcc --slr \
	-l SimpleMath -v 1.0.0 \
	-nq \
	--dfa \
	"$script_path/simplemath.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[3/1] Output LL(1) DFA:"
./ictcc --ll \
	-l SimpleMath -v 1.0.0 \
	-nq \
	--dfa \
	"$script_path/simplemath.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"