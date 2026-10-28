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
INPUT_ACTIONS = set('slow click interact move raw wait walk run run-start run-dir run-stop walk-start walk-dir walk-stop run-to-world walk-to-world wait-map wait-player-reloaded select-manual-campaign-save click-dialog-yes click-npc-dialog-done assert-quickbar-spell assert-spell-set-row assert-server-inventory-count observe-player-target capture-quest-map-audit-frame cast-at-nearest-monster-input pickup-named-object-input pickup-map-object-input screen quit'.split())
ROUTES = {'objects': '/debug/nox/objects', 'script': '/debug/nox/script/legacy', 'walls': '/debug/nox/walls'}
SUITE_FIELDS = {'version', 'name', 'coverage', 'seed_contract', 'cases'}
CASE_FIELDS = {'name', 'scenario', 'expected_maps', 'manual_save', 'require_poison_free',
               'timeout_seconds', 'required_checkpoints', 'expected_health', 'expected_mana',
               'expected_maximum_health', 'expected_maximum_mana', 'expected_stat_inputs',
               'minimum_health', 'minimum_mana',
               'expect_damage', 'created_type', 'minimum_created', 'minimum_dead_created',
               'required_defeated_script_ids', 'require_monster_health_zero', 'required_enable_transitions',
               'required_position_transitions', 'required_plate_holds', 'required_map_pickups',
               'required_swap_pairs'}
STAT_INPUT_WORDS = {'base_bits32', 'maximum_bits32', 'warrior_bits32', 'multipliers_bits32'}


def validate_stat_inputs(value):
    if (not isinstance(value, dict) or set(value) != STAT_INPUT_WORDS | {'class', 'level'}
            or type(value.get('class')) is not int or not 0 <= value['class'] <= 2
            or type(value.get('level')) is not int or not 0 <= value['level'] <= 255
            or any(not isinstance(value[key], list) or len(value[key]) != 4
                   or any(type(word) is not int or not 0 <= word < 1 << 32 for word in value[key])
                   for key in STAT_INPUT_WORDS)):
        raise ValueError('expected_stat_inputs requires class, level and four exact uint32 stat-word arrays')


def validate_suite_fields(suite):
    # A misspelled requirement must never silently omit an outcome check.
    unknown = set(suite) - SUITE_FIELDS
    if unknown:
        raise ValueError('unknown campaign suite fields: ' + ', '.join(sorted(unknown)))
    for case in suite.get('cases', []):
        unknown = set(case) - CASE_FIELDS
        if unknown:
            raise ValueError('unknown campaign case fields: ' + ', '.join(sorted(unknown)))
        if 'expected_stat_inputs' in case:
            validate_stat_inputs(case['expected_stat_inputs'])
        if 'require_monster_health_zero' in case:
            identities = case.get('required_defeated_script_ids')
            if (type(case['require_monster_health_zero']) is not bool or not isinstance(identities, list) or not identities
                    or any(type(v) is not int or not 0 < v < 1 << 31 for v in identities)):
                raise ValueError('require_monster_health_zero needs a boolean and required original monster identities')
        if 'required_swap_pairs' in case:
            pairs = case['required_swap_pairs']
            if not isinstance(pairs, list) or not pairs:
                raise ValueError('required_swap_pairs must be a nonempty list')
            for pair in pairs:
                if (not isinstance(pair, dict) or set(pair) != {'player_script_id', 'target_script_ids', 'key', 'tolerance', 'minimum_distance', 'player_start'}
                        or type(pair['player_script_id']) is not int or not 0 < pair['player_script_id'] < 1 << 31
                        or not isinstance(pair['target_script_ids'], list) or not pair['target_script_ids']
                        or any(type(v) is not int or not 0 < v < 1 << 31 or v == pair['player_script_id'] for v in pair['target_script_ids'])
                        or len(set(pair['target_script_ids'])) != len(pair['target_script_ids'])
                        or type(pair['key']) is not int or pair['key'] != 33
                        or not isinstance(pair['player_start'], list) or len(pair['player_start']) != 2
                        or any(type(v) not in (int, float) or not math.isfinite(v) for v in pair['player_start'])
                        or any(type(pair[k]) not in (int, float) or not math.isfinite(pair[k]) for k in ('tolerance', 'minimum_distance'))
                        or pair['tolerance'] < 0 or pair['minimum_distance'] <= 2 * pair['tolerance']):
                    raise ValueError('required_swap_pairs needs distinct original identities, F key33 and separated finite position bounds')
        if 'required_map_pickups' in case:
            pickups = case['required_map_pickups']
            if not isinstance(pickups, list) or not pickups:
                raise ValueError('required_map_pickups must be a nonempty list')
            if len(case.get('expected_maps', [])) != 1:
                raise ValueError('required_map_pickups requires one expected map')
            extents = set()
            for pickup in pickups:
                if (not isinstance(pickup, dict) or set(pickup) != {'extent', 'item'}
                        or type(pickup['extent']) is not int or pickup['extent'] <= 0
                        or not isinstance(pickup['item'], str)
                        or not re.fullmatch(r'[A-Za-z0-9_]+', pickup['item'])):
                    raise ValueError('required_map_pickups needs positive extents and item type names')
                if pickup['extent'] in extents:
                    raise ValueError('required_map_pickups contains a duplicate extent')
                extents.add(pickup['extent'])
        for field in ('health', 'mana'):
            minimum = case.get('minimum_' + field)
            if 'minimum_' + field in case:
                if type(minimum) is not int or not 0 <= minimum <= 0xffff:
                    raise ValueError('minimum_' + field + ' must be a uint16 value')
                if 'expected_' + field in case and case['expected_' + field] < minimum:
                    raise ValueError('expected_' + field + ' is below its required minimum')


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


