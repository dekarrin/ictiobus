#!/bin/bash

script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"

echo "[1/5] FISHI with no actions"
./ictcc -nsqS all "$script_path/no-actions.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[2/5] FISHI with no grammar"
./ictcc -nsqS all "$script_path/no-grammar.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[3/5] FISHI with no tokens"
./ictcc -nsqS all "$script_path/no-tokens.md" || { echo "FAIL" >&2 ; exit 1 ; }
echo "(done)"

echo "[4/5] Invalid FISHI should fail (all code blocks are empty)"
./ictcc -nsqS all "$script_path/empty-fishi-block.md"
echo "(done)"

echo "[5/5] Invalid FISHI should fail (no code blocks are present)"
./ictcc -nsqS all "$script_path/no-fishi-blocks.md"
echo "(done)"
