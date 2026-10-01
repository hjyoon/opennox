#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 || $# -gt 2 ]]; then
	echo "usage: $0 NOX_DATA_DIR [OUTPUT_DIR]" >&2
	exit 2
fi
for command in curl jq; do
	if ! command -v "$command" >/dev/null; then
		echo "error: $command is required" >&2
		exit 1
	fi
done

script_dir="$(cd "$(dirname "$0")" && pwd -P)"
repo_dir="$(cd "$script_dir/.." && pwd -P)"
go_cmd="$script_dir/go.sh"
data_dir="$(cd "$1" && pwd -P)"
output_dir="${2:-${TMPDIR:-/tmp}/opennox-headless-server-e2e}"
mkdir -p "$output_dir"
output_dir="$(cd "$output_dir" && pwd -P)"
if [[ ! -f "$data_dir/GAME.EXE" || ! -f "$data_dir/thing.bin" ]]; then
	echo "error: $data_dir is not a complete Nox data directory" >&2
	exit 1
fi
if [[ "$("$go_cmd" env GOVERSION)" != "go1.26.5" ]]; then
	echo "error: Go 1.26.5 is required" >&2
	exit 1
fi

server_port="${NOX_E2E_SERVER_PORT:-18610}"
metrics_port="${NOX_E2E_SERVER_METRICS_PORT:-6062}"
for port in "$server_port" "$metrics_port"; do
	case "$port" in
	'' | *[!0-9]*) echo "error: invalid test port: $port" >&2; exit 2 ;;
	esac
	if ((10#$port < 1 || 10#$port > 65535)); then
		echo "error: test port is out of range: $port" >&2
		exit 2
	fi
done
if [[ "$server_port" == "$metrics_port" ]]; then
	echo "error: server and metrics ports must differ" >&2
	exit 2
fi

server_binary="${NOX_E2E_SERVER_BINARY:-}"
if [[ -z "$server_binary" ]]; then
	target_os="$("$go_cmd" env GOOS)"
	target_arch="$("$go_cmd" env GOARCH)"
	"$go_cmd" -C "$repo_dir/src" run ./internal/noxbuild \
		-go="$go_cmd" -os="$target_os" -arch="$target_arch" \
		-o="$output_dir" server
	server_binary="$output_dir/opennox-server"
fi
server_binary="$(cd "$(dirname "$server_binary")" && pwd -P)/$(basename "$server_binary")"
if [[ ! -x "$server_binary" ]]; then
	echo "error: server binary is not executable: $server_binary" >&2
	exit 1
fi

# Match the GUI runner's isolated, shallow map view: no personal saves/config,
# and no generated user.rul writes through a maps-directory symlink.
runtime_data_dir="$(mktemp -d "$output_dir/runtime-data.XXXXXX")"
while IFS= read -r -d '' source_path; do
	entry_name="$(basename "$source_path")"
	ln -s "$source_path" "$runtime_data_dir/$entry_name"
done < <(find "$data_dir" -mindepth 1 -maxdepth 1 \
	! -iname save ! -iname nox.cfg ! -iname opennox.yml \
	! -iname nc.obj ! -iname maps -print0)
mkdir "$runtime_data_dir/Save" "$runtime_data_dir/maps"
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

base_url="http://127.0.0.1:$server_port"
metrics_url="http://127.0.0.1:$metrics_port/metrics"
# This token authorizes only this isolated test server; it is not a user secret.
test_token="opennox-e2e-local-$$-$RANDOM"
server_log="$output_dir/server.log"
server_pid=""
# shellcheck disable=SC2329 # Invoked indirectly by the EXIT trap below.
cleanup_server() {
	if [[ -n "$server_pid" ]] && kill -0 "$server_pid" 2>/dev/null; then
		kill -TERM "$server_pid" 2>/dev/null || true
		wait "$server_pid" || true
	fi
}
trap 'cleanup_server' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# Refuse to send map/quit requests to an already-running service.
if curl --silent --max-time 1 "$base_url/api/v0/game/info" >/dev/null 2>&1 \
	|| curl --silent --max-time 1 "$metrics_url" >/dev/null 2>&1; then
	echo "error: test ports already have an HTTP service; choose unused ports" >&2
	exit 1
fi
(
	cd "$runtime_data_dir"
	exec env -u NOX_DATA \
		NOX_E2E="$script_dir/e2e/dedicated-server-idle.yaml" \
		NOX_E2E_SEAT=headless NOX_E2E_AUDIO=mock NOX_E2E_OVERRIDE=true \
		NOX_XWIS=false NOX_NET_NAT=false NOX_API_TOKEN="$test_token" \
		"$server_binary" -config "$runtime_data_dir/opennox.yml" \
		-data "$runtime_data_dir" -port "$server_port" \
		-pprof "127.0.0.1:$metrics_port" -noaudio
) >"$server_log" 2>&1 &
server_pid=$!
echo "server runtime data: $runtime_data_dir"
echo "server log: $server_log"

assert_map() {
	local expected_map="$1" expected_mode="$2" game_info attempt
	for ((attempt = 0; attempt < 150; attempt++)); do
		if ! kill -0 "$server_pid" 2>/dev/null; then
			echo "error: server exited before $expected_map was ready" >&2
			return 1
		fi
		if game_info="$(curl --fail --silent --max-time 1 "$base_url/api/v0/game/info")" \
			&& jq -e --arg map "$expected_map" --arg mode "$expected_mode" \
			'.map == $map and .mode == $mode and .players.cur == 0 and .players.max == 32' \
			<<<"$game_info" >/dev/null; then
			echo "game info: $game_info"
			return 0
		fi
		sleep 0.1
	done
	echo "error: timed out waiting for map $expected_map in mode $expected_mode" >&2
	return 1
}
counter_value() {
	curl --fail --silent --show-error --max-time 3 "$metrics_url" \
		| awk -v metric="$1" '$1 == metric { print $2; found = 1 } END { if (!found) exit 1 }'
}
assert_loop_progress() {
	local before_loop after_loop before_tick after_tick
	before_loop="$(counter_value nox_mainloop)"
	before_tick="$(counter_value nox_game_tick)"
	sleep 0.3
	after_loop="$(counter_value nox_mainloop)"
	after_tick="$(counter_value nox_game_tick)"
	if ! awk -v before_loop="$before_loop" -v after_loop="$after_loop" \
		-v before_tick="$before_tick" -v after_tick="$after_tick" \
		'BEGIN { exit !(after_loop > before_loop && after_tick > before_tick) }'; then
		echo "error: dedicated loop/game tick did not advance: $before_loop/$before_tick -> $after_loop/$after_tick" >&2
		return 1
	fi
	echo "main loop/game tick advance: $before_loop/$before_tick -> $after_loop/$after_tick"
}
assert_map_played() {
	local map="$1" expected="$2" count
	count="$(counter_value "nox_map_played{map=\"$map.map\"}")"
	if [[ "$count" != "$expected" ]]; then
		echo "error: map $map load count is $count, expected $expected" >&2
		return 1
	fi
	echo "map load count: $map=$count"
}
post_api() {
	curl --fail --silent --show-error --max-time 3 -X POST \
		-H "X-Token: $test_token" --data "$2" "$base_url/api/v0/game/$1"
}

assert_map so_beach chat
assert_loop_progress
assert_map_played so_beach 1
post_api map estate
assert_map estate arena
assert_loop_progress
assert_map_played estate 1
post_api map trilevel
assert_map trilevel arena
assert_loop_progress
assert_map_played trilevel 1
post_api map estate
assert_map estate arena
assert_loop_progress
assert_map_played estate 2

# Real console quit must terminate the process, not merely stop game updates.
post_api cmd quit
for ((attempt = 0; attempt < 100; attempt++)); do
	if ! kill -0 "$server_pid" 2>/dev/null; then
		if wait "$server_pid"; then
			server_pid=""
			echo "dedicated server E2E passed: startup, three map transitions, quit exit=0"
			exit 0
		fi
		echo "error: dedicated server quit returned a nonzero exit status" >&2
		exit 1
	fi
	sleep 0.1
done
echo "error: dedicated server kept running after console quit" >&2
exit 1
