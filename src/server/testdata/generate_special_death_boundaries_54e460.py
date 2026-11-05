"""Add whole-memory observations to the sealed original-only producer.

The original source is hash-bound and remains unchanged. Only observation
statements are inserted; the producer still checks every original instruction,
write, complete heap/image/helper region, caller, register and FPU invariant.
No native implementation or native output is used.
"""
from pathlib import Path
import argparse
import gzip
import hashlib
import json
import struct
import sys

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--reference-dir', type=Path, default=Path(__file__).resolve().parent)
parser.add_argument('--exe', type=Path, required=True)
parser.add_argument('--control', required=True, choices=['0xe3f', '0x23f'])
parser.add_argument('--original-output', type=Path, required=True)
parser.add_argument('--original-proof', type=Path, required=True)
parser.add_argument('--output', type=Path, required=True)
parser.add_argument('--proof', type=Path, required=True)
options = parser.parse_args()
source_path = options.reference_dir / 'original_special_death_54e460.py'
source = source_path.read_bytes()
source_sha = hashlib.sha256(source).hexdigest()
assert source_sha == '495b201aab6158e286759e41a50027f19126e41c15e0754c78919bf5847d7f4e'
code = source.decode()

def replace(old, new):
    global code
    assert code.count(old) == 1, old
    code = code.replace(old, new)

replace('outputs, details = [], []', '''outputs, details = [], []
boundary_outputs, boundaries, image_reads = [], [], []
image_regions = [(mass_at-16, 64), (names_at-16, 88),
                 (index_at, 4), (fps_at, 4), (frame_at, 4)]
def capture_boundary():
    boundaries.append(dict(event_cursor=len(events),
        heap_sha256=sha(bytes(u.mem_read(heap, 0x10000))),
        image_sha256=sha(bytes(u.mem_read(base, image_size))),
        image_regions=[bytes(u.mem_read(address, size)).hex()
                       for address, size in image_regions],
        original_writes=[dict(item) for item in writes]))
''')
replace('        events.append(event)\n        mutate(plan)',
        '        capture_boundary()\n        events.append(event)\n        mutate(plan)')
replace('    seen.add(address)', '''    if address in (0x54e500, 0x54e58f):
        current_index = u.reg_read(UC_X86_REG_EDX)
        # The effective original address, including DWORD wrap, is observed
        # directly from the decoded instruction's actual index register.
        read_address = ((names_at + 4*current_index) if address == 0x54e500
                        else (mass_at + current_index)) & 0xffffffff
        size = 4 if address == 0x54e500 else 1
        try:
            data = bytes(u.mem_read(read_address, size)).hex()
        except UcError:
            data = None
        image_reads.append(dict(pc=hex(address), address=hex(read_address), data=data))
    seen.add(address)''')
replace('    events, frames, occurrences, writes = [], [], {}, []',
        '    events, frames, occurrences, writes = [], [], {}, []\n    boundaries, image_reads = [], []')
replace('    outputs.append(dict(name=current[\'name\']', '''    capture_boundary()
    boundary_outputs.append(dict(name=current['name'], boundaries=boundaries,
                                 image_reads=image_reads))
    outputs.append(dict(name=current['name']''')

sys.argv = [str(source_path), '--exe', str(options.exe), '--control', options.control,
            '--output', str(options.original_output), '--proof', str(options.original_proof)]
namespace = {'__name__': '__main__', '__file__': str(source_path)}
exec(compile(code, str(source_path), 'exec'), namespace)

# Capture immutable image transport independently from the native code. The
# named regions include all image writes performed by this reference's worlds.
original_image = namespace['original_image']
base = namespace['base']
names_at, mass_at = namespace['names_at'], namespace['mass_at']
regions = namespace['image_regions']
loaded_name_values = set(namespace['name_addresses'])
for address in range(names_at-16, names_at+56, 4):
    loaded_name_values.add(struct.unpack_from('<I', original_image, address-base)[0])
names = {}
for pointer in sorted(loaded_name_values):
    if base <= pointer and pointer+48 <= base+len(original_image):
        names[hex(pointer)] = original_image[pointer-base:pointer-base+48].hex()
transport = dict(base=base, image_size=len(original_image),
    original_image_sha256=hashlib.sha256(original_image).hexdigest(),
    regions=[dict(address=address, bytes=original_image[address-base:address-base+size].hex())
             for address, size in regions],
    names=names, named_addresses=namespace['name_addresses'],
    outputs=namespace['boundary_outputs'])
decoded = json.dumps(transport, sort_keys=True, separators=(',', ':')).encode()+b'\n'
encoded = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0)); encoded[9] = 0xff
with options.output.open('xb') as stream:
    stream.write(encoded)
proof = dict(original_producer_sha256=source_sha,
    original_output_sha256=hashlib.sha256(options.original_output.read_bytes()).hexdigest(),
    original_proof_sha256=hashlib.sha256(options.original_proof.read_bytes()).hexdigest(),
    input_cases=len(transport['outputs']),
    whole_original_boundaries=sum(len(row['boundaries']) for row in transport['outputs']),
    original_image_sha256=transport['original_image_sha256'],
    encoded_sha256=hashlib.sha256(encoded).hexdigest(),
    decoded_sha256=hashlib.sha256(decoded).hexdigest(),
    control=options.control, native_source_used=False, native_comparison_credit=False)
with options.proof.open('x') as stream:
    json.dump(proof, stream, indent=2); stream.write('\n')
print(json.dumps(proof), flush=True)
