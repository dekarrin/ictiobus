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

echo "[1/1] No statement shark!"
./testdiag-eval -C "(0-2) / 0     <o^><"
echo "(done)"