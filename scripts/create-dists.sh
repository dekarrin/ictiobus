#!/bin/bash

# this file builds distributions. By default, for 3 major operating systems:
# Mac (darwin), Windows, and Linux

# fail immediately on first error
set -eo pipefail

# assumes we are in a dir called 'scripts' in the repo root:
cd "$(dirname "$0")/.."

if [ -z "$PLATFORMS" ]
then
  PLATFORMS="
darwin/amd64
windows/amd64
linux/amd64"
fi


# only do skip tests if tests have already been done.
[ "$1" = "--skip-tests" ] && skip_tests=1

BINARY_NAME="ictcc"
MAIN_PACKAGE_PATH="./cmd/ictcc"
ARCHIVE_NAME="ictiobus"

tar_cmd=tar
if [ "$(uname -s)" = "Darwin" ]
then
	if tar --version | grep bsdtar >/dev/null 2>&1
	then
		if ! gtar --version >/dev/null 2>&1
		then
			echo "You appear to be running on a mac where 'tar' is BSD tar." >&2
			echo "This will cause issues due to its adding of non-standard headers." >&2
			echo "" >&2
			echo "Please install GNU tar and make it available as 'gtar' with:" >&2
			echo "  brew install gnu-tar" >&2
			echo "And then try again" >&2
			exit 1
		else
			tar_cmd=gtar
		fi
	fi
fi


version="$(go run $MAIN_PACKAGE_PATH --version | awk '{print $NF;}')"
if [ -z "$version" ]
then
	echo "could not get version number; abort" >&2
	exit 1
fi

echo "Creating distributions for $ARCHIVE_NAME version $version"

rm -rf "$BINARY_NAME" "$BINARY_NAME.exe"
rm -rf "source.tar.gz"
rm -rf *-source/

if [ -z "$skip_tests" ]
then
  go clean
  go get ./... || { echo "could not install dependencies; abort" >&2 ; exit 1 ; }
  echo "Running unit tests..."
  if go test -count 1 -timeout 30s ./...
  then
    echo "Unit tests passed"
  else
    echo "Unit tests failed; fix the tests and then try again" >&2
    exit 1
  fi
  echo "Running integration tests..."
  if tests/run-all.sh
  then
    echo "Integration tests passed"
  else
    echo "Integration tests failed; fix the tests and then try again" >&2
    exit 1
  fi
  echo "Running example tests..."
  if examples/run-all-tests.sh
  then
    echo "Example tests passed"
  else
    echo "Example tests failed; fix the tests and then try again" >&2
  fi
else
  echo "Skipping tests due to --skip-tests flag; make sure they are executed elsewhere"
fi

source_dir="$ARCHIVE_NAME-$version-source"
git archive --format=tar --prefix="$source_dir/" HEAD | "$tar_cmd" xf -
"$tar_cmd" czf "source.tar.gz" "$source_dir"
rm -rf "$source_dir"

for p in $PLATFORMS
do
  current_os="${p%/*}"
  current_arch="${p#*/}"
  echo "Building for $current_os on $current_arch..."

  dist_bin_name="$BINARY_NAME"
  if [ "$current_os" = "windows" ]
  then
    dist_bin_name="${BINARY_NAME}.exe"
  fi

  go clean
  env CGO_ENABLED=0 GOOS="$current_os" GOARCH="$current_arch" go build -o "$dist_bin_name" "$MAIN_PACKAGE_PATH" || { echo "build failed; abort" >&2 ; exit 1 ; }

  dist_versioned_name="$ARCHIVE_NAME-$version-$current_os-$current_arch"
  dist_latest_name="$ARCHIVE_NAME-latest-$current_os-$current_arch"

  distfolder="$dist_versioned_name"
  rm -rf "$distfolder" "$dist_latest_name.tar.gz" "$dist_versioned_name.tar.gz"
  mkdir "$distfolder"
  mkdir "$distfolder/docs"
  mkdir "$distfolder/examples"
  cp docs/*.md "$distfolder/docs"

  mkdir "$distfolder"/examples/fishimath-immediate
  cp -R examples/fishimath-immediate/fmhooks "$distfolder/examples/fishimath-immediate"
  cp -R examples/fishimath-immediate/fm-eval.md "$distfolder/examples/fishimath-immediate"
  cp -R examples/fishimath-immediate/go.mod "$distfolder/examples/fishimath-immediate"
  cp -R examples/fishimath-immediate/go.sum "$distfolder/examples/fishimath-immediate"
  cp -R examples/fishimath-immediate/build-diag-fm.sh "$distfolder/examples/fishimath-immediate"
  cp -R examples/fishimath-immediate/README.md "$distfolder/examples/fishimath-immediate"
  cp -R examples/fishimath-immediate/eights.fm "$distfolder/examples/fishimath-immediate"
  mkdir "$distfolder"/examples/fishimath-ast
  cp -R examples/fishimath-ast/fmhooks "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/fmfront "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/fm "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/cmd "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/fm-ast.md "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/build-fmi.sh "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/go.mod "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/go.sum "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/README.md "$distfolder/examples/fishimath-ast"
  cp -R examples/fishimath-ast/eights.fm "$distfolder/examples/fishimath-ast"
  cp README.md RELEASES.md LICENSE source.tar.gz "$distfolder"

  # do some magic in our example directories to get a go.mod that points to the
  # actual current version, if we are building one
  re='^v[0-9]+(\.[0-9]+)*$'
  if [[ $version =~ $re ]]; then
    echo "Distribution is for a release version; updating example go.mod files"

    exec_dir="$(pwd)"
    for example_dir in $(cd "$distfolder/examples" ; echo */)
    do
      echo "Update example $example_dir..."
      cd "$distfolder/examples/$example_dir"
      rm go.sum
      env GOWORK=off go mod edit -require=github.com/dekarrin/ictiobus@$version
      env GOWORK=off go mod tidy
      cd "$exec_dir"
    done
  fi
  
  if [ "$current_os" != "windows" ]
  then
    # no need to set executable bit on windows
    chmod +x "$dist_bin_name"
  fi
  mv $dist_bin_name "$distfolder/"
  $tar_cmd czf "$dist_versioned_name.tar.gz" "$distfolder"
  rm -rf "$distfolder"

  echo "$dist_versioned_name.tar.gz"
  cp "$dist_versioned_name.tar.gz" "$dist_latest_name.tar.gz"
  echo "$dist_latest_name.tar.gz"
done

rm -rf source.tar.gz
