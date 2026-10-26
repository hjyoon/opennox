#!/usr/bin/env python3
"""Run saved campaign regression scenarios against one prebuilt binary.

Example:
  python3 scripts/run-campaign-e2e.py \
    --suite scripts/e2e/campaign/wizard2.json --binary ./bin/opennox \
    --data /path/to/Nox --seed /path/to/earned/Save --output ./artifacts/wizard2

The supplied Save is a read-only checkpoint earned through normal gameplay.
Each case gets its own copy. Runtime saves, screenshots, logs and read-only
world/script/wall samples stay under output, including failures. A diagnostic
case can allow poison; it is never reported as a healthy continuation.
"""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import queue
import re
import shutil
import signal
import socket
import subprocess
import sys
import threading
import time
import urllib.request

MAPS = set('war01a war02a war02b war03a war03b war03c war03d con01a con02a con03a con03b wiz01a wiz02a wiz02b wiz02c wiz03a wiz03b wiz03c'.split())
INPUT_ACTIONS = set('slow click interact move raw wait walk run run-start run-dir run-stop walk-start walk-dir walk-stop run-to-world walk-to-world wait-map wait-player-reloaded select-manual-campaign-save click-dialog-yes click-npc-dialog-done assert-quickbar-spell assert-spell-set-row observe-player-target capture-quest-map-audit-frame cast-at-nearest-monster-input screen quit'.split())
ROUTES = {'objects': '/debug/nox/objects', 'script': '/debug/nox/script/legacy', 'walls': '/debug/nox/walls'}