def data_manifest(path):
    files = [path / 'GAME.EXE', path / 'thing.bin']
    for directory in (path / 'maps').iterdir():
        if directory.is_dir() and directory.name.lower() in MAPS:
            files.extend(file for file in directory.rglob('*') if file.is_file())
    return {str(file): sha(file) for file in sorted(files)}


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
        if step.get('action') == 'pickup-map-object-input':
            extent = step.get('extent')
            if (type(extent) is not int or not 0 < extent <= 0xffffffff
                    or not isinstance(step.get('item'), str) or not step['item']):
                raise ValueError('map pickup requires a nonzero ABI32 extent and expected item type')
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


def required_monster_health_zero(samples, required_ids):
    required = set(required_ids)

    def health(obj):
        value = obj.get('health')
        if not isinstance(value, dict):
            return None
        current, maximum = value.get('current'), value.get('maximum')
        if (type(current) is not int or type(maximum) is not int
                or not 0 <= current <= 0xffff or not 0 < maximum <= 0xffff):
            return None
        return current

    first, last = (samples[0], samples[-1]) if samples else ([], [])
    alive = {obj.get('script_id') for obj in first
             if isinstance(obj.get('class'), list) and isinstance(obj.get('flags'), list)
             and 'MONSTER' in obj['class']
             and not {'DEAD', 'DESTROYED'} & set(obj.get('flags') or [])
             and health(obj) is not None and health(obj) > 0} & required
    dead = {obj.get('script_id') for obj in last
            if isinstance(obj.get('class'), list) and isinstance(obj.get('flags'), list)
            and 'MONSTER' in obj['class']
            and 'DEAD' in (obj.get('flags') or []) and health(obj) == 0} & alive
    return {'required_script_ids': sorted(required), 'initial_alive_health_script_ids': sorted(alive),
            'final_dead_zero_health_script_ids': sorted(dead), 'passed': bool(required) and required <= dead}


def required_map_pickups(samples, log, requirements, *, actual_maps, expected_maps):
    # For consumed map items, require the original object, ordinary Pickup
    # cursor input for that exact identity, and its removal from a live scene.
    first, last = (samples[0], samples[-1]) if samples else ([], [])
    final_player = any('PLAYER' in (obj.get('class') or []) for obj in last)
    map_scope = len(expected_maps) == 1 and set(actual_maps) == set(expected_maps)
    inputs = {(int(extent), item) for item, extent in re.findall(
        r'NAMED PICKUP ACTUAL INPUT: id="([A-Za-z0-9_]+) extent=(\d+)" [^\n]*\bcursor=2\b', log)}

    def present(objects, extent, item):
        return any(obj.get('ind') == extent and (obj.get('type') or {}).get('id') == item
                   for obj in objects)

    rows = []
    for requirement in requirements:
        extent, item = requirement['extent'], requirement['item']
        initial = present(first, extent, item)
        clicked = (extent, item) in inputs
        removed = final_player and not present(last, extent, item)
        rows.append(dict(extent=extent, item=item, initially_present=initial,
                         pickup_input_observed=clicked, removed_from_final_scene=removed,
                         passed=initial and clicked and removed))
    return dict(objects=rows, final_player_present=final_player, map_scope_passed=map_scope,
                passed=map_scope and bool(requirements) and all(row['passed'] for row in rows))


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


