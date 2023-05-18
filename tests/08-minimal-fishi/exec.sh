#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[1/2] Preprocess file with comments"
./ictcc -Pnq "$script_path/commented-fishi.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[2/2] Final spec of file with comments"
./ictcc -snq "$script_path/commented-fishi.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"