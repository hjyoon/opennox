"""Independent scenarios for four raw death entries; no native implementation input."""
from pathlib import Path
import argparse, gzip, hashlib, json

ap = argparse.ArgumentParser(description=__doc__)
ap.add_argument('--output', type=Path, required=True)
a = ap.parse_args()
positions = [[0, 0], [0x80000000, 0], [0x3f800000, 0xbf800000], [0x43a00000, 0x43a50000], [0x7f800000, 0xff800000], [0x7fc12345, 0xffc54321], [1, 0x80000001], [0x7f7fffff, 0xff7fffff]]
rows = []
def add(entry, suffix, **changes):
    row = dict(name=f'{entry:08x}/{suffix}', entry=entry, unit=0,
               positions=[positions[3], positions[2]], flag_result=0, read_positions=True, script={})
    row.update(changes); rows.append(row)
for entry in [0x549d80, 0x549e00, 0x549e70, 0x54a250]:
    for unit in [0, 1]:
        for i, pos in enumerate(positions):
            for flag in [0, 1, 0x80000000, 0xffffffff]:
                add(entry, f'unit{unit}/position{i}/flag{flag:08x}', unit=unit, positions=[pos, list(reversed(pos))], flag_result=flag)
    services = 6 if entry == 0x549d80 else 5 if entry == 0x549e00 else 1
    for index in range(services):
        add(entry, f'live-position-after-service{index}', script={str(index): dict(positions=[positions[(index + 1) % len(positions)], positions[(index + 2) % len(positions)]])})
    for inner in [0x549d80, 0x549e00, 0x549e70, 0x54a250]:
        index = 1 if entry == 0x549d80 else 0
        add(entry, f'reenter-{inner:08x}', script={str(index): dict(positions=[positions[2], positions[3]], reenter=dict(entry=inner, unit=1), after_positions=[positions[4], positions[5]])})
    for unit in [-1, -2]:
        for read in [True, False]:
            add(entry, f'bad-unit{unit}/world-read{int(read)}', unit=unit, read_positions=read)
    add(entry, 'flag-world-result-changed-after-query', flag_result=0x80000000, script={'0': dict(flag_result=0)})
assert len({r['name'] for r in rows}) == len(rows)
decoded = json.dumps(rows, separators=(',', ':'), sort_keys=True).encode() + b'\n'
encoded = gzip.compress(decoded, mtime=0)
with a.output.open('xb') as stream:
    stream.write(encoded)
print(json.dumps(dict(cases=len(rows), sha256=hashlib.sha256(encoded).hexdigest(), decoded_sha256=hashlib.sha256(decoded).hexdigest(), no_engine_source_consulted=True)))
