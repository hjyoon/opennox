#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(dirname "$script_dir")
required_go=$(tr -d '[:space:]' < "$repo_dir/toolchain/go-version.txt")
go_command=${GO:-go}

unset GOROOT
export GOTOOLCHAIN="$required_go"
export GOEXPERIMENT=

actual_go=$("$go_command" env GOVERSION)
if [ "$actual_go" != "$required_go" ]; then
	echo "Go toolchain mismatch: required $required_go, got $actual_go" >&2
	exit 2
fi

exec "$go_command" "$@"
