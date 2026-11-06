"""Independent initial PE bytes and native-field/read correspondence.

The ONLY input file read is the previously chosen independent input gzip.
No original output, emulator, native engine or expected root result is read.
This generator performs no Readiness slot-selection or callback algorithm.
"""
from pathlib import Path
import argparse, gzip, hashlib, json

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--inputs', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
options = parser.parse_args()
encoded_inputs = options.inputs.read_bytes()
assert hashlib.sha256(encoded_inputs).hexdigest() == 'cb6330a5622ee63f15cc64b2979d3521fac88df4b3184651517bdc2e88987d4f'
decoded_inputs = gzip.decompress(encoded_inputs)
assert hashlib.sha256(decoded_inputs).hexdigest() == 'daca8b8c2e4e0ed897a33394dba93657b15cfbfa78dc0bc3ffb3f09f22e8413f'
inputs = json.loads(decoded_inputs)
assert len(inputs) == 113
base, bad = 0x21000000, 0x41000000
rows = []

for case in inputs:
    shapes = {role: case.get(role, 'valid') for role in
              ['item', 'table', 'first', 'second', 'replacement', 'alternate_table']}
    roles = {}
    for role, shape in shapes.items():
        if shape == 'null':
            pointer = 0
        elif shape == 'bad':
            pointer = bad
        elif shape.startswith('last-'):
            # Independently chosen PE read extents; native equivalents are
            # placed using actual offsetof/sizeof in the Go/C controls.
            extent = {'last-init': 692+4, 'last-class': 8+4,
                      'last-slot2': 2*4+4, 'last-slot3': 3*4+4,
                      'last-callback': 40+4, 'last-strength': 48+4}[shape]
            pointer = base+65536-extent
        else:
            assert shape == 'valid'
            pointer = base+{'item': 0x3000, 'table': 0x5000, 'first': 0x7000,
                            'second': 0x9000, 'replacement': 0xb000,
                            'alternate_table': 0xd000}[role]
        roles[role] = pointer

    fieldspec = [('item', 0, 0, 692, 'pointer', roles['table']),
                 ('item', 1, 0, 8, 'word', case['obj_class'])]
    for index, value in enumerate(case['slots']):
        pointer = 0 if value == 'null' else bad if value == 'bad' else roles[value]
        fieldspec.append(('table', 2, index, 4*index, 'pointer', pointer))
    for index, value in enumerate([bad, bad, roles['replacement'], 0]):
        fieldspec.append(('alternate_table', 2, index, 4*index, 'pointer', value))
    for role in ['first', 'second', 'replacement']:
        fieldspec += [(role, 3, 0, 40, 'pointer', case['callbacks'][role]),
                      (role, 4, 0, 48, 'word', case['strengths'][role])]
    assert len(fieldspec) == 16
    memory = bytearray(((i*37+case['seed']*19) ^ (i >> 3)) & 255 for i in range(65536))
    native_fields, initializer_words, reads = [], [], []
    for role, kind, index, offset, typ, value in fieldspec:
        address = roles[role]+offset
        if base <= address and address+4 <= base+65536:
            memory[address-base:address-base+4] = value.to_bytes(4, 'little')
            native_fields.append(dict(role=roles[role], address=address, kind=kind,
                                      index=index, type=typ, value=value))
            initializer_words.append(dict(address=address, value=value))
    for role, kind, index, offset, typ, value in fieldspec:
        address = roles[role]+offset
        want = int.from_bytes(memory[address-base:address-base+4], 'little') if base <= address and address+4 <= base+65536 else None
        reads.append(dict(role=roles[role], kind=kind, index=index, want=want))
    rows.append(dict(name=case['name'], seed=case['seed'], roles=roles, shapes=shapes,
                     initializer_words=initializer_words, native_fields=native_fields,
                     independent_reads=reads, callback_values=sorted(set(case['callbacks'].values())),
                     initial_heap_sha256=hashlib.sha256(memory).hexdigest()))

decoded = json.dumps(rows, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded = gzip.compress(decoded, mtime=0)
with options.output.open('xb') as stream:
    stream.write(encoded)
print(json.dumps(dict(cases=len(rows), fields_per_case=16,
                     independent_read_domains=sum(len(row['independent_reads']) for row in rows),
                     independent_read_fault_domains=sum(read['want'] is None for row in rows for read in row['independent_reads']),
                     encoded_sha256=hashlib.sha256(encoded).hexdigest(),
                     decoded_sha256=hashlib.sha256(decoded).hexdigest(),
                     original_outputs_read=0, original_or_native_engine_executions=0)), flush=True)
