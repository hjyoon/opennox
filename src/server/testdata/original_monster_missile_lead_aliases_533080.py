"""Execute all 64 original prediction bytes without replacing FP operations."""
import argparse, collections, gzip, hashlib, json, random, struct
from pathlib import Path
from capstone import Cs, CS_ARCH_X86, CS_MODE_32
from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_WRITE, UC_PROT_READ, UC_PROT_EXEC
from unicorn.x86_const import *

parser = argparse.ArgumentParser(description=__doc__)
for key in ['game', 'fixture', 'proof', 'input-seal']:
    parser.add_argument('--' + key, type=Path, required=True)
parser.add_argument('--control', type=lambda x: int(x, 0), choices=[0xe3f, 0x23f], required=True)
args = parser.parse_args()
sha = lambda raw: hashlib.sha256(raw).hexdigest()
blob = args.game.read_bytes()
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
original = bytes(u.mem_read(0x533080, 64))
assert sha(original) == '1aa0090f3d9bded2c626084381c21b52dbc357d735a3ea6353e4fa423ef6ca68'
u.mem_protect(base, size, UC_PROT_READ | UC_PROT_EXEC)
heap, stack, helpers = 0x21000000, 0x22000000, 0x23000000
u.mem_map(heap, 0x1000)
u.mem_map(stack, 0x10000)
u.mem_map(helpers, 0x1000)
unit, target, out = heap + 0x100, heap + 0x300, heap + 0x510
sp, stop, init, cw = stack + 0xe000, helpers + 0x100, helpers + 0x200, helpers + 0x300
u.mem_write(init, b'\xdb\xe3\xd9\x2d' + struct.pack('<I', cw))
u.mem_write(cw, struct.pack('<H', args.control))
bits = lambda v: struct.unpack('<I', struct.pack('<f', v))[0]
positions = [[bits(v) for v in row] for row in [
    [0, 0, 0, 0], [0, 0, 3, 4], [300, 300, 320, 300],
    [-300, 300, 320, -300], [16777216, 1, 16777218, -1],
    [0.125, -0.125, 0.25, -0.25], [1210.1971, 4426.965, 1930, 4915],
]]
positions += [[0, 0, v, v ^ 0x80000000] for v in [
    1, 0x800000, 0x7f7fffff, 0x7f800000, 0xff800000, 0x7fc12345, 0x7fa12345]]
speeds = [bits(v) for v in [0, -0.0, 1, -1, 0.1, 12, 27, 1e-20, 1e20]] + [1, 0x7f7fffff, 0x7f800000, 0x7fc12345]
velocities = [[bits(v) for v in row] for row in [[0, 0], [-0.0, 0], [1, -1], [12, 4], [0.1, -0.1], [1e20, -1e20]]]
velocities += [[1, 0x80000001], [0x7f7fffff, 0x7f7fffff], [0x7f800000, 0xff800000], [0x7fc12345, 0x7fa12345]]
patterns = [
 ([bits(100),bits(80),bits(148),bits(104)], [bits(2),bits(-1)], bits(12)),
 ([0,0,0,0], [bits(1),bits(2)], 0),
 ([0,0,bits(3),bits(4)], [0x7fc12345,0x7fa12345], 0),
 ([0,0,0x7fc12345,0xffc12345], [bits(1),bits(-1)], bits(12)),
 ([0,0,0x7fa12345,0xffa12345], [bits(1),bits(-1)], bits(12)),
 ([bits(-1),bits(2),bits(17),bits(-9)], [bits(-3),bits(7)], 0x7fa12345),
 ([1,0x80000001,0x00800000,0x80800000], [1,0x80000001], 1),
 ([bits(1e20),bits(-1e20),bits(-1e20),bits(1e20)], [bits(1e-20),bits(-1e-20)], bits(0.1)),
]
aliases = {'separate':heap+0x810,'unit-position':unit+56,'unit-position-y':unit+60,'target-position':target+56,'target-position-y':target+60,'target-velocity':target+80,'target-velocity-y':target+84}
inputs = [dict(position_bits=p,velocity_bits=v,speed_bits=s,alias=name)for name in aliases for p,v,s in patterns]
sealed = json.dumps(inputs, separators=(',', ':')).encode()
if args.input_seal.exists():
    assert args.input_seal.read_bytes() == sealed
else:
    with args.input_seal.open('xb') as f:
        f.write(sealed)
saved = {UC_X86_REG_EBX: 0x12345678, UC_X86_REG_ESI: 0x23456789,
         UC_X86_REG_EDI: 0x3456789a, UC_X86_REG_EBP: 0x456789ab}