def required_position_transitions(samples, requirements):
    def point(value):
        if not isinstance(value, dict):
            return None
        xy = (value.get('x'), value.get('y'))
        return xy if all(type(v) in (int, float) and math.isfinite(v) for v in xy) else None

    def position(objects, ident):
        matches = [obj for obj in objects if obj.get('script_id') == ident]
        return point(matches[0].get('pos')) if len(matches) == 1 else None

    transitions = []
    for required in requirements:
        ident = required['script_id']
        before, after = point(required['before']), point(required['after'])
        tolerance = required.get('tolerance', 0)
        valid = (before is not None and after is not None
                 and type(tolerance) in (int, float) and math.isfinite(tolerance)
                 and tolerance >= 0 and math.dist(before, after) > 2 * tolerance)
        positions = [position(objects, ident) for objects in samples]
        initial, final = (positions[0], positions[-1]) if positions else (None, None)
        passed = bool(valid and positions and all(p is not None for p in positions)
                      and math.dist(initial, before) <= tolerance
                      and math.dist(final, after) <= tolerance)
        transitions.append(dict(required, initial_position=initial, final_position=final, passed=passed))
    return {'transitions': transitions,
            'passed': bool(transitions) and all(row['passed'] for row in transitions)}


def required_swap_pairs(samples, log, requirements):
    # One native frame must exchange both original positions after the actual
    # F input. Moving targets have no fixed world-space landing coordinate.
    def point(value):
        return (isinstance(value, list) and len(value) == 2
                and all(type(v) in (int, float) and math.isfinite(v) for v in value))

    traces = []
    malformed = False
    for line in log.splitlines():
        if 'MONSTER QUICKBAR POSITION: ' not in line:
            continue
        try:
            row = json.loads(line.split('MONSTER QUICKBAR POSITION: ', 1)[1])
            valid = (isinstance(row, dict) and point(row.get('player')) and point(row.get('target'))
                     and all(type(row.get(k)) is int and 0 <= row[k] <= 0xffffffff
                             for k in ('frame', 'cast_frame', 'player_script_id', 'script_id', 'extent', 'wire', 'key'))
                     and type(row.get('key_queued')) is bool)
        except (ValueError, TypeError):
            valid = False
        if valid:
            traces.append(row)
        else:
            malformed = True
    actual_inputs = {(int(wire, 16), int(extent), int(ident), int(key))
                     for wire, extent, ident, key in re.findall(
                         r'MONSTER QUICKBAR ACTUAL INPUT: type="[A-Za-z0-9_]+" wire=(0x[0-9a-f]+) extent=(\d+) script_id=(\d+) key=(\d+) key_queued=true\b', log)}
    pairs = []
    for requirement in requirements:
        player_id, target_ids = requirement['player_script_id'], requirement['target_script_ids']
        alive_ids = set()
        for target_id in target_ids:
            initial = [o for o in samples[0] if o.get('script_id') == target_id] if samples else []
            if len(initial) == 1 and not {'DEAD', 'DESTROYED'} & set(initial[0].get('flags') or []):
                alive_ids.add(target_id)
        initial_player = [o.get('pos') for o in samples[0] if o.get('script_id') == player_id] if samples else []
        initial_point = ([initial_player[0].get('x'), initial_player[0].get('y')]
                         if len(initial_player) == 1 and isinstance(initial_player[0], dict) else None)
        starts_inside = point(initial_point) and math.dist(initial_point, requirement['player_start']) <= requirement['tolerance']
        final = [o.get('pos') for o in samples[-1] if o.get('script_id') == player_id] if samples else []
        final_point = ([final[0].get('x'), final[0].get('y')]
                       if len(final) == 1 and isinstance(final[0], dict) else None)
        rows = [v for v in traces if v['player_script_id'] == player_id and v['script_id'] in alive_ids
                and v['key'] == requirement['key'] and v['key_queued']]
        exchange = None
        for before, after in zip(rows, rows[1:]):
            identity = lambda v: (v['wire'], v['extent'], v['script_id'], v['key'])
            tolerance = requirement['tolerance']
            if ((after['frame'] - before['frame']) & 0xffffffff == 1
                    and before['cast_frame'] == after['cast_frame'] and identity(before) == identity(after)
                    and identity(after) in actual_inputs
                    and math.dist(before['player'], after['player']) >= requirement['minimum_distance']
                    and math.dist(before['player'], after['target']) <= tolerance
                    and math.dist(before['target'], after['player']) <= tolerance
                    and point(final_point) and math.dist(final_point, after['player']) <= tolerance):
                exchange = {'before': before, 'after': after}
                break
        selected_id = exchange['after']['script_id'] if exchange else None
        pairs.append(dict(requirement, initial_alive_target_script_ids=sorted(alive_ids), selected_target_script_id=selected_id, initial_player_position=initial_point,
                          initial_player_matches=bool(starts_inside), position_exchange=exchange,
                          final_player_position=final_point, passed=bool(len(alive_ids) == len(target_ids) and starts_inside and exchange and not malformed)))
    return {'pairs': pairs, 'malformed_position_observation': malformed,
            'passed': bool(pairs) and all(p['passed'] for p in pairs)}


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


