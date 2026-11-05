"""Independent raw Skull/ArrowTrap inputs; reads no native source or outputs."""
import argparse
import copy
import gzip
import hashlib
import json
from pathlib import Path

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
options = parser.parse_args()

# Roles 0..3 are complete object records, 4..7 are complete data records.
# -1 is nil; -2 is an unmapped original domain. Native correspondence must
# be independently validated before any unchanged-C fault or port credit.
base = dict(entry='54f9a0', unit=0, target=1, projectile=2, update=4,
            collide=5, flags=0x01000000, source_type=85, direction=1,
            position=[0x43fa0001, 0xc3fa0001], radius_bits=0x41200001,
            speed_bits=0x3f800001, timers=[0, 0], state=1, enabled=1,
            projectile_type=73, targets=[1, 3], target_class=6,
            target_flags=0, enemy=1, map_check=1, in_front=1,
            caches=[17, 0, 0, 0, 0, 0], fps=30,
            name_results={'MercArcherArrow': 73, 'ArrowTrap1': 85, 'ArrowTrap2': 86},
            damage_bits=0x4236abcd, script={})
rows = []


def add(name, **changes):
    row = copy.deepcopy(base)
    row.update(changes)
    row['name'] = name
    rows.append(row)


entries = ['54f9a0', '54fa80', '54fbb0', '54fbf0', '54fc50']
for entry in entries:
    add(entry + '-default', entry=entry)
    for unit in [-2, -1, 0, 1, 2, 3]:
        add(f'{entry}-source-{unit}', entry=entry, unit=unit)
    for target in [-2, -1, 0, 1, 2, 3]:
        if entry == '54fc50':
            add(f'{entry}-candidate-{target}', entry=entry, target=target)

for flags in [0, 1, 0x01000000, 0x80000000, 0xffffffff]:
    for timer in [0, 1, 30, 0x80000000, 0xffffffff]:
        for cooldown in [0, 1, 30, 0xffffffff]:
            for enabled in [0, 1, 0xff]:
                add(f'update-flags-{flags:08x}-timers-{timer:08x}-{cooldown:08x}-enabled-{enabled}',
                    flags=flags, timers=[timer, cooldown], enabled=enabled)
for state in [0, 1, 2, 0xff]:
    add(f'update-state-{state}', state=state, timers=[1, 0])
    add(f'wrapper-state-{state}-accept', entry='54fbb0', state=state, timers=[17, 19])
    add(f'wrapper-state-{state}-reject', entry='54fbb0', state=state, timers=[17, 19], enemy=0)
for fps in [0, 1, 30, 0x7fffffff, 0x80000000, 0xffffffff]:
    add(f'update-fps-{fps:08x}', fps=fps)
for entry in ['54f9a0', '54fa80']:
    for value in [0, 1, 85, 86, 0xffff]:
        add(f'{entry}-source-type-{value}', entry=entry, source_type=value)
    for cache in [[17, 73, 85, 86, 85, 86], [17, 1, 2, 3, 4, 5],
                  [17, 0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff, 0xffffffff],
                  [17, 73, 85, 85, 85, 85], [17, 73, 0, 0, 0, 0]]:
        add(entry + '-cache-' + '-'.join(f'{v:08x}' for v in cache), entry=entry, caches=cache)
    for name in ['MercArcherArrow', 'ArrowTrap1', 'ArrowTrap2']:
        for value in [0, 1, 0x7fffffff, 0x80000000, 0xffffffff]:
            names = dict(base['name_results'])
            names[name] = value
            add(f'{entry}-name-{name}-{value:08x}', entry=entry, name_results=names)
for entry in ['54f9a0', '54fbb0']:
    for update in [-2, -1, 0, 2, 4, 6]:
        add(f'{entry}-update-{update}', entry=entry, update=update)
for projectile in [-2, -1, 0, 1, 2, 3]:
    add(f'projectile-result-{projectile}', entry='54fa80', projectile=projectile)
for collide in [-2, -1, 0, 2, 4, 5, 7]:
    add(f'projectile-collide-{collide}', entry='54fa80', collide=collide)
for projectile_type in [0, 1, 73, 85, 86, 0x7fffffff, 0x80000000, 0xffffffff]:
    add(f'projectile-argument-{projectile_type:08x}', entry='54fa80', projectile_type=projectile_type)
for direction in [0, 1, 63, 64, 127, 128, 255, 256, 0x7fff, 0x8000, 0xffff]:
    add(f'projectile-direction-word-{direction:04x}', entry='54fa80', direction=direction)
