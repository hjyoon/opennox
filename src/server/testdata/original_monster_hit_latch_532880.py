"""Execute unchanged 532880 with full guards and no service substitutions."""
import argparse, gzip, hashlib, itertools, json, struct
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_32
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_READ, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for name in ['game', 'fixture', 'proof']:
    parser.add_argument('--' + name, type=Path, required=True)
parser.add_argument('--control', type=lambda s: int(s, 0), choices=[0xe3f, 0x23f], required=True)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda value: hashlib.sha256(value).hexdigest()
assert sha(blob) == '0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a'
pe = struct.unpack_from('<I', blob, 60)[0]
opt = pe + 24
base = struct.unpack_from('<I', blob, opt + 28)[0]
size = (struct.unpack_from('<I', blob, opt + 56)[0] + 4095) & ~4095
tab = opt + struct.unpack_from('<H', blob, pe + 20)[0]
u = Uc(UC_ARCH_X86, UC_MODE_32)
u.mem_map(base, size)
for i in range(struct.unpack_from('<H', blob, pe + 6)[0]):
    _, rva, length, offset = struct.unpack_from('<IIII', blob, tab + i * 40 + 8)
    if length:
        u.mem_write(base + rva, blob[offset:offset + length])
entry, extent, frame_address = 0x532880, 48, 0x84ea04
original_code = bytes(u.mem_read(entry, extent))
assert sha(original_code[:33]) == '73e524dcc8ceaeacaa15850d264310cd0465da1085324a7de94b6564b55e5b7c'
assert original_code[32:] == b'\xc3' + b'\x90' * 15
instructions = {v.address: bytes(v.bytes) for v in Cs(CS_ARCH_X86, CS_MODE_32).disasm(original_code, entry)}
assert sum(map(len, instructions.values())) == extent
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x8000)
u.mem_map(stack, 0x10000)
u.mem_map(code, 0x1000)
unit, records = heap + 0x1000, [heap + 0x3000, heap + 0x5000]
sp, stop, init, cw = stack + 0xe000, code + 0x100, code + 0x200, code + 0x300
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
wr = lambda address, value: u.mem_write(address, struct.pack('<I', value & 0xffffffff))
rd = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
active, visited, current_writes, current_frame_reads = False, set(), [], []

def on_code(uc, address, length, _):
    if active:
        assert address in instructions and bytes(uc.mem_read(address, length)) == instructions[address]
        visited.add(address)

def on_write(uc, access, address, length, value, _):
    if active:
        assert (address, length) == (selected + 520, 4), (hex(address), length)
        current_writes.append({'size': length, 'value': value & 0xffffffff})

def on_read(uc, access, address, length, value, _):
    if active and address == frame_address:
        assert length == 4
        current_frame_reads.append(rd(address))

u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)
u.hook_add(UC_HOOK_MEM_READ, on_read)
numbers = [0, 1, 117, 0xffff, 0x10000, 0x7fffffff, 0x80000000, 0xffffffff]
classes = [0, 1, 2, 4, 6, 0xffffffff]
flags = [0, 1, 0x8000, 0xffffffff]
rows = []
for frame, initial, cls, bits, slot in itertools.product(numbers, numbers, classes, flags, [0, 1]):
    selected = records[slot]
    u.mem_write(heap, b'\xa5' * 0x8000)
    u.mem_write(stack, b'\xcc' * 0x10000)
    u.mem_write(frame_address - 64, b'\x6b' * 132)
    wr(unit + 8, cls)
    wr(unit + 16, bits)
    wr(unit + 748, selected)
    wr(selected + 520, initial)
    wr(records[1 - slot] + 520, initial ^ 0xffffffff)
    wr(frame_address, frame)
    wr(sp, stop)
    wr(sp + 4, unit)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(init, init + 8, count=4)
    assert u.reg_read(UC_X86_REG_FPCW) == args.control
    fpu = [u.reg_read(v) for v in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]]
    for reg, value in saved.items():
        u.reg_write(reg, value)
    heap_before = bytearray(u.mem_read(heap, 0x8000))
    stack_before = bytes(u.mem_read(stack, 0x10000))
    frame_before = bytes(u.mem_read(frame_address - 64, 132))
    current_writes, current_frame_reads = [], []
    active = True
    u.emu_start(entry, stop, count=100)
    active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
    assert u.reg_read(UC_X86_REG_EAX) == selected
    assert all(u.reg_read(reg) == value for reg, value in saved.items())
    assert [u.reg_read(v) for v in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]] == fpu
    result = rd(selected + 520)
    struct.pack_into('<I', heap_before, selected + 520 - heap, result)
    assert bytes(u.mem_read(heap, 0x8000)) == heap_before
    assert bytes(u.mem_read(stack, 0x10000)) == stack_before
    assert bytes(u.mem_read(frame_address - 64, 132)) == frame_before
    assert bytes(u.mem_read(entry, extent)) == original_code
    assert len(current_writes) == (1 if initial == 0 else 0)
    assert len(current_frame_reads) == (1 if initial == 0 else 0)
    rows.append({'name': f'{len(rows):04d}', 'input': {'frame': frame, 'initial': initial,
        'class': cls, 'flags': bits, 'record_slot': slot, 'other_initial': initial ^ 0xffffffff},
        'output': {'return_kind': 'selected_update', 'timestamp': result,
        'frame_reads': current_frame_reads, 'timestamp_writes': current_writes,
        'complete_other_heap_stack_frame_guards_code_nonvolatile_and_FPU_unchanged': True}})
assert len(rows) == 3072
executed = sum(len(instructions[address]) for address in visited)
assert executed == 33
fixture = {'schema': 'original-monster-hit-latch-532880-v1', 'original_sha256': sha(blob),
    'entry': entry, 'extent': extent, 'body_bytes': 33, 'code_sha256': sha(original_code),
    'frame_address': frame_address, 'control_word': args.control, 'declared_services': [], 'rows': rows}
raw = (json.dumps(fixture, sort_keys=True, separators=(',', ':')) + '\n').encode()
with args.fixture.open('xb') as output:
    with gzip.GzipFile(filename='', mode='wb', fileobj=output, compresslevel=9, mtime=0) as stream:
        stream.write(raw)
assert sha(args.game.read_bytes()) == sha(blob)
proof = {'status': 'passed', 'contexts': len(rows), 'control_word': hex(args.control),
    'original_sha256': sha(blob), 'entry': hex(entry), 'extent': extent, 'executed_original_bytes': executed,
    'padding_bytes_not_executed': 15, 'code_sha256': sha(original_code), 'frame_address': hex(frame_address),
    'no_service_or_result_substitution': True, 'full_heap_stack_frame_guards_code_nonvolatile_and_FPU_preserved': True,
    'conditional_frame_read_and_exact_DWORD_write_observed': True,
    'fixture_sha256': sha(args.fixture.read_bytes()), 'raw_fixture_sha256': sha(raw),
    'raw_fixture_bytes': len(raw), 'original_unchanged': True, 'overall_porting_100_percent': False}
with args.proof.open('x') as stream:
    json.dump(proof, stream, indent=2); stream.write('\n')
print(json.dumps(proof))