def player_value_failures(player, case):
    values = {'expected_health': player['player']['health'],
              'expected_mana': player['mana'],
              'expected_maximum_health': player['player'].get('maximum_health'),
              'expected_maximum_mana': player.get('maximum_mana')}
    minimums = {'minimum_health': player['player'].get('health'),
                'minimum_mana': player.get('mana')}
    failures = [{'check': key, 'expected': case[key], 'actual': value}
                for key, value in values.items() if key in case and value != case[key]]
    failures.extend({'check': key, 'minimum': case[key], 'actual': value}
                    for key, value in minimums.items()
                    if key in case and not (type(value) is int and value >= case[key]))
    if 'expected_stat_inputs' in case:
        actual = player.get('stat_inputs')
        matches = False
        if (isinstance(actual, dict) and actual.get('present') is True
                and type(actual.get('game_flags')) is int and 0 <= actual['game_flags'] < 1 << 32):
            selected = {key: actual.get(key) for key in STAT_INPUT_WORDS | {'class', 'level'}}
            try:
                validate_stat_inputs(selected)
                matches = selected == case['expected_stat_inputs']
            except ValueError:
                pass
        if not matches:
            failures.append({'check': 'expected_stat_inputs', 'expected': case['expected_stat_inputs'], 'actual': actual})
    return failures


def expected_player_values(player, case):
    return not player_value_failures(player, case)


def load_sequence(path):
    sequence = json.loads(path.read_text())
    unknown = set(sequence) - {'version', 'name', 'seed_contract', 'coverage', 'stages'}
    if unknown or sequence.get('version') != 1 or not sequence.get('stages'):
        raise ValueError('invalid campaign sequence fields or version')
    stages = []
    names = set()
    for entry in sequence['stages']:
        if set(entry) != {'name', 'suite'}:
            raise ValueError('sequence stages require name and suite')
        name = entry['name']
        if (not re.fullmatch(r'[A-Za-z0-9._-]+', name) or name in ('.', '..')
                or name in names):
            raise ValueError('sequence stage names must be unique simple directory names')
        names.add(name)
        suite_path = (path.parent / entry['suite']).resolve()
        suite = json.loads(suite_path.read_text())
        validate_suite_fields(suite)
        if suite.get('version') != 1 or len(suite.get('cases', [])) != 1:
            raise ValueError('each sequence stage requires one campaign case')
        case = suite['cases'][0]
        if (not re.fullmatch(r'[A-Za-z0-9._-]+', case['name']) or case['name'] in ('.', '..')
                or not case.get('require_poison_free', True)
                or not case['expected_maps']
                or not {m.lower() for m in case['expected_maps']} <= MAPS):
            raise ValueError('sequence stages require healthy chapter 1 through 3 cases')
        scenario = (suite_path.parent / case['scenario']).resolve()
        validate_scenario(scenario)
        stages.append({'name': name, 'suite': suite_path, 'case': case,
                       'suite_sha256': sha(suite_path), 'scenario_sha256': sha(scenario)})
    return sequence, stages


