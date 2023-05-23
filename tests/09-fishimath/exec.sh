#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

old_dir="$(pwd)"

cd "$script_path/.immediate"

"$old_dir/ictcc" --clr \
	--ir '[]github.com/dekarrin/fishimath/fmhooks.FMValue' \
	-l FISHIMath -v 1.0 \
	-d "$script_path/fmc-eval" \
	--hooks "fmhooks" \
	--dev \
	-S all \
	-nq \
	"$script_path/fm-eval.md" > /dev/null || { echo "FAIL" >&2 ; exit 1 ; }


#echo "[5/5] Invalid FISHI should fail (no code blocks are present)"
#./ictcc -nsqS all "$script_path/no-fishi-blocks.md" 2>&1 | sed 's%'"$script_path"'%(TEST_PATH)%g'
#echo "(done)"
