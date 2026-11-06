#!/usr/bin/env python3
"""Choose independent native field correspondence from chosen inputs only.

No expected golden, engine, predicate algorithm, original execution, Go or C
source is read. This computes initial bytes and individual field read domains,
not a predicate return, branch, read order, root fault or native expected result.
"""
from pathlib import Path
import argparse
import gzip
import hashlib
import json

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--inputs', type=Path, default=Path(__file__).resolve().parent/'damage_predicates_independent_inputs_4e1400.json.gz')
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
sha = lambda data: hashlib.sha256(data).hexdigest()
encoded = args.inputs.read_bytes()
assert sha(encoded) == '1f96b7f8b86e8fd9f37cd9ee8c71993b8f1c9a2093ca0e5ba9afb0586f6fdac6'
decoded = gzip.decompress(encoded)
assert sha(decoded) == 'a6f5a269759c38d3e850436c5acea2b61b629a17e18fd2eb6d4c27026971be55'
inputs = json.loads(decoded)
assert len(inputs) == 487
base, end, bad = 0x21000000, 0x21010000, 0x41000000
# Read kinds are source.class DWORD, source.subclass BYTE, weapon.class DWORD,
# weapon.subclass DWORD, weapon.UseData pointer, WandUseData.flags BYTE.
specs = [('source', 8, 4), ('source', 12, 1), ('weapon', 8, 4),
         ('weapon', 12, 4), ('weapon', 736, 4), ('use', 96, 1)]
rows = []
for case in inputs:
    def pointer(role):
        shape = case[role]
        if shape == 'null': return 0
        if shape == 'bad': return bad
        if shape == 'weapon': return pointer('weapon')
        edges = {'last-class': 12, 'last-subclass-byte': 13,
                 'last-subclass': 16, 'last-use-pointer': 740,
                 'last-flags-byte': 97}
        if shape in edges: return end-edges[shape]
        value = base+{'source': 0x1000, 'weapon': 0x3000, 'use': 0x5000}[role]
        if shape.startswith('unaligned'): return value+int(shape[len('unaligned'):])
        assert shape == 'valid'
        return value
    roles = {role: pointer(role) for role in ['source', 'weapon', 'use']}
    memory = bytearray(((i*37+case['seed']*19)^(i>>3))&255 for i in range(65536))
    initializers = []
    field_specs = [
        (0, case['source_class']), (1, case['source_subclass']),
        (2, case['weapon_class']), (3, case['weapon_subclass']),
        (4, roles['use']), (5, case['use_flags']),
    ]
    # Input initialization preserves all four bytes when readable, including
    # high bytes which a BYTE-only root read must ignore. Last-byte payloads
    # initialize exactly one byte and keep following bytes protected.
    initialized_fields = []
    for kind, value in field_specs:
        role, offset, _ = specs[kind]
        address, width = roles[role]+offset, 4
        if not (base <= address and address+width <= end):
            if kind in (1, 5) and base <= address < end: width = 1
            else: continue
        value &= (1 << (8*width))-1
        memory[address-base:address-base+width] = value.to_bytes(width, 'little')
        initializers.append(dict(address=address, width=width, value=value))
        initialized_fields.append((kind, address, width))
    fields = []
    seen = set()
    for kind, address, width in initialized_fields:
        # Explicit source=weapon alias has one actual shared native field.
        if (address, width) in seen: continue
        seen.add((address, width))
        role = specs[kind][0]
        fields.append(dict(kind=kind, role=roles[role], address=address,
                           pe_width=width, pointer=kind == 4,
                           value=int.from_bytes(memory[address-base:address-base+width], 'little')))
    reads = []
    for kind, (role, offset, width) in enumerate(specs):
        address = roles[role]+offset
        value = int.from_bytes(memory[address-base:address-base+width], 'little') if base <= address and address+width <= end else None
        reads.append(dict(role=roles[role], kind=kind, want=value))
    rows.append(dict(name=case['name'], entry=case['entry'], seed=case['seed'],
                     roles=roles, shapes={role: case[role] for role in roles},
                     initializer_words=initializers, native_fields=fields,
                     independent_reads=reads, initial_heap_sha256=sha(memory)))
raw = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode()+b'\n'
packed = gzip.compress(raw, mtime=0)
with args.output.open('xb') as output: output.write(packed)
print(json.dumps(dict(cases=len(rows), entries={entry:sum(row['entry']==entry for row in rows) for entry in ['attack','friendly']},
    encoded_sha256=sha(packed), decoded_sha256=sha(raw),
    individually_unreadable_field_domains=sum(read['want'] is None for row in rows for read in row['independent_reads']),
    original_outputs_or_native_sources_read=False, predicate_results_or_fault_paths_calculated=False)), flush=True)