def continuation_seed(report, stage_dir, case, binary_hash, source_seed):
    # A healthy save from a FAILED objective is not a passing sequence link.
    if (report.get('passed') is not True or report.get('binary_sha256') != binary_hash
            or report.get('source_seed_manifest') != source_seed
            or len(report.get('cases', [])) != 1):
        raise ValueError('stage failed or binary/save lineage does not match')
    result = report['cases'][0]
    if (result.get('passed') is not True or result.get('continuation_eligible') is not True
            or result.get('manual_save_written') is not True):
        raise ValueError('stage did not produce a passing healthy manual save')
    runtime = Path(result['runtime_data']).resolve()
    if not runtime.is_relative_to(stage_dir.resolve()) or runtime == stage_dir.resolve():
        raise ValueError('stage runtime must stay inside its report directory')
    seed = runtime / 'Save'
    saved = (seed / case.get('manual_save', 'SAVE0001/Player.plr')).resolve()
    if (not saved.is_relative_to(seed) or not saved.is_file()
            or sha(saved) != result.get('saved_player_sha256')):
        raise ValueError('stage manual save is absent or has changed')
    manifest = tree(seed)
    if not manifest:
        raise ValueError('stage save is empty')
    return seed, manifest


def resume_stages(args, stages, binary_hash, initial_seed):
    previous = getattr(args, 'resume_from', None)
    if previous is None:
        return []
    # Validate everything before creating output or launching a game. An old
    # success is reusable evidence only for this exact executable and inputs.
    if sha(previous / 'opennox-tested') != binary_hash:
        raise ValueError('resume requires the same tested executable')
    assets = data_manifest(args.data)
    config_hash = sha(args.config_seed) if args.config_seed else None
    source_seed = initial_seed
    reused = []
    for line in (previous / 'stages.jsonl').read_text().splitlines(keepends=True):
        if not line.endswith('\n'):
            break  # Interrupted progress writes cannot certify a stage.
        row = json.loads(line)
        if row.get('passed') is not True:
            break
        index = len(reused)
        if index >= len(stages) or row['name'] != stages[index]['name'] or row['exit_code'] != 0:
            raise ValueError('resume stage order or successful exit does not match')
        stage = stages[index]
        stage_dir = Path(row['report']).resolve().parent
        child = json.loads((stage_dir / 'result.json').read_text())
        inputs = json.loads((stage_dir / 'inputs.json').read_text())
        current_scenario = (stage['suite'].parent / stage['case']['scenario']).resolve()
        if (sha(stage['suite']) != stage['suite_sha256']
                or sha(current_scenario) != stage['scenario_sha256']
                or inputs['suite_sha256'] != stage['suite_sha256']
                or sha(stage_dir / 'suite.source.json') != stage['suite_sha256']
                or inputs['binary_sha256'] != binary_hash
                or sha(stage_dir / 'opennox-tested') != binary_hash
                or inputs['seed'] != source_seed or tree(Path(row['seed'])) != source_seed
                or inputs['config_sha256'] != config_hash or inputs['data'] != assets
                or not all(child.get(key) is True for key in
                           ('seed_unchanged', 'data_inputs_unchanged', 'config_unchanged', 'binary_unchanged'))):
            raise ValueError('resume suite, executable, seed, config or data inputs changed')
        if len(child.get('cases', [])) != 1:
            raise ValueError('resume requires exactly one case per stage')
        case = child['cases'][0]
        scenario = Path(case['scenario']).resolve()
        if (not scenario.is_relative_to(stage_dir)
                or case['scenario_sha256'] != stage['scenario_sha256']
                or sha(scenario) != stage['scenario_sha256'] or not case['snapshots']):
            raise ValueError('resume scenario or observation evidence changed')
        for sample in case['snapshots']:
            if set(sample['routes']) != set(ROUTES):
                raise ValueError('resume requires all read-only observation routes')
            for name, evidence in sample['routes'].items():
                path = Path(evidence['path']).resolve()
                if (not path.is_relative_to(stage_dir) or evidence['status'] != 200
                        or evidence['method'] != 'GET' or evidence['route'] != ROUTES[name]
                        or evidence['observation_only'] is not True
                        or path.stat().st_size != evidence['bytes'] or sha(path) != evidence['sha256']):
                    raise ValueError('resume observation evidence is absent or changed')
        seed, source_seed = continuation_seed(child, stage_dir, stage['case'], binary_hash, source_seed)
        if source_seed != row['next_seed_manifest'] or seed != Path(row['next_seed']).resolve():
            raise ValueError('resume earned Save files changed')
        reused.append(dict(row, reused=True, reused_from=str(previous)))
    return reused


