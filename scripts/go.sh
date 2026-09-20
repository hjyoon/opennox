#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(dirname "$script_dir")
required_go=$(tr -d '[:space:]' < "$repo_dir/toolchain/go-version.txt")
go_command=${GO:-go}

unset GOROOT
export GOTOOLCHAIN="$required_go"
export GOEXPERIMENT=

# The legacy C sources intentionally use ABI flags such as -fshort-wchar and
# -fsigned-char. Keep direct go build/test invocations consistent with
# internal/noxbuild and the Makefile while still allowing callers to provide a
# stricter policy explicitly.
if [ -z "${CGO_CFLAGS_ALLOW:-}" ]; then
	export CGO_CFLAGS_ALLOW='-f.*'
fi

# Homebrew keeps openal-soft's pkg-config metadata in its opt prefix instead
# of the default Apple Silicon search path. Make direct Go commands behave the
# same way as the macOS launcher while preserving an explicit caller setting.
if [ -z "${PKG_CONFIG_PATH:-}" ] && [ -d /opt/homebrew/opt/openal-soft/lib/pkgconfig ]; then
	export PKG_CONFIG_PATH=/opt/homebrew/opt/openal-soft/lib/pkgconfig
fi

actual_go=$("$go_command" env GOVERSION)
if [ "$actual_go" != "$required_go" ]; then
	echo "Go toolchain mismatch: required $required_go, got $actual_go" >&2
	exit 2
fi

exec "$go_command" "$@"
