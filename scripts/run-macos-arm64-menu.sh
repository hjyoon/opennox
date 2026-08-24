#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
	echo "usage: $0 NOX_DATA_DIR [OUTPUT_DIR]" >&2
	exit 2
fi

if [[ "$(uname -s)" != "Darwin" || "$(uname -m)" != "arm64" ]]; then
	echo "error: this smoke launcher requires macOS on arm64" >&2
	exit 1
fi

script_dir="$(cd "$(dirname "$0")" && pwd -P)"
repo_dir="$(cd "$script_dir/.." && pwd -P)"
src_dir="$repo_dir/src"
go_cmd="$script_dir/go.sh"
data_dir="$(cd "$1" && pwd -P)"
output_dir="${2:-${TMPDIR:-/tmp}/opennox-macos-arm64-menu}"

if [[ ! -f "$data_dir/GAME.EXE" || ! -f "$data_dir/thing.bin" ]]; then
	echo "error: $data_dir is not a complete Nox data directory" >&2
	exit 1
fi

go_version="$($go_cmd version)"
if [[ "$go_version" != *"go1.26.5 "* ]]; then
	echo "error: expected Go 1.26.5, got: $go_version" >&2
	exit 1
fi

if [[ -z "${PKG_CONFIG_PATH:-}" && -d /opt/homebrew/opt/openal-soft/lib/pkgconfig ]]; then
	export PKG_CONFIG_PATH=/opt/homebrew/opt/openal-soft/lib/pkgconfig
fi

cd "$src_dir"
"$go_cmd" run ./internal/noxbuild \
	-go="$go_cmd" \
	-os=darwin \
	-arch=arm64 \
	-o="$output_dir" \
	client

cd "$data_dir"
exec "$output_dir/opennox" \
	-config "$output_dir/opennox.yml" \
	-data "$data_dir" \
	-window \
	-noaudio
