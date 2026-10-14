"""Execute the unchanged GAME.EXE palette initializer with one declared RGB sink.

This records RGB bytes and stores; it is not an original Windows framebuffer.
"""
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
optional = h + 24
base = struct.unpack_from("<I", pe, optional + 28)[0]
image_size = struct.unpack_from("<I", pe, optional + 56)[0]
sections_at = optional + struct.unpack_from("<H", pe, h + 20)[0]
cpu = Uc(UC_ARCH_X86, UC_MODE_32)
cpu.mem_map(base, (image_size + 4095) & ~4095)
for i in range(struct.unpack_from("<H", pe, h + 6)[0]):
    _, virtual, length, raw = struct.unpack_from("<IIII", pe, sections_at + 40*i + 8)
    cpu.mem_write(base + virtual, pe[raw:raw+length])
cpu.mem_map(0x20000000, 0x10000)
ENTRY, RGB, TABLE, STACK, STOP = 0x4b5cd0, 0x4344a0, 0x714c88, 0x20008000, 0x20009000
code = bytes(cpu.mem_read(ENTRY, 319))
assert code[-1] == 0xc3
entries, faults, calls, stores = 0, [], [], []

def get(at):
    return struct.unpack("<I", cpu.mem_read(at, 4))[0]

def hook(_cpu, address, _length, _):
    global entries
    if address == ENTRY:
        entries += 1
    if ENTRY <= address < ENTRY + len(code) or address == STOP:
        return
    assert address == RGB, "undeclared execution " + hex(address)
    sp = cpu.reg_read(UC_X86_REG_ESP)
    r, g, b = (v & 255 for v in struct.unpack("<III", cpu.mem_read(sp + 4, 12)))
    # A distinctive controlled return proves each original store location and
    # the initializer's return independently of Windows color conversion.
    token = (r << 24) | (g << 16) | (b << 8) | len(calls)
    calls.append([r, g, b, token])
    cpu.reg_write(UC_X86_REG_EAX, token)
    cpu.reg_write(UC_X86_REG_EIP, get(sp))
    cpu.reg_write(UC_X86_REG_ESP, sp + 4)

def invalid(_cpu, access, address, length, value, _):
    faults.append(dict(address=address, length=length, pc=cpu.reg_read(UC_X86_REG_EIP)))
    return False

cpu.hook_add(UC_HOOK_CODE, hook)
cpu.hook_add(UC_HOOK_MEM_INVALID, invalid)
rows = []
for pattern in (0, 0x5a, 0xff):
    calls = []
    cpu.mem_write(TABLE - 32, bytes([pattern]) * (1024 + 64))
    cpu.mem_write(STACK - 64, bytes([pattern]) * 128)
    cpu.mem_write(STACK, struct.pack("<I", STOP))
    cpu.reg_write(UC_X86_REG_ESP, STACK)
    cpu.emu_start(ENTRY, STOP, count=20000)
    assert cpu.reg_read(UC_X86_REG_EIP) == STOP and cpu.reg_read(UC_X86_REG_ESP) == STACK + 4
    assert len(calls) == 256 and not faults
    assert bytes(cpu.mem_read(ENTRY, len(code))) == code
    assert bytes(cpu.mem_read(TABLE - 32, 32)) == bytes([pattern]) * 32
    assert bytes(cpu.mem_read(TABLE + 1024, 32)) == bytes([pattern]) * 32
    rows.append(dict(stack_and_guard_pattern=pattern, calls=calls,
                     table_words=list(struct.unpack("<256I", cpu.mem_read(TABLE, 1024))),
                     returned=cpu.reg_read(UC_X86_REG_EAX)))
assert entries == 3 and rows[0]["calls"] == rows[1]["calls"] == rows[2]["calls"]
assert rows[0]["table_words"] == rows[1]["table_words"] == rows[2]["table_words"]
print(json.dumps(dict(original_sha256=hashlib.sha256(pe).hexdigest(),
    entry=hex(ENTRY), code_bytes=len(code), code_sha256=hashlib.sha256(code).hexdigest(),
    original_entries=entries, declared_RGB_calls=256*entries, memory_faults=faults,
    code_and_guards_unchanged=True, rows=rows,
    scope="unhooked original integer loop/division/store order with declared RGB-byte/token sink; not Windows RGB5551/raster/full world"), separators=(",", ":")))