def sha(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def tree(path):
    result = {}
    for file in sorted(path.rglob('*')):
        if file.is_symlink():
            raise ValueError(f'input contains a symbolic link: {file}')
        if file.is_file():
            result[str(file.relative_to(path))] = {'bytes': file.stat().st_size, 'sha256': sha(file)}
    return result


def put(path, value):
    with path.open('x') as stream:
        json.dump(value, stream, indent=2, ensure_ascii=False)
        stream.write('\n')


def validate_scenario(path):
    # JSON flow mappings are valid YAML and let this runner validate inputs
    # without a second YAML parser or any Go compilation.
    steps = []
    for line in path.read_text().splitlines():
        if not line.strip() or line.lstrip().startswith('#') or line == 'steps:':
            continue
        if not line.startswith('  - {'):
            raise ValueError(f'campaign scenario requires JSON flow steps: {path}')
        step = json.loads(line[4:])
        if step.get('action') not in INPUT_ACTIONS:
            raise ValueError(f'non-input or unknown campaign action: {step.get("action")}')
        if 'dir' in step:
            raise ValueError('movement uses ang, not dir')
        if step.get('action') == 'wait-map' and step.get('map', '').lower() not in MAPS:
            raise ValueError('campaign map must be in chapters 1 through 3')
        steps.append(step)
    if not steps or steps[-1]['action'] != 'quit':
        raise ValueError('campaign scenario must finish with quit')
    return steps


def check_ports(port):
    sockets = []
    try:
        for number in (port, port + 1):
            if number == 22000 or not 1024 <= number <= 65535:
                raise ValueError('invalid or reserved game/debug port')
            for kind in (socket.SOCK_STREAM, socket.SOCK_DGRAM):
                sock = socket.socket(socket.AF_INET, kind)
                sockets.append(sock)
                sock.bind(('127.0.0.1', number))
    finally:
        for sock in sockets:
            sock.close()


def created_monster_history(samples, type_id):
    initial_ids = {obj.get('script_id') for obj in samples[0]}
    created = set()
    dead = set()
    observed_dead = set()
    for objects in samples[1:]:
        for obj in objects:
            ident = obj.get('script_id')
            if ident is None or ident in initial_ids or obj.get('type', {}).get('id') != type_id:
                continue
            created.add(ident)
            if 'DEAD' in (obj.get('flags') or []):
                dead.add(ident)
                observed_dead.add(ident)
            else:
                dead.discard(ident)
    final_ids = {obj.get('script_id') for obj in samples[-1]}
    return {'type': type_id, 'count': len(created), 'dead': len(dead),
            'script_ids': sorted(created), 'observed_dead_script_ids': sorted(observed_dead),
            'last_observed_dead_script_ids': sorted(dead),
            'absent_from_final_script_ids': sorted(created - final_ids)}


def required_monster_deaths(samples, required_ids):
    required = set(required_ids)
    initial_alive = {obj.get('script_id') for obj in samples[0]
                     if 'MONSTER' in (obj.get('class') or [])
                     and 'DEAD' not in (obj.get('flags') or [])} & required
    dead = set()
    observed_dead = set()
    for objects in samples[1:]:
        for obj in objects:
            ident = obj.get('script_id')
            if ident not in initial_alive:
                continue
            if 'DEAD' in (obj.get('flags') or []):
                dead.add(ident)
                observed_dead.add(ident)
            else:
                dead.discard(ident)
    return {'required_script_ids': sorted(required),
            'initial_alive_script_ids': sorted(initial_alive),
            'observed_dead_script_ids': sorted(observed_dead),
            'last_observed_dead_script_ids': sorted(dead),
            'passed': bool(required) and required <= dead}


def required_enable_transitions(samples, requirements):
    def enabled(objects, ident):
        matches = [obj for obj in objects if obj.get('script_id') == ident]
        return matches[0].get('enabled') if len(matches) == 1 else None

    transitions = []
    for required in requirements:
        ident = required['script_id']
        initial = enabled(samples[0], ident) if samples else None
        final = enabled(samples[-1], ident) if samples else None
        before, after = required['before'], required['after']
        passed = (type(before) is bool and type(after) is bool and before != after
                  and initial is before and final is after)
        transitions.append({'script_id': ident, 'expected_initial_enabled': before,
                            'expected_final_enabled': after, 'initial_enabled': initial,
                            'final_enabled': final, 'passed': passed})
    return {'transitions': transitions,
            'passed': bool(transitions) and all(row['passed'] for row in transitions)}


def required_plate_holds(samples, requirements):
    def position(objects, ident):
        matches = [obj for obj in objects if obj.get('script_id') == ident]
        point = matches[0].get('pos', {}) if len(matches) == 1 else {}
        values = (point.get('x'), point.get('y'))
        return values if all(type(v) in (int, float) and math.isfinite(v) for v in values) else None

    holds = []
    for required in requirements:
        maximum = required['maximum_weight_distance']
        minimum = required['minimum_player_distance']
        drift = required['maximum_weight_drift']
        valid_bounds = (all(type(v) in (int, float) and math.isfinite(v) for v in (maximum, minimum, drift))
                        and 0 < maximum < minimum and drift >= 0)
        observations = []
        for objects in samples:
            weight = position(objects, required['weight_script_id'])
            plate = position(objects, required['plate_script_id'])
            player = position(objects, required['player_script_id'])
            observations.append((weight, math.dist(weight, plate), math.dist(player, plate))
                                if weight and plate and player else None)
        initial, final = (observations[0], observations[-1]) if observations else (None, None)
        separated = [i for i, row in enumerate(observations)
                     if valid_bounds and row and row[1] <= maximum and row[2] >= minimum]
        stable = bool(final and separated) and all(
            row and row[1] <= maximum and math.dist(row[0], final[0]) <= drift
            for row in observations[separated[0]:])
        passed = bool(valid_bounds and initial and final and initial[1] > maximum
                      and final[1] <= maximum and final[2] >= minimum
                      and len(separated) >= 2 and stable)
        holds.append(dict(required, initial_weight_distance=initial[1] if initial else None,
                          final_weight_distance=final[1] if final else None,
                          final_player_distance=final[2] if final else None,
                          separated_samples=len(separated), stable_after_separation=stable, passed=passed))
    return {'holds': holds, 'passed': bool(holds) and all(row['passed'] for row in holds)}


def run_case(args, binary, suite_dir, case, port):
    case_dir = args.output / case['name']
    case_dir.mkdir()
    frames = case_dir / 'frames'
    frames.mkdir()
    scenario = (suite_dir / case['scenario']).resolve()
    validate_scenario(scenario)
    check_ports(port)
    env = {k: v for k, v in os.environ.items() if not k.startswith('NOX_E2E') and k != 'NOX_DATA'}
    env.update(NOX_E2E_BINARY=str(binary), NOX_E2E_SAVE_SEED=str(args.seed),
               NOX_E2E_KEEP_RUNTIME='true', NOX_E2E_MAP_PROFILE='campaign1to3',
               NOX_E2E_CLOCK='ordinary', NOX_E2E_OVERRIDE='false', NOX_E2E_FAILFAST='true',
               NOX_E2E_SEAT='headless', NOX_E2E_AUDIO='mock', NOX_E2E_PORT=str(port),
               NOX_NET_NAT='false', NOX_XWIS='false', TMPDIR=str(frames), GOMAXPROCS='2')
    if args.config_seed:
        env['NOX_E2E_CONFIG_SEED'] = str(args.config_seed)
    command = ['/bin/bash', str(Path(__file__).with_name('run-headless-gui-e2e.sh')),
               str(args.data), str(scenario), str(case_dir)]
    if args.sandbox_profile:
        command = ['/usr/bin/sandbox-exec', '-f', str(args.sandbox_profile), *command]
    report = {'name': case['name'], 'scenario': str(scenario), 'scenario_sha256': sha(scenario),
              'port': port, 'clock': 'ordinary', 'command': command, 'snapshots': [],
              'player_observations': [], 'observed_maps': [], 'forced_termination': False}
    http = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    captures = queue.Queue()
    checkpoints = []
    errors = []

    def capture():
        while True:
            line = captures.get()
            try:
                if line is None:
                    return
                number = len(report['snapshots'])
                sample = {'after_line': line.strip(), 'routes': {}}
                for label, route in ROUTES.items():
                    row = {'method': 'GET', 'route': route, 'observation_only': True}
                    try:
                        with http.open(f'http://127.0.0.1:{port + 1}{route}', timeout=3) as response:
                            data = response.read(16 * 1024 * 1024 + 1)
                            row['status'] = response.status
                        if row['status'] != 200 or len(data) > 16 * 1024 * 1024:
                            raise ValueError('incomplete snapshot response')
                        json.loads(data)
                        path = case_dir / f'{label}-{number}.json'
                        path.write_bytes(data)
                        row.update(path=str(path), bytes=len(data), sha256=sha(path))
                    except Exception as error:
                        row['error'] = str(error)
                    sample['routes'][label] = row
                report['snapshots'].append(sample)
            finally:
                captures.task_done()

    started = time.monotonic()
    proc = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                            env=env, start_new_session=True)
    report['owned_process_group'] = proc.pid
    worker = threading.Thread(target=capture, daemon=True)
    worker.start()

    def read_output():
        with (case_dir / 'game.log').open('xb') as log:
            for raw in iter(proc.stdout.readline, b''):
                log.write(raw)
                log.flush()
                line = raw.decode(errors='replace')
                if 'PLAYER TARGET OBSERVATION: ' in line:
                    try:
                        report['player_observations'].append(json.loads(line.split('PLAYER TARGET OBSERVATION: ', 1)[1]))
                    except ValueError:
                        errors.append('invalid player observation')
                match = re.search(r'MAP READY: map="([^"]+)"', line)
                if match:
                    report['observed_maps'].append(match[1].lower())
                if 'QUEST MAP AUDIT FRAME:' in line:
                    checkpoints.append(line)
                    captures.put(line)
                if 'panic:' in line or 'fatal error:' in line or 'not ported' in line or 'event: "PlayerDeath"' in line:
                    errors.append(line.strip())

    reader = threading.Thread(target=read_output, daemon=True)
    reader.start()
    heartbeat = started
    while proc.poll() is None:
        elapsed = time.monotonic() - started
        if errors or elapsed > case.get('timeout_seconds', 420):
            report['forced_termination'] = True
            report['abort_reason'] = 'runtime_error' if errors else 'timeout'
            try:
                os.killpg(proc.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                try:
                    os.killpg(proc.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
            break
        if time.monotonic() - heartbeat >= 30:
            print(json.dumps({'case': case['name'], 'status': 'running', 'elapsed_seconds': round(elapsed, 1)}), flush=True)
            heartbeat = time.monotonic()
        time.sleep(.2)
    report['exit_code'] = proc.wait()
    reader.join(timeout=5)
    captures.put(None)
    worker.join(timeout=12)
    report['elapsed_seconds'] = round(time.monotonic() - started, 3)
    report['errors'] = errors
    report['checkpoints'] = checkpoints
    player = report['player_observations'][-1] if report['player_observations'] else None
    actual_maps = set(report['observed_maps'])
    expected_maps = {m.lower() for m in case['expected_maps']}
    snapshots_ok = bool(report['snapshots']) and all(
        row.get('status') == 200 and 'error' not in row
        for sample in report['snapshots'] for row in sample['routes'].values())
    healthy = bool(player and player['player'].get('health', 0) > 0)
    poison_free = bool(player and player.get('poison') == 0)
    runtime_file = case_dir / 'runtime-data-path.txt'
    if runtime_file.exists():
        runtime = Path(runtime_file.read_text().strip())
        report['runtime_data'] = str(runtime)
        saved = runtime / 'Save' / case.get('manual_save', 'SAVE0001/Player.plr')
        report['manual_save_written'] = saved.is_file() and sha(saved) != sha(args.seed / case.get('manual_save', 'SAVE0001/Player.plr'))
        if saved.is_file():
            report['saved_player_sha256'] = sha(saved)
    required = all(any(text in line for line in checkpoints) for text in case.get('required_checkpoints', []))
    outcomes_ok = True
    if player:
        outcomes_ok = (player['player']['health'] == case.get('expected_health', player['player']['health'])
                       and player['mana'] == case.get('expected_mana', player['mana']))
    if case.get('expect_damage'):
        outcomes_ok = outcomes_ok and any(p['player']['health'] < p['player']['maximum_health'] for p in report['player_observations'])
    if (case.get('created_type') or case.get('required_defeated_script_ids')
            or case.get('required_enable_transitions') or case.get('required_plate_holds')) and snapshots_ok:
        samples = [json.loads(Path(sample['routes']['objects']['path']).read_text())
                   for sample in report['snapshots']]
    if case.get('created_type') and snapshots_ok:
        history = created_monster_history(samples, case['created_type'])
        report['original_created_monsters'] = history
        outcomes_ok = (outcomes_ok and history['count'] >= case['minimum_created']
                       and history['dead'] >= case['minimum_dead_created'])
    if case.get('required_defeated_script_ids') and snapshots_ok:
        history = required_monster_deaths(samples, case['required_defeated_script_ids'])
        report['required_original_monsters'] = history
        outcomes_ok = outcomes_ok and history['passed']
    if case.get('required_enable_transitions') and snapshots_ok:
        transitions = required_enable_transitions(samples, case['required_enable_transitions'])
        report['required_original_enable_transitions'] = transitions
        outcomes_ok = outcomes_ok and transitions['passed']
    if case.get('required_plate_holds') and snapshots_ok:
        holds = required_plate_holds(samples, case['required_plate_holds'])
        report['required_original_plate_holds'] = holds
        outcomes_ok = outcomes_ok and holds['passed']
    report['continuation_eligible'] = healthy and poison_free and bool(report.get('manual_save_written')) and not errors
    report['passed'] = (report['exit_code'] == 0 and not report['forced_termination'] and not errors
                        and not reader.is_alive() and not worker.is_alive() and snapshots_ok
                        and actual_maps <= MAPS and expected_maps <= actual_maps and required and outcomes_ok
                        and healthy and (poison_free or not case.get('require_poison_free', True))
                        and bool(report.get('manual_save_written')))
    put(case_dir / 'result.json', report)
    print(json.dumps({'case': case['name'], 'passed': report['passed'],
                      'continuation_eligible': report['continuation_eligible'], 'exit_code': report['exit_code'],
                      'elapsed_seconds': report['elapsed_seconds']}), flush=True)
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ('suite', 'binary', 'data', 'seed', 'output'):
        parser.add_argument('--' + name, required=True, type=Path)
    parser.add_argument('--case', action='append', dest='cases')
    parser.add_argument('--port', type=int, default=23173)
    parser.add_argument('--config-seed', type=Path)
    parser.add_argument('--sandbox-profile', type=Path)
    args = parser.parse_args()
    for name in ('suite', 'binary', 'data', 'seed', 'output'):
        setattr(args, name, getattr(args, name).resolve())
    if args.config_seed:
        args.config_seed = args.config_seed.resolve()
    if args.sandbox_profile:
        args.sandbox_profile = args.sandbox_profile.resolve()
    if args.output.exists() or args.output.is_relative_to(args.data) or args.output.is_relative_to(args.seed):
        parser.error('output must be a new directory outside data and seed')
    suite = json.loads(args.suite.read_text())
    if suite.get('version') != 1:
        parser.error('unsupported campaign suite version')
    if not args.binary.is_file() or not os.access(args.binary, os.X_OK):
        parser.error('binary must be an executable file')
    if args.config_seed and not args.config_seed.is_file():
        parser.error('config seed must be an existing nox.cfg file')
    cases = [case for case in suite['cases'] if not args.cases or case['name'] in args.cases]
    if not cases or args.cases and set(args.cases) != {case['name'] for case in cases}:
        parser.error('unknown or empty case selection')
    names = [case['name'] for case in cases]
    if len(set(names)) != len(names) or any(not re.fullmatch(r'[A-Za-z0-9._-]+', name) for name in names):
        parser.error('case names must be unique simple directory names')
    if any(not set(m.lower() for m in case['expected_maps']) <= MAPS for case in cases):
        parser.error('only chapter 1 through 3 maps are admitted')
    for case in cases:
        validate_scenario((args.suite.parent / case['scenario']).resolve())
    before_seed = tree(args.seed)
    if not before_seed:
        parser.error('seed must contain an earned campaign save')
    config_hash = sha(args.config_seed) if args.config_seed else None
    data_inputs = [args.data / 'GAME.EXE', args.data / 'thing.bin']
    for directory in (args.data / 'maps').iterdir():
        if directory.is_dir() and directory.name.lower() in MAPS:
            data_inputs.extend(path for path in directory.rglob('*') if path.is_file())
    data_hashes = {str(path): sha(path) for path in data_inputs}
    binary_hash = sha(args.binary)
    args.output.mkdir(parents=True)
    binary = args.output / 'opennox-tested'
    shutil.copy2(args.binary, binary)
    if sha(binary) != binary_hash:
        raise RuntimeError('binary copy changed')
    result = {'suite': suite['name'], 'binary_sha256': binary_hash, 'built_for_each_case': False,
              'source_seed_manifest': before_seed, 'cases': [], 'overall_porting_100_percent': False}
    put(args.output / 'inputs.json', dict(suite_sha256=sha(args.suite), binary_sha256=binary_hash,
                                       seed=before_seed, config_sha256=config_hash, data=data_hashes))
    try:
        for index, case in enumerate(cases):
            result['cases'].append(run_case(args, binary, args.suite.parent, case, args.port + 4 * index))
            if not result['cases'][-1]['passed']:
                break
    finally:
        result['seed_unchanged'] = tree(args.seed) == before_seed
        result['data_inputs_unchanged'] = all(sha(Path(path)) == value for path, value in data_hashes.items())
        result['config_unchanged'] = not args.config_seed or sha(args.config_seed) == config_hash
        result['binary_unchanged'] = sha(args.binary) == binary_hash and sha(binary) == binary_hash
        result['passed'] = (len(result['cases']) == len(cases) and all(case['passed'] for case in result['cases'])
                            and result['seed_unchanged'] and result['data_inputs_unchanged']
                            and result['config_unchanged'] and result['binary_unchanged'])
        put(args.output / 'result.json', result)
    return 0 if result['passed'] else 1


if __name__ == '__main__':
    sys.exit(main())
