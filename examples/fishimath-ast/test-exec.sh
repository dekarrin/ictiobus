#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

old_dir="$(pwd)"

export ICTIOBUS_SOURCE="$old_dir"

cd "$script_path"

echo "[PRE] Building diagnostic binary..."

"$old_dir/ictcc" --clr \
	--ir 'github.com/dekarrin/ictfishimath_ast/fmhooks.AST' \
	-l FISHIMath -v 1.0 \
	-d "$script_path/testdiag" \
	--hooks "fmhooks" \
	-S all \
    --dev \
	-nq \
	"$script_path/fm-ast.md" || { echo "FAIL" >&2 ; exit 1 ; }

echo "(done)"

echo "[1/7] int arithmetic"
./testdiag -C "2 / 3 + 3384 * >{16 - 20'}             <o^><"
echo "(done)"

echo "[2/7] float arithmetic"
./testdiag -C "
2 / 3 + 3384.2 * >{16 - 20.24'}  <o^><
0.1 + 0.2                        <o^><
"
echo "(done)"

echo "[3/7] variable"
./testdiag -C "
vriska =o 4 * 2  <o^><
vriska * 2       <o^><
"
echo "(done)"

echo "[4/7] Divide positive by zero"
./testdiag -C "2 / 0          <o^><"
echo "(done)"

echo "[5/7] Divide negative by zero"
./testdiag -C ">{0-2'} / 0    <o^><"
echo "(done)"

echo "[6/7] Regular input"
./testdiag -C ">{0-2'} / 0    <o^><"
echo "(done)"

echo "[7/7] Missing statement shark gives error"
./testdiag -C ">{0-2'} / 0"
echo "(done)"