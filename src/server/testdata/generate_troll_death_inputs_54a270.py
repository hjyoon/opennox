from pathlib import Path
import argparse, copy, gzip, hashlib, json, struct

p = Path(__file__).resolve().parent
ap = argparse.ArgumentParser(description='Regenerate the frozen Troll death input columns, without native source or expected output.')
ap.add_argument('--output', type=Path, required=True)
args = ap.parse_args()
old_bytes = (p / 'troll_dead_cloud_54a270_startup_original.json').read_bytes()
assert hashlib.sha256(old_bytes).hexdigest() == '33bab941bd4b36f77eefed01b25e7894735c8965a00b093ac8a12a2ab5289e21'
old = json.loads(old_bytes)
assert len(old['rows']) == 30
rows = []
base = dict(unit=0, new=2, cache=0, resolved=71, fps=30,
            lifetime=struct.unpack('<Q', struct.pack('<d', 1.25))[0],
            positions=[[0x40e00000, 0x41000000], [0x41100000, 0x41200000], [0x42c80000, 0x43480000], [0x43960000, 0x43c80000]],
            updates=[4, 5, 4, 5], script={})
def add(name, **changes):
    row = copy.deepcopy(base); row.update(changes); row['name'] = name; rows.append(row)

# Historical inputs are reused, never their expected output or native implementation.
for item in old['rows']:
    lifetime = struct.unpack('<f', struct.pack('<I', item['lifetime_bits']))[0]
    script = {}
    mutation = item['mutation']
    if mutation:
        stage, action = mutation.split('-', 1)
        index = {'create': 2, 'audio': 3, 'balance': 4}[stage]
        if action in ['replace', 'nil']: script[str(index)] = {'update': [2, 5 if action == 'replace' else -1]}
        elif mutation == 'create-position': script[str(index)] = {'position': [0, 0x42c60000, 0x42ca0000]}
        elif mutation == 'balance-fps': script[str(index)] = {'fps': 31}
        else: raise AssertionError(mutation)
    add('historical-input-' + item['name'], lifetime=struct.unpack('<Q', struct.pack('<d', lifetime))[0], fps=item['fps'], new=-1 if item['allocation_failure'] else 2, script=script)

for cache in [0, 17, 0xffffffff]:
    for unit in [0, 1, -1, -2]:
        for new in [-1, 2]:
            add(f'cache-{cache:08x}-unit-{unit}-new-{new}', cache=cache, unit=unit, new=new)
for resolved in [0, 0xffffffff]: add(f'resolve-{resolved:08x}', resolved=resolved)
for new in [0, 1, 3, -2]: add(f'constructor-role-{new}', new=new)
for data in [-1, -2, 5]:
    updates = base['updates'].copy(); updates[2] = data
    add(f'initial-update-{data}', updates=updates)
for stage, index in [('new', 1), ('create', 2), ('audio', 3), ('balance', 4)]:
    add(stage + '-moves-source', script={str(index): {'position': [0, 0x447a0000, 0xc47a0000]}})
    for data in [-1, -2, 5]: add(stage + f'-replaces-update-{data}', script={str(index): {'update': [2, data]}})
for bits in [0, 0x8000000000000000, 0x3fe0000000000000, 0xbfe0000000000000, 0x3ff8000000000000, 0x3ff0000000000001, 0x3fefffffffffffff, 1, 0x0010000000000000, 0x7fefffffffffffff, 0x7ff0000000000000, 0xfff0000000000000, 0x7ff8000000001234]:
    for fps in [1, 30, 0x80000000, 0xffffffff]: add(f'double-{bits:016x}-fps-{fps:08x}', lifetime=bits, fps=fps)
for index in [0, 1, 2, 3, 4]: add(f'reentrant-service-{index}', script={str(index): {'reenter': 1}})
assert len({row['name'] for row in rows}) == len(rows)
decoded = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode() + b'\n'
encoded = bytearray(gzip.compress(decoded, mtime=0))
# Retain the canonical first-produced gzip OS byte across Python/zlib versions.
encoded[9] = 0x13
encoded = bytes(encoded)
with args.output.open('xb') as stream: stream.write(encoded)
print(json.dumps(dict(cases=len(rows), input_sha256=hashlib.sha256(encoded).hexdigest(), decoded_sha256=hashlib.sha256(decoded).hexdigest(), no_native_source_or_expected_output_used=True)))
