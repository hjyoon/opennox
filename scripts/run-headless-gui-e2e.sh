#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 || $# -gt 3 ]]; then
	echo "usage: $0 NOX_DATA_DIR E2E_YAML [OUTPUT_DIR]" >&2
	exit 2
fi

script_dir="$(cd "$(dirname "$0")" && pwd -P)"
repo_dir="$(cd "$script_dir/.." && pwd -P)"
src_dir="$repo_dir/src"
go_cmd="$script_dir/go.sh"
data_dir="$(cd "$1" && pwd -P)"
scenario_dir="$(cd "$(dirname "$2")" && pwd -P)"
scenario="$scenario_dir/$(basename "$2")"
output_dir="${3:-${TMPDIR:-/tmp}/opennox-headless-gui-e2e}"

if [[ ! -f "$data_dir/GAME.EXE" || ! -f "$data_dir/thing.bin" ]]; then
	echo "error: $data_dir is not a complete Nox data directory" >&2
	exit 1
fi
if [[ ! -f "$scenario" ]]; then
	echo "error: E2E scenario not found: $scenario" >&2
	exit 1
fi
if [[ "$($go_cmd env GOVERSION)" != "go1.26.5" ]]; then
	echo "error: Go 1.26.5 is required" >&2
	exit 1
fi

target_os="$($go_cmd env GOOS)"
target_arch="$($go_cmd env GOARCH)"

"$go_cmd" -C "$src_dir" run ./internal/noxbuild \
	-go="$go_cmd" \
	-os="$target_os" \
	-arch="$target_arch" \
	-o="$output_dir" \
	client

cd "$data_dir"
runtime_args=(
	-config "$output_dir/opennox.yml"
	-data "$data_dir"
	-window
)
if [[ "${NOX_E2E_AUDIO_HANDLES:-}" != "true" ]]; then
	runtime_args+=(-noaudio)
fi
NOX_E2E="$scenario" exec "$output_dir/opennox" "${runtime_args[@]}"
