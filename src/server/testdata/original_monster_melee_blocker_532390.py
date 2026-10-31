"""Execute the unchanged melee blocker callback; no service or FP substitution."""
import argparse, collections, gzip, hashlib, json, math, struct
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_32
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for key in ['game', 'fixture', 'proof']:
    parser.add_argument('--' + key, type=Path, required=True)
parser.add_argument('--control', type=lambda value: int(value, 0), choices=[0xe3f, 0x23f], required=True)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda raw: hashlib.sha256(raw).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
optional = pe + 24
base = struct.unpack_from('<I', blob, optional + 28)[0]
size = (struct.unpack_from('<I', blob, optional + 56)[0] + 4095) & ~4095
table = optional + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, size)
for i in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, rva, length, offset = struct.unpack_from('<IIII', blob, table + i * 40 + 8)
    if length:
        u.mem_write(base + rva, blob[offset:offset + length])
entry, extent, selection, distance = 0x532390, 176, 0x833c20, 0x833c24
code = bytes(u.mem_read(entry, extent))
assert sha(code) == '30557bd8dc8d0b0af82c46c58092dd560181d620e431ca1b866677c20b393eef'
instructions = {v.address: bytes(v.bytes) for v in Cs(CS_ARCH_X86, CS_MODE_32).disasm(code, entry)}
assert sum(map(len, instructions.values())) == extent and code[160:] == b'\xc3' + b'\x90' * 15
readonly = [(entry, code), (0x5b6658, bytes(u.mem_read(0x5b6658, 2048))),
            (0x583d40, bytes(u.mem_read(0x583d40, 4))), (0x58350c, bytes(u.mem_read(0x58350c, 4)))]
assert sha(readonly[1][1]) == 'fbfdbd07ab3c6fbec2eb3768e69f1c32eb63c77b6e80e9957b9567c78d61bfcf'
heap, stack, helpers = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x6000)
u.mem_map(stack, 0x4000)
u.mem_map(helpers, 0x1000)
unit, candidate, prior = heap + 0x1000, heap + 0x2000, heap + 0x3000
sp, stop, init, cw = stack + 0x3000, helpers + 0x100, helpers + 0x200, helpers + 0x300
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
wr = lambda address, value: u.mem_write(address, struct.pack('<I', value & 0xffffffff))
rd = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
bits = lambda value: struct.unpack('<I', struct.pack('<f', value))[0]
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
active, visited, writes = False, set(), []
def on_code(uc, address, length, _):
    if active:
        assert address in instructions and bytes(uc.mem_read(address, length)) == instructions[address]
        visited.add(address)
def on_write(uc, access, address, length, value, _):
    if active:
        assert length == 4 and address in [sp - 4, sp + 4, selection, distance], (hex(address), length)
        writes.append((address, value & 0xffffffff))
u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)
inputs = []
def add(family, direction, ux, uy, cx, cy, threshold, cls=4, flags=0, initial=1):
    inputs.append({'family': family, 'direction': direction, 'unit_x': ux, 'unit_y': uy,
                   'candidate_x': cx, 'candidate_y': cy, 'threshold': threshold,
                   'class': cls, 'flags': flags, 'initial_selection': initial})
zero = bits(0)
points = [(0, 0), (1, 0), (-1, 0), (0, 1), (0, -1), (1, 1), (-1, 1), (1, -1),
          (49.9999, 0), (50, 0), (50.9999, 0), (51, 0), (3, 4), (1e-20, 1e-20)]
positions = [(bits(x), bits(y)) for x, y in points]
positions += [(bits(1), value) for value in [0x3fddb3d6, 0x3fddb3d7, 0x3fddb3d8]]
for direction in range(256):
    for cx, cy in positions:
        for threshold in [bits(0), bits(5.0001), bits(51)]:
            add('directions-and-boundaries', direction, zero, zero, cx, cy, threshold)
for cls in [0, 1, 2, 4, 6, 8, 0x100, 0x102, 0x104, 0x80000000, 0xffffffff]:
    for flags in [0, 1, 0x7fff, 0x8000, 0x8001, 0xffff7fff, 0xffff8000, 0xffffffff]:
        for initial in [0, 1]:
            add('class-flags-and-sticky-selection', 0, zero, zero, bits(1), zero, bits(51), cls, flags, initial)
