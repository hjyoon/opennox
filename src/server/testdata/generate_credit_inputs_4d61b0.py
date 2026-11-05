"""Independent integer/alias inputs for the original generator-credit dependency."""
from pathlib import Path
import argparse
import copy
import gzip
import hashlib
import json

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
options = parser.parse_args()
base = dict(unit='normal', flags=0xabcdef00, update='normal', player='normal',
            counter=3, player_flags=0x12345670, alias='none')
rows = []
def add(name, **changes):
    row = copy.deepcopy(base)
    row.update(changes)
    row['name'] = name
    rows.append(row)
for flag in range(256):
    add(f'flags-low-byte-{flag:02x}', flags=0xabcdef00 | flag)
for counter in [0, 1, 0x7ffffffe, 0x7fffffff, 0x80000000, 0xfffffffe, 0xffffffff]:
    add(f'counter-{counter:08x}', counter=counter)
for flags in [0, 8, 0x20, 0x80000000, 0xffffffff]:
    add(f'player-flags-{flags:08x}', player_flags=flags)
for unit in ['nil', 'unmapped']:
    add(f'unit-{unit}', unit=unit)
for update in ['nil', 'unmapped', 'unit']:
    add(f'update-{update}', update=update)
    add(f'destroyed-skips-update-{update}', flags=0xffffffff, update=update)
for player in ['nil', 'unmapped', 'unit']:
    add(f'player-{player}', player=player)
    add(f'destroyed-skips-player-{player}', flags=0x20, player=player)
for alias in ['counter-is-update-player', 'counter-is-unit-update',
              'counter-is-unit-flags', 'player-flags-is-update-player']:
    add('alias-'+alias, alias=alias)
assert len({row['name'] for row in rows}) == len(rows)
decoded = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0))
encoded[9] = 0xff
with options.output.open('xb') as stream:
    stream.write(encoded)
print(json.dumps(dict(cases=len(rows), sha256=hashlib.sha256(encoded).hexdigest(),
                      decoded_sha256=hashlib.sha256(decoded).hexdigest(), native_source_used=False)))
