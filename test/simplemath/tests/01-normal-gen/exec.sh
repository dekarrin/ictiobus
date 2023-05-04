#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"
simplemath_path="$(cd "$script_path/../.." >/dev/null ; pwd -P)"

./ictcc --clr \
	--ir 'int' \
	-l SimpleMath -v 1.0.0 \
	-d "$script_path/testdiag" \
    --hooks "$script_path/hooks" \
	--dev \
	-n \
	"$simplemath_path/simplemath.md"
