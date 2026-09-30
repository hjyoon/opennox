#!/bin/sh
set -eu

# go.sh deliberately clears GOEXPERIMENT for reproducible release builds.
# Resolve its pinned toolchain first, then enable strict CGo checks only for
# tests. This entry point never invokes go build or internal/noxbuild.
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(dirname "$script_dir")
required_go=$(tr -d '[:space:]' < "$repo_dir/toolchain/go-version.txt")
test_go_root=$("$script_dir/go.sh" env GOROOT)
test_go_command="$test_go_root/bin/go"

unset GOROOT
export GOTOOLCHAIN=local
export GOEXPERIMENT=cgocheck2
export CGO_ENABLED=1

if [ -z "${CGO_CFLAGS_ALLOW:-}" ]; then
	export CGO_CFLAGS_ALLOW='-f.*'
fi
if [ -z "${PKG_CONFIG_PATH:-}" ] && [ -d /opt/homebrew/opt/openal-soft/lib/pkgconfig ]; then
	export PKG_CONFIG_PATH=/opt/homebrew/opt/openal-soft/lib/pkgconfig
fi

actual_go=$("$test_go_command" env GOVERSION)
actual_experiment=$("$test_go_command" env GOEXPERIMENT)
actual_cgo=$("$test_go_command" env CGO_ENABLED)
if [ "$actual_go" != "$required_go" ] || [ "$actual_experiment" != cgocheck2 ] || [ "$actual_cgo" != 1 ]; then
	echo "Strict CGo test configuration mismatch: $actual_go GOEXPERIMENT=$actual_experiment CGO_ENABLED=$actual_cgo" >&2
	exit 2
fi
echo "Strict CGo tests: $actual_go GOEXPERIMENT=$actual_experiment CGO_ENABLED=$actual_cgo" >&2

if [ "$#" -eq 0 ]; then
	set -- ./...
fi
exec "$test_go_command" -C "$repo_dir/src" test "$@"