def run_sequence(args, sequence, stages):
    initial_seed = tree(args.seed)
    if not initial_seed:
        raise ValueError('seed must contain an earned campaign save')
    binary_hash = sha(args.binary)
    reused = resume_stages(args, stages, binary_hash, initial_seed)
    args.output.mkdir(parents=True)
    shutil.copyfile(args.sequence, args.output / 'sequence.source.json')
    binary = args.output / 'opennox-tested'
    shutil.copy2(args.binary, binary)
    result = {'sequence': sequence['name'], 'binary_sha256': binary_hash,
              'built_for_each_case': False, 'source_seed_manifest': initial_seed,
              'stages': list(reused), 'reused_stages': len(reused), 'executed_stages': 0,
              'fresh_replay': not reused, 'overall_porting_100_percent': False}
    if getattr(args, 'resume_from', None):
        result['resume_from'] = str(args.resume_from)
    seed, source_seed = args.seed, initial_seed
    if reused:
        seed = Path(reused[-1]['next_seed'])
        source_seed = reused[-1]['next_seed_manifest']
    started = time.monotonic()
    try:
        with (args.output / 'stages.jsonl').open('x') as progress:
            for row in reused:
                progress.write(json.dumps(row, ensure_ascii=False) + '\n')
                print(json.dumps({'stage': row['name'], 'passed': True, 'reused': True,
                                  'report': row['report']}), flush=True)
            progress.flush()
            for index in range(len(reused), len(stages)):
                stage = stages[index]
                stage_dir = args.output / stage['name']
                command = [sys.executable, str(Path(__file__).resolve()),
                           '--suite', str(stage['suite']), '--binary', str(binary),
                           '--data', str(args.data), '--seed', str(seed),
                           '--output', str(stage_dir), '--port', str(args.port + 4 * index)]
                for flag in ('config_seed', 'sandbox_profile'):
                    if getattr(args, flag):
                        command += ['--' + flag.replace('_', '-'), str(getattr(args, flag))]
                row = {'name': stage['name'], 'suite': str(stage['suite']),
                       'seed': str(seed), 'report': str(stage_dir / 'result.json'),
                       'command': command, 'passed': False, 'reused': False}
                result['stages'].append(row)
                try:
                    scenario = (stage['suite'].parent / stage['case']['scenario']).resolve()
                    if (sha(stage['suite']) != stage['suite_sha256']
                            or sha(scenario) != stage['scenario_sha256']):
                        raise ValueError('sequence input changed after validation')
                    result['executed_stages'] += 1
                    row['exit_code'] = subprocess.run(command, check=False).returncode
                    child = json.loads((stage_dir / 'result.json').read_text())
                    row['failed_checks'] = [dict(failure, case=case.get('name'))
                                            for case in child.get('cases', [])
                                            for failure in case.get('failed_checks', [])]
                    if row['exit_code'] != 0 or child.get('passed') is not True:
                        raise ValueError('stage failed; see failed_checks and the stage report')
                    if tree(seed) != source_seed:
                        raise ValueError('stage changed its source save')
                    # Verify the exact suite and scenario that the child executed.
                    inputs = json.loads((stage_dir / 'inputs.json').read_text())
                    if (inputs['suite_sha256'] != stage['suite_sha256']
                            or child['cases'][0]['scenario_sha256'] != stage['scenario_sha256']):
                        raise ValueError('executed sequence input does not match validation')
                    seed, source_seed = continuation_seed(child, stage_dir, stage['case'],
                                                          binary_hash, source_seed)
                    row.update(passed=True, next_seed=str(seed), next_seed_manifest=source_seed)
                except (OSError, ValueError, KeyError, IndexError) as error:
                    row['error'] = str(error)
                progress.write(json.dumps(row, ensure_ascii=False) + '\n')
                progress.flush()
                print(json.dumps({'stage': row['name'], 'passed': row['passed'],
                                  'report': row['report'], 'error': row.get('error'),
                                  'failed_checks': row.get('failed_checks', [])}), flush=True)
                if not row['passed']:
                    break
    finally:
        result['elapsed_seconds'] = round(time.monotonic() - started, 3)
        result['seed_unchanged'] = tree(args.seed) == initial_seed
        result['binary_unchanged'] = sha(args.binary) == binary_hash and sha(binary) == binary_hash
        result['passed'] = (len(result['stages']) == len(stages)
                            and all(row['passed'] for row in result['stages'])
                            and result['seed_unchanged'] and result['binary_unchanged'])
        if result['passed']:
            result['final_seed'] = str(seed)
            result['final_seed_manifest'] = source_seed
        put(args.output / 'result.json', result)
    return 0 if result['passed'] else 1


