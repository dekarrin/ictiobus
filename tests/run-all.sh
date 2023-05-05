#!/bin/bash

# builds ictcc and then runs all tests.

set -o pipefail

script_name="$(basename "$0")"
script_path="$(cd "$(dirname "$0")" >/dev/null ; pwd -P)"
repo_root="$(cd "$script_path/.." >/dev/null ; pwd -P)"

cd "$repo_root"

rundir=".testruns/$(date '+%Y%m%d%H%M%S')"
mkdir -p "$rundir" || { echo "Cannot create rundir; FAIL" >&2 ; exit 1 ; }

echo "Putting test output in $rundir..."

./build.sh >/dev/null || { echo "Could not build ictcc bin; FAIL" >&2 ; exit 1 ; }

any_test_failed=

# assumes we dont put spaces in subdirs of "tests".
for f in $(cd tests ; echo */)
do
  test_failed=
  testdir="$rundir/$f"
  mkdir -p "$testdir"

  echo "--------------------------------"
  echo "STARTING TEST $f..."
  echo "--------------------------------"
  tests/$f/exec.sh 2>&1 | tee "$testdir/actual.txt" || test_failed=1
  
  if [ -z "$test_failed" ]
  then
    # do a diff on actual vs expected:
    if ! diff -u "tests/$f/expect.txt" "$testdir/actual.txt"
    then
      echo "Output does not match expected: FAIL" >&2
      test_failed=1
    fi
  fi

  [ -z "$test_failed" ] || any_test_failed=1
done

echo "----------------------------------"
[ -z "$any_test_failed" ] || { echo "One or more tests failed" >&2 ; exit 2 ; }

echo "ALL TESTS PASSED"
