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

# Keep persistent player saves and generated configuration out of the source
# data tree. Every other top-level entry remains a symlink to the exact input
# data, so E2E still exercises the caller-supplied GAME.EXE and game assets.
temp_root="$(cd "${TMPDIR:-/tmp}" && pwd -P)"
runtime_data_dir="$(mktemp -d "$temp_root/opennox-headless-e2e-data.XXXXXX")"
cleanup_runtime_data() {
	case "$runtime_data_dir" in
		"$temp_root"/opennox-headless-e2e-data.*)
			rm -rf -- "$runtime_data_dir"
			;;
	esac
}
trap cleanup_runtime_data EXIT

while IFS= read -r -d '' source_path; do
	entry_name="$(basename "$source_path")"
	ln -s "$source_path" "$runtime_data_dir/$entry_name"
done < <(find "$data_dir" -mindepth 1 -maxdepth 1 \
	! -iname save ! -iname nox.cfg ! -iname opennox.yml -print0)
mkdir "$runtime_data_dir/Save"

cd "$runtime_data_dir"
runtime_args=(
	-config "$output_dir/opennox.yml"
	-data "$runtime_data_dir"
	-window
)
if [[ "${NOX_E2E_AUDIO_HANDLES:-}" != "true" ]]; then
	runtime_args+=(-noaudio)
fi
NOX_E2E="$scenario" "$output_dir/opennox" "${runtime_args[@]}"
