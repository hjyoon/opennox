"""Execute the unchanged PE32 video callback setter and descriptor callback.

No service is substituted. This is not a Windows renderer or framebuffer test.
"""
import collections
import hashlib
import json
import struct
import sys
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EIP, UC_X86_REG_ESP

pe = Path(sys.argv[1]).read_bytes()
assert hashlib.sha256(pe).hexdigest() == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
o = h + 24
base = struct.unpack_from("<I", pe, o + 28)[0]
size = struct.unpack_from("<I", pe, o + 56)[0]
s = o + struct.unpack_from("<H", pe, h + 20)[0]
sections = [struct.unpack_from("<IIII", pe, s + 40*i + 8) for i in range(struct.unpack_from("<H", pe, h + 6)[0])]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for _, virtual, length, raw in sections:
    cpu.mem_write(base + virtual, pe[raw:raw + length])
cpu.mem_map(0x20000000, 0x10000)
RECORD, STACK, STOP = 0x20000020, 0x20008000, 0x20009000
SLOT, COUNTER, TABLE = 0x6f7bb4, 0x9753a0, 0x9756e0
ranges = {0x48a210: 10, 0x48a21a: 6, 0x47fce0: 130, 0x47fd62: 14}
sealed = {at: bytes(cpu.mem_read(at, length)) for at, length in ranges.items()}
assert sealed[0x48a21a] == b"\x90"*6 and sealed[0x47fd62] == b"\x90"*14
entries, faults = collections.Counter(), []

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def hook(_cpu, address, _size, _):
    if address in (0x48a210, 0x47fce0):
        entries[hex(address)] += 1
    if address != STOP and not any(at <= address < at+length for at, length in ranges.items()):
        raise AssertionError("undeclared execution "+hex(address))

def fault(_cpu, access, address, size, value, _):
    faults.append([access, address, size, value])
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, fault)

def call(at, *args):
    cpu.mem_write(STACK-0x100, b"\xa5"*16)
    cpu.mem_write(STACK, struct.pack("<"+"I"*(len(args)+1), STOP, *(x & 0xffffffff for x in args)))
    cpu.mem_write(STACK+4*(len(args)+1), b"\x5a"*16)
    before = bytes(cpu.mem_read(STACK, 4*(len(args)+1)+16))
    cpu.reg_write(UC_X86_REG_ESP, STACK)
    cpu.emu_start(at, STOP, count=10000)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP
    assert bytes(cpu.mem_read(STACK, len(before))) == before
    assert bytes(cpu.mem_read(STACK-0x100, 16)) == b"\xa5"*16
    assert all(bytes(cpu.mem_read(a, n)) == sealed[a] for a, n in ranges.items())
    return cpu.reg_read(UC_X86_REG_EAX)

setter_rows = []
values = [0, 1, 0x47fce0, 0x20000020, 0x7fffffff, 0x80000000, 0x89abcdef, 0xffffffff]
for initial in (0, 0x11223344, 0xffffffff):
    for value in values:
        guard = b"\x13\x27\x3b\x4f" + struct.pack("<I", initial) + b"\x51\x65\x79\x8d"
        cpu.mem_write(SLOT-4, guard)
        result = call(0x48a210, value)
        after = bytes(cpu.mem_read(SLOT-4, len(guard)))
        assert result == value and get(SLOT) == value
        assert after[:4] == guard[:4] and after[8:] == guard[8:]
        setter_rows.append(dict(initial=initial, value=value, result=result, stored=get(SLOT), neighbors_unchanged=True))

descriptor_rows = []
cpu.mem_write(TABLE-16, b"\x3c"*16 + b"\xa7"*128 + b"\xc3"*16)
put(COUNTER, 0)
descriptor_values = [
    [640, 480, 32, 0],
    [640, 480, 32, 0],
    [800, 600, 16, 1],
    [640, 480, 32, 0],
    [0xffffffff, 0x80000000, 0x12345678, 0x89abcdef],
    [0xffffffff, 0x80000000, 0x12345678, 0x89abcdef],
]
for i, values in enumerate(descriptor_values):
    fields = [((0x10203040*(j+1)) ^ (i*0x11111111)) & 0xffffffff for j in range(27)]
    for word, value in zip((3, 2, 21, 26), values):
        fields[word] = value
    record = struct.pack("<27I", *fields)
    cpu.mem_write(RECORD-16, b"\xd6"*16 + record + b"\x6d"*16)
    result = call(0x47fce0, RECORD, [0, 1, 0xffffffff, 0x80000000, 7, 0][i])
    assert result == 1
    assert bytes(cpu.mem_read(RECORD-16, 140)) == b"\xd6"*16 + record + b"\x6d"*16
    assert bytes(cpu.mem_read(TABLE-16, 16)) == b"\x3c"*16
    assert bytes(cpu.mem_read(TABLE+128, 16)) == b"\xc3"*16
    descriptor_rows.append(dict(fields=fields, ignored_arg=[0, 1, 0xffffffff, 0x80000000, 7, 0][i],
        result=result, count=get(COUNTER), table_hex=bytes(cpu.mem_read(TABLE, 128)).hex(), input_unchanged=True))

assert [r["count"] for r in descriptor_rows] == [1, 1, 2, 2, 3, 3]
assert not faults
print(json.dumps(dict(
    executable_sha256=hashlib.sha256(pe).hexdigest(),
    ranges=[dict(address=hex(a), bytes=n, sha256=hashlib.sha256(sealed[a]).hexdigest()) for a, n in ranges.items()],
    callback_slot=hex(SLOT), descriptor_counter=hex(COUNTER), descriptor_table=hex(TABLE),
    setter_rows=setter_rows, descriptor_rows=descriptor_rows,
    unhooked_entries=dict(entries), declared_services={}, memory_faults=faults,
    executable_bytes_unchanged=True, guarded_inputs_unchanged=True,
    original_Windows_renderer_executed=False,
), indent=2))