active, executed, all_executed, writes = False, set(), set(), []
flags = collections.Counter()
def on_code(uc, address, length, _):
    if not active:
        return
    if address == stop:
        uc.emu_stop()
        return
    assert 0x533080 <= address and address + length <= 0x5330c0, hex(address)
    executed.update(range(address, address + length))
def on_write(uc, access, address, length, value, _):
    if active:
        assert out <= address and address + length <= out + 8, (hex(address), length)
        writes.append([address - out, length])
u.hook_add(UC_HOOK_CODE, on_code)
u.hook_add(UC_HOOK_MEM_WRITE, on_write)
rows = []
for index, case in enumerate(inputs):
    out = aliases[case["alias"]]
    guard = bytes((i * 37 + 0x65) & 255 for i in range(0x1000))
    u.mem_write(heap, guard)
    u.mem_write(unit, bytes(0x100)); u.mem_write(target, bytes(0x100))
    p = case['position_bits']
    u.mem_write(unit + 56, struct.pack('<II', *p[:2]))
    u.mem_write(target + 56, struct.pack('<II', *p[2:]))
    u.mem_write(target + 80, struct.pack('<II', *case['velocity_bits']))
    heap_before = bytes(u.mem_read(heap, 0x1000))
    u.mem_write(sp, struct.pack('<IIIII', stop, unit, target, case['speed_bits'], out))
    stack_before = bytes(u.mem_read(stack, 0x10000))
    u.reg_write(UC_X86_REG_ESP, sp)
    for reg, value in saved.items():
        u.reg_write(reg, value)
    u.emu_start(init, init + 8, count=8)
    assert u.reg_read(UC_X86_REG_FPCW) == args.control
    executed, writes, active = set(), [], True
    u.emu_start(0x533080, stop + 1, count=128)
    active = False
    assert u.reg_read(UC_X86_REG_EIP) == stop and u.reg_read(UC_X86_REG_ESP) == sp + 4
    assert u.reg_read(UC_X86_REG_EAX) == target
    assert all(u.reg_read(reg) == value for reg, value in saved.items())
    assert u.reg_read(UC_X86_REG_FPCW) == args.control
    assert bytes(u.mem_read(stack, 0x10000)) == stack_before
    after = bytes(u.mem_read(heap, 0x1000))
    left, right = out - heap, out + 8 - heap
    assert after[:left] == heap_before[:left] and after[right:] == heap_before[right:]
    assert writes == [[0, 4], [4, 4]] and len(executed) == 64
    all_executed.update(executed)
    status = u.reg_read(UC_X86_REG_FPSW)
    flags[str(status)] += 1
    rows.append(dict(index=index, **case, output_bits=list(struct.unpack('<II', u.mem_read(out, 8))), return_role='target', original_FPSW=status, unit_position_after_bits=list(struct.unpack('<II',u.mem_read(unit+56,8))), target_position_after_bits=list(struct.unpack('<II',u.mem_read(target+56,8))), target_velocity_after_bits=list(struct.unpack('<II',u.mem_read(target+80,8)))))
assert len(all_executed) == 64 and bytes(u.mem_read(0x533080, 64)) == original
fixture = dict(entry='0x533080', control=args.control, rows=rows, input_sha256=sha(sealed))
raw = json.dumps(fixture, separators=(',', ':')).encode()
with args.fixture.open('xb') as f:
    with gzip.GzipFile(filename='', mode='wb', fileobj=f, mtime=0, compresslevel=6) as z:
        z.write(raw)
proof = dict(status='passed', control=hex(args.control), contexts=len(rows),
    original_GAME_EXE_sha256=sha(blob), original_code_sha256=sha(original),
    original_body_bytes_executed=64, padding_bytes=0, input_sha256=sha(sealed),
    fixture_sha256=sha(args.fixture.read_bytes()), fixture_uncompressed_sha256=sha(raw),
    all_calls_execute_unchanged_original_code=True, no_services_or_FP_result_substitution=True,
    native_pointer_return_role='target', full_heap_except_output_and_stack_and_nonvolatile_guards=True,
    FPCW_preserved=True, original_FPSW_distribution=dict(flags),
    FPSW_scope='Recorded as original diagnostic evidence; no native status-word equivalence claimed.',
    overall_porting_100_percent=False)
with args.proof.open('x') as f:
    json.dump(proof, f, indent=2)
    f.write('\n')
print(json.dumps(proof))