for origin in [(320, 300), (-320, -300), (1e8, -1e8), (1e-20, -1e-20)]:
    for dx, dy in points:
        for direction in [0, 31, 64, 127, 128, 191, 192, 255]:
            add('nonzero-origins', direction, bits(origin[0]), bits(origin[1]), bits(origin[0] + dx), bits(origin[1] + dy), bits(51))
for value in [0x00000001, 0x007fffff, 0x00800000, 0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000, 0x7fc00001, 0x7fa00001]:
    for direction in [0, 32, 64, 128, 192, 255]:
        for threshold in [bits(-1), bits(51), 0x7f800000, 0x7fc00001]:
            add('nonfinite-and-extreme-x', direction, zero, zero, value, bits(1), threshold)
            add('nonfinite-and-extreme-y', direction, zero, zero, bits(1), value, threshold)
rows = []
for item in inputs:
    u.mem_write(heap, b'\xa5' * 0x6000)
    u.mem_write(stack, b'\xcc' * 0x4000)
    u.mem_write(selection - 64, b'\x6b' * 136)
    for offset, value in [(8, item['class']), (16, item['flags']), (56, item['candidate_x']), (60, item['candidate_y'])]:
        wr(candidate + offset, value)
    wr(unit + 56, item['unit_x']); wr(unit + 60, item['unit_y'])
    u.mem_write(unit + 124, struct.pack('<h', item['direction']))
    initial_pointer = prior if item['initial_selection'] else 0
    wr(selection, initial_pointer); wr(distance, item['threshold'])
    wr(sp, stop); wr(sp + 4, candidate); wr(sp + 8, unit)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(init, init + 8, count=4)
    assert u.reg_read(UC_X86_REG_FPCW) == args.control
    for reg, value in saved.items(): u.reg_write(reg, value)
    heap_before = bytes(u.mem_read(heap, 0x6000))
    stack_before = bytearray(u.mem_read(stack, 0x4000))
    global_before = bytearray(u.mem_read(selection - 64, 136))
    writes = []; active = True
    u.emu_start(entry, stop, count=100)
    active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
    assert all(u.reg_read(reg) == value for reg, value in saved.items())
    assert u.reg_read(UC_X86_REG_FPCW) == args.control
    assert bytes(u.mem_read(heap, 0x6000)) == heap_before
    for address, value in writes:
        target, offset = (stack_before, address - stack) if address in [sp - 4, sp + 4] else (global_before, address - selection + 64)
        struct.pack_into('<I', target, offset, value)
    assert bytes(u.mem_read(stack, 0x4000)) == stack_before
    assert bytes(u.mem_read(selection - 64, 136)) == global_before
    selected = rd(selection)
    assert selected in [0, prior, candidate]
    rows.append({'name': f'{len(rows):05d}', 'input': item,
                 'output': {'selection': 'candidate' if selected == candidate else 'prior' if selected == prior else 'zero',
                            'distance': rd(distance), 'global_writes': [{'field': 'selection' if a == selection else 'distance', 'value': 'candidate' if a == selection else v} for a, v in writes if a in [selection, distance]]}})
for address, raw in readonly: assert bytes(u.mem_read(address, len(raw))) == raw
assert sha(args.game.read_bytes()) == sha(blob)
fixture = {'schema': 'original-monster-melee-blocker-532390-v1', 'original_sha256': sha(blob),
           'control_word': args.control, 'entry': entry, 'extent': extent, 'declared_services': [], 'rows': rows}
raw = (json.dumps(fixture, sort_keys=True, separators=(',', ':')) + '\n').encode()
with args.fixture.open('xb') as output:
    with gzip.GzipFile(filename='', mode='wb', fileobj=output, compresslevel=9, mtime=0) as stream: stream.write(raw)
proof = {'status': 'passed', 'contexts': len(rows), 'per_family': dict(collections.Counter(x['family'] for x in inputs)),
         'control_word': hex(args.control), 'original_sha256': sha(blob), 'original_code_sha256': sha(code),
         'loaded_original_bytes': extent, 'executed_original_bytes': sum(len(instructions[a]) for a in visited),
         'code_table_constants_heap_stack_global_guards_and_nonvolatile_checked': True,
         'FPU_control_unchanged_status_flags_are_outputs_not_forced': True, 'no_service_or_numeric_substitution': True,
         'fixture_sha256': sha(args.fixture.read_bytes()), 'raw_fixture_sha256': sha(raw), 'original_unchanged': True}
with args.proof.open('x') as output: json.dump(proof, output, indent=2); output.write('\n')
print(json.dumps(proof))
