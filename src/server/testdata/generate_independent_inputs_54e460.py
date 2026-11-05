"""Generate raw original inputs without reading any native implementation or outputs."""
from pathlib import Path
import argparse
import copy
import gzip
import hashlib
import json
import struct

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
options = parser.parse_args()
bits = lambda number: struct.unpack('<Q', struct.pack('<d', number))[0]
base = dict(entry='54e4b0', unit=0, index=0, divisor=12, fps=30, frame=123456,
            parts=2, seconds=13, float_bits=bits(1.25), quest=1,
            owner_return=1, source_owner=1, parent_update=6, source_update=7,
            parent_class=4, slots=[0, 5, -1, 0],
            new={'BigRock':2, 'MediumRock':2, 'SmallRock':2, 'DestroyedGenerator':3},
            script={})
rows = []
def add(name, **changes):
    row = copy.deepcopy(base)
    row.update(changes)
    row['name'] = name
    rows.append(row)

services = {
    '54e460':['owner', 'pointfx', 'delete'],
    '54e4b0':['audio', 'pointfx', 'int', 'new', 'point', 'create', 'float', 'raise', 'force', 'decay'],
    '54e630':['timer', 'mode', 'script', 'audio', 'generatorfx', 'flags', 'owner', 'credit', 'new', 'create', 'delete'],
}
for entry, kinds in services.items():
    add(entry+'-baseline', entry=entry)
    for unit in [-1, -2, 1, 4]:
        add(f'{entry}-source-{unit}', entry=entry, unit=unit)
        if entry != '54e460':
            add(f'{entry}-source-{unit}-allocation-fails', entry=entry, unit=unit, new={})
    if entry != '54e460':
        for result in [-2, 0, 1, 3, 5]:
            add(f'{entry}-allocation-alias-{result}', entry=entry,
                new={name:result for name in base['new']})
        add(entry+'-allocation-fails', entry=entry, new={})
    for kind in kinds:
        add(f'{entry}-position-at-{kind}', entry=entry,
            script={kind+':0':{'position':[0, 0x447a0000, 0xc47a0000]}})
        add(f'{entry}-same-entry-reentry-at-{kind}', entry=entry,
            script={kind+':0':{'reenter':[entry, 4]}})
    for nested in services:
        add(f'{entry}-cross-entry-reentry-{nested}', entry=entry,
            script={kinds[0]+':0':{'reenter':[nested, 4]}})

for slot in range(4):
    values = [5]*4
    values[slot] = 0
    add(f'marker-first-match-{slot}', entry='54e460', slots=values)
add('marker-all-duplicates', entry='54e460', slots=[0]*4)
add('marker-no-match', entry='54e460', slots=[5]*4)
for parent in [-1, -2, 0, 5]:
    add(f'marker-parent-{parent}', entry='54e460', owner_return=parent)
for update in [-1, -2, 7]:
    add(f'marker-parent-update-{update}', entry='54e460', parent_update=update)
add('marker-parent-callback-replaces-update', entry='54e460',
    script={'owner:0':{'update':[1, 7]}})
add('marker-first-slot-reloaded-after-owner', entry='54e460', slots=[5]*4,
    script={'owner:0':{'slot':[0, 0]}})
add('marker-effect-observes-cleared-first-slot', entry='54e460', slots=[0]*4,
    script={'pointfx:0':{'slot':[0, 5]}})

for parts in [-2, -1, 0, 1, 13, 31]:
    add(f'boulder-count-{parts}', parts=parts)
for value in [0, 0x8000000000000000, 1, 0x0010000000000000,
              0x3ff0000000000001, 0xbff0000000000001, 0x3fefffffffffffff,
              0x7fefffffffffffff, 0xffefffffffffffff,
              0x7ff0000000000000, 0xfff0000000000000, 0x7ff8000000001234]:
    add(f'boulder-float-{value:016x}', float_bits=value)
for fps in [0, 1, 0x7fffffff, 0x80000000, 0xffffffff]:
    add(f'boulder-fps-{fps:08x}', fps=fps, seconds=0x7fffffff)
for index, divisor in [(0, 1), (1, 12), (11, 12), (0, 0), (0, -1), (0, -2),
                       (12, 12), (0xffffffff, 12), (0x7fffffff, -1),
                       (0x80000000, 12), (0x10000000, 12)]:
    add(f'boulder-index-{index:08x}-divisor-{divisor}', index=index, divisor=divisor, parts=1)
for kind in services['54e4b0']:
    add(f'boulder-live-index-divisor-at-{kind}',
        script={kind+':0':{'index':11, 'divisor':-2, 'fps':0xffffffff}})
for index in [1, 11, 12, 0xffffffff, 0x10000000]:
    add(f'boulder-mass-index-{index:08x}', script={'new:0':{'index':index}})
for value in [0, 128, 255]:
    add(f'boulder-unsigned-mass-{value}', script={'new:0':{'mass':[0, value]}})
for occurrence in [0, 1]:
    add(f'boulder-allocation-fails-at-{occurrence}', script={f'new:{occurrence}':{'return_role':-1}})
add('boulder-live-next-name', script={'decay:0':{'table':[1, 0]}})
add('boulder-live-next-null-name', script={'decay:0':{'table':[1, -1]}})
add('boulder-returned-float-precedes-live-flags', script={'float:1':{'flags':[2, 0x01020304]}})
add('boulder-signed-divide-overflow', script={'decay:0':{'index':0x7fffffff, 'divisor':-1}})
add('boulder-no-source-read-empty-null', unit=-1, parts=0)
add('boulder-no-source-read-empty-unmapped', unit=-2, parts=0)

for quest in [0, 1, 2, -1]:
    for owner in [-1, 1]:
        add(f'generator-quest-{quest}-owner-{owner}', entry='54e630', quest=quest, source_owner=owner)
for parent in [-1, -2, 0, 4]:
    add(f'generator-parent-{parent}', entry='54e630', owner_return=parent)
for value in [0, 1, 2, 4, 0x84, 0xff]:
    add(f'generator-parent-low-class-{value:02x}', entry='54e630', parent_class=0xabcdef00 | value)
for update in [-1, -2, 6]:
    add(f'generator-initial-update-{update}', entry='54e630', source_update=update)
for kind in ['timer', 'mode', 'script', 'audio', 'generatorfx', 'flags', 'owner', 'credit', 'new']:
    add(f'generator-live-state-at-{kind}', entry='54e630',
        script={kind+':0':{'update':[0, 6], 'owner':[0, -1], 'frame':0xffffffff, 'quest':0}})
add('generator-returned-parent-class-is-live', entry='54e630',
    script={'owner:0':{'class':[1, 0]}})
add('generator-cached-update-despite-timer-reentry', entry='54e630',
    script={'timer:0':{'update':[0, 6], 'reenter':['54e460', 4]}})

assert len({row['name'] for row in rows}) == len(rows)
raw = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode() + b'\n'
packed = bytearray(gzip.compress(raw, compresslevel=9, mtime=0))
packed[9] = 0xff
with options.output.open('xb') as stream:
    stream.write(packed)
print(json.dumps(dict(cases=len(rows), sha256=hashlib.sha256(packed).hexdigest(),
                      decoded_sha256=hashlib.sha256(raw).hexdigest(), native_source_used=False)))
