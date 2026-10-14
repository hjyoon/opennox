"""Execute unchanged PE32 RGB5551 setup, unpack, and material functions.

The setup prefix selects the original format table; this is not Windows
video initialization or a framebuffer oracle. The MMX lane cache is disabled.
"""
import collections
import hashlib
import json
import struct
import sys
from pathlib import Path

from unicorn import Uc, UC_ARCH_X86, UC_MODE_32, UC_HOOK_CODE, UC_HOOK_MEM_INVALID
from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EBX, UC_X86_REG_ECX, UC_X86_REG_EDX, UC_X86_REG_EDI, UC_X86_REG_ESI, UC_X86_REG_EBP, UC_X86_REG_EIP, UC_X86_REG_ESP

pe = Path(sys.argv[1]).read_bytes()
original = hashlib.sha256(pe).hexdigest()
assert original == "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a"
h = struct.unpack_from("<I", pe, 0x3c)[0]
o = h + 24
base, size = struct.unpack_from("<I", pe, o + 28)[0], struct.unpack_from("<I", pe, o + 56)[0]
s = o + struct.unpack_from("<H", pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (size + 4095) & ~4095)
for i in range(struct.unpack_from("<H", pe, h + 6)[0]):
    _, virtual, length, raw = struct.unpack_from("<IIII", pe, s + 40*i + 8)
    cpu.mem_write(base + virtual, pe[raw:raw + length])
cpu.mem_map(0x20000000, 0x10000)
DATA, OUT, STACK, STOP = 0x20000100, 0x20001000, 0x20008000, 0x20009000
DATA_SIZE = 1068
ranges = {0x4338d0: 137, 0x4340a0: 292, 0x4341d0: 333, 0x4351c0: 56, 0x435280: 81}
sealed = {at: bytes(cpu.mem_read(at, n)) for at, n in ranges.items()}
table = bytes(cpu.mem_read(0x583828, 160))
entries, faults = collections.Counter(), []

def put(at, value):
    cpu.mem_write(at, struct.pack("<I", value & 0xffffffff))

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def hook(_cpu, at, _length, _):
    if at in ranges:
        entries[hex(at)] += 1
    if at == STOP or any(start <= at < start + n for start, n in ranges.items()):
        return
    raise AssertionError("undeclared execution " + hex(at))

