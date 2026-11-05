"""Scalar and graph cases chosen before a native predicate implementation."""
from pathlib import Path
import gzip
import hashlib
import itertools
import json

stage = Path(__file__).resolve().parent
cases = []
strengths = [0, 1, 2, 255, 65536, 0x7fffffff, 0x80000000,
             0x80000001, 0xffff0000, 0xfffffffe, 0xffffffff]
for entry in ['inversion', 'grip']:
    for strength, output in itertools.product(strengths, ['separate', 'field', 'header', 'unaligned']):
        cases.append(dict(name=f'{entry}/{strength:08x}/{output}', entry=entry,
                          strength=strength, effect='valid', output=output))
    for effect, output in [('null', 'separate'), ('bad', 'separate'),
                           ('null', 'bad'), ('bad', 'null'),
                           ('valid', 'null'), ('valid', 'bad'),
                           ('last-field', 'separate')]:
        cases.append(dict(name=f'{entry}/{effect}-{output}', entry=entry,
                          strength=0xffffffff, effect=effect, output=output))

def scan(name, **kwargs):
    case = dict(name='scan/'+name, entry='scan', item='valid', obj_class=0x3001000,
                data='valid', slots=['none', 'none'], strengths=[0, 0], alias=False)
    case.update(kwargs)
    cases.append(case)

scan('null-item', item='null', data='bad', slots=['bad', 'bad'])
scan('bad-item', item='bad')
for value in [0, 1, 4, 0x80000000]:
    scan(f'class-reject-{value:08x}', obj_class=value, data='bad', slots=['bad', 'bad'])
for value in [0x1000, 0x1000000, 0x2000000, 0x3001000, 0xffffffff]:
    scan(f'class-admit-{value:08x}', obj_class=value, slots=['none', 'grip'], strengths=[0, 1])
for data in ['null', 'bad']:
    scan('data-'+data, data=data)
for left, right in itertools.product(['none', 'inversion', 'null-fn', 'alien-fn', 'grip'], repeat=2):
    scan(f'identities-{left}-{right}', slots=[left, right], strengths=[0xffffffff, 1])
for left, right in itertools.product(strengths, repeat=2):
    scan(f'signed-{left:08x}-{right:08x}', slots=['grip', 'grip'], strengths=[left, right])
for position in [0, 1]:
    for value in strengths:
        slots = ['none', 'none']; slots[position] = 'grip'
        values = [0, 0]; values[position] = value
        scan(f'only-slot-{position+2}-{value:08x}', slots=slots, strengths=values)
scan('bad-first-slot', slots=['bad', 'grip'], strengths=[0, 1])
scan('bad-second-slot', slots=['none', 'bad'])
scan('early-success-skips-bad', slots=['grip', 'bad'], strengths=[1, 0])
scan('negative-continues-to-bad', slots=['grip', 'bad'], strengths=[0xffffffff, 0])
for value in strengths:
    scan(f'aliased-slots-{value:08x}', slots=['grip', 'grip'], strengths=[value, value], alias=True)
for number, case in enumerate(cases):
    case['seed'] = number + 1
    # These intentionally unrelated words exercise ignored callback arguments.
    case['opaque_args'] = [0x81000000 | number, 0xffffffff - number, 0x12345678 ^ number, 0x41000000 + 16*number]
assert len({case['name'] for case in cases}) == len(cases)
encoded = json.dumps(cases, sort_keys=True, separators=(',', ':')).encode()+b'\n'
import argparse
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--output', type=Path, required=True)
target = parser.parse_args().output
with target.open('xb') as stream:
    stream.write(gzip.compress(encoded, mtime=0))
print(json.dumps(dict(cases=len(cases), sha256=hashlib.sha256(target.read_bytes()).hexdigest())))
