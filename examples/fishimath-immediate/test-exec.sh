#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

old_dir="$(pwd)"

export ICTIOBUS_SOURCE="$old_dir"

cd "$script_path"

echo "[PRE] Building diagnostic binary..."

"$old_dir/ictcc" --clr \
	--ir '[]github.com/dekarrin/fishimath/fmhooks.FMValue' \
	-l FISHIMath -v 1.0 \
	-d "$script_path/testdiag-eval" \
	--hooks "fmhooks" \
	-S all \
    --dev \
	-nq \
	"$script_path/fm-eval.md" || { echo "FAIL" >&2 ; exit 1 ; }

echo "(done)"

echo "[1/2] int arithmetic"
./testdiag-eval -C "2 / 3 + 3384 * >{16 - 20'}             <o^><"
echo "(done)"

echo "[1/2] float arithmetic"
./testdiag-eval -C "
2 / 3 + 3384.2 * >{16 - 20.24'}  <o^><
0.1 + 0.2                        <o^><
"
echo "(done)"

echo "[1/2] variable"
./testdiag-eval -s -C "
vriska =o 4 <o^><
vriska * 2       <o^><
"
echo "(done)"

echo "[1/2] Divide negative by zero"
./testdiag-eval -C ">{0-2'} / 0    <o^><"
echo "(done)"

echo "[1/2] Regular input"
./testdiag-eval -C ">{0-2'} / 0    <o^><"
echo "(done)"

echo "[2/1] No statement shark gives error"
./testdiag-eval -C ">{0-2'} / 0"
echo "(done)"