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
case "${NOX_E2E_MAP_PROFILE:-all}" in
all|campaign1to3) ;;
*) echo "error: NOX_E2E_MAP_PROFILE must be all or campaign1to3" >&2; exit 2 ;;
esac
seat_backend="${NOX_E2E_SEAT:-headless}"
case "$seat_backend" in
headless | sdl) ;;
*)
	echo "error: NOX_E2E_SEAT must be headless or sdl (got: $seat_backend)" >&2
	exit 2
	;;
esac
audio_backend="${NOX_E2E_AUDIO:-mock}"
case "$audio_backend" in
mock | openal) ;;
*)
	echo "error: NOX_E2E_AUDIO must be mock or openal (got: $audio_backend)" >&2
	exit 2
	;;
esac
if [[ "$audio_backend" == openal && "${NOX_E2E_AUDIO_HANDLES:-}" == true ]]; then
	echo "error: NOX_E2E_AUDIO=openal cannot be combined with mock NOX_E2E_AUDIO_HANDLES" >&2
	exit 2
fi
client_target="${NOX_E2E_CLIENT_TARGET:-client}"
case "$client_target" in
client)
	client_binary="opennox"
	;;
client-hd)
	client_binary="opennox-hd"
	;;
*)
	echo "error: NOX_E2E_CLIENT_TARGET must be client or client-hd (got: $client_target)" >&2
	exit 2
	;;
esac

echo "E2E seat: $seat_backend"
echo "E2E audio: $audio_backend"

if [[ -n "${NOX_E2E_BINARY:-}" ]]; then
	client_executable="$(cd "$(dirname "$NOX_E2E_BINARY")" && pwd -P)/$(basename "$NOX_E2E_BINARY")"
	if [[ ! -f "$client_executable" || ! -x "$client_executable" ]]; then
		echo "error: NOX_E2E_BINARY must be an executable file" >&2
		exit 2
	fi
	echo "E2E existing binary: $client_executable"
else
	required_go="$(tr -d '[:space:]' < "$repo_dir/toolchain/go-version.txt")"
	if [[ "$("$go_cmd" env GOVERSION)" != "$required_go" ]]; then
		echo "error: $required_go is required" >&2
		exit 1
	fi
	target_os="$($go_cmd env GOOS)"
	target_arch="$($go_cmd env GOARCH)"
	"$go_cmd" -C "$src_dir" run ./internal/noxbuild \
		-go="$go_cmd" -os="$target_os" -arch="$target_arch" \
		-o="$output_dir" "$client_target"
	client_executable="$output_dir/$client_binary"
fi

# Keep persistent player saves and generated configuration out of the source
# data tree. Map assets get private copies so runtime saves cannot write through
# symlinks and the map HTTP server can archive regular files. Other immutable
# files still resolve to the exact input GAME.EXE and game assets.
temp_root="$(cd "${TMPDIR:-/tmp}" && pwd -P)"
if [[ "${NOX_E2E_KEEP_RUNTIME:-false}" == true ]]; then
	runtime_data_dir="$(mktemp -d "$output_dir/runtime-data.XXXXXX")"
else
	runtime_data_dir="$(mktemp -d "$temp_root/opennox-headless-e2e-data.XXXXXX")"
fi
cleanup_runtime_data() {
	if [[ "${NOX_E2E_KEEP_RUNTIME:-false}" == true ]]; then return; fi
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
	if [[ "${NOX_E2E_MAP_PROFILE:-all}" == campaign1to3 ]]; then
		case "$(printf '%s' "$map_name" | tr '[:upper:]' '[:lower:]')" in
		war01a|war02a|war02b|war03a|war03b|war03c|war03d|con01a|con02a|con03a|con03b|wiz01a|wiz02a|wiz02b|wiz02c|wiz03a|wiz03b|wiz03c) ;;
		*) continue ;;
		esac
	fi
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
		cp -Rp "$source_map_file" "$runtime_data_dir/maps/$map_name/$file_name"
	done < <(find "$source_map_path" -mindepth 1 -maxdepth 1 -print0)
done < <(find "$data_dir/maps" -mindepth 1 -maxdepth 1 -print0)

if [[ -n "${NOX_E2E_SAVE_SEED:-}" ]]; then
	save_seed="$(cd "$NOX_E2E_SAVE_SEED" && pwd -P)"
	if [[ -n "$(find "$save_seed" -type l -print -quit)" ]]; then
		echo "error: save seed must contain regular files and directories" >&2
		exit 2
	fi
	cp -Rp "$save_seed/." "$runtime_data_dir/Save/"
fi
if [[ -n "${NOX_E2E_CONFIG_SEED:-}" ]]; then
	if [[ ! -f "$NOX_E2E_CONFIG_SEED" || -L "$NOX_E2E_CONFIG_SEED" ]]; then
		echo "error: NOX_E2E_CONFIG_SEED must be a regular nox.cfg file" >&2
		exit 2
	fi
	cp -p "$NOX_E2E_CONFIG_SEED" "$runtime_data_dir/nox.cfg"
fi

cd "$runtime_data_dir"
printf '%s\n' "$runtime_data_dir" > "$output_dir/runtime-data-path.txt"
runtime_args=(
	-config "$output_dir/opennox.yml"
	-data "$runtime_data_dir"
	-window
)
if [[ -n "${NOX_E2E_PORT:-}" ]]; then
	if [[ ! "$NOX_E2E_PORT" =~ ^[0-9]+$ ]] || (( NOX_E2E_PORT < 1024 || NOX_E2E_PORT >= 65535 )); then
		echo "error: NOX_E2E_PORT must be 1024..65534" >&2
		exit 2
	fi
	runtime_args+=(-port "$NOX_E2E_PORT" -clientport 0 -pprof "127.0.0.1:$((NOX_E2E_PORT+1))" -rcon "" -autoexec "")
fi
if [[ "$audio_backend" == mock && "${NOX_E2E_AUDIO_HANDLES:-}" != "true" ]]; then
	runtime_args+=(-noaudio)
fi

# A scenario that exercises loading an existing save may name a second E2E
# scenario which creates that save in the same isolated runtime directory.
# This keeps the result deterministic and prevents tests from importing the
# caller's personal Save directory.
seed_scenario="$(awk '/^seed-scenario:[[:space:]]*/ { sub(/^[^:]*:[[:space:]]*/, ""); sub(/[[:space:]]*(#.*)?$/, ""); print; exit }' "$scenario")"
if [[ -n "$seed_scenario" && -n "${NOX_E2E_SAVE_SEED:-}" ]]; then
	echo "error: select either seed-scenario or NOX_E2E_SAVE_SEED" >&2
	exit 2
fi
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
	NOX_E2E="$seed_scenario_path" "$client_executable" "${runtime_args[@]}"
	seed_player="$(find "$runtime_data_dir/Save" -type f -path '*/AUTOSAVE/*' -iname 'player.plr' -size +0c -print -quit)"
	if [[ -z "$seed_player" ]]; then
		echo "error: seed scenario did not create a non-empty AUTOSAVE/Player.plr" >&2
		exit 1
	fi
fi
NOX_E2E="$scenario" "$client_executable" "${runtime_args[@]}"
