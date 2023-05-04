#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

./ictcc --clr \
	--ir 'int' \
	-l SimpleMath -v 1.0.0 \
	-d "$script_path/testdiag" \
    --hooks "$script_path/hooks" \
	--dev \
	-S all \
	-nq \
	"$script_path/simplemath.md" || { echo "FAIL" >&2 ; exit 1 ; }

# above should produce warnings but no actual issues. we can verify types in
# manual simulation

"$script_path/testdiag" --sim