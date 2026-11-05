"""Reproduce independent PE initial-heap and direction transport, without native source."""
import argparse
import gzip
import hashlib
import json
import struct
from pathlib import Path

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--exe', type=Path, required=True)
p.add_argument('--initial-heap-output', type=Path, required=True)
p.add_argument('--direction-output', type=Path, required=True)
args = p.parse_args()
sha = lambda b: hashlib.sha256(b).hexdigest()
packed = Path(__file__).with_name('skull_raw_independent_inputs_54f9a0.json.gz').read_bytes()
assert sha(packed) == '2fa912b92071a20903b0596adce24a80942874efeb6e80fe5b8d8f6def6dcc36'
inputs = json.loads(gzip.decompress(packed))
assert len(inputs) == 753

def pointer(role):
    if role == -1:
        return 0
    if role == -2:
        return 0x41000000
    assert 0 <= role < 8
    return 0x21000000 + (role+1)*0x1000

heaps = []
for current in inputs:
    data = bytearray((n*17+(n >> 7)*29) & 255 for n in range(0x10000))
    def put(role, offset, value, width=4):
        at = (role+1)*0x1000+offset
        data[at:at+width] = (value & ((1 << (8*width))-1)).to_bytes(width, 'little')
    for n in range(4):
        for offset, value, width in [
            (4, current['source_type'], 2), (8, current['target_class'], 4),
            (16, current['target_flags'] if n == 1 else current['flags'], 4),
            (56, current['position'][0]+n, 4), (60, current['position'][1]+n, 4),
            (124, current['direction'], 2), (176, current['radius_bits'], 4),
            (544, current['speed_bits'], 4), (700, pointer(current['collide']), 4),
            (748, pointer(6 if n == 3 else current['update']), 4),
        ]:
            put(n, offset, value, width)
    for role, offset, value, width in [
        (4, 0, current['timers'][0], 4), (4, 4, current['timers'][1], 4),
        (4, 8, current['state'], 1), (4, 12, current['projectile_type'], 4),
        (4, 48, current['enabled'], 1), (6, 0, 17, 4), (6, 4, 19, 4),
        (6, 8, 1, 1), (6, 12, 73, 4), (6, 48, 1, 1),
    ]:
        put(role, offset, value, width)
    heaps.append(dict(name=current['name'], heap_sha256=sha(data)))
decoded = json.dumps(heaps, sort_keys=True, separators=(',', ':')).encode()+b'\n'
output = bytearray(gzip.compress(decoded, compresslevel=9, mtime=0))
output[9] = 0xff
assert sha(output) == 'b4c2147a417a21e7583c8a049ab57c9c9d337f2eb5a9fe9b8a0f6eb0aebf3e7c'
with args.initial_heap_output.open('xb') as f:
    f.write(output)

raw = args.exe.read_bytes()
assert sha(raw) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', raw, 60)[0]
optional = pe+24
base = struct.unpack_from('<I', raw, optional+28)[0]
table = optional+struct.unpack_from('<H', raw, pe+20)[0]
image = bytearray(struct.unpack_from('<I', raw, optional+56)[0])
sections = []
for n in range(struct.unpack_from('<H', raw, pe+6)[0]):
    _, va, size, offset = struct.unpack_from('<IIII', raw, table+40*n+8)
    flags = struct.unpack_from('<I', raw, table+40*n+36)[0]
    image[va:va+size] = raw[offset:offset+size]
    sections.append((base+va, base+va+size, flags))
rows = []
for direction in [-32768, -1, *range(256), 256, 32767]:
    address = 0x5b6658 + direction*8
    file_sections = [flags for start, end, flags in sections if start <= address and address+8 <= end]
    rows.append(dict(address=f'0x{address:08X}', direction=direction,
                     executable_section=bool(file_sections and file_sections[0] & 0x20000000),
                     file_backed=bool(file_sections),
                     words=list(struct.unpack_from('<II', image, address-base))))
result = dict(original_exe_sha256=sha(raw), rows=rows, schema='original-skull-direction-data-v1',
              scope='Original file-backed pairs plus explicitly declared initialized-PE virtual-zero pair; no arbitrary native named-field alias certification.')
output = json.dumps(result, sort_keys=True, separators=(',', ':')).encode()+b'\n'
assert sha(output) == '3e1fe81326cb6c663fc63e12dcc35ac13e78b5c7c00a088ee78fc946e329d93d'
with args.direction_output.open('xb') as f:
    f.write(output)
print(json.dumps(dict(cases=len(heaps), directions=len(rows), native_source_used=False)), flush=True)