def run_case(args, binary, suite_dir, case, port):
    case_dir = args.output / case['name']
    case_dir.mkdir()
    frames = case_dir / 'frames'
    frames.mkdir()
    source_scenario = (suite_dir / case['scenario']).resolve()
    # Keep the exact executed input with successes and failures. Editing the
    # next route must not change an earlier report's replay or its evidence.
    scenario = case_dir / 'scenario.yaml'
    shutil.copyfile(source_scenario, scenario)
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
    report = {'name': case['name'], 'source_scenario': str(source_scenario),
              'scenario': str(scenario), 'scenario_sha256': sha(scenario),
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
        with proc.stdout, (case_dir / 'game.log').open('xb') as log:
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
    failures = []
    outcomes_ok = True
    if player:
        outcomes_ok = expected_player_values(player, case)
        failures.extend(player_value_failures(player, case))
    if case.get('expect_damage'):
        outcomes_ok = outcomes_ok and any(p['player']['health'] < p['player']['maximum_health'] for p in report['player_observations'])
    if (case.get('created_type') or case.get('required_defeated_script_ids')
            or case.get('required_enable_transitions') or case.get('required_position_transitions')
            or case.get('required_plate_holds') or case.get('required_map_pickups') or case.get('required_swap_pairs')) and snapshots_ok:
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
        if not history['passed']:
            failures.append({'check': 'required_defeated_script_ids',
                             'missing': sorted(set(history['required_script_ids'])
                                               - set(history['last_observed_dead_script_ids'])),
                             'initial_alive': history['initial_alive_script_ids'],
                             'observed_dead': history['last_observed_dead_script_ids']})
    if case.get('require_monster_health_zero'):
        native_health = required_monster_health_zero(samples if snapshots_ok else [], case['required_defeated_script_ids'])
        report['required_native_monster_health'] = native_health
        outcomes_ok = outcomes_ok and native_health['passed']
        if not native_health['passed']:
            failures.append({'check': 'require_monster_health_zero', **native_health})
    if case.get('required_map_pickups'):
        log_path = case_dir / 'game.log'
        pickups = required_map_pickups(samples if snapshots_ok else [],
                                       log_path.read_text(errors='replace') if log_path.exists() else '',
                                       case['required_map_pickups'], actual_maps=actual_maps,
                                       expected_maps=expected_maps)
        report['required_original_map_pickups'] = pickups
        outcomes_ok = outcomes_ok and pickups['passed']
        if not pickups['passed']:
            failures.append({'check': 'required_map_pickups',
                             'missing': [row for row in pickups['objects'] if not row['passed']],
                             'final_player_present': pickups['final_player_present'],
                             'map_scope_passed': pickups['map_scope_passed']})
    if case.get('required_enable_transitions') and snapshots_ok:
        transitions = required_enable_transitions(samples, case['required_enable_transitions'])
        report['required_original_enable_transitions'] = transitions
        outcomes_ok = outcomes_ok and transitions['passed']
    if case.get('required_position_transitions') and snapshots_ok:
        transitions = required_position_transitions(samples, case['required_position_transitions'])
        report['required_original_position_transitions'] = transitions
        outcomes_ok = outcomes_ok and transitions['passed']
    if case.get('required_swap_pairs'):
        swaps = required_swap_pairs(samples if snapshots_ok else [],
                                    (case_dir / 'game.log').read_text(errors='replace'), case['required_swap_pairs'])
        report['required_original_swap_pairs'] = swaps
        outcomes_ok = outcomes_ok and swaps['passed']
        if not swaps['passed']:
            failures.append({'check': 'required_swap_pairs', 'pairs': swaps['pairs'],
                             'malformed_position_observation': swaps['malformed_position_observation']})
    if case.get('required_plate_holds') and snapshots_ok:
        holds = required_plate_holds(samples, case['required_plate_holds'])
        report['required_original_plate_holds'] = holds
        outcomes_ok = outcomes_ok and holds['passed']
    report['continuation_eligible'] = healthy and poison_free and bool(report.get('manual_save_written')) and not errors
    checks = {'successful_exit': report['exit_code'] == 0,
              'natural_exit': not report['forced_termination'], 'runtime_errors_absent': not errors,
              'output_reader_finished': not reader.is_alive(), 'snapshot_reader_finished': not worker.is_alive(),
              'snapshots_complete': snapshots_ok, 'maps_in_scope': actual_maps <= MAPS,
              'expected_maps_observed': expected_maps <= actual_maps,
              'required_checkpoints': required, 'expected_outcomes': outcomes_ok,
              'player_alive': healthy, 'poison_free': poison_free or not case.get('require_poison_free', True),
              'manual_save_written': bool(report.get('manual_save_written'))}
    has_outcome_details = bool(failures)
    failures.extend({'check': name} for name, passed in checks.items()
                    if not passed and (name != 'expected_outcomes' or not has_outcome_details))
    report['failed_checks'] = failures
    report['passed'] = all(checks.values())
    put(case_dir / 'result.json', report)
    print(json.dumps({'case': case['name'], 'passed': report['passed'],
                      'continuation_eligible': report['continuation_eligible'], 'exit_code': report['exit_code'],
                      'elapsed_seconds': report['elapsed_seconds'], 'failed_checks': failures}), flush=True)
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    selection = parser.add_mutually_exclusive_group(required=True)
    selection.add_argument('--suite', type=Path)
    selection.add_argument('--sequence', type=Path,
                           help='replay ordered suites, passing each healthy earned Save to the next')
    for name in ('binary', 'data', 'seed', 'output'):
        parser.add_argument('--' + name, required=True, type=Path)
    parser.add_argument('--case', action='append', dest='cases')
    parser.add_argument('--port', type=int, default=23173)
    parser.add_argument('--config-seed', type=Path)
    parser.add_argument('--sandbox-profile', type=Path)
    parser.add_argument('--resume-from', type=Path,
                        help='reuse a sequence report\'s unchanged passing prefix and replay the rest')
    args = parser.parse_args()
    for name in ('suite', 'sequence', 'binary', 'data', 'seed', 'output', 'resume_from'):
        if getattr(args, name):
            setattr(args, name, getattr(args, name).resolve())
    if args.config_seed:
        args.config_seed = args.config_seed.resolve()
    if args.sandbox_profile:
        args.sandbox_profile = args.sandbox_profile.resolve()
    if args.output.exists() or args.output.is_relative_to(args.data) or args.output.is_relative_to(args.seed):
        parser.error('output must be a new directory outside data and seed')
    if args.resume_from and not args.sequence:
        parser.error('--resume-from requires --sequence')
    if args.resume_from and args.output.is_relative_to(args.resume_from):
        parser.error('resume output must stay outside the previous report')
    if args.sequence:
        if args.cases:
            parser.error('--case applies to a single suite, not a sequence')
        if not args.binary.is_file() or not os.access(args.binary, os.X_OK):
            parser.error('binary must be an executable file')
        try:
            sequence, stages = load_sequence(args.sequence)
            if not tree(args.seed):
                parser.error('seed must contain an earned campaign save')
            return run_sequence(args, sequence, stages)
        except (OSError, ValueError, KeyError, TypeError) as error:
            parser.error(str(error))
    suite = json.loads(args.suite.read_text())
    try:
        validate_suite_fields(suite)
    except ValueError as error:
        parser.error(str(error))
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
    data_hashes = data_manifest(args.data)
    binary_hash = sha(args.binary)
    args.output.mkdir(parents=True)
    shutil.copyfile(args.suite, args.output / 'suite.source.json')
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