for value in [0, 0x80000000, 1, 0x007fffff, 0x00800000, 0x3f800001,
              0x4b800001, 0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000,
              0x7fc01234]:
    add(f'scan-position-{value:08x}', entry='54fbf0', position=[value, value])
    add(f'projectile-radius-{value:08x}', entry='54fa80', radius_bits=value)
    add(f'projectile-speed-{value:08x}', entry='54fa80', speed_bits=value)
    add(f'projectile-damage-{value:08x}', entry='54fa80', damage_bits=value)
for target_class in [0, 1, 2, 4, 6, 0x10000, 0x10002, 0x80000000, 0xffffffff]:
    for target_flags in [0, 0x20, 0x8000, 0x8020, 0x80000000, 0xffffffff]:
        add(f'callback-class-{target_class:08x}-flags-{target_flags:08x}', entry='54fc50',
            target_class=target_class, target_flags=target_flags)
for field in ['enemy', 'map_check', 'in_front']:
    for value in [0, 1, 2, 3, 0x80000000, 0xffffffff]:
        add(f'callback-{field}-{value:08x}', entry='54fc50', **{field: value})
for unit in [-2, -1]:
    for target_class, target_flags, enemy, map_check in [(0, 0, 1, 1), (6, 0x20, 1, 1),
                                                       (6, 0, 0, 1), (6, 0, 1, 0),
                                                       (6, 0, 1, 1)]:
        add(f'callback-source-{unit}-short-{target_class}-{target_flags}-{enemy}-{map_check}',
            entry='54fc50', unit=unit, target_class=target_class, target_flags=target_flags,
            enemy=enemy, map_check=map_check)
for targets in [[], [1], [3, 1], [1, 1], [0], [-1], [-2]]:
    add('scan-targets-' + '-'.join(map(str, targets)), entry='54fbf0', targets=targets)

# Each mutation/reentry is prescribed independently of the native code.
# Offsets below are original PE32 record offsets, never native struct offsets.
service_entries = {'name': '54f9a0', 'new': '54fa80', 'create': '54fa80',
                   'damage': '54fa80', 'audio': '54fa80', 'fx': '54f9a0',
                   'enumerate': '54fbf0', 'enemy': '54fc50',
                   'map': '54fc50', 'front': '54fc50'}
for service, entry in service_entries.items():
    plans = {
        'source-position': {'words': [[0, 56, 0x447a0001], [0, 60, 0xc47a0001]]},
        'source-direction': {'words16': [[0, 124, 128]]},
        'source-type': {'words16': [[0, 4, 86]]},
        'source-flags': {'words': [[0, 16, 0x80008020]]},
        'replace-update': {'pointers': [[0, 748, 6]]},
        'nil-update': {'pointers': [[0, 748, -1]]},
        'replace-collide': {'pointers': [[2, 700, 7]]},
        'nil-collide': {'pointers': [[2, 700, -1]]},
        'timer-state': {'words': [[4, 0, 0xffffffff], [4, 4, 19]], 'bytes': [[4, 8, 0]]},
        'live-caches': {'globals': [[i, 0x87654321 + i] for i in range(6)]},
        'live-fps': {'fps': 0xffffffff},
    }
    for label, plan in plans.items():
        add(f'{entry}-at-{service}-{label}', entry=entry, script={service + ':0': plan})
    for nested in entries:
        add(f'{entry}-at-{service}-reenter-{nested}', entry=entry, script={service + ':0': {
            'reenter': {'entry': nested, 'unit': 3, 'target': 1, 'projectile_type': 73}}})
add('scan-outer-accept-then-nested-empty-scan-resets-shared-marker', entry='54fbf0', script={
    'enemy:1': {'result': 0, 'reenter': {'entry': '54fbf0', 'unit': 3}},
    'enumerate:1': {'targets': []}})
add('update-reentry-same-source-replaces-cache-and-live-cooldown', script={
    'fx:0': {'reenter': {'entry': '54f9a0', 'unit': 0}, 'globals': [[4, 0], [5, 86]]},
    'front:2': {'result': 0}})
add('projectile-balance-nested-callback-replaces-cached-collide', entry='54fa80', script={
    'damage:0': {'reenter': {'entry': '54fc50', 'unit': 0, 'target': 1}},
    'front:0': {'pointers': [[2, 700, 7]]}})

assert len({r['name'] for r in rows}) == len(rows)
decoded = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode() + b'\n'
encoded = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0))
encoded[9] = 0x13
with options.output.open('xb') as stream:
    stream.write(encoded)
print(json.dumps(dict(cases=len(rows), sha256=hashlib.sha256(encoded).hexdigest(),
                      decoded_sha256=hashlib.sha256(decoded).hexdigest(),
                      native_source_used=False, original_execution_credit=False)))
