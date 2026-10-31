"""Execute unchanged532040 with guarded inputs and no substituted services."""
import argparse, collections, gzip, hashlib, itertools, json, struct
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_32
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for name in ['game', 'fixture', 'proof']:
    parser.add_argument('--' + name, type=Path, required=True)
parser.add_argument('--control', type=lambda value: int(value, 0), choices=[0xe3f, 0x23f], required=True)
args = parser.parse_args()
blob = args.game.read_bytes()
sha = lambda value: hashlib.sha256(value).hexdigest()
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
entry, extent, global_address = 0x532040, 48, 0x833c1c
original_code = bytes(u.mem_read(entry, extent))
instructions = {v.address: bytes(v.bytes) for v in Cs(CS_ARCH_X86, CS_MODE_32).disasm(original_code, entry)}
assert sum(map(len, instructions.values())) == extent
assert original_code[41:] == b'\xc3' + b'\x90' * 6
assert global_address == 0x5d4594 + 2487944
heap, stack, code = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x8000)
u.mem_map(stack, 0x10000)
u.mem_map(code, 0x1000)
candidate, context, other_update = heap + 0x1000, heap + 0x3000, heap + 0x5000
sp, stop, init, cw = stack + 0xe000, code + 0x100, code + 0x200, code + 0x300
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
wr = lambda address, value: u.mem_write(address, struct.pack('<I', value & 0xffffffff))
rd = lambda address: struct.unpack('<I', u.mem_read(address, 4))[0]
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
active, visited, current_writes = False, set(), []

def on_code(uc, address, length, _):
    if not active:
        return
    assert address in instructions and bytes(uc.mem_read(address, length)) == instructions[address]
    visited.add(address)

def on_write(uc, access, address, length, value, _):
    if not active:
        return
    assert (address, length) in [(global_address, 4), (sp - 4, 4)], (hex(address), length)
    if address == global_address:
        current_writes.append({'size': length, 'value': value & 0xffffffff})

u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)
numbers = [0, 1, 117, 0x7fffffff, 0x80000000, 0xffffffff]
flags = [0, 1, 0x7fff, 0x8000, 0x8001, 0xffff, 0x10000, 0xffff7fff, 0xffff8000, 0xffffffff]
classes = [0, 2, 4, 6, 0xffffffff]
rows = []
for number, wanted, bits, initial, cls in itertools.product(numbers, numbers, flags, numbers, classes):
    u.mem_write(heap, b'\xa5' * 0x8000)
    u.mem_write(stack, b'\xcc' * 0x10000)
    u.mem_write(global_address - 64, b'\x6b' * 132)
    wr(candidate + 8, cls)
    wr(candidate + 16, bits)
    wr(candidate + 36, number)
    # This is deliberately a different valid record. The callback must use
    # its explicit context argument rather than the object's UpdateData link.
    wr(candidate + 748, other_update)
    wr(context + 1200, wanted)
    wr(other_update + 1200, wanted ^ 0xffffffff)
    wr(global_address, initial)
    wr(sp, stop)
    wr(sp + 4, candidate)
    wr(sp + 8, context)
    u.reg_write(UC_X86_REG_ESP, sp)
    u.emu_start(init, init + 8, count=4)
    assert u.reg_read(UC_X86_REG_FPCW) == args.control
    fpu = [u.reg_read(v) for v in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]]
    for reg, value in saved.items():
        u.reg_write(reg, value)
    heap_before = bytes(u.mem_read(heap, 0x8000))
    stack_before = bytearray(u.mem_read(stack, 0x10000))
    guards_before = bytearray(u.mem_read(global_address - 64, 132))
    current_writes = []
    active = True
    u.emu_start(entry, stop, count=100)
    active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
    assert u.reg_read(UC_X86_REG_EAX) == candidate
    assert all(u.reg_read(reg) == value for reg, value in saved.items())
    assert [u.reg_read(v) for v in [UC_X86_REG_FPCW, UC_X86_REG_FPSW, UC_X86_REG_FPTAG]] == fpu
    assert bytes(u.mem_read(heap, 0x8000)) == heap_before
    struct.pack_into('<I', stack_before, sp - 4 - stack, saved[UC_X86_REG_ESI])
    assert bytes(u.mem_read(stack, 0x10000)) == stack_before
    result = rd(global_address)
    struct.pack_into('<I', guards_before, 64, result)
    assert bytes(u.mem_read(global_address - 64, 132)) == guards_before
    assert bytes(u.mem_read(entry, extent)) == original_code
    rows.append({'name': f'{len(rows):05d}', 'input': {'candidate_netcode': number, 'context_netcode': wanted,
        'flags': bits, 'initial_global': initial, 'class': cls, 'other_update_netcode': wanted ^ 0xffffffff},
        'output': {'return_value': {'kind': 'candidate'}, 'global': result, 'global_writes': current_writes,
                   'complete_candidate_context_other_record_and_guards_unchanged': True}})

assert len(rows) == 10800
executed = sum(len(instructions[address]) for address in visited)
assert executed == 42
fixture = {'schema': 'original-monster-dead-target-532040-v1', 'original_sha256': sha(blob),
    'entry': entry, 'extent': extent, 'code_sha256': sha(original_code), 'global_address': global_address,
    'control_word': args.control, 'declared_services': [], 'rows': rows}
raw = (json.dumps(fixture, sort_keys=True, separators=(',', ':')) + '\n').encode()
with args.fixture.open('xb') as output:
    with gzip.GzipFile(filename='', mode='wb', fileobj=output, compresslevel=9, mtime=0) as stream:
        stream.write(raw)
assert sha(args.game.read_bytes()) == sha(blob)
proof = {'status': 'passed', 'contexts': len(rows), 'control_word': hex(args.control),
    'original_sha256': sha(blob), 'entry': hex(entry), 'extent': extent, 'executed_original_bytes': executed,
    'padding_bytes_not_executed': 6, 'code_sha256': sha(original_code), 'global_address': hex(global_address),
    'no_service_or_result_substitution': True, 'full_heap_stack_global_guards_code_nonvolatile_and_FPU_preserved': True,
    'distinct_explicit_context_and_object_update_link': True, 'fixture_sha256': sha(args.fixture.read_bytes()),
    'raw_fixture_sha256': sha(raw), 'raw_fixture_bytes': len(raw), 'original_unchanged': True,
    'overall_porting_100_percent': False}
with args.proof.open('x') as stream:
    json.dump(proof, stream, indent=2)
    stream.write('\n')
print(json.dumps(proof))
