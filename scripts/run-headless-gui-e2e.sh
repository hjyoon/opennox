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
mkdir -p "$output_dir"
output_dir="$(cd "$output_dir" && pwd -P)"

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
# data tree. Map directories get a shallow symlink view so the game can create
# user.rul without writing through a maps/ symlink into the caller's data. All
# immutable files still resolve to the exact input GAME.EXE and game assets.
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
	! -iname save ! -iname nox.cfg ! -iname opennox.yml \
	! -iname nc.obj ! -iname maps -print0)
mkdir "$runtime_data_dir/Save"
mkdir "$runtime_data_dir/maps"
while IFS= read -r -d '' source_map_path; do
	map_name="$(basename "$source_map_path")"
	if [[ ! -d "$source_map_path" ]]; then
		ln -s "$source_map_path" "$runtime_data_dir/maps/$map_name"
		continue
	fi
	mkdir "$runtime_data_dir/maps/$map_name"
	while IFS= read -r -d '' source_map_file; do
		file_name="$(basename "$source_map_file")"
		case "$file_name" in
		[Uu][Ss][Ee][Rr].[Rr][Uu][Ll]) continue ;;
		esac
		ln -s "$source_map_file" "$runtime_data_dir/maps/$map_name/$file_name"
	done < <(find "$source_map_path" -mindepth 1 -maxdepth 1 -print0)
done < <(find "$data_dir/maps" -mindepth 1 -maxdepth 1 -print0)

cd "$runtime_data_dir"
runtime_args=(
	-config "$output_dir/opennox.yml"
	-data "$runtime_data_dir"
	-window
)
if [[ "${NOX_E2E_AUDIO_HANDLES:-}" != "true" ]]; then
	runtime_args+=(-noaudio)
fi

# A scenario that exercises loading an existing save may name a second E2E
# scenario which creates that save in the same isolated runtime directory.
# This keeps the result deterministic and prevents tests from importing the
# caller's personal Save directory.
seed_scenario="$(awk '/^seed-scenario:[[:space:]]*/ { sub(/^[^:]*:[[:space:]]*/, ""); sub(/[[:space:]]*(#.*)?$/, ""); print; exit }' "$scenario")"
if [[ -n "$seed_scenario" ]]; then
	case "$seed_scenario" in
	/*) seed_scenario_path="$seed_scenario" ;;
	*) seed_scenario_path="$scenario_dir/$seed_scenario" ;;
	esac
	if [[ ! -f "$seed_scenario_path" ]]; then
		echo "error: seed scenario not found: $seed_scenario_path" >&2
		exit 1
	fi
	echo "seeding isolated E2E state with: $seed_scenario_path"
	NOX_E2E="$seed_scenario_path" "$output_dir/opennox" "${runtime_args[@]}"
	seed_player="$(find "$runtime_data_dir/Save" -type f -path '*/AUTOSAVE/*' -iname 'player.plr' -size +0c -print -quit)"
	if [[ -z "$seed_player" ]]; then
		echo "error: seed scenario did not create a non-empty AUTOSAVE/Player.plr" >&2
		exit 1
	fi
fi
NOX_E2E="$scenario" "$output_dir/opennox" "${runtime_args[@]}"