def invalid(_cpu, access, address, length, value, _):
    faults.append(dict(access=access, address=address, length=length, eip=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
put(0x973fdc, 1)
cpu.reg_write(UC_X86_REG_ESP, STACK)
cpu.emu_start(0x4338d0, 0x433959, count=1000)
assert cpu.reg_read(UC_X86_REG_EIP) == 0x433959
assert bytes(cpu.mem_read(0x975260, 160)) == table
assert get(0x69a014) == 0x4351c0 and get(0x975240) == 0x435280
mode = dict(index=get(0x973fdc), pack=get(0x69a014), unpack=get(0x975240),
            masks_and_shifts=list(struct.unpack("<6I", cpu.mem_read(0x975260, 24))))
put(0x973fa8, DATA)
put(0x974860, 0)

def call(at, *args):
    cpu.mem_write(STACK - 1024, bytes([0xa5])*1024)
    cpu.mem_write(STACK, struct.pack("<" + "I"*(len(args)+1), STOP, *(v & 0xffffffff for v in args)))
    for reg in (UC_X86_REG_EAX, UC_X86_REG_ECX, UC_X86_REG_EDX):
        cpu.reg_write(reg, 0x5aa55aa5)
    saved = {UC_X86_REG_EBX: 0x11223344, UC_X86_REG_ESI: 0x55667788, UC_X86_REG_EDI: 0x99aabbcc, UC_X86_REG_EBP: 0xddeeff11}
    for reg, value in saved.items():
        cpu.reg_write(reg, value)
    cpu.reg_write(UC_X86_REG_ESP, STACK)
    cpu.emu_start(at, STOP, count=1000)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
    assert all(cpu.reg_read(reg) == value for reg, value in saved.items())
    assert not faults

split_digest = hashlib.sha256()
samples = []
sample_words = {0, 1, 0x001f, 0x03e0, 0x7c00, 0x4210, 0x7fff, 0x8000, 0xffff}
for word in range(65536):
    cpu.mem_write(OUT, bytes([0xa5])*16)
    call(0x435280, word, OUT+1, OUT+6, OUT+11)
    result = bytes(cpu.mem_read(OUT, 16))
    assert all(v == 0xa5 for i, v in enumerate(result) if i not in (1, 6, 11))
    rgb = bytes(result[i] for i in (1, 6, 11))
    split_digest.update(rgb)
    if word in sample_words:
        samples.append(dict(word=word, rgb=list(rgb)))

def reset_data():
    cpu.mem_write(DATA - 16, bytes([0x5a])*(DATA_SIZE + 32))

def data_bytes():
    return bytes(cpu.mem_read(DATA - 16, DATA_SIZE + 32))

def material_rgb(index):
    return list(struct.unpack("<3I", cpu.mem_read(DATA + 0x108 + index*48 + 24, 12)))

def check_writes(before, index):
    after = data_bytes()
    mutable = set()
    if 0 <= index < 16:
        start = 16 + 0x108 + index*48
        mutable = {start + offset for offset in (*range(24, 36), *range(40, 44))}
    assert all(a == b for i, (a, b) in enumerate(zip(before, after)) if i not in mutable)

materials = []
for index in (0, 1, 6, 15, -1, 16):
    for word in (0, 1, 0x001f, 0x03e0, 0x7c00, 0x4210, 0x7fff, 0xffff):
        color32 = word | word << 16
        reset_data()
        before = data_bytes()
        call(0x4341d0, index, color32)
        check_writes(before, index)
        row = dict(index=index, color32=color32, rgb=material_rgb(index) if 0 <= index < 16 else None)
        if 0 <= index < 16:
            assert get(DATA + 0x108 + index*48 + 40) == color32
        same = data_bytes()
        call(0x4341d0, index, color32)
        assert data_bytes() == same
        row["cached_and_neighbors_unchanged"] = True
        materials.append(row)

rgb_rows = []
for index in (0, 1, 6, 15, -1, 16):
    for rgb in ((0, 0, 0), (255, 255, 255), (248, 248, 248), (31, 63, 127), (256, -1, 0x1234), (0x123456, 0x100020, 0x4321)):
        reset_data()
        before = data_bytes()
        call(0x4340a0, index, *rgb)
        check_writes(before, index)
        valid = 0 <= index < 16
        rgb_rows.append(dict(index=index, input=list(rgb), rgb=material_rgb(index) if valid else None,
                             color32=get(DATA + 0x108 + index*48 + 40) if valid else None))

assert all(bytes(cpu.mem_read(at, n)) == sealed[at] for at, n in ranges.items())
assert bytes(cpu.mem_read(0x583828, 160)) == table
assert bytes(cpu.mem_read(0x975260, 160)) == table
assert get(0x69a014) == 0x4351c0 and get(0x975240) == 0x435280 and get(0x974860) == 0
print(json.dumps(dict(original_sha256=original,
    functions={hex(at): dict(bytes=n, sha256=hashlib.sha256(sealed[at]).hexdigest()) for at, n in ranges.items()},
    format_table=dict(address="0x583828", bytes=160, sha256=hashlib.sha256(table).hexdigest()), mode=mode,
    split=dict(count=65536, rgb_sha256=split_digest.hexdigest(), samples=samples),
    materials=materials, rgb_materials=rgb_rows, unhooked_entries=dict(entries), memory_faults=faults,
    executable_and_table_unchanged=True,
    scope="unchanged RGB5551 setup prefix and original scalar split/material bodies, MMX lane cache disabled; no Windows initialization/driver/framebuffer"), separators=(",", ":")))
